package costcomparecmd

import (
	stdctx "context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	hxcontext "github.com/ropeixoto/harnessx/internal/context"
)

type ModelPrice struct {
	Model         string
	InputUSDPerM  float64
	OutputUSDPerM float64
	Vendor        string
	Notes         string
}

var BundledModelPrices = []ModelPrice{
	{"claude-opus-4-7-1m", 15.00, 75.00, "anthropic", "1M context, most capable; default Claude Code Opus tier"},
	{"claude-opus-4-6-fast", 15.00, 75.00, "anthropic", "Fast mode (/fast); Opus 4.6 latency-optimized, same pricing"},
	{"claude-sonnet-4-6", 3.00, 15.00, "anthropic", "everyday tasks; 200k context"},
	{"claude-sonnet-4-6-1m", 6.00, 22.50, "anthropic", "1M context tier; billed 2x at $6/$22.50"},
	{"claude-haiku-4-5", 0.80, 4.00, "anthropic", "fastest, quick answers, cheapest Claude"},
	{"gpt-5", 1.25, 10.00, "openai", "Codex default; large reasoning models tier"},
	{"gpt-5-mini", 0.25, 2.00, "openai", "cheap codex; default for short tasks"},
	{"gpt-5-nano", 0.05, 0.40, "openai", "smallest, edge"},
	{"gemini-2.5-pro", 1.25, 5.00, "google", "≤200k input tier (>200k: $2.50/$15)"},
	{"gemini-2.5-flash", 0.30, 2.50, "google", "fast multimodal"},
	{"gemini-2.5-flash-lite", 0.10, 0.40, "google", "edge / batch"},
	{"kimi-k2-instruct", 0.60, 2.50, "moonshot", "256k context; strong on planning + search"},
	{"kimi-k2-thinking", 0.60, 2.50, "moonshot", "reasoning tier; price same as instruct"},
	{"deepseek-v3", 0.27, 1.10, "deepseek", "cheap large code model"},
	{"llama-3.1-70b", 0.00, 0.00, "ollama-local", "local (no API spend); compute time only"},
}

type Effort string

const (
	EffortLow    Effort = "low"
	EffortMedium Effort = "medium"
	EffortHigh   Effort = "high"
)

type EffortProfile struct {
	Level             Effort
	OutputMultiplier  float64
	ReasoningOverhead int
	Rationale         string
}

var BundledEfforts = []EffortProfile{
	{EffortLow, 1.0, 0, "no extended reasoning; fast turn-around; cheap"},
	{EffortMedium, 2.0, 1500, "extended thinking ~1.5k reasoning tokens; default for code-gen"},
	{EffortHigh, 4.0, 6000, "deep reasoning ~6k reasoning tokens; only when low/medium failed"},
}

type TaskKind string

const (
	TaskReadOnly     TaskKind = "read_only"
	TaskQuickEdit    TaskKind = "quick_edit"
	TaskFeature      TaskKind = "feature"
	TaskRefactor     TaskKind = "refactor"
	TaskCodebaseScan TaskKind = "codebase_scan"
	TaskDesign       TaskKind = "design"
	TaskSecurity     TaskKind = "security"
	TaskDebugDeep    TaskKind = "debug_deep"
	TaskCheapReview  TaskKind = "cheap_review"
	TaskPlanning     TaskKind = "planning"
	TaskFrontendImpl TaskKind = "frontend_impl"
)

type Recommendation struct {
	Task   TaskKind
	Model  string
	Effort Effort
	Why    string
}

var BundledRecommendations = []Recommendation{
	{TaskReadOnly, "(no LLM — deterministic)", EffortLow, "harness ci/ask/route/explain/plan/cost-compare all run zero-token. Use the binary, save the spend."},
	{TaskQuickEdit, "claude-haiku-4-5", EffortLow, "rename / typo / one-liner. Haiku is fastest + cheapest Claude; effort=low because no reasoning needed."},
	{TaskFeature, "claude-sonnet-4-6", EffortMedium, "default code-gen sweet spot. $3/$15 is the cheap Sonnet tier; medium effort buys ~1.5k reasoning tokens."},
	{TaskRefactor, "claude-sonnet-4-6", EffortMedium, "structural changes across files; Sonnet 4.6 handles 200k context, medium effort prevents cargo-culting."},
	{TaskCodebaseScan, "kimi-k2-instruct", EffortLow, "Kimi 256k context + cheap = lowest-cost path for 'where is X defined' across big repos. Beats Sonnet 4-5x on tokens-per-dollar for pure retrieval."},
	{TaskDesign, "claude-opus-4-7-1m", EffortHigh, "design-to-product needs deep reasoning + large context. Opus + high effort = best per-token quality for design conformance."},
	{TaskSecurity, "claude-opus-4-7-1m", EffortHigh, "security review wants chain-of-thought; Opus high effort. Pay 30x Haiku, get 10x the catch rate."},
	{TaskDebugDeep, "claude-opus-4-7-1m", EffortHigh, "only when Sonnet + medium has already failed twice. Opus high reasoning ~6k tokens to chase the bug."},
	{TaskCheapReview, "gemini-2.5-flash", EffortLow, "lint-style PR review; Gemini Flash is cheapest per output token. Low effort = no reasoning overhead."},
	{TaskPlanning, "kimi-k2-thinking", EffortMedium, "planning benefits from extended-context retrieval (Kimi 256k) AND reasoning. Cheaper than Opus for the same spec quality."},
	{TaskFrontendImpl, "claude-sonnet-4-6", EffortMedium, "TS/React typing is non-trivial; Sonnet covers JSX semantics well at medium reasoning."},
}

