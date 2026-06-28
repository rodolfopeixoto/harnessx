package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/learncmd"
)

func newMemoryLearnCmd() *cobra.Command {
	var write, apply bool
	c := &cobra.Command{
		Use:   "learn",
		Short: "Analyze .harness/runs/* and surface deterministic optimization patterns (no LLM)",
		Long: `Walks every recorded run, aggregates per-adapter token + cost + status
metrics, and prints actionable patterns:

  - dominant adapter + worktree leakage detection (avg files/run > 50)
  - recurring error_types with the deterministic fix command
  - waiting_approval backlog
  - cheaper adapter recommendation by per-run cost
  - orphan run dirs (no meta.json) — prunable

The output is fully deterministic. With --write the result is also
persisted to .harness/memory/learned-patterns.json for future reads.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			_, err = learncmd.Run(cmd.OutOrStdout(), learncmd.Options{Root: root, WriteFile: write, Apply: apply})
			return err
		},
	}
	c.Flags().BoolVar(&write, "write", false, "also persist the patterns to .harness/memory/learned-patterns.json")
	c.Flags().BoolVar(&apply, "apply", false, "execute the deterministic fixes (prune orphans, refresh worktree excludes, pin autonomy)")
	return c
}
