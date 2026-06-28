package main

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/costcomparecmd"
)

func newCostCompareCmd() *cobra.Command {
	var (
		outBudgetTokens int
		effort          string
		showCriteria    bool
	)
	c := &cobra.Command{
		Use:   "cost-compare <prompt>",
		Short: "Honest tokens + $ — harness vs every bundled LLM, with effort tier + task-to-model criteria",
		Long: `Builds the deterministic context pack the harness would send, applies an
effort tier (low / medium / high — controls output multiplier + extended
thinking overhead), and prints the cost per bundled model (Claude Opus 4.7,
Opus 4.6 Fast, Sonnet 4.6 + 1M, Haiku 4.5, GPT-5/mini/nano, Gemini 2.5
Pro/Flash/Flash-Lite, Kimi K2 instruct/thinking, DeepSeek-V3, Llama 3.1
local) sorted by total USD.

Pass --criteria to also print the task -> model + effort recommendation
table with the criteria the bundled router uses (read-only, quick edit,
feature, refactor, codebase scan, design, security, debug-deep, cheap
review, planning, frontend impl).

No agent is invoked. Harness column is always $0 — deterministic, 0 LLM tokens.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			opts := costcomparecmd.Options{
				Root:         root,
				Prompt:       strings.Join(args, " "),
				OutputTokens: outBudgetTokens,
				Effort:       costcomparecmd.Effort(effort),
				ShowCriteria: showCriteria,
			}
			res, err := costcomparecmd.Estimate(cmd.Context(), opts)
			if err != nil {
				return err
			}
			return costcomparecmd.Render(cmd.OutOrStdout(), res, opts)
		},
	}
	c.Flags().IntVar(&outBudgetTokens, "out-tokens", 4000, "estimated LLM output budget before effort multiplier (tokens)")
	c.Flags().StringVar(&effort, "effort", "medium", "reasoning effort: low | medium | high (controls output multiplier + extended-thinking overhead)")
	c.Flags().BoolVar(&showCriteria, "criteria", false, "also print task -> (model, effort) recommendation table with rationale")
	return c
}
