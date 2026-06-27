package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	hxcontext "github.com/ropeixoto/harnessx/internal/context"
)

type modelPrice struct {
	Model         string
	InputUSDPerM  float64
	OutputUSDPerM float64
}

var bundledModelPrices = []modelPrice{
	{"claude-opus-4-7", 15.00, 75.00},
	{"claude-sonnet-4-6", 3.00, 15.00},
	{"claude-haiku-4-5", 0.80, 4.00},
	{"gpt-5", 5.00, 25.00},
	{"gpt-5-mini", 0.40, 1.60},
	{"gemini-2.5-pro", 3.50, 17.50},
	{"gemini-2.5-flash", 0.30, 1.20},
}

func newCostCompareCmd() *cobra.Command {
	var prompt string
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
			prompt = strings.Join(args, " ")
			root, err := cwd()
			if err != nil {
				return err
			}
			start := time.Now()
			pack, err := hxcontext.Build(cmd.Context(), hxcontext.Options{Root: root, Task: prompt})
			if err != nil {
				return err
			}
			elapsed := time.Since(start)
			out := cmd.OutOrStdout()
			return renderCostCompare(out, pack.Stats.EstimatedTokens, outBudgetTokens, pack.Stats.FilesCount, elapsed)
		},
	}
	c.Flags().IntVar(&outBudgetTokens, "out-tokens", 4000, "estimated LLM output budget (tokens)")
	return c
}

func renderCostCompare(out io.Writer, contextTokens, outputTokens, files int, dur time.Duration) error {
	fmt.Fprintf(out, "harness cost-compare\n")
	fmt.Fprintf(out, "  deterministic context pack: %d files, ~%d input tokens, built in %s\n", files, contextTokens, dur.Round(time.Millisecond))
	fmt.Fprintf(out, "  output budget assumed for the LLM column: %d tokens\n\n", outputTokens)
	fmt.Fprintf(out, "%-22s %12s %14s %14s %14s\n", "PATH", "INPUT_TOK", "OUTPUT_TOK", "USD", "NOTES")
	fmt.Fprintf(out, "%-22s %12d %14d %14s %s\n", "harness (no LLM)", 0, 0, "$0.0000", "deterministic")
	for _, m := range bundledModelPrices {
		inCost := float64(contextTokens) / 1e6 * m.InputUSDPerM
		outCost := float64(outputTokens) / 1e6 * m.OutputUSDPerM
		total := inCost + outCost
		fmt.Fprintf(out, "%-22s %12d %14d %14s %s\n", m.Model, contextTokens, outputTokens, fmt.Sprintf("$%.4f", total), "direct LLM call")
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Notes:")
	fmt.Fprintln(out, "  - Token estimates are deterministic (BPE approximation, not exact tokenizer).")
	fmt.Fprintln(out, "  - LLM column is what a direct API call would cost, not what harness charges.")
	fmt.Fprintln(out, "  - For commands that DO call an LLM (--agent), real cost lands in `harness metrics`.")
	return nil
}
