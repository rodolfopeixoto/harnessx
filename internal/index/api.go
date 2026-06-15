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

func BuildAPIMap(root string, stacks []Stack) APIMap {
	a := APIMap{Confidence: ConfidenceLow}
	if hasStack(stacks, "rails") {
		a.Routes = append(a.Routes, collectRailsRoutes(root)...)
	}
	if hasStack(stacks, "nextjs") {
		a.Routes = append(a.Routes, collectNextRoutes(root)...)
	}
	if hasStack(stacks, "go") {
		a.Routes = append(a.Routes, collectGoRoutes(root)...)
	}
	sortRoutes(a.Routes)
	return a
}

func hasStack(stacks []Stack, name string) bool {
	for _, s := range stacks {
		if s.Name == name {
			return true
		}
	}
	return false
}

func collectRailsRoutes(root string) []RouteEntry {
	b, err := os.ReadFile(filepath.Join(root, "config", "routes.rb"))
	if err != nil {
		return nil
	}
	matches := railsRouteRe.FindAllStringSubmatch(string(b), -1)
	out := make([]RouteEntry, 0, len(matches))
	for _, m := range matches {
		out = append(out, RouteEntry{
			Method: normalizeRailsMethod(m[1]),
			Path:   "/" + strings.TrimPrefix(m[2], "/"),
			Source: "config/routes.rb",
			Stack:  "rails",
		})
	}
	return out
}

func normalizeRailsMethod(method string) string {
	upper := strings.ToUpper(method)
	if upper == "RESOURCE" || upper == "RESOURCES" {
		return "RESOURCES"
	}
	return upper
}

func collectNextRoutes(root string) []RouteEntry {
	var out []RouteEntry
	for _, baseDir := range []string{"pages", "app"} {
		abs := filepath.Join(root, baseDir)
		if _, err := os.Stat(abs); err != nil {
			continue
		}
		out = append(out, walkNextRoutes(abs, baseDir)...)
	}
	return out
}

func walkNextRoutes(abs, baseDir string) []RouteEntry {
	var out []RouteEntry
	_ = filepath.WalkDir(abs, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !isNextRouteFile(d.Name()) {
			return nil
		}
		rel, _ := filepath.Rel(abs, p)
		out = append(out, RouteEntry{
			Method: "GET",
			Path:   nextRoutePath(rel),
			Source: filepath.Join(baseDir, rel),
			Stack:  "nextjs",
		})
		return nil
	})
	return out
}

func isNextRouteFile(name string) bool {
	switch name {
	case "page.tsx", "page.ts", "page.jsx", "page.js":
		return true
	}
	return strings.HasSuffix(name, ".tsx") || strings.HasSuffix(name, ".jsx")
}

func nextRoutePath(rel string) string {
	noExt := strings.TrimSuffix(rel, filepath.Ext(rel))
	noPage := strings.TrimSuffix(noExt, "/page")
	routePath := "/" + noPage
	if routePath == "/index" {
		return "/"
	}
	return routePath
}

func collectGoRoutes(root string) []RouteEntry {
	var out []RouteEntry
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
			out = append(out, RouteEntry{Path: m[1], Source: rel, Stack: "go"})
		}
		return nil
	})
	return out
}

func sortRoutes(routes []RouteEntry) {
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Stack != routes[j].Stack {
			return routes[i].Stack < routes[j].Stack
		}
		return routes[i].Path < routes[j].Path
	})
}
