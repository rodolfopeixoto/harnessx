// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

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
			runs, err := execution.ListRuns(root)
			if err != nil {
				return err
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(runs)
			}
			return renderRunsTable(cmd.OutOrStdout(), runs)
		},
	}
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	return c
}

func renderRunsTable(out io.Writer, runs []execution.Result) error {
	if len(runs) == 0 {
		fmt.Fprintln(out, "no runs")
		return nil
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RUN ID\tAGENT\tMODE\tSTATUS\tFILES\tCOST")
	for _, r := range runs {
		mode := r.AgentID
		_ = mode
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t$%.4f\n",
			r.RunID, r.AgentID, "-", r.Status, len(r.ChangedFiles), r.EstimatedCostUSD)
	}
	return w.Flush()
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
			r, err := execution.LoadRun(root, args[0])
			if err != nil {
				return err
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(r)
			}
			return renderRunDetail(cmd.OutOrStdout(), r)
		},
	}
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	return c
}

func renderRunDetail(out io.Writer, r execution.Result) error {
	fmt.Fprintf(out, "Run:        %s\nAgent:      %s\nStatus:     %s\nStarted:    %s\nFinished:   %s\nWorktree:   %s\nStdout:     %s\nStderr:     %s\nDiff:       %s\nReport:     %s\nFiles:      %d\nTokens:     in=%d out=%d\nCost (est): $%.4f (exact=%t)\n",
		r.RunID, r.AgentID, r.Status, r.StartedAt.Format("2006-01-02 15:04:05"), r.FinishedAt.Format("2006-01-02 15:04:05"),
		r.WorktreePath, r.StdoutPath, r.StderrPath, r.DiffPath, r.ReportPath,
		len(r.ChangedFiles), r.InputTokens, r.OutputTokens, r.EstimatedCostUSD, r.ExactUsageAvailable)
	if len(r.Sensors) > 0 {
		fmt.Fprintln(out, "Sensors:")
		for _, s := range r.Sensors {
			fmt.Fprintf(out, "  - %s [%s] %dms\n", s.ID, s.Status, s.DurationMs)
		}
	}
	if len(r.MCPDetectedNotActive) > 0 {
		fmt.Fprintf(out, "\nMCP configs detected but not injected yet (P32): %d\n", len(r.MCPDetectedNotActive))
	}
	if len(r.HooksDetectedNotActive) > 0 {
		fmt.Fprintf(out, "Hooks detected but not executed yet (P32): %d\n", len(r.HooksDetectedNotActive))
	}
	if r.ErrorMessage != "" {
		fmt.Fprintf(out, "\nError: %s: %s\n", r.ErrorType, r.ErrorMessage)
	}
	return nil
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
			r, err := execution.LoadRun(root, args[0])
			if err != nil {
				return err
			}
			data, err := os.ReadFile(r.ReportPath)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
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
			r, err := execution.LoadRun(root, args[0])
			if err != nil {
				return err
			}
			if len(r.Sensors) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no sensors recorded")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tSTATUS\tMS\tOUTPUT")
			for _, s := range r.Sensors {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", s.ID, s.Status, s.DurationMs, s.Output)
			}
			return w.Flush()
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
