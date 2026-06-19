// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/activeagent"
	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/app/workflow"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
	"github.com/ropeixoto/harnessx/internal/router"
	"github.com/ropeixoto/harnessx/internal/scaffoldpkg"
	"github.com/ropeixoto/harnessx/internal/taskgraph"
)

func newDoCmd() *cobra.Command {
	var (
		yes           bool
		deterministic bool
		budget        float64
		maxTasks      int
		autonomy      string
		image         string
		asJSON        bool
		agentOverride string
	)
	c := &cobra.Command{
		Use:   "do \"<prompt>\"",
		Short: "Decompose prompt into tasks and route each to the best adapter",
		Long: `Splits the prompt into sub-tasks ("scaffold X and add Y" → 2 tasks),
picks the best adapter per task by matching task tags against adapter
strengths, and runs them in order. Deterministic tasks (scaffold, lint,
test, secrets-scan) skip the LLM by default.

  harness do "scaffold python and add a /healthz endpoint"
  harness do "..." --yes --budget-usd 0.30 --deterministic`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDo(cmd.Context(), cmd.OutOrStdout(), strings.Join(args, " "), doOpts{
				yes: yes, det: deterministic, budget: budget, maxTasks: maxTasks,
				autonomy: autonomy, image: image, asJSON: asJSON, agentOverride: agentOverride,
			})
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "skip plan confirmation prompt")
	c.Flags().BoolVar(&deterministic, "deterministic", true, "prefer scaffold/sensor over LLM where possible")
	c.Flags().Float64Var(&budget, "budget-usd", 1.0, "max USD across all routed tasks")
	c.Flags().IntVar(&maxTasks, "max-tasks", 10, "hard cap on decomposed tasks")
	c.Flags().StringVar(&autonomy, "autonomy", "safe_execute", "autonomy level for LLM tasks")
	c.Flags().StringVar(&image, "image", "", "attach an image; auto-adds vision tag for routing")
	c.Flags().BoolVar(&asJSON, "json", false, "emit final result as JSON (implies --yes)")
	c.Flags().StringVar(&agentOverride, "agent", "", "force a specific adapter id (overrides router + active pin)")
	return c
}

func newRouteCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "route",
		Short: "Inspect routing decisions without executing",
	}
	c.AddCommand(routeShowCmd())
	return c
}

func routeShowCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "show \"<prompt>\"",
		Short: "Dry-run: print the task graph + chosen adapter per task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRouteShow(cmd.Context(), cmd.OutOrStdout(), strings.Join(args, " "), asJSON)
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "emit JSON for programmatic consumers (IDE plugins, etc.)")
	return c
}

type doOpts struct {
	yes           bool
	det           bool
	budget        float64
	maxTasks      int
	autonomy      string
	image         string
	asJSON        bool
	agentOverride string
}

type plannedStep struct {
	task   taskgraph.Task
	choice router.Choice
	chosen string // "deterministic:scaffold" | "adapter:claude" | "none"
}

