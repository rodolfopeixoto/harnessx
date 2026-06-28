package learncmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/execution"
)

type Pattern struct {
	Title         string
	Detail        string
	Severity      string
	Suggestion    string
	Evidence      []string
	EstSavings    float64
	Deterministic bool
}

type Options struct {
	Root      string
	WriteFile bool
}

type Result struct {
	Patterns     []Pattern
	RunsAnalyzed int
	TokensTotal  int
	CostTotal    float64
	WrittenPath  string
}

func Run(out io.Writer, opts Options) (Result, error) {
	runs, err := execution.ListRuns(opts.Root)
	if err != nil {
		return Result{}, err
	}
	res := Result{RunsAnalyzed: len(runs)}
	stats := analyze(runs, &res)
	res.Patterns = derivePatterns(stats, runs)
	render(out, res)
	if opts.WriteFile {
		path, err := writeMemoryFile(opts.Root, res)
		if err != nil {
			fmt.Fprintf(out, "  warn: write memory file: %v\n", err)
		} else {
			res.WrittenPath = path
			fmt.Fprintf(out, "  memory: written to %s\n", path)
		}
	}
	return res, nil
}

type runStats struct {
	byAdapter         map[string]int
	byStatus          map[string]int
	byErrorType       map[string]int
	tokensByAdapter   map[string]int
	costByAdapter     map[string]float64
	avgFilesByAdapter map[string]float64
}

func analyze(runs []execution.Result, res *Result) runStats {
	s := runStats{
		byAdapter:         map[string]int{},
		byStatus:          map[string]int{},
		byErrorType:       map[string]int{},
		tokensByAdapter:   map[string]int{},
		costByAdapter:     map[string]float64{},
		avgFilesByAdapter: map[string]float64{},
	}
	filesAccum := map[string]int{}
	for _, r := range runs {
		s.byAdapter[r.AgentID]++
		s.byStatus[string(r.Status)]++
		if r.ErrorType != "" {
			s.byErrorType[r.ErrorType]++
		}
		s.tokensByAdapter[r.AgentID] += r.InputTokens + r.OutputTokens
		s.costByAdapter[r.AgentID] += r.EstimatedCostUSD
		filesAccum[r.AgentID] += len(r.ChangedFiles)
		res.TokensTotal += r.InputTokens + r.OutputTokens
		res.CostTotal += r.EstimatedCostUSD
	}
	for adapter, count := range s.byAdapter {
		if count > 0 {
			s.avgFilesByAdapter[adapter] = float64(filesAccum[adapter]) / float64(count)
		}
	}
	return s
}

func derivePatterns(s runStats, runs []execution.Result) []Pattern {
	var out []Pattern
	if dom, count := topKey(s.byAdapter); count >= 3 {
		share := float64(count) / float64(sumValues(s.byAdapter)) * 100
		out = append(out, Pattern{
			Title:    fmt.Sprintf("Most-used adapter: %s (%d runs, %.0f%% of total)", dom, count, share),
			Detail:   fmt.Sprintf("Adapter %s handled %d of %d runs.", dom, count, sumValues(s.byAdapter)),
			Severity: "info",
		})
		if avg, ok := s.avgFilesByAdapter[dom]; ok && avg > 50 {
			out = append(out, Pattern{
				Title:         fmt.Sprintf("Worktree leakage with %s — avg %.0f files/run", dom, avg),
				Detail:        "More than 50 files per run usually means .venv / node_modules / __pycache__ leaked into the worktree diff.",
				Severity:      "high",
				Suggestion:    "Update .harness/worktrees/<id>/.git/info/exclude with venv/, node_modules/, __pycache__/ — `harness fix-env` adds these automatically.",
				Deterministic: true,
			})
		}
	}
	if fail, count := topKey(s.byErrorType); count >= 2 {
		out = append(out, Pattern{
			Title:         fmt.Sprintf("Recurring failure: %s (%d times)", fail, count),
			Detail:        "Same error_type fired more than once — pin a deterministic fix instead of paying for retries.",
			Severity:      "high",
			Suggestion:    lookupFixForError(fail),
			Deterministic: true,
		})
	}
	if waiting := s.byStatus[string(execution.StatusWaitingApproval)]; waiting >= 3 {
		out = append(out, Pattern{
			Title:         fmt.Sprintf("%d runs stuck in waiting_approval", waiting),
			Detail:        "Autonomy gate is deferring apply. Either approve them in batch or raise autonomy to safe_execute for low-risk modes.",
			Severity:      "medium",
			Suggestion:    "Run `harness runs list` to triage; `harness runs approve <id>` per item, or `harness autonomy set safe_execute` to auto-apply.",
			Deterministic: true,
		})
	}
	if best, savings := cheapestAdapter(s); best != "" && savings > 0.10 {
		out = append(out, Pattern{
			Title:      fmt.Sprintf("Cheaper adapter available: %s saves ~$%.2f vs current dominant choice", best, savings),
			Detail:     "Per-run cost is significantly lower with this adapter for similar workloads.",
			Severity:   "info",
			Suggestion: fmt.Sprintf("Pin with `harness use %s` or route per task via `harness config set <task> %s`.", best, best),
			EstSavings: savings,
		})
	}
	if pendingTODOs := countOrphanReports(runs); pendingTODOs >= 5 {
		out = append(out, Pattern{
			Title:         fmt.Sprintf("%d orphan run dirs (no meta.json)", pendingTODOs),
			Detail:        "Likely from a previous version that wrote report.md without meta.json. Safe to prune.",
			Severity:      "low",
			Suggestion:    "Run `harness runs prune` to clean.",
			Deterministic: true,
		})
	}
	return out
}

