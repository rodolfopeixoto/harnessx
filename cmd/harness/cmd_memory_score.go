package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/learncmd"
)

func newMemoryScoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "score-skills",
		Short: "Score every skill by run outcome telemetry (no LLM)",
		Long: `Walks .harness/runs/* and joins each run's enhancement.json
(skill_prefixes) with the run's final status and recovery.retries.
Prints a ranked table (SUCCESS_RATE, RUNS, AVG_RETRIES). Writes
.harness/memory/skill-scores.json for downstream consumption.

A skill with high success_rate + low retries is a deterministic win.
A skill with low success_rate is a candidate for disabling via
.harness/config/skills.disabled.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			_, _, err = learncmd.ScoreSkills(root, cmd.OutOrStdout())
			return err
		},
	}
}