type Options struct {
	Root         string
	Prompt       string
	OutputTokens int
	Models       []ModelPrice
	Effort       Effort
	ShowCriteria bool
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
	effort := opts.Effort
	if effort == "" {
		effort = EffortMedium
	}
	profile := profileFor(effort)
	outputTokens := opts.OutputTokens + profile.ReasoningOverhead
	multiplier := profile.OutputMultiplier

	fmt.Fprintf(out, "harness cost-compare\n")
	fmt.Fprintf(out, "  context pack:           %d files, ~%d input tokens, built in %s (deterministic)\n", res.Files, res.ContextTokens, res.Duration.Round(time.Millisecond))
	fmt.Fprintf(out, "  output budget assumed:  %d tokens\n", opts.OutputTokens)
	fmt.Fprintf(out, "  effort tier:            %s (output × %.1f, +%d reasoning tokens — %s)\n", effort, multiplier, profile.ReasoningOverhead, profile.Rationale)
	fmt.Fprintf(out, "  effective output:       %d tokens (%d budget × %.1f + %d reasoning)\n\n", int(float64(opts.OutputTokens)*multiplier)+profile.ReasoningOverhead, opts.OutputTokens, multiplier, profile.ReasoningOverhead)

	fmt.Fprintln(out, "Formula:  cost = in_tokens/1M × in$/Mtok  +  out_tokens/1M × out$/Mtok")
	fmt.Fprintf(out, "          in_tokens = %d (context pack)\n", res.ContextTokens)
	fmt.Fprintf(out, "          out_tokens = %d × %.1f + %d (reasoning) = %d\n\n", opts.OutputTokens, multiplier, profile.ReasoningOverhead, int(float64(opts.OutputTokens)*multiplier)+profile.ReasoningOverhead)

	rows := computeRows(res, outputTokens, multiplier, models)
	fmt.Fprintf(out, "%-22s %-10s %10s %10s %14s  %s\n", "MODEL", "VENDOR", "IN$/Mtok", "OUT$/Mtok", "USD", "NOTES")
	fmt.Fprintf(out, "%-22s %-10s %10s %10s %14s  %s\n", "harness (no LLM)", "local", "-", "-", "$0.0000", "deterministic — zero tokens")
	for _, r := range rows {
		fmt.Fprintf(out, "%-22s %-10s %10s %10s %14s  %s\n",
			r.Model, r.Vendor,
			fmt.Sprintf("$%.2f", r.InputUSDPerM),
			fmt.Sprintf("$%.2f", r.OutputUSDPerM),
			fmt.Sprintf("$%.4f", r.Total),
			r.Notes)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Notes:")
	fmt.Fprintln(out, "  - Token estimates are deterministic (BPE approximation, not exact tokenizer).")
	fmt.Fprintln(out, "  - LLM column = direct API call (your card), not what harness charges.")
	fmt.Fprintln(out, "  - For runs that DO call an LLM (--agent), real cost lands in `harness metrics`.")
	fmt.Fprintln(out, "  - LOWER-BOUND ONLY: real code-gen typically consumes 3-6× more input tokens than the context-pack count above.")

	if opts.ShowCriteria {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Task → model + effort criteria (bundled router defaults):")
		fmt.Fprintf(out, "%-15s %-22s %-7s %s\n", "TASK", "MODEL", "EFFORT", "WHY")
		for _, r := range BundledRecommendations {
			fmt.Fprintf(out, "%-15s %-22s %-7s %s\n", r.Task, r.Model, r.Effort, r.Why)
		}
		fmt.Fprintln(out, "\nOverride per task: harness config set <task> <adapter> --model <name> --effort <low|medium|high>")
	}
	return nil
}

type rowResult struct {
	Model         string
	Vendor        string
	Notes         string
	InputUSDPerM  float64
	OutputUSDPerM float64
	Total         float64
}

func computeRows(res Result, outTokens int, mult float64, models []ModelPrice) []rowResult {
	out := make([]rowResult, 0, len(models))
	for _, m := range models {
		inCost := float64(res.ContextTokens) / 1e6 * m.InputUSDPerM
		outCost := float64(outTokens) / 1e6 * m.OutputUSDPerM
		out = append(out, rowResult{
			Model: m.Model, Vendor: m.Vendor, Notes: m.Notes,
			InputUSDPerM: m.InputUSDPerM, OutputUSDPerM: m.OutputUSDPerM,
			Total: inCost + outCost,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Total < out[j].Total })
	_ = mult
	_ = strings.Builder{}
	return out
}

func profileFor(e Effort) EffortProfile {
	for _, p := range BundledEfforts {
		if p.Level == e {
			return p
		}
	}
	return BundledEfforts[1]
}

func RecommendationFor(task TaskKind) (Recommendation, bool) {
	for _, r := range BundledRecommendations {
		if r.Task == task {
			return r, true
		}
	}
	return Recommendation{}, false
}
