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
	Apply     bool
}

type Result struct {
	Patterns     []Pattern
	RunsAnalyzed int
	TokensTotal  int
	CostTotal    float64
	WrittenPath  string
	Applied      []string
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
	if opts.Apply {
		applied := applyDeterministic(out, opts.Root, res.Patterns, runs)
		res.Applied = applied
		if len(applied) > 0 {
			fmt.Fprintf(out, "  applied %d deterministic fix(es): %s\n", len(applied), strings.Join(applied, ", "))
		} else {
			fmt.Fprintln(out, "  no deterministic fixes available for the surfaced patterns")
		}
	}
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

func applyDeterministic(out io.Writer, root string, patterns []Pattern, runs []execution.Result) []string {
	var applied []string
	for _, p := range patterns {
		if !p.Deterministic {
			continue
		}
		switch {
		case strings.Contains(p.Title, "orphan run dirs"):
			n := pruneOrphanRuns(root, runs)
			if n > 0 {
				fmt.Fprintf(out, "  ✓ pruned %d orphan run dir(s)\n", n)
				applied = append(applied, "prune_orphans")
			}
		case strings.Contains(p.Title, "Worktree leakage"):
			fmt.Fprintln(out, "  ✓ refreshed .git/info/exclude in every worktree under .harness/worktrees")
			refreshWorktreeExcludes(root)
			applied = append(applied, "refresh_worktree_excludes")
		case strings.Contains(p.Title, "waiting_approval"):
			fmt.Fprintln(out, "  ✓ pinned autonomy=safe_execute (was manual); next runs auto-apply low-risk diffs")
			_ = writeFile(filepath.Join(root, ".harness", "config", "autonomy"), "safe_execute\n")
			applied = append(applied, "autonomy_safe_execute")
		}
	}
	return applied
}

func pruneOrphanRuns(root string, runs []execution.Result) int {
	n := 0
	for _, r := range runs {
		if r.Status == execution.StatusIncomplete {
			path := filepath.Join(root, ".harness", "runs", r.RunID)
			if err := os.RemoveAll(path); err == nil {
				n++
			}
		}
	}
	return n
}

func refreshWorktreeExcludes(root string) {
	wtRoot := filepath.Join(root, ".harness", "worktrees")
	entries, err := os.ReadDir(wtRoot)
	if err != nil {
		return
	}
	body := "# harness memory-learn refresh\n"
	for _, d := range []string{".venv", "venv", "__pycache__", "node_modules", "target", "dist", "build", "_build", "deps", ".gradle", ".idea", ".pytest_cache", ".mypy_cache", ".ruff_cache", "bin", "obj", "coverage"} {
		body += d + "/\n"
	}
	body += "*.tsbuildinfo\n*.log\n"
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info := filepath.Join(wtRoot, e.Name(), ".git", "info")
		_ = os.MkdirAll(info, 0o755)
		path := filepath.Join(info, "exclude")
		existing, _ := os.ReadFile(path)
		_ = os.WriteFile(path, []byte(string(existing)+"\n"+body), 0o644)
	}
}

func writeFile(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

type Incremental struct {
	GeneratedAt   time.Time      `json:"generated_at"`
	RunsSeen      int            `json:"runs_seen"`
	LastRunID     string         `json:"last_run_id"`
	TokensTotal   int            `json:"tokens_total"`
	CostTotal     float64        `json:"cost_total"`
	ByAdapter     map[string]int `json:"by_adapter"`
	ByStatus      map[string]int `json:"by_status"`
	ByErrorType   map[string]int `json:"by_error_type,omitempty"`
	TokensPerAdpt map[string]int `json:"tokens_per_adapter,omitempty"`
}

func LoadIncremental(root string) (Incremental, error) {
	path := filepath.Join(root, ".harness", "memory", "incremental.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Incremental{
				ByAdapter:     map[string]int{},
				ByStatus:      map[string]int{},
				ByErrorType:   map[string]int{},
				TokensPerAdpt: map[string]int{},
			}, nil
		}
		return Incremental{}, err
	}
	var inc Incremental
	if err := json.Unmarshal(data, &inc); err != nil {
		return Incremental{}, err
	}
	if inc.ByAdapter == nil {
		inc.ByAdapter = map[string]int{}
	}
	if inc.ByStatus == nil {
		inc.ByStatus = map[string]int{}
	}
	if inc.ByErrorType == nil {
		inc.ByErrorType = map[string]int{}
	}
	if inc.TokensPerAdpt == nil {
		inc.TokensPerAdpt = map[string]int{}
	}
	return inc, nil
}

func UpdateIncremental(root string, run execution.Result) (Incremental, string, error) {
	inc, err := LoadIncremental(root)
	if err != nil {
		return Incremental{}, "", err
	}
	if run.RunID == inc.LastRunID {
		return inc, "", nil
	}
	inc.GeneratedAt = time.Now().UTC()
	inc.RunsSeen++
	inc.LastRunID = run.RunID
	inc.TokensTotal += run.InputTokens + run.OutputTokens
	inc.CostTotal += run.EstimatedCostUSD
	inc.ByAdapter[run.AgentID]++
	inc.ByStatus[string(run.Status)]++
	if run.ErrorType != "" {
		inc.ByErrorType[run.ErrorType]++
	}
	inc.TokensPerAdpt[run.AgentID] += run.InputTokens + run.OutputTokens
	path := filepath.Join(root, ".harness", "memory", "incremental.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return inc, "", err
	}
	body, err := json.MarshalIndent(inc, "", "  ")
	if err != nil {
		return inc, "", err
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return inc, "", err
	}
	return inc, path, nil
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
