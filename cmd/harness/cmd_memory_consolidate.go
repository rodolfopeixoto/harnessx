package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/learncmd"
)

func newMemoryConsolidateCmd() *cobra.Command {
	var output string
	c := &cobra.Command{
		Use:   "consolidate",
		Short: "Roll every project's .harness/memory/incremental.json into one global file (no LLM)",
		Long: `Walks the workspace registry, reads each project's incremental memory
file, and writes a roll-up to ~/.config/harness/memory.json (override
with --output). The global file aggregates runs, tokens, cost and
per-adapter counters across every registered project so the user can
spot cross-project optimisation opportunities (e.g. one adapter
dominates every project, time to negotiate a bulk rate).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _, err := learncmd.Consolidate(cmd.Context(), cmd.OutOrStdout(), learncmd.ConsolidateOptions{OutputPath: output})
			return err
		},
	}
	c.Flags().StringVar(&output, "output", "", "explicit output path (default ~/.config/harness/memory.json)")
	return c
}
