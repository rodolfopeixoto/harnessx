// SPDX-License-Identifier: MIT

package workflow

import (
	stdctx "context"
	"fmt"
	"io"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	hxcontext "github.com/ropeixoto/harnessx/internal/context"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/execution"
	"github.com/ropeixoto/harnessx/internal/intent"
	"github.com/ropeixoto/harnessx/internal/promptenh"
	"github.com/ropeixoto/harnessx/internal/router"
	"github.com/ropeixoto/harnessx/internal/sensors"
	"github.com/ropeixoto/harnessx/internal/ui"
)

// runWithExecutor migrates the agentic step onto execution.DefaultExecutor
// so feature/bugfix/run share the same loop as `harness execute`. Falls
// back to the legacy executeAgents path when no AgentID is supplied (the
// caller already routed there before reaching this point).
//
// P34 additions: enriches the prompt via promptenh.Enhance and selects
// the adapter model with router.PickModel based on intent.Complexity.
func runWithExecutorAndComplexity(ctx stdctx.Context, rc runtimeCtx, mode domain.Mode, opts Options, complexity intent.Complexity, pack *hxcontext.Pack, out io.Writer) (execution.Result, error) {
	reg, _, err := agentcmd.LoadAll(rc.root)
	if err != nil {
		return execution.Result{}, fmt.Errorf("load agents: %w", err)
	}
	adapter, ok := reg.Get(opts.AgentID)
	if !ok {
		return execution.Result{}, fmt.Errorf("agent %q not registered", opts.AgentID)
	}
	autonomyLevel := opts.Autonomy
	if autonomyLevel == "" {
		autonomyLevel = "safe_execute"
	}
	ex := execution.NewDefaultExecutor(rc.root, adapter, defaultSensors(opts.NoSensors), rc.profile)
	ex.Status = func(msg string) { fmt.Fprintf(out, "  [agent] %s\n", msg) }
	ex.LiveOut = newAgentLivePrefix(out)
	enhancement := promptenh.Enhance(opts.Prompt, mode, pack, nil)
	model := router.PickModel(adapter, complexity)
	if model != "" {
		fmt.Fprintf(out, "Routing: complexity=%s -> model=%s\n", complexity, model)
	}
	req := execution.Request{
		ProjectPath:    rc.root,
		Prompt:         opts.Prompt,
		EnhancedPrompt: enhancement.Enhanced,
		Mode:           mapMode(mode),
		AgentID:        opts.AgentID,
		Apply:          opts.Apply,
		PlanOnly:       opts.PlanOnly,
		Autonomy:       execution.AutonomyLevel(autonomyLevel),
		BudgetUSD:      opts.BudgetUSD,
		Model:          model,
	}
	res, err := ex.Execute(ctx, req)
	if res.RunID != "" {
		if _, werr := promptenh.Write(fmt.Sprintf("%s/.harness/runs/%s", rc.root, res.RunID), enhancement); werr == nil {
			fmt.Fprintf(out, "  enhancement: %s/.harness/runs/%s/enhancement.json\n", rc.root, res.RunID)
		}
		fmt.Fprintf(out, "Execute: run=%s status=%s files=%d cost=$%.4f\n",
			res.RunID, res.Status, len(res.ChangedFiles), res.EstimatedCostUSD)
		if res.DiffPath != "" {
			fmt.Fprintf(out, "  %s diff: %s\n", ui.MarkInfo(), res.DiffPath)
			printDiffPreview(out, res.DiffPath, res.DiffStatPath)
		}
		if res.ReportPath != "" {
			fmt.Fprintf(out, "  report: %s\n", res.ReportPath)
		}
		if res.Status == execution.StatusWaitingApproval {
			fmt.Fprintf(out, "  next: harness runs approve %s | harness runs discard %s\n", res.RunID, res.RunID)
		}
	}
	return res, err
}

func mapMode(m domain.Mode) execution.Mode {
	switch m {
	case domain.ModeFeature:
		return execution.ModeFeature
	case domain.ModeBugfix:
		return execution.ModeBugfix
	case domain.ModeQuestion:
		return execution.ModeAsk
	default:
		return execution.ModeFeature
	}
}

func defaultSensors(disabled bool) []sensors.Sensor {
	if disabled {
		return nil
	}
	return nil
}
