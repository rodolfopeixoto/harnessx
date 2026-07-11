// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/activeagent"
	"github.com/ropeixoto/harnessx/internal/app/workflow"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
	"github.com/ropeixoto/harnessx/internal/scaffoldpkg"
)

// executeStep dispatches a planned step to its adapter (deterministic
// scaffold, deterministic sensor hint, or workflow-driven agent call).
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

func runDeterministicScaffold(_ context.Context, out io.Writer, dir, lang string) string {
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

func runDeterministicSensor(_ context.Context, out io.Writer, _, kind string) string {
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
	return fmt.Sprintf("workflow-status:%s cost=$%.4f", res.ExecutionStatus, res.ExecutionCostUSD)
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
	totalCost := 0.0
	for _, r := range results {
		if idx := strings.Index(r, "cost=$"); idx >= 0 {
			var c float64
			if _, err := fmt.Sscanf(r[idx+6:], "%f", &c); err == nil {
				totalCost += c
			}
		}
	}
	fmt.Fprintf(&b, "cost_usd: $%.4f\n\n", totalCost)
	for i, s := range steps {
		fmt.Fprintf(&b, "## task %d — %s\n\n", i+1, s.task.Kind)
		fmt.Fprintf(&b, "- routing: %s\n- tags: %s\n- prompt: %s\n- result: %s\n\n",
			s.chosen, strings.Join(s.task.Tags, ","), s.task.Prompt, results[i])
	}
	return path, os.WriteFile(path, []byte(b.String()), 0o644)
}
