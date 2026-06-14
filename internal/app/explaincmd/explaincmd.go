// SPDX-License-Identifier: MIT

// Package explaincmd implements `harness explain <prompt>` — dry-run the
// intent classifier + router so the user can see what a real `run` would do.
package explaincmd

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/intent"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
	"github.com/ropeixoto/harnessx/internal/router"
)

type Options struct {
	StartDir string
	Prompt   string
}

func Run(out io.Writer, opts Options) error {
	if opts.Prompt == "" {
		return fmt.Errorf("explain: empty prompt")
	}
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	cls := intent.Classify(opts.Prompt)
	task := taskFor(cls.Mode)
	fmt.Fprintf(out, "Intent: %s (confidence %.2f)\n", cls.Mode, cls.Confidence)
	for _, r := range cls.Reasons {
		fmt.Fprintf(out, "  reason: %s\n", r)
	}
	fmt.Fprintf(out, "Routed task: %s\n", task)

	reg, _, err := agentcmd.LoadAll(root)
	if err != nil {
		return err
	}
	routes := router.Defaults(reg)
	if user, err := router.LoadConfig(filepath.Join(root, ".harness", "config", "routes.yaml")); err == nil && user != nil {
		for k, v := range user {
			routes[k] = v
		}
	}
	r := router.New(reg, routes)
	d, err := r.Select(task)
	if err != nil {
		fmt.Fprintf(out, "Chain: (unresolved — %v)\n", err)
		return nil
	}
	fmt.Fprintf(out, "Chain: ")
	for i, a := range d.Chain {
		if i > 0 {
			fmt.Fprint(out, " → ")
		}
		fmt.Fprint(out, a.ID())
	}
	fmt.Fprintf(out, "  (budget $%.2f)\n", d.BudgetUSD)
	fmt.Fprintln(out, "No agent ran — pass `--execute --yes` on `harness run` to dispatch.")
	return nil
}

// taskFor mirrors workflow.taskFor so explain matches the executor.
func taskFor(mode domain.Mode) string {
	switch mode {
	case domain.ModeQuestion:
		return "codebase_exploration"
	case domain.ModeBugfix:
		return "implementation"
	case domain.ModeOptimization:
		return "resource_optimization"
	case domain.ModeAudit:
		return "dependency_audit"
	case domain.ModeReview:
		return "security_review"
	case domain.ModeDesignToProduct:
		return "design_to_product"
	default:
		return "implementation"
	}
}
