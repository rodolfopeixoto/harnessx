// SPDX-License-Identifier: MIT

package workflow

import (
	stdctx "context"
	"fmt"
	"io"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/execution"
	"github.com/ropeixoto/harnessx/internal/sensors"
)

// runWithExecutor migrates the agentic step onto execution.DefaultExecutor
// so feature/bugfix/run share the same loop as `harness execute`. Falls
// back to the legacy executeAgents path when no AgentID is supplied (the
// caller already routed there before reaching this point).
func runWithExecutor(ctx stdctx.Context, rc runtimeCtx, mode domain.Mode, opts Options, out io.Writer) (execution.Result, error) {
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
	req := execution.Request{
		ProjectPath: rc.root,
		Prompt:      opts.Prompt,
		Mode:        mapMode(mode),
		AgentID:     opts.AgentID,
		Apply:       opts.Apply,
		PlanOnly:    opts.PlanOnly,
		Autonomy:    execution.AutonomyLevel(autonomyLevel),
		BudgetUSD:   opts.BudgetUSD,
	}
	res, err := ex.Execute(ctx, req)
	if res.RunID != "" {
		fmt.Fprintf(out, "Execute: run=%s status=%s files=%d cost=$%.4f\n",
			res.RunID, res.Status, len(res.ChangedFiles), res.EstimatedCostUSD)
		if res.DiffPath != "" {
			fmt.Fprintf(out, "  diff: %s\n", res.DiffPath)
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
