// SPDX-License-Identifier: MIT

package workflow

import (
	stdctx "context"
	"fmt"
	"io"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/execution"
	"github.com/ropeixoto/harnessx/internal/index"
)

// askAgent is the trivial-prompt fast path. It invokes the chosen agent
// directly inside the project root (no worktree, no diff, no sensors,
// no autonomy gate beyond plan_and_ask defaults). The agent's final
// message becomes the run report's Answer section.
//
// Used when intent.Classify returns Complexity = Trivial. Saves the cost
// of writing spec + plan + creating a worktree for a one-shot Q&A.
func askAgent(ctx stdctx.Context, rc runtimeCtx, opts Options, out io.Writer) (execution.Result, error) {
	reg, _, err := agentcmd.LoadAll(rc.root)
	if err != nil {
		return execution.Result{}, fmt.Errorf("load agents: %w", err)
	}
	adapter, ok := reg.Get(opts.AgentID)
	if !ok {
		return execution.Result{}, fmt.Errorf("agent %q not registered", opts.AgentID)
	}
	ex := execution.NewDefaultExecutor(rc.root, adapter, nil, index.Profile{})
	req := execution.Request{
		ProjectPath: rc.root,
		Prompt:      opts.Prompt,
		Mode:        execution.ModeAsk,
		AgentID:     opts.AgentID,
		Apply:       false,
		PlanOnly:    false,
		Autonomy:    execution.AutonomyLevel(defaultAutonomy(opts.Autonomy)),
		BudgetUSD:   opts.BudgetUSD,
	}
	res, err := ex.Execute(ctx, req)
	if res.RunID != "" {
		fmt.Fprintf(out, "Ask: run=%s status=%s cost=$%.4f\n", res.RunID, res.Status, res.EstimatedCostUSD)
		if res.ReportPath != "" {
			fmt.Fprintf(out, "  report: %s\n", res.ReportPath)
		}
	}
	return res, err
}

func defaultAutonomy(s string) string {
	if s == "" {
		return "plan_and_ask"
	}
	return s
}
