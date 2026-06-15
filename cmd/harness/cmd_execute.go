// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/execution"
	"github.com/ropeixoto/harnessx/internal/index"
)

// newExecuteCmd is the direct path into the real agentic execution loop
// while feature/bugfix workflow integration is staged (P31 phase 7). It
// invokes the chosen adapter through DefaultExecutor and prints the
// resulting run id + paths.
func newExecuteCmd() *cobra.Command {
	var (
		agentID      string
		mode         string
		apply        bool
		planOnly     bool
		autonomy     string
		budget       float64
		sandbox      string
		sandboxImage string
	)
	c := &cobra.Command{
		Use:   "execute <prompt>",
		Short: "Run the agentic execution loop (P31 direct path)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			reg, _, err := agentcmd.LoadAll(root)
			if err != nil {
				return fmt.Errorf("load agents: %w", err)
			}
			adapter, ok := reg.Get(agentID)
			if !ok {
				return fmt.Errorf("agent %q not registered", agentID)
			}
			ex := execution.NewDefaultExecutor(root, adapter, nil, index.Profile{})
			req := execution.Request{
				ProjectPath:  root,
				Prompt:       strings.Join(args, " "),
				Mode:         execution.Mode(mode),
				AgentID:      agentID,
				Apply:        apply,
				PlanOnly:     planOnly,
				Autonomy:     execution.AutonomyLevel(autonomy),
				BudgetUSD:    budget,
				Sandbox:      sandbox,
				SandboxImage: sandboxImage,
			}
			res, err := ex.Execute(cmd.Context(), req)
			if err != nil && res.Status == "" {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Run:      %s\n", res.RunID)
			fmt.Fprintf(out, "Agent:    %s\n", res.AgentID)
			fmt.Fprintf(out, "Status:   %s\n", res.Status)
			if res.WorktreePath != "" {
				fmt.Fprintf(out, "Worktree: %s\n", res.WorktreePath)
			}
			if res.DiffPath != "" {
				fmt.Fprintf(out, "Diff:     %s\n", res.DiffPath)
			}
			fmt.Fprintf(out, "Report:   %s\n", res.ReportPath)
			if len(res.ChangedFiles) > 0 {
				fmt.Fprintf(out, "Files:    %d\n", len(res.ChangedFiles))
			}
			if res.ErrorMessage != "" {
				fmt.Fprintf(out, "Error:    %s: %s\n", res.ErrorType, res.ErrorMessage)
			}
			fmt.Fprintln(out)
			fmt.Fprintf(out, "Next:\n  harness runs inspect %s\n  harness runs report %s\n",
				filepath.Base(res.RunID), filepath.Base(res.RunID))
			if res.Status == execution.StatusWaitingApproval {
				fmt.Fprintf(out, "  harness runs approve %s\n  harness runs discard %s\n", res.RunID, res.RunID)
			}
			return err
		},
	}
	c.Flags().StringVar(&agentID, "agent", "fake-real", "agent id from the registry")
	c.Flags().StringVar(&mode, "mode", "feature", "feature|bugfix|ask|review")
	c.Flags().BoolVar(&apply, "apply", false, "apply diff to project root after gate allow")
	c.Flags().BoolVar(&planOnly, "plan-only", false, "skip agent invocation")
	c.Flags().StringVar(&autonomy, "autonomy", "safe_execute", "manual|plan_and_ask|safe_execute|full_project_loop|scheduled_maintenance")
	c.Flags().Float64Var(&budget, "budget-usd", 1.0, "budget cap in USD")
	c.Flags().StringVar(&sandbox, "sandbox", "host", "host|container — run the agent inside the selected runtime")
	c.Flags().StringVar(&sandboxImage, "sandbox-image", "", "image to use when --sandbox container (default alpine:3.20)")
	return c
}
