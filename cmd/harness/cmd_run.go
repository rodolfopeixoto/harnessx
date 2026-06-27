package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/runscmd"
	"github.com/ropeixoto/harnessx/internal/execution"
)

func newRunsCmd() *cobra.Command {
	c := &cobra.Command{Use: "runs", Short: "Inspect / approve / discard agentic runs"}
	c.AddCommand(newRunListCmd(), newRunInspectCmd(), newRunReportCmd(), newRunSensorsCmd(), newRunApproveCmd(), newRunDiscardCmd(), newRunsPruneCmd())
	return c
}

func newRunListCmd() *cobra.Command {
	var jsonOut bool
	c := &cobra.Command{
		Use:   "list",
		Short: "List agentic runs persisted under .harness/runs/",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			return runscmd.List(cmd.OutOrStdout(), runscmd.Options{Root: root, JSON: jsonOut})
		},
	}
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	return c
}

func newRunInspectCmd() *cobra.Command {
	var jsonOut bool
	c := &cobra.Command{
		Use:   "inspect <run-id>",
		Short: "Show one run's metadata, paths, sensors, and cost",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			return runscmd.Inspect(cmd.OutOrStdout(), runscmd.Options{Root: root, JSON: jsonOut, RunID: args[0]})
		},
	}
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	return c
}

func newRunReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report <run-id>",
		Short: "Print the rendered Markdown report for a run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			return runscmd.Report(cmd.OutOrStdout(), runscmd.Options{Root: root, RunID: args[0]})
		},
	}
}

func newRunSensorsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sensors <run-id>",
		Short: "List sensor outcomes for a run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			return runscmd.Sensors(cmd.OutOrStdout(), runscmd.Options{Root: root, RunID: args[0]})
		},
	}
}

func newRunApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <run-id>",
		Short: "Approve a waiting_approval run and apply the worktree diff",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			r, err := execution.LoadRun(root, args[0])
			if err != nil {
				return err
			}
			if r.Status != execution.StatusWaitingApproval {
				return fmt.Errorf("run is not waiting_approval (status=%s)", r.Status)
			}
			if r.WorktreePath == "" {
				return fmt.Errorf("run has no worktree to apply")
			}
			wt := execution.Worktree{RunID: r.RunID, Kind: "git_worktree", Path: r.WorktreePath}
			runDir := filepath.Join(root, ".harness", "runs", r.RunID)
			if err := execution.ApplyWorktreeDiff(cmd.Context(), root, wt, runDir); err != nil {
				return fmt.Errorf("apply: %w", err)
			}
			mgr := execution.NewManager(root)
			_ = mgr.Cleanup(cmd.Context(), wt)
			r.Status = execution.StatusApplied
			r.WorktreePath = ""
			data, _ := json.MarshalIndent(r, "", "  ")
			_ = os.WriteFile(filepath.Join(runDir, "meta.json"), data, 0o644)
			fmt.Fprintln(cmd.OutOrStdout(), "Applied. Worktree removed.")
			return nil
		},
	}
}

func newRunDiscardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "discard <run-id>",
		Short: "Discard a run's worktree without applying",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			r, err := execution.LoadRun(root, args[0])
			if err != nil {
				return err
			}
			if r.WorktreePath != "" {
				wt := execution.Worktree{RunID: r.RunID, Kind: "git_worktree", Path: r.WorktreePath}
				if err := execution.NewManager(root).Cleanup(cmd.Context(), wt); err != nil {
					return err
				}
			}
			r.Status = execution.StatusDiscarded
			r.WorktreePath = ""
			runDir := filepath.Join(root, ".harness", "runs", r.RunID)
			data, _ := json.MarshalIndent(r, "", "  ")
			_ = os.WriteFile(filepath.Join(runDir, "meta.json"), data, 0o644)
			fmt.Fprintln(cmd.OutOrStdout(), "Discarded.")
			return nil
		},
	}
}
