// SPDX-License-Identifier: MIT

package optimize

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// callSitePatterns are the noisy log invocations we look for. Each entry
// is matched anywhere on the line and tagged with the language.
var callSitePatterns = []struct{ needle, kind string }{
	{"console.log(", "console.log"},
	{"console.debug(", "console.debug"},
	{"console.info(", "console.info"},
	{"puts ", "puts"},
	{"println!(", "println!"},
	{"print!(", "print!"},
	{"fmt.Println(", "fmt.Println"},
	{"fmt.Printf(", "fmt.Printf"},
	{"print(", "print"},
}

var scanLogExts = map[string]bool{
	".js": true, ".jsx": true, ".ts": true, ".tsx": true,
	".rb": true, ".go": true, ".rs": true, ".py": true,
}

var scanLogIgnoredDirs = map[string]bool{
	".git": true, ".harness": true, "node_modules": true, "vendor": true,
	"target": true, "dist": true, "build": true, "bin": true,
	".venv": true, "venv": true, ".tox": true,
	"__pycache__": true, ".pytest_cache": true, ".mypy_cache": true, ".ruff_cache": true,
	"site-packages": true,
}

func scanLogCallSites(root string) []LogCallSite {
	const maxHits = 200
	var hits []LogCallSite
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || len(hits) >= maxHits {
			return nil
		}
		if d.IsDir() {
			if scanLogIgnoredDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, "_test.go") ||
			strings.HasSuffix(name, ".test.ts") || strings.HasSuffix(name, ".test.tsx") ||
			strings.HasSuffix(name, ".test.js") || strings.HasSuffix(name, ".test.jsx") ||
			strings.HasSuffix(name, "_spec.rb") || strings.HasSuffix(name, ".spec.ts") ||
			strings.HasPrefix(name, "test_") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if !scanLogExts[ext] {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()
		rel, _ := filepath.Rel(root, p)
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1024*1024), 4*1024*1024)
		for lineNo := 1; sc.Scan() && len(hits) < maxHits; lineNo++ {
			line := sc.Text()
			for _, pat := range callSitePatterns {
				if strings.Contains(line, pat.needle) {
					hits = append(hits, LogCallSite{
						Path: rel, Line: lineNo,
						Snippet: trimSnippet(line, 120), Kind: pat.kind,
					})
					break
				}
			}
		}
		return nil
	})
	return hits
}

func trimSnippet(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
