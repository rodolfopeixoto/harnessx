// SPDX-License-Identifier: MIT

// Package workflow contains the cross-mode orchestrators that glue
// intent → context → spec → plan → (optional execution) → report.
// Each public function maps to one CLI subcommand (ask, plan, run,
// feature, bugfix). They are deliberately deterministic-first: agent
// execution is opt-in via Options.Execute so the framework runs
// end-to-end even without a CLI binary installed.
package workflow

import (
	stdctx "context"
	"fmt"
	"io"

	"github.com/ropeixoto/harnessx/internal/app/reportcmd"
	hxcontext "github.com/ropeixoto/harnessx/internal/context"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/intent"
	"github.com/ropeixoto/harnessx/internal/plan"
	"github.com/ropeixoto/harnessx/internal/platform/budget"
	"github.com/ropeixoto/harnessx/internal/platform/i18n"
	"github.com/ropeixoto/harnessx/internal/spec"
)

type Options struct {
	StartDir  string
	Prompt    string
	ModeHint  domain.Mode
	AutoYes   bool
	BudgetUSD float64
	Execute   bool
	NoSensors bool
}

type Result struct {
	Mode          domain.Mode
	Intent        intent.Classification
	SpecPath      string
	PlanPath      string
	ReportPath    string
	ContextHash   string
	SessionID     string
	RunID         string
	SensorSummary string
	Confirmed     bool
}

// Ask is Question mode (spec §7). It builds context, surfaces evidence,
// and writes a short report. Never modifies project files.
func Ask(ctx stdctx.Context, opts Options, out io.Writer) (Result, error) {
	rc, err := newRC(opts.StartDir)
	if err != nil {
		return Result{}, err
	}
	res := Result{Mode: domain.ModeQuestion}
	res.Intent = intent.Classification{Mode: domain.ModeQuestion, Confidence: 1, Reasons: []string{"explicit ask"}}

	sess, run, repo := openTelemetry(ctx, rc, domain.ModeQuestion, domain.Stage("ask"))
	defer closeRepo(repo)
	res.SessionID, res.RunID = sess.ID, run.ID

	pack, err := hxcontext.Build(ctx, hxcontext.Options{Root: rc.root, Task: opts.Prompt})
	if err != nil {
		return res, err
	}
	res.ContextHash = pack.Hash

	fmt.Fprintf(out, "I detected a Question Mode request.\n")
	fmt.Fprintf(out, "Context pack hash: %s (%d files, ~%d tokens)\n", pack.Hash[:12], pack.Stats.FilesCount, pack.Stats.EstimatedTokens)
	if len(pack.RelevantFiles) > 0 {
		fmt.Fprintln(out, "Relevant evidence:")
		for _, f := range pack.RelevantFiles {
			fmt.Fprintf(out, "  - %s (%s)\n", f.Path, f.Reason)
		}
	} else {
		fmt.Fprintln(out, "No deterministic evidence located. Use `harness context build` to inspect, then re-ask with a more specific prompt.")
	}

	res.ReportPath, err = writeReport(rc.root, reportcmd.Input{
		SessionID: sess.ID, RunID: run.ID, Mode: domain.ModeQuestion,
		Intent:   opts.Prompt,
		Evidence: filesOf(pack),
	})
	finishTelemetry(ctx, repo, sess, run, domain.StatusSucceeded)
	return res, err
}

// Plan classifies, builds the spec + plan, prints them, and stops short
// of execution. Matches `harness plan "<prompt>"`.
func Plan(ctx stdctx.Context, opts Options, out io.Writer) (Result, error) {
	return planThenMaybeExecute(ctx, opts, false, out)
}

// Run is the natural form. Classifies, plans, optionally executes (when
// opts.Execute is true), and always emits a report.
func Run(ctx stdctx.Context, opts Options, out io.Writer) (Result, error) {
	return planThenMaybeExecute(ctx, opts, opts.Execute, out)
}

// Feature pins ModeHint to ModeFeature before delegating.
func Feature(ctx stdctx.Context, opts Options, out io.Writer) (Result, error) {
	opts.ModeHint = domain.ModeFeature
	return planThenMaybeExecute(ctx, opts, opts.Execute, out)
}

// Bugfix pins ModeHint to ModeBugfix before delegating.
func Bugfix(ctx stdctx.Context, opts Options, out io.Writer) (Result, error) {
	opts.ModeHint = domain.ModeBugfix
	return planThenMaybeExecute(ctx, opts, opts.Execute, out)
}

