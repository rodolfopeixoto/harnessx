package costcomparecmd

import (
	stdctx "context"
	"fmt"
	"io"
	"time"

	hxcontext "github.com/ropeixoto/harnessx/internal/context"
)

type ModelPrice struct {
	Model         string
	InputUSDPerM  float64
	OutputUSDPerM float64
}

var BundledModelPrices = []ModelPrice{
	{"claude-opus-4-7", 15.00, 75.00},
	{"claude-sonnet-4-6", 3.00, 15.00},
	{"claude-haiku-4-5", 0.80, 4.00},
	{"gpt-5", 5.00, 25.00},
	{"gpt-5-mini", 0.40, 1.60},
	{"gemini-2.5-pro", 3.50, 17.50},
	{"gemini-2.5-flash", 0.30, 1.20},
}

type Options struct {
	Root         string
	Prompt       string
	OutputTokens int
	Models       []ModelPrice
}

type Result struct {
	ContextTokens int
	Files         int
	Duration      time.Duration
}

func Estimate(ctx stdctx.Context, opts Options) (Result, error) {
	start := time.Now()
	pack, err := hxcontext.Build(ctx, hxcontext.Options{Root: opts.Root, Task: opts.Prompt})
	if err != nil {
		return Result{}, err
	}
	return Result{
		ContextTokens: pack.Stats.EstimatedTokens,
		Files:         pack.Stats.FilesCount,
		Duration:      time.Since(start),
	}, nil
}

func Render(out io.Writer, res Result, opts Options) error {
	models := opts.Models
	if len(models) == 0 {
		models = BundledModelPrices
	}
	fmt.Fprintf(out, "harness cost-compare\n")
	fmt.Fprintf(out, "  deterministic context pack: %d files, ~%d input tokens, built in %s\n", res.Files, res.ContextTokens, res.Duration.Round(time.Millisecond))
	fmt.Fprintf(out, "  output budget assumed for the LLM column: %d tokens\n\n", opts.OutputTokens)
	fmt.Fprintf(out, "%-22s %12s %14s %14s %14s\n", "PATH", "INPUT_TOK", "OUTPUT_TOK", "USD", "NOTES")
	fmt.Fprintf(out, "%-22s %12d %14d %14s %s\n", "harness (no LLM)", 0, 0, "$0.0000", "deterministic")
	for _, m := range models {
		inCost := float64(res.ContextTokens) / 1e6 * m.InputUSDPerM
		outCost := float64(opts.OutputTokens) / 1e6 * m.OutputUSDPerM
		total := inCost + outCost
		fmt.Fprintf(out, "%-22s %12d %14d %14s %s\n", m.Model, res.ContextTokens, opts.OutputTokens, fmt.Sprintf("$%.4f", total), "direct LLM call")
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Notes:")
	fmt.Fprintln(out, "  - Token estimates are deterministic (BPE approximation, not exact tokenizer).")
	fmt.Fprintln(out, "  - LLM column is what a direct API call would cost, not what harness charges.")
	fmt.Fprintln(out, "  - For commands that DO call an LLM (--agent), real cost lands in `harness metrics`.")
	fmt.Fprintln(out, "  - LOWER-BOUND ONLY: real code-gen runs typically consume 3-6× more input tokens than")
	fmt.Fprintln(out, "    the context-pack count above (full session history + system prompts + tool messages).")
	fmt.Fprintln(out, "    The harness $0 row applies to read-only / deterministic work — for code-gen the LLM")
	fmt.Fprintln(out, "    is required and real cost will be 3-6× the figure above. Trust `harness metrics`.")
	return nil
}