func render(out io.Writer, res Result) {
	fmt.Fprintf(out, "memory learn — %d runs analyzed\n", res.RunsAnalyzed)
	fmt.Fprintf(out, "  total tokens: %d (in+out)\n", res.TokensTotal)
	fmt.Fprintf(out, "  total cost:   $%.4f\n", res.CostTotal)
	if len(res.Patterns) == 0 {
		fmt.Fprintln(out, "  no patterns surfaced")
		return
	}
	for _, p := range res.Patterns {
		fmt.Fprintf(out, "[%s] %s\n", p.Severity, p.Title)
		if p.Detail != "" {
			fmt.Fprintf(out, "  detail: %s\n", p.Detail)
		}
		if p.Suggestion != "" {
			fmt.Fprintf(out, "  fix:    %s\n", p.Suggestion)
		}
		if p.Deterministic {
			fmt.Fprintln(out, "  kind:   deterministic (no LLM tokens spent on the fix)")
		}
	}
}

func writeMemoryFile(root string, res Result) (string, error) {
	dir := filepath.Join(root, ".harness", "memory")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "learned-patterns.json")
	body, err := json.MarshalIndent(struct {
		GeneratedAt time.Time `json:"generated_at"`
		Result      Result    `json:"result"`
	}{time.Now().UTC(), res}, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func topKey(m map[string]int) (string, int) {
	bestK, bestV := "", 0
	for k, v := range m {
		if k == "" {
			continue
		}
		if v > bestV {
			bestK, bestV = k, v
		}
	}
	return bestK, bestV
}

func sumValues(m map[string]int) int {
	t := 0
	for _, v := range m {
		t += v
	}
	return t
}

func cheapestAdapter(s runStats) (string, float64) {
	type entry struct {
		adapter string
		costPer float64
		runs    int
	}
	var entries []entry
	for adapter, runs := range s.byAdapter {
		if runs < 2 || adapter == "" {
			continue
		}
		entries = append(entries, entry{adapter, s.costByAdapter[adapter] / float64(runs), runs})
	}
	if len(entries) < 2 {
		return "", 0
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].costPer < entries[j].costPer })
	cheapest := entries[0]
	mostUsed := entries[0]
	for _, e := range entries {
		if e.runs > mostUsed.runs {
			mostUsed = e
		}
	}
	if cheapest.adapter == mostUsed.adapter {
		return "", 0
	}
	savings := (mostUsed.costPer - cheapest.costPer) * float64(mostUsed.runs)
	if savings <= 0 {
		return "", 0
	}
	return cheapest.adapter, savings
}

func countOrphanReports(runs []execution.Result) int {
	count := 0
	for _, r := range runs {
		if r.Status == execution.StatusIncomplete {
			count++
		}
	}
	return count
}

func lookupFixForError(errType string) string {
	switch strings.ToLower(errType) {
	case "budget_exceeded":
		return "Raise --budget-usd OR pin a cheaper adapter (`harness use claude-haiku-4-5`)."
	case "pre_hook_blocked":
		return "Inspect `harness hook list`; the pre-tool-use hook returned non-zero. Fix the script or disable for this run."
	case "worktree_prepare":
		return "Check that the project root is a git repo OR has write permissions for `.harness/worktrees/`."
	}
	return "Inspect the run report (`harness runs report <id>`) and the agent stderr for the root cause."
}