func planThenMaybeExecute(ctx stdctx.Context, opts Options, execute bool, out io.Writer) (Result, error) {
	rc, err := newRC(opts.StartDir)
	if err != nil {
		return Result{}, err
	}
	res := Result{}
	res.Intent = intent.Classify(opts.Prompt)
	mode := res.Intent.Mode
	if opts.ModeHint != "" {
		mode = opts.ModeHint
		res.Intent.Mode = mode
		res.Intent.Reasons = append(res.Intent.Reasons, "mode forced by caller")
	}
	res.Mode = mode

	sess, run, repo := openTelemetry(ctx, rc, mode, domain.StagePlan)
	defer closeRepo(repo)
	res.SessionID, res.RunID = sess.ID, run.ID

	fmt.Fprintf(out, "Detected intent: %s (confidence %.2f)\n", mode, res.Intent.Confidence)
	for _, r := range res.Intent.Reasons {
		fmt.Fprintf(out, "  reason: %s\n", r)
	}

	sp := spec.NewFromPrompt(opts.Prompt, mode)
	specPath, err := sp.Write(rc.root)
	if err != nil {
		finishTelemetry(ctx, repo, sess, run, domain.StatusFailed)
		return res, err
	}
	res.SpecPath = specPath
	fmt.Fprintf(out, "Spec written: %s\n", specPath)

	pack, err := hxcontext.Build(ctx, hxcontext.Options{
		Root: rc.root, Task: opts.Prompt,
		Providers: hxcontext.AutoLSP(rc.root),
	})
	if err == nil {
		res.ContextHash = pack.Hash
	}

	pl := plan.New(sp.ID, mode)
	pl.Summary = opts.Prompt
	pl.DetectedStack = stackNames(rc.profile)
	pl.RelevantContext = filesOf(pack)
	pl.SensorsToRun = []string{"forbidden_files", "forbidden_commands", "secrets_scan", "stack rule pack"}
	pl.SecurityChecks = []string{"secrets_scan", "forbidden_files", "stack-specific security sensor"}
	pl.EstimatedCostUSD = estimateCost(opts.BudgetUSD)
	pl.EstimatedTime = "deterministic; agent execution skipped in Phase 6 baseline"
	pl.AgentChain = []string{"deterministic (no agent run requested)"}
	pl.Risks = riskHints(mode)
	if opts.AutoYes {
		pl.ConfirmationStatus = "approved (auto-yes)"
	} else {
		pl.ConfirmationStatus = "pending"
	}
	planPath, err := pl.Write(rc.root)
	if err != nil {
		finishTelemetry(ctx, repo, sess, run, domain.StatusFailed)
		return res, err
	}
	res.PlanPath = planPath
	fmt.Fprintf(out, "Plan written: %s\n", planPath)

	g := budget.New(opts.BudgetUSD)
	_ = g.Charge(pl.EstimatedCostUSD)
	fmt.Fprintf(out, "Budget: $%.2f / $%.2f remaining\n", g.Remaining(), opts.BudgetUSD)

	res.Confirmed = opts.AutoYes
	if !opts.AutoYes && execute {
		if confirmInteractive(out, i18n.T("plan.confirm")) {
			res.Confirmed = true
		} else {
			fmt.Fprintln(out, i18n.T("plan.not_approved"))
			execute = false
		}
	}

	if execute {
		_ = executeAgents(ctx, rc, mode, opts.Prompt, opts.BudgetUSD, pack, out)
	}

	reportPath, err := writeReport(rc.root, reportcmd.Input{
		SessionID: sess.ID, RunID: run.ID, Mode: mode,
		Intent:   opts.Prompt,
		SpecPath: specPath, PlanPath: planPath,
		Risks:    pl.Risks,
		Evidence: filesOf(pack),
	})
	if err != nil {
		finishTelemetry(ctx, repo, sess, run, domain.StatusFailed)
		return res, err
	}
	res.ReportPath = reportPath
	fmt.Fprintf(out, "Report written: %s\n", reportPath)

	finishTelemetry(ctx, repo, sess, run, domain.StatusSucceeded)
	return res, nil
}

func writeReport(root string, in reportcmd.Input) (string, error) {
	return reportcmd.Build(root, in)
}
