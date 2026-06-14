// SPDX-License-Identifier: MIT

package index

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// frameworkOf maps a path to the test framework it belongs to, or "".
func frameworkOf(rel string) string {
	base := filepath.Base(rel)
	switch {
	case strings.HasSuffix(base, "_test.go"):
		return "go-test"
	case strings.HasSuffix(base, "_spec.rb"):
		return "rspec"
	case strings.HasSuffix(base, ".test.ts"), strings.HasSuffix(base, ".test.tsx"),
		strings.HasSuffix(base, ".test.js"), strings.HasSuffix(base, ".test.jsx"),
		strings.HasSuffix(base, ".spec.ts"), strings.HasSuffix(base, ".spec.tsx"),
		strings.HasSuffix(base, ".spec.js"), strings.HasSuffix(base, ".spec.jsx"):
		return "vitest"
	case strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py"),
		strings.HasSuffix(base, "_test.py"):
		return "pytest"
	}
	return ""
}

// excludedTestDirs are skipped wholesale during test discovery.
var excludedTestDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	".git":         true,
	".harness":     true,
	"target":       true,
	"dist":         true,
	"build":        true,
	"bin":          true,
}

func BuildTestMap(root string) TestMap {
	buckets := map[string][]string{}
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if excludedTestDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		if fw := frameworkOf(rel); fw != "" {
			buckets[fw] = append(buckets[fw], rel)
		}
		return nil
	})

	tm := TestMap{}
	frameworks := make([]string, 0, len(buckets))
	for k := range buckets {
		frameworks = append(frameworks, k)
	}
	sort.Strings(frameworks)
	for _, fw := range frameworks {
		files := buckets[fw]
		sort.Strings(files)
		tm.Suites = append(tm.Suites, TestSuite{Framework: fw, Files: files})
		tm.TotalFiles += len(files)
	}
	return tm
}
