// SPDX-License-Identifier: MIT

package optimize

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

// WriteSnapshotReport renders a markdown summary of one snapshot under
// .harness/artifacts/reports/perf-<id>.md.
func WriteSnapshotReport(root string, s Snapshot) (string, error) {
	dir := filepath.Join(paths.HarnessDir(root), "artifacts", "reports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "perf-"+ids.New()+".md")
	var b strings.Builder
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }
	w("# Executive Summary")
	w(fmt.Sprintf("Snapshot `%s` captured at %s.", s.ID, s.CapturedAt.Format(time.RFC3339)))
	w("")
	w("# What Changed")
	w("_no comparison snapshot provided; this is a baseline._")
	w("")
	w("# What Did Not Change and Why")
	if len(s.Deps.KeepReasons) == 0 {
		w("_no kept-for-operational-safety entries recorded._")
	} else {
		w("| Dependency | Reason |")
		w("|---|---|")
		for _, k := range s.Deps.KeepReasons {
			w(fmt.Sprintf("| `%s` | %s |", k.Name, k.Reason))
		}
	}
	w("")
	w("# Metrics")
	w("| Metric | Value |")
	w("|---|---|")
	w(fmt.Sprintf("| deps_total | %d |", s.Deps.Total))
	w(fmt.Sprintf("| removal_candidates | %d |", len(s.Deps.Candidates)))
	w(fmt.Sprintf("| noisy_log_call_sites | %d |", s.Logs.TotalCallSites))
	w(fmt.Sprintf("| jsonl_log_bytes | %d |", s.Logs.JSONLBytes))
	w(fmt.Sprintf("| harness_dir_bytes | %d |", s.Disk.HarnessBytes))
	w(fmt.Sprintf("| project_bytes | %d |", s.Disk.ProjectBytes))
	if s.Dockerfile != nil {
		w(fmt.Sprintf("| dockerfile_run_steps | %d |", s.Dockerfile.RunSteps))
		w(fmt.Sprintf("| dockerfile_findings | %d |", len(s.Dockerfile.Findings)))
	}
	w("")
	w("# Files / Dependencies")
	if len(s.Deps.Candidates) == 0 {
		w("_no removal candidates._")
	} else {
		w("| Dependency | Ecosystem | Reason | Risk | Rollback |")
		w("|---|---|---|---|---|")
		for _, c := range s.Deps.Candidates {
			w(fmt.Sprintf("| `%s` | %s | %s | medium | restore from lockfile |", c.Name, c.Ecosystem, c.Reason))
		}
	}
	w("")
	w("# Dockerfile Findings")
	switch {
	case s.Dockerfile == nil:
		w("_no Dockerfile present._")
	case len(s.Dockerfile.Findings) == 0:
		w("_clean._")
	default:
		w("| ID | Severity | Message |")
		w("|---|---|---|")
		for _, f := range s.Dockerfile.Findings {
			w(fmt.Sprintf("| `%s` | %s | %s |", f.ID, f.Severity, f.Message))
		}
	}
	w("")
	w("# Risks")
	w("_changes flagged here are recommendations only; nothing was removed automatically._")
	w("")
	w("# Rollback Plan")
	w("- Revert any dependency removal via the project's package manager lockfile.")
	w("- Revert log/runtime changes by restoring the previous commit.")
	w("")
	w("# Next Steps With Real ROI")
	w("- Apply the highest-severity Dockerfile findings first (pin `:latest`, add USER).")
	w("- Capture another snapshot after each safe change and run `harness perf-compare`.")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// WriteCompareReport renders a markdown diff under
// .harness/artifacts/reports/perf-compare-<id>.md.
func WriteCompareReport(root string, d Diff) (string, error) {
	dir := filepath.Join(paths.HarnessDir(root), "artifacts", "reports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "perf-compare-"+ids.New()+".md")
	var b strings.Builder
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }
	w("# Executive Summary")
	w(fmt.Sprintf("Comparing snapshot `%s` (%s) → `%s` (%s).",
		d.From.ID, d.From.CapturedAt.Format(time.RFC3339),
		d.To.ID, d.To.CapturedAt.Format(time.RFC3339)))
	w("")
	w("# Metrics")
	w("| Metric | Before | After | Delta | Status |")
	w("|---|---|---|---|---|")
	for _, r := range d.Rows {
		w(fmt.Sprintf("| %s | %v | %v | %s | %s |", r.Metric, r.Before, r.After, r.Delta, r.Status))
	}
	w("")
	w("# Risks")
	w("_review every metric flagged `regressed`; numeric improvements should be paired with sensor pass evidence before celebrating._")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
