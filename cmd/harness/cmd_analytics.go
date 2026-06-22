// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/analytics"
)

func newAnalyticsCmd() *cobra.Command {
	var (
		jsonOut bool
		sinceD  string
		roots   []string
	)
	c := &cobra.Command{
		Use:   "analytics",
		Short: "Aggregate chat session spend per stack / adapter / day",
		Long: `Walks one or more project roots (default: cwd) and aggregates the
chat-session ledger so the user sees where token spend lands. Reads
.harness/sessions/*.jsonl + .meta.json; respects --since for a
rolling window.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(roots) == 0 {
				dir, err := cwd()
				if err != nil {
					return err
				}
				roots = []string{dir}
			}
			var since time.Time
			if sinceD != "" {
				d, err := time.ParseDuration(sinceD)
				if err != nil {
					return fmt.Errorf("analytics: --since %q: %w", sinceD, err)
				}
				since = time.Now().Add(-d)
			}
			rep, err := analytics.Walk(roots, since)
			if err != nil {
				return err
			}
			if jsonOut {
				return analytics.RenderJSON(cmd.OutOrStdout(), rep)
			}
			analytics.Render(cmd.OutOrStdout(), rep)
			return nil
		},
	}
	c.Flags().StringSliceVar(&roots, "root", nil, "project root(s) to walk (default: cwd)")
	c.Flags().BoolVar(&jsonOut, "json", false, "emit a single JSON envelope")
	c.Flags().StringVar(&sinceD, "since", "", "only include turns within the last duration (e.g. 24h, 7d→168h)")
	return c
}
