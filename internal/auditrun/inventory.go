// SPDX-License-Identifier: MIT

package auditrun

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileStat struct {
	Path string `json:"path"`
	Loc  int    `json:"loc"`
}

type Inventory struct {
	GeneratedAt  time.Time  `json:"generated_at"`
	GoFiles      int        `json:"go_files"`
	GoTestFiles  int        `json:"go_test_files"`
	GoLoc        int        `json:"go_loc"`
	TSXFiles     int        `json:"tsx_files"`
	TSXTestFiles int        `json:"tsx_test_files"`
	TSXLoc       int        `json:"tsx_loc"`
	ShellScripts int        `json:"shell_scripts"`
	ShellTests   int        `json:"shell_tests"`
	SpecFiles    int        `json:"spec_files"`
	E2EScripts   int        `json:"e2e_scripts"`
	LargestFiles []FileStat `json:"largest_files"`
}

func BuildInventory(root string) Inventory {
	inv := Inventory{GeneratedAt: time.Now().UTC()}
	var stats []FileStat
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if skipDir(filepath.Base(path)) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		loc := countLines(path)
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go":
			inv.GoFiles++
			inv.GoLoc += loc
			if strings.HasSuffix(path, "_test.go") {
				inv.GoTestFiles++
			}
		case ".tsx", ".ts":
			inv.TSXFiles++
			inv.TSXLoc += loc
			if strings.Contains(filepath.Base(path), ".test.") {
				inv.TSXTestFiles++
			}
		case ".sh":
			inv.ShellScripts++
			if strings.HasPrefix(filepath.Base(path), "test-") {
				inv.ShellTests++
			}
			if strings.HasPrefix(filepath.Base(path), "e2e-") {
				inv.E2EScripts++
			}
		case ".md":
			if strings.Contains(rel, ".harness/artifacts/specs/") {
				inv.SpecFiles++
			}
		}
		stats = append(stats, FileStat{Path: rel, Loc: loc})
		return nil
	})
	inv.LargestFiles = topByLoc(stats, 20)
	return inv
}

func skipDir(name string) bool {
	switch name {
	case "node_modules", ".git", "dist", "coverage", "bin", "tmp", ".harness", "docs":
		return true
	}
	return false
}

func countLines(path string) int {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	return strings.Count(string(b), "\n") + 1
}

func topByLoc(stats []FileStat, n int) []FileStat {
	for i := 1; i < len(stats); i++ {
		j := i
		for j > 0 && stats[j-1].Loc < stats[j].Loc {
			stats[j-1], stats[j] = stats[j], stats[j-1]
			j--
		}
	}
	if len(stats) > n {
		return stats[:n]
	}
	return stats
}
