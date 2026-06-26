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
	"strings"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/app/reportcmd"
	hxcontext "github.com/ropeixoto/harnessx/internal/context"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/input"
	"github.com/ropeixoto/harnessx/internal/intent"
	"github.com/ropeixoto/harnessx/internal/plan"
	"github.com/ropeixoto/harnessx/internal/platform/budget"
	"github.com/ropeixoto/harnessx/internal/platform/i18n"
	"github.com/ropeixoto/harnessx/internal/spec"
)

type Options struct {
	StartDir     string
	Prompt       string
	ModeHint     domain.Mode
	AutoYes      bool
	BudgetUSD    float64
	Execute      bool
	EvidenceOnly bool
	NoSensors    bool
	AgentID      string
	Apply        bool
	PlanOnly     bool
	Autonomy     string
	PromptFile   string
	PDF          string
	Image        string
}

type Result struct {
	Mode              domain.Mode
	Intent            intent.Classification
	SpecPath          string
	PlanPath          string
	ReportPath        string
	ContextHash       string
	SessionID         string
	RunID             string
	SensorSummary     string
	Confirmed         bool
	ExecutionRunID    string
	ExecutionStatus   string
	ExecutionDiffPath string
	ExecutionCostUSD  float64
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

	// Audit BUG-7: previously `harness ask` only listed evidence files and
	// never produced a textual answer, breaking the "question mode" UX.
	// When --agent is set (and --evidence-only is not), route the evidence
	// through the executor with a strict cite-only system prompt and the
	// caller's budget cap (default 0.05 USD).
	if opts.AgentID != "" && !opts.EvidenceOnly {
		answer, askErr := answerFromEvidence(ctx, rc, opts, pack, out)
		if askErr != nil {
			fmt.Fprintf(out, "ask: answer step failed (%v); falling back to evidence-only output\n", askErr)
		} else {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Answer:")
			fmt.Fprintln(out, answer)
		}
	}

	res.ReportPath, err = writeReport(rc.root, reportcmd.Input{
		SessionID: sess.ID, RunID: run.ID, Mode: domain.ModeQuestion,
		Intent:   opts.Prompt,
		Evidence: filesOf(pack),
	})
	finishTelemetry(ctx, repo, sess, run, domain.StatusSucceeded)
	return res, err
}

// answerFromEvidence enriches the user's question with the deterministic
// evidence pack and asks the chosen adapter to respond using ONLY that
// evidence, citing each claim with `[path:line]`. Cost is capped by
// opts.BudgetUSD which `harness ask` exposes as --budget-usd.
func answerFromEvidence(ctx stdctx.Context, rc runtimeCtx, opts Options, pack *hxcontext.Pack, out io.Writer) (string, error) {
	reg, _, err := agentcmd.LoadAll(rc.root)
	if err != nil {
		return "", err
	}
	adapter, ok := reg.Get(opts.AgentID)
	if !ok {
		return "", fmt.Errorf("agent %q not registered", opts.AgentID)
	}
	var sb strings.Builder
	sb.WriteString("You are answering a question about a codebase.\n")
	sb.WriteString("Rules:\n")
	sb.WriteString("  1. Use ONLY the evidence below.\n")
	sb.WriteString("  2. Cite each claim with a [path:line] reference.\n")
	sb.WriteString("  3. If the evidence is insufficient, say so and stop.\n\n")
	sb.WriteString("Question: ")
	sb.WriteString(opts.Prompt)
	sb.WriteString("\n\nEvidence:\n")
	for _, f := range pack.RelevantFiles {
		fmt.Fprintf(&sb, "- %s (%s)\n", f.Path, f.Reason)
	}
	req := agents.AgentRequest{
		Prompt: sb.String(), WorkingDir: rc.root,
	}
	result := adapter.Run(ctx, req)
	if result.Err != nil {
		return "", result.Err
	}
	answer := strings.TrimSpace(result.Output.FinalMessage)
	if opts.BudgetUSD > 0 && result.Usage.EstimatedCostUSD > opts.BudgetUSD {
		fmt.Fprintf(out, "ask: warning — cost $%.4f exceeded --budget-usd $%.4f\n",
			result.Usage.EstimatedCostUSD, opts.BudgetUSD)
	}
	return answer, nil
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
	if opts.PromptFile != "" || opts.PDF != "" || opts.Image != "" {
		assembled, aerr := input.Assemble(input.Sources{
			Positional: opts.Prompt, PromptFile: opts.PromptFile, PDF: opts.PDF, Image: opts.Image,
		})
		if aerr != nil {
			return Result{}, aerr
		}
		opts.Prompt = assembled.Prompt
		fmt.Fprintf(out, "Input: %d attachments (%v)\n", len(assembled.Attachments), assembled.Notes)
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
	if pl.EstimatedCostUSD > 0 {
		fmt.Fprintf(out, "Budget: $%.2f / $%.2f remaining\n", g.Remaining(), opts.BudgetUSD)
	}

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
		switch {
		case opts.AgentID != "" && res.Intent.Complexity == intent.ComplexityTrivial && !opts.Apply:
			fmt.Fprintln(out, "Routing: trivial prompt — using AskAgent fast path (no diff, no worktree)")
			exRes, exErr := askAgent(ctx, rc, opts, out)
			if exErr != nil {
				fmt.Fprintf(out, "Ask: %v\n", exErr)
			}
			res.ExecutionRunID = exRes.RunID
			res.ExecutionStatus = string(exRes.Status)
			res.ExecutionCostUSD = exRes.EstimatedCostUSD
		case opts.AgentID != "":
			exRes, exErr := runWithExecutorAndComplexity(ctx, rc, mode, opts, res.Intent.Complexity, pack, out)
			if exErr != nil {
				fmt.Fprintf(out, "Execute: %v\n", exErr)
			}
			res.ExecutionRunID = exRes.RunID
			res.ExecutionStatus = string(exRes.Status)
			res.ExecutionDiffPath = exRes.DiffPath
			res.ExecutionCostUSD = exRes.EstimatedCostUSD
		default:
			// Audit BUG-5: when --plan-only is set the user expects no LLM
			// invocation. Skip the legacy agent chain (the --agent path
			// already honours PlanOnly inside the executor).
			if opts.PlanOnly {
				fmt.Fprintln(out, "Execute: --plan-only set; skipping agent chain (no LLM cost)")
				break
			}
			_ = executeAgents(ctx, rc, mode, opts.Prompt, opts.BudgetUSD, pack, out)
		}
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
