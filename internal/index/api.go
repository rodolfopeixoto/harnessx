// SPDX-License-Identifier: MIT

package index

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	railsRouteRe = regexp.MustCompile(`(?m)^\s*(get|post|put|patch|delete|resources?|root)\s+['":/]([^\s,'"]+)['"]?`)
	goHTTPRe     = regexp.MustCompile(`(?:HandleFunc|Handle|GET|POST|PUT|DELETE|PATCH)\s*\(\s*"([^"]+)"`)
)

// BuildAPIMap extracts best-effort route hints. Always returns
// Confidence=low for inferred routes; callers should treat the data as a
// hint rather than a contract.
func BuildAPIMap(root string, stacks []Stack) APIMap {
	a := APIMap{Confidence: ConfidenceLow}
	hasStack := func(name string) bool {
		for _, s := range stacks {
			if s.Name == name {
				return true
			}
		}
		return false
	}

	if hasStack("rails") {
		if b, err := os.ReadFile(filepath.Join(root, "config", "routes.rb")); err == nil {
			for _, m := range railsRouteRe.FindAllStringSubmatch(string(b), -1) {
				method := strings.ToUpper(m[1])
				if method == "RESOURCE" || method == "RESOURCES" {
					method = "RESOURCES"
				}
				a.Routes = append(a.Routes, RouteEntry{
					Method: method, Path: "/" + strings.TrimPrefix(m[2], "/"),
					Source: "config/routes.rb", Stack: "rails",
				})
			}
		}
	}

	if hasStack("nextjs") {
		// Pages router + App router conventions.
		for _, baseDir := range []string{"pages", "app"} {
			abs := filepath.Join(root, baseDir)
			if _, err := os.Stat(abs); err != nil {
				continue
			}
			_ = filepath.WalkDir(abs, func(p string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				name := d.Name()
				if name == "page.tsx" || name == "page.ts" || name == "page.jsx" || name == "page.js" ||
					strings.HasSuffix(name, ".tsx") || strings.HasSuffix(name, ".jsx") {
					rel, _ := filepath.Rel(abs, p)
					routePath := "/" + strings.TrimSuffix(strings.TrimSuffix(rel, filepath.Ext(rel)), "/page")
					if routePath == "/index" {
						routePath = "/"
					}
					a.Routes = append(a.Routes, RouteEntry{
						Method: "GET", Path: routePath, Source: filepath.Join(baseDir, rel), Stack: "nextjs",
					})
				}
				return nil
			})
		}
	}

	if hasStack("go") {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() && excludedTestDirs[d.Name()] {
				return filepath.SkipDir
			}
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(root, p)
			for _, m := range goHTTPRe.FindAllStringSubmatch(string(b), -1) {
				a.Routes = append(a.Routes, RouteEntry{
					Path: m[1], Source: rel, Stack: "go",
				})
			}
			return nil
		})
	}

	sort.Slice(a.Routes, func(i, j int) bool {
		if a.Routes[i].Stack != a.Routes[j].Stack {
			return a.Routes[i].Stack < a.Routes[j].Stack
		}
		return a.Routes[i].Path < a.Routes[j].Path
	})
	return a
}
