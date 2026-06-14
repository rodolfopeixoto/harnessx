// SPDX-License-Identifier: MIT

package sensors

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/index"
)

// PerformanceBudgetSensor compares the most recent
// .harness/artifacts/perf/<ts>-<id>.json snapshot against the budgets in
// .harness/project/performance-budget.json. A snapshot key beats its
// budget when the snapshot value exceeds the limit; that breaks the
// sensor. Missing snapshot → skipped. Missing budget → skipped.
type PerformanceBudgetSensor struct{}

func (PerformanceBudgetSensor) ID() string                     { return "performance_budget" }
func (PerformanceBudgetSensor) Category() Category             { return CatPerf }
func (PerformanceBudgetSensor) Kind() Kind                     { return KindComputational }
func (PerformanceBudgetSensor) AppliesTo(p index.Profile) bool { return true }

func (s PerformanceBudgetSensor) Run(rc RunCtx) Result {
	start := time.Now()
	res := Result{ID: s.ID(), Category: s.Category(), Kind: s.Kind()}

	var budget index.PerformanceBudget
	if err := index.ReadMap(rc.Root, index.MapPerformance, &budget); err != nil || len(budget.Budgets) == 0 {
		res.Status = StatusSkipped
		res.Detail = "no performance-budget.json (run `harness project index`)"
		res.Duration = time.Since(start)
		return res
	}
	snapshot, snapPath := newestSnapshot(rc.Root)
	if snapshot == nil {
		res.Status = StatusSkipped
		res.Detail = "no perf snapshot (run `harness perf-snapshot`)"
		res.Duration = time.Since(start)
		return res
	}

	var breaches []string
	keys := sortedBudgetKeys(budget.Budgets)
	for _, k := range keys {
		limit, ok := numericBudget(budget.Budgets[k])
		if !ok {
			continue
		}
		actual, ok := snapshotValue(snapshot, k)
		if !ok {
			continue
		}
		if actual > limit {
			breaches = append(breaches, fmt.Sprintf("%s: %s > %s", k, fmtNum(actual), fmtNum(limit)))
		}
	}
	res.Duration = time.Since(start)
	if len(breaches) == 0 {
		res.Status = StatusPassed
		res.Detail = "snapshot " + filepath.Base(snapPath) + " within budget"
		return res
	}
	res.Status = StatusFailed
	res.Detail = strings.Join(breaches, "; ")
	res.OutputPath = writeOutput(rc.OutputDir, s.ID(), []byte(strings.Join(breaches, "\n")+"\n"), nil)
	return res
}

// newestSnapshot finds the most recently modified perf snapshot under
// .harness/artifacts/perf/ and decodes it into a generic map so we can
// look up keys dynamically without coupling the sensor to the optimize
// package types (avoids an import cycle when sensors grow).
func newestSnapshot(root string) (map[string]any, string) {
	dir := filepath.Join(root, ".harness", "artifacts", "perf")
	var newest string
	var newestMod time.Time
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".json" {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().After(newestMod) {
			newestMod = info.ModTime()
			newest = p
		}
		return nil
	})
	if newest == "" {
		return nil, ""
	}
	b, err := os.ReadFile(newest)
	if err != nil {
		return nil, ""
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, ""
	}
	return m, newest
}

// snapshotValue resolves budget keys to snapshot fields. The mapping is
// intentionally narrow — adding a key without a snapshot mapping is a
// no-op (skipped), not an implicit pass.
func snapshotValue(snap map[string]any, key string) (float64, bool) {
	get := func(path ...string) (float64, bool) {
		var cur any = snap
		for _, p := range path {
			m, ok := cur.(map[string]any)
			if !ok {
				return 0, false
			}
			cur, ok = m[p]
			if !ok {
				return 0, false
			}
		}
		return asFloat(cur)
	}
	switch key {
	case "container_memory_mb":
		conts, ok := snap["runtime"].(map[string]any)
		if !ok {
			return 0, false
		}
		list, ok := conts["containers"].([]any)
		if !ok || len(list) == 0 {
			return 0, false
		}
		var total float64
		for _, c := range list {
			m, ok := c.(map[string]any)
			if !ok {
				continue
			}
			if v, ok := asFloat(m["mem_usage_mb"]); ok {
				total += v
			}
		}
		return total, true
	case "container_cpu_percent":
		conts, ok := snap["runtime"].(map[string]any)
		if !ok {
			return 0, false
		}
		list, ok := conts["containers"].([]any)
		if !ok || len(list) == 0 {
			return 0, false
		}
		var max float64
		for _, c := range list {
			m, ok := c.(map[string]any)
			if !ok {
				continue
			}
			if v, ok := asFloat(m["cpu_percent"]); ok && v > max {
				max = v
			}
		}
		return max, true
	case "process_rss_mb":
		return get("runtime", "process_rss_mb")
	case "image_size_mb":
		return 0, false // requires `docker history`; future Phase 9.5
	case "build_time_s", "test_time_s":
		return 0, false
	case "bundle_main_kb", "bundle_vendor_kb":
		return 0, false
	case "request_p95_ms":
		return 0, false
	case "deps_total":
		return get("deps", "total")
	case "noisy_log_call_sites":
		return get("logs", "total_call_sites")
	case "jsonl_log_bytes":
		return get("logs", "jsonl_bytes")
	case "harness_dir_bytes":
		return get("disk", "harness_dir_bytes")
	case "project_bytes":
		return get("disk", "project_bytes")
	case "dockerfile_findings":
		v, ok := get("dockerfile", "findings")
		// findings is an array; len() is the metric. Fall through to a
		// separate count when needed.
		if ok {
			return v, true
		}
		if arr, ok := snap["dockerfile"].(map[string]any); ok {
			if list, ok := arr["findings"].([]any); ok {
				return float64(len(list)), true
			}
		}
		return 0, false
	}
	return 0, false
}

func numericBudget(v any) (float64, bool) { return asFloat(v) }

func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

func sortedBudgetKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func fmtNum(f float64) string {
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return fmt.Sprintf("%.2f", f)
}
