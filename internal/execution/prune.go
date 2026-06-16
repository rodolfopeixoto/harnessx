// SPDX-License-Identifier: MIT

package execution

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// PruneCandidates returns run directories under .harness/runs that
// match the retention policy. Caller decides dry-run vs actual delete.
// Sorted oldest-first by RunID (lexicographic; RunID is a ULID).
func PruneCandidates(projectRoot string, olderThan time.Duration, keepLast int) ([]string, error) {
	dir := filepath.Join(projectRoot, ".harness", "runs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	type run struct {
		id      string
		modTime time.Time
		path    string
	}
	var runs []run
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		runs = append(runs, run{id: e.Name(), modTime: info.ModTime(), path: filepath.Join(dir, e.Name())})
	}
	sort.SliceStable(runs, func(i, j int) bool { return runs[i].id < runs[j].id })

	var candidates []string
	cutoff := time.Now().Add(-olderThan)
	for i, r := range runs {
		olderOK := olderThan > 0 && r.modTime.Before(cutoff)
		keepOK := keepLast > 0 && i < len(runs)-keepLast
		if olderOK || keepOK {
			candidates = append(candidates, r.path)
		}
	}
	return candidates, nil
}

// DeletePaths removes each path (RemoveAll). Returns total bytes freed
// best-effort.
func DeletePaths(paths []string) (int64, error) {
	var freed int64
	for _, p := range paths {
		if size, err := dirSize(p); err == nil {
			freed += size
		}
		if err := os.RemoveAll(p); err != nil {
			return freed, fmt.Errorf("remove %s: %w", p, err)
		}
	}
	return freed, nil
}

func dirSize(root string) (int64, error) {
	var total int64
	err := filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}