func planDo(ctx context.Context, dir, prompt string, opts doOpts) ([]plannedStep, error) {
	tasks := taskgraph.Decompose(prompt, taskgraph.Options{})
	if opts.image != "" {
		for i := range tasks {
			tasks[i].Tags = append(tasks[i].Tags, "vision")
		}
	}
	if len(tasks) > opts.maxTasks {
		tasks = tasks[:opts.maxTasks]
	}
	reg, _, err := agentcmd.LoadAll(dir)
	if err != nil {
		return nil, err
	}
	override := activeagent.ResolveAgentID(dir, opts.agentOverride)
	steps := make([]plannedStep, 0, len(tasks))
	for _, t := range tasks {
		step := plannedStep{task: t}
		if opts.det && deterministicMatch(t) != "" {
			step.chosen = "deterministic:" + deterministicMatch(t)
		} else if c, ok := router.Pick(t.Tags, reg); ok {
			step.choice = c
			step.chosen = "adapter:" + c.AdapterID
			if override != "" && override != c.AdapterID {
				if _, registered := reg.Get(override); registered {
					step.choice = router.Choice{AdapterID: override}
					step.chosen = "adapter:" + override + " (--agent override)"
				}
			}
		} else if override != "" {
			if _, registered := reg.Get(override); registered {
				step.choice = router.Choice{AdapterID: override}
				step.chosen = "adapter:" + override + " (--agent override)"
			} else {
				step.chosen = "none (no adapter matches tags)"
			}
		} else {
			step.chosen = "none (no adapter matches tags)"
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func deterministicMatch(t taskgraph.Task) string {
	switch t.Kind {
	case taskgraph.KindScaffold:
		if t.Lang != "" {
			return "scaffold:" + t.Lang
		}
		return ""
	case taskgraph.KindLint, taskgraph.KindTest, taskgraph.KindSecrets:
		return "sensor:" + string(t.Kind)
	}
	return ""
}

func runRouteShow(ctx context.Context, out io.Writer, prompt string, asJSON bool) error {
	dir, err := cwd()
	if err != nil {
		return err
	}
	steps, err := planDo(ctx, dir, prompt, doOpts{det: true, maxTasks: 10})
	if err != nil {
		return err
	}
	if asJSON {
		return emitJSON(out, prompt, steps)
	}
	printPlan(out, steps)
	return nil
}

type jsonStep struct {
	Index      int      `json:"index"`
	Kind       string   `json:"kind"`
	Tags       []string `json:"tags"`
	Routing    string   `json:"routing"`
	AdapterID  string   `json:"adapter_id,omitempty"`
	Prompt     string   `json:"prompt"`
	Confidence float64  `json:"confidence"`
	Lang       string   `json:"lang,omitempty"`
}

// JSONSchemaVersion is bumped whenever the shape of jsonPlan or
// jsonDoResult changes. v1 covers v0.39/v0.40 layout plus this version
// field (added in v0.43). Consumers should refuse versions they do not
// know.
const JSONSchemaVersion = 1

type jsonPlan struct {
	SchemaVersion int        `json:"schema_version"`
	Prompt        string     `json:"prompt"`
	Steps         []jsonStep `json:"steps"`
}

func emitJSON(out io.Writer, prompt string, steps []plannedStep) error {
	js := jsonPlan{SchemaVersion: JSONSchemaVersion, Prompt: prompt}
	for i, s := range steps {
		js.Steps = append(js.Steps, jsonStep{
			Index: i + 1, Kind: string(s.task.Kind), Tags: s.task.Tags,
			Routing: s.chosen, AdapterID: s.choice.AdapterID,
			Prompt: s.task.Prompt, Confidence: s.task.Confidence, Lang: s.task.Lang,
		})
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(js)
}

func printPlan(out io.Writer, steps []plannedStep) {
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "#\tKIND\tCONF\tTAGS\tROUTING\tPROMPT")
	lowConf := false
	for i, s := range steps {
		fmt.Fprintf(tw, "%d\t%s\t%.2f\t%s\t%s\t%s\n", i+1, s.task.Kind, s.task.Confidence, strings.Join(s.task.Tags, ","), s.chosen, truncStr(s.task.Prompt, 50))
		if s.task.Confidence < 0.5 {
			lowConf = true
		}
	}
	_ = tw.Flush()
	if lowConf {
		fmt.Fprintln(out, "⚠ one or more tasks have low classification confidence — review before --yes")
	}
}

// handoff builds a "Past steps in this run" block that gets prepended
// to every task after the first. Lets multi-adapter pipelines share
// state via the prompt without a new shared-memory abstraction.
func handoff(originalPrompt string, prev []plannedStep, results []string) string {
	var b strings.Builder
	b.WriteString("# Past steps in this run\n\n")
	fmt.Fprintf(&b, "Original request: %s\n\n", originalPrompt)
	for i, s := range prev {
		fmt.Fprintf(&b, "- step %d (%s via %s): %s\n", i+1, s.task.Kind, s.chosen, results[i])
	}
	b.WriteString("\nUse these as context. Do not redo work that already succeeded.\n\n# This step\n")
	return b.String()
}

func truncStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func runDo(ctx context.Context, out io.Writer, prompt string, opts doOpts) error {
	dir, err := cwd()
	if err != nil {
		return err
	}
	steps, err := planDo(ctx, dir, prompt, opts)
	if err != nil {
		return err
	}
	logOut := out
	if opts.asJSON {
		opts.yes = true
		logOut = os.Stderr
	}
	fmt.Fprintln(logOut, "plan:")
	printPlan(logOut, steps)
	if !opts.yes && !confirmYes(logOut, "execute? [y/N] ") {
		fmt.Fprintln(logOut, "cancelled")
		return nil
	}
	results := make([]string, len(steps))
	for i, s := range steps {
		fmt.Fprintf(logOut, "\n[task %d/%d] %s — %s\n", i+1, len(steps), s.task.Kind, s.chosen)
		if i > 0 {
			s.task.Prompt = handoff(prompt, steps[:i], results[:i]) + "\n\n" + s.task.Prompt
		}
		results[i] = executeStep(ctx, logOut, dir, s, opts)
	}
	rp, _ := writeDoReport(dir, prompt, steps, results)
	if rp != "" {
		fmt.Fprintf(logOut, "\ndo report: %s\n", rp)
	}
	if opts.asJSON {
		return emitDoJSON(out, prompt, steps, results, rp)
	}
	return nil
}

type jsonDoResult struct {
	SchemaVersion int        `json:"schema_version"`
	Prompt        string     `json:"prompt"`
	ReportPath    string     `json:"report_path"`
	Steps         []jsonStep `json:"steps"`
	Results       []string   `json:"results"`
}

func emitDoJSON(out io.Writer, prompt string, steps []plannedStep, results []string, reportPath string) error {
	js := jsonDoResult{SchemaVersion: JSONSchemaVersion, Prompt: prompt, ReportPath: reportPath, Results: results}
	for i, s := range steps {
		js.Steps = append(js.Steps, jsonStep{
			Index: i + 1, Kind: string(s.task.Kind), Tags: s.task.Tags,
			Routing: s.chosen, AdapterID: s.choice.AdapterID,
			Prompt: s.task.Prompt, Confidence: s.task.Confidence, Lang: s.task.Lang,
		})
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(js)
}

func confirmYes(out io.Writer, prompt string) bool {
	fmt.Fprint(out, prompt)
	var raw string
	_, _ = fmt.Scanln(&raw)
	return raw == "y" || raw == "Y" || raw == "yes"
}

func executeStep(ctx context.Context, out io.Writer, dir string, s plannedStep, opts doOpts) string {
	switch {
	case strings.HasPrefix(s.chosen, "deterministic:scaffold:"):
		return runDeterministicScaffold(ctx, out, dir, s.task.Lang)
	case strings.HasPrefix(s.chosen, "deterministic:sensor:"):
		return runDeterministicSensor(ctx, out, dir, string(s.task.Kind))
	case strings.HasPrefix(s.chosen, "adapter:"):
		agentID := activeagent.ResolveAgentID(dir, opts.agentOverride)
		if agentID == "" {
			agentID = s.choice.AdapterID
		}
		return runWorkflowFeature(ctx, out, dir, s.task.Prompt, agentID, opts)
	default:
		fmt.Fprintln(out, "  ✗ no adapter matched (consider 'harness agent install <id>')")
		return "skipped"
	}
}

func runDeterministicScaffold(ctx context.Context, out io.Writer, dir, lang string) string {
	if lang == "" {
		fmt.Fprintln(out, "  ✗ scaffold task missing language")
		return "missing-lang"
	}
	m, err := scaffoldpkg.Load(lang)
	if err != nil {
		fmt.Fprintf(out, "  ✗ %v\n", err)
		return "load-error"
	}
	res, err := scaffoldpkg.Apply(m, scaffoldpkg.ApplyOptions{Root: dir, Name: filepath.Base(dir)})
	if err != nil {
		fmt.Fprintf(out, "  ✗ %v\n", err)
		return "apply-error"
	}
	fmt.Fprintf(out, "  ✓ scaffold %s — %d files (dry-run; pass --apply via harness scaffold to write)\n", lang, len(res.Created))
	return "scaffold-dry"
}

func runDeterministicSensor(ctx context.Context, out io.Writer, dir, kind string) string {
	fmt.Fprintf(out, "  → run: harness sensor run %s_scan (or harness check)\n", kind)
	return "sensor-hint"
}

func runWorkflowFeature(ctx context.Context, out io.Writer, dir, prompt, agentID string, opts doOpts) string {
	res, err := workflow.Feature(ctx, workflow.Options{
		StartDir: dir, Prompt: prompt, AgentID: agentID, Execute: true,
		AutoYes: true, BudgetUSD: opts.budget, Autonomy: opts.autonomy,
		PlanOnly: false, Apply: true,
	}, out)
	if err != nil {
		fmt.Fprintf(out, "  ✗ %v\n", err)
		return "error: " + err.Error()
	}
	return "workflow-status:" + res.ExecutionStatus
}

func writeDoReport(dir, prompt string, steps []plannedStep, results []string) (string, error) {
	root := paths.HarnessDir(dir)
	outDir := filepath.Join(root, "runs", "_do")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(outDir, fmt.Sprintf("do-%s.md", time.Now().UTC().Format("20060102-150405")))
	var b strings.Builder
	b.WriteString("# harness do report\n\n")
	fmt.Fprintf(&b, "prompt: %s\n\n", prompt)
	for i, s := range steps {
		fmt.Fprintf(&b, "## task %d — %s\n\n", i+1, s.task.Kind)
		fmt.Fprintf(&b, "- routing: %s\n- tags: %s\n- prompt: %s\n- result: %s\n\n",
			s.chosen, strings.Join(s.task.Tags, ","), s.task.Prompt, results[i])
	}
	return path, os.WriteFile(path, []byte(b.String()), 0o644)
}
