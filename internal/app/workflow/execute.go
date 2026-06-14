// SPDX-License-Identifier: MIT

package workflow

import (
	stdctx "context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	hxcontext "github.com/ropeixoto/harnessx/internal/context"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/budget"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/router"
)

// execEvidence summarises one agent execution for telemetry + report.
type execEvidence struct {
	SelectedAgent string
	Model         string
	FallbackFrom  string
	ErrorType     string
	Usage         agents.Usage
	Summary       string
}

// defaultRoutes is router.Defaults; kept as a local alias so executeAgents
// reads naturally.
var defaultRoutes = router.Defaults

// executeAgents resolves the agent chain via routes.yaml (or a bundled
// default chain when none exists) and runs it through router.Execute.
func executeAgents(ctx stdctx.Context, rc runtimeCtx, mode domain.Mode, prompt string, budgetUSD float64, pack *hxcontext.Pack, out io.Writer) *execEvidence {
	reg, _, err := agentcmd.LoadAll(rc.root)
	if err != nil {
		fmt.Fprintf(out, "Execute: agent load failed (%v); skipping\n", err)
		return nil
	}
	if len(reg.IDs()) == 0 {
		fmt.Fprintln(out, "Execute: no agent adapters registered; skipping")
		return nil
	}

	routes := defaultRoutes(reg)
	if user, err := router.LoadConfig(filepath.Join(rc.root, ".harness", "config", "routes.yaml")); err == nil && user != nil {
		for k, v := range user {
			routes[k] = v
		}
	}
	if repoStats, err := sqlite.Open(rc.dbPath); err == nil {
		if stats, err := router.LoadStats(ctx, repoStats.DB()); err == nil && len(stats) > 0 {
			for k, v := range routes {
				routes[k] = router.ApplyStats(v, stats)
			}
		}
		_ = repoStats.Close()
	}
	r := router.New(reg, routes)

	task := taskFor(mode)
	g := budget.New(budgetUSD)
	req := agents.AgentRequest{
		Prompt: prompt, WorkingDir: rc.root,
		Timeout: constants.DefaultAgentTimeout,
	}
	d, err := r.Select(task)
	if err != nil {
		fmt.Fprintf(out, "Execute: router failed (%v); skipping\n", err)
		return nil
	}
	fmt.Fprintf(out, "Execute: task=%s chain=", task)
	for i, a := range d.Chain {
		if i > 0 {
			fmt.Fprint(out, " → ")
		}
		fmt.Fprint(out, a.ID())
	}
	fmt.Fprintf(out, " (budget $%.2f)\n", g.Remaining())

	exec, _ := r.Execute(ctx, task, req, nil)
	ev := &execEvidence{
		SelectedAgent: idOrEmpty(exec.Selected),
		Model:         exec.Result.Output.FinalMessage[:0],
		Usage:         exec.Result.Usage,
	}
	if len(exec.Fallbacks) > 0 {
		ev.FallbackFrom = exec.Fallbacks[0].From
		ev.ErrorType = string(exec.Fallbacks[0].Failure)
	}
	if exec.Succeeded {
		ev.Summary = fmt.Sprintf("agent=%s ok; tokens=%d/%d cost=$%.4f",
			ev.SelectedAgent, ev.Usage.InputTokens, ev.Usage.OutputTokens, ev.Usage.EstimatedCostUSD)
		fmt.Fprintf(out, "  %s\n", ev.Summary)
	} else {
		ev.Summary = fmt.Sprintf("agent chain failed; last=%s failure=%s",
			ev.SelectedAgent, exec.Result.Failure)
		fmt.Fprintf(out, "  %s\n", ev.Summary)
	}
	_ = pack
	_ = time.Now // keep import stable when slimming this file later
	return ev
}

func idOrEmpty(a agents.AgentAdapter) string {
	if a == nil {
		return ""
	}
	return a.ID()
}
