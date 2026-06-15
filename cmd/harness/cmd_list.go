// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/app/projectcmd"
	"github.com/ropeixoto/harnessx/internal/execution"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Composite read-only view: projects + recent runs + agents",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			dir, err := cwd()
			if err != nil {
				return err
			}

			fmt.Fprintln(out, "## projects")
			if err := projectcmd.List(ctx, projectcmd.Options{}, false, out); err != nil {
				fmt.Fprintf(out, "  (project registry unavailable: %v)\n", err)
			}

			fmt.Fprintln(out, "\n## recent runs")
			runs, err := execution.ListRuns(dir)
			if err != nil || len(runs) == 0 {
				fmt.Fprintln(out, "  (no runs yet)")
			} else {
				renderRecentRuns(out, runs, 10)
			}

			fmt.Fprintln(out, "\n## agents")
			if err := agentcmd.List(out, dir); err != nil {
				fmt.Fprintf(out, "  (agent registry unavailable: %v)\n", err)
			}
			return nil
		},
	}
}

func renderRecentRuns(out io.Writer, runs []execution.Result, limit int) {
	if limit > len(runs) {
		limit = len(runs)
	}
	fmt.Fprintf(out, "  %-32s %-12s %-18s %s\n", "RUN ID", "AGENT", "STATUS", "FILES")
	for i := 0; i < limit; i++ {
		r := runs[i]
		fmt.Fprintf(out, "  %-32s %-12s %-18s %d\n", r.RunID, r.AgentID, r.Status, len(r.ChangedFiles))
	}
}
