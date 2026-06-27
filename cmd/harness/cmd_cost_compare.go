package main

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/costcomparecmd"
)

func newCostCompareCmd() *cobra.Command {
	var outBudgetTokens int
	c := &cobra.Command{
		Use:   "cost-compare <prompt>",
		Short: "Compare deterministic harness vs a real LLM run, in honest tokens and dollars",
		Long: `Builds the deterministic context pack the harness would actually send,
estimates the input tokens, applies an output-token budget, and prints the
$ each bundled model would cost for the same job. No agent is invoked.

The harness column is the actual local work: 0 LLM tokens. The model
columns are the equivalent direct-LLM cost.`,
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
			}
			res, err := costcomparecmd.Estimate(cmd.Context(), opts)
			if err != nil {
				return err
			}
			return costcomparecmd.Render(cmd.OutOrStdout(), res, opts)
		},
	}
	c.Flags().IntVar(&outBudgetTokens, "out-tokens", 4000, "estimated LLM output budget (tokens)")
	return c
}
