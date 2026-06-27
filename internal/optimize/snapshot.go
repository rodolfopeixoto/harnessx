// SPDX-License-Identifier: MIT

package optimize

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type SnapshotOptions struct {
	Root  string
	Label string
}

// Capture takes a fresh snapshot. Side-effect: writes JSON to
// .harness/artifacts/perf/<timestamp>-<id>.json.
func Capture(opts SnapshotOptions) (Snapshot, string, error) {
	if opts.Root == "" {
		return Snapshot{}, "", fmt.Errorf("optimize: empty root")
	}
	now := time.Now().UTC()
	s := Snapshot{
		ID: ids.New(), Label: opts.Label,
		CapturedAt: now, Root: opts.Root,
		Project:    detectProject(opts.Root),
		Dockerfile: scanDockerfile(opts.Root),
		Deps:       computeDeps(opts.Root),
		Logs:       computeLogs(opts.Root),
		Disk:       computeDisk(opts.Root),
		Runtime:    captureRuntime(),
		System:     SystemInfo{OS: runtime.GOOS, Arch: runtime.GOARCH},
	}
	path, err := writeSnapshot(opts.Root, s)
	if err != nil {
		return s, "", err
	}
	return s, path, nil
}

func detectProject(root string) Project {
	p := Project{Name: filepath.Base(root)}
	var prof index.Profile
	if err := index.ReadMap(root, index.MapProfile, &prof); err == nil {
		for _, st := range prof.Stacks {
			p.Stacks = append(p.Stacks, st.Name)
		}
	}
	return p
}

func computeDeps(root string) DepsMetrics {
	d := DepsMetrics{ByEcosystem: map[string]int{}}
	var deps index.Dependencies
	if err := index.ReadMap(root, index.MapDependencies, &deps); err != nil || len(deps.Ecosystems) == 0 {
		deps = index.BuildDependencies(root, nil)
	}
	for eco, e := range deps.Ecosystems {
		d.ByEcosystem[eco] = e.Count
		d.Total += e.Count
		all := append([]index.DependencyEntry{}, e.Runtime...)
		all = append(all, e.Dev...)
		for _, dep := range all {
			if cand, reason := removalCandidate(eco, dep.Name); cand {
				d.Candidates = append(d.Candidates, Candidate{
					Ecosystem: eco, Name: dep.Name, Reason: reason,
				})
				continue
			}
			if reason, ok := keepReason(eco, dep.Name); ok {
				d.KeepReasons = append(d.KeepReasons, KeepReason{Name: dep.Name, Reason: reason})
			}
		}
	}
	sort.Slice(d.Candidates, func(i, j int) bool { return d.Candidates[i].Name < d.Candidates[j].Name })
	sort.Slice(d.KeepReasons, func(i, j int) bool { return d.KeepReasons[i].Name < d.KeepReasons[j].Name })
	return d
}

// removalCandidate flags conservatively. Returns true only when we have
// high confidence a dependency is unused in production runtime.
func removalCandidate(eco, name string) (bool, string) {
	// Conservative list: only flag classic dev-tool duplicates landing in
	// the runtime ecosystem section. The user inspects every flag.
	dupDevTools := map[string]bool{
		"eslint": true, "prettier": true, "typescript": true,
		"webpack": true, "rollup": true, "babel": true,
	}
	if eco == "node" && dupDevTools[name] {
		return true, "dev tool — confirm not used at runtime before removal"
	}
	return false, ""
}

func keepReason(eco, name string) (string, bool) {
	observability := map[string]bool{
		"sentry-go": true, "@sentry/node": true, "@sentry/react": true,
		"opentelemetry": true, "prometheus": true, "datadog": true,
		"newrelic": true, "rollbar": true,
	}
	security := map[string]bool{
		"helmet": true, "argon2": true, "bcrypt": true, "rack-protection": true, "brakeman": true,
	}
	debugging := map[string]bool{
		"pry": true, "byebug": true, "delve": true, "debugpy": true,
	}
	switch {
	case observability[name]:
		return "observability — keep for operational visibility", true
	case security[name]:
		return "security — keep for safe defaults", true
	case debugging[name]:
		return "debugger — keep for incident response", true
	}
	_ = eco
	return "", false
}

func computeLogs(root string) LogsMetrics {
	m := LogsMetrics{}
	jsonlPath := filepath.Join(root, ".harness", "logs", "events.jsonl")
	if info, err := os.Stat(jsonlPath); err == nil {
		m.JSONLBytes = info.Size()
	}
	hits := scanLogCallSites(root)
	m.NoisyCallSites = hits
	m.TotalCallSites = len(hits)
	return m
}

func computeDisk(root string) DiskMetrics {
	d := DiskMetrics{}
	d.HarnessBytes = dirSize(filepath.Join(root, ".harness"))
	d.ProjectBytes = dirSize(root)
	return d
}

func dirSize(p string) int64 {
	var total int64
	const cap = 2 * 1024 * 1024 * 1024 // 2 GiB cap for sanity
	_ = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "target" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		if total > cap {
			return filepath.SkipAll
		}
		return nil
	})
	return total
}

func writeSnapshot(root string, s Snapshot) (string, error) {
	dir := filepath.Join(paths.HarnessDir(root), "artifacts", "perf")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := s.CapturedAt.Format("20060102T150405Z") + "-" + s.ID + ".json"
	path := filepath.Join(dir, name)
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// LatestTwo loads the two most-recent snapshots from disk. Useful for the
// default `harness perf-compare` invocation with no args.
func LatestTwo(root string) (Snapshot, Snapshot, string, string, error) {
	dir := filepath.Join(paths.HarnessDir(root), "artifacts", "perf")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Snapshot{}, Snapshot{}, "", "", err
	}
	type pick struct {
		name string
		mod  time.Time
	}
	var picks []pick
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		picks = append(picks, pick{name: e.Name(), mod: info.ModTime()})
	}
	if len(picks) < 2 {
		return Snapshot{}, Snapshot{}, "", "", fmt.Errorf("optimize: need at least 2 snapshots (have %d) — run `harness perf-snapshot` twice", len(picks))
	}
	sort.Slice(picks, func(i, j int) bool { return picks[i].mod.Before(picks[j].mod) })
	a := picks[len(picks)-2]
	b := picks[len(picks)-1]
	aSnap, err := LoadSnapshot(filepath.Join(dir, a.name))
	if err != nil {
		return Snapshot{}, Snapshot{}, "", "", err
	}
	bSnap, err := LoadSnapshot(filepath.Join(dir, b.name))
	if err != nil {
		return Snapshot{}, Snapshot{}, "", "", err
	}
	return aSnap, bSnap, a.name, b.name, nil
}

func LoadSnapshot(path string) (Snapshot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, err
	}
	var s Snapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return Snapshot{}, err
	}
	return s, nil
}
