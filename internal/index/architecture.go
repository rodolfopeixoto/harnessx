// SPDX-License-Identifier: MIT

package index

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// purposeByDir maps common top-level dir names to a coarse purpose tag.
// "unknown" is honest — better than guessing.
var purposeByDir = map[string]string{
	"src":          "src",
	"lib":          "src",
	"app":          "src",
	"pkg":          "src",
	"internal":     "src",
	"cmd":          "src",
	"apps":         "src",
	"packages":     "src",
	"server":       "src",
	"client":       "src",
	"web":          "src",
	"frontend":     "src",
	"backend":      "src",
	"api":          "src",
	"test":         "tests",
	"tests":        "tests",
	"spec":         "tests",
	"e2e":          "tests",
	"playwright":   "tests",
	"cypress":      "tests",
	"docs":         "docs",
	"doc":          "docs",
	"examples":     "docs",
	"infra":        "infra",
	"deploy":       "infra",
	"deployment":   "infra",
	"terraform":    "infra",
	"k8s":          "infra",
	"helm":         "infra",
	"ops":          "infra",
	"scripts":      "infra",
	".github":      "infra",
	"assets":       "assets",
	"static":       "assets",
	"public":       "assets",
	"images":       "assets",
	"config":       "config",
	"configs":      "config",
	"settings":     "config",
	"node_modules": "deps",
	"vendor":       "deps",
	"target":       "build",
	"dist":         "build",
	"build":        "build",
	"bin":          "build",
	".harness":     "harnessx",
}

func BuildArchitecture(root string) Architecture {
	entries, err := os.ReadDir(root)
	if err != nil {
		return Architecture{}
	}
	var out []ArchDirEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") && name != ".github" && name != ".harness" {
			continue
		}
		purpose := purposeByDir[name]
		if purpose == "" {
			purpose = "unknown"
		}
		files := countFiles(filepath.Join(root, name), 2)
		out = append(out, ArchDirEntry{Path: name, Purpose: purpose, Files: files})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return Architecture{TopLevel: out}
}

// countFiles counts regular files under dir up to maxDepth (1 = direct
// children only). Bounded so huge node_modules trees don't stall.
func countFiles(dir string, maxDepth int) int {
	const cap = 5000
	var n int
	var walk func(p string, depth int)
	walk = func(p string, depth int) {
		if n >= cap {
			return
		}
		entries, err := os.ReadDir(p)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				if depth < maxDepth {
					walk(filepath.Join(p, e.Name()), depth+1)
				}
				continue
			}
			n++
			if n >= cap {
				return
			}
		}
	}
	walk(dir, 1)
	return n
}
