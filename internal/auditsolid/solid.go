// SPDX-License-Identifier: MIT

// Package auditsolid scans Go sources for SOLID-violation smells:
// god files (LOC > LimitLOC), high fan-in/out (> LimitFan).
package auditsolid

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Violation struct {
	Path   string
	Kind   string
	Metric int
	Limit  int
}

type Opts struct {
	LimitLOC int
	LimitFan int
}

func Default() Opts { return Opts{LimitLOC: 400, LimitFan: 15} }

func Scan(root string, opts Opts) ([]Violation, error) {
	var out []Violation
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		loc, err := countLines(path)
		if err != nil {
			return err
		}
		if loc > opts.LimitLOC {
			out = append(out, Violation{Path: path, Kind: "loc", Metric: loc, Limit: opts.LimitLOC})
		}
		fan, err := fanOut(path)
		if err != nil {
			return nil
		}
		if fan > opts.LimitFan {
			out = append(out, Violation{Path: path, Kind: "fan-out", Metric: fan, Limit: opts.LimitFan})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func skipDir(name string) bool {
	switch name {
	case "vendor", "node_modules", ".git", "dist", "bin":
		return true
	}
	return false
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 1<<20), 1<<24)
	n := 0
	for s.Scan() {
		n++
	}
	return n, s.Err()
}

func fanOut(path string) (int, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return 0, err
	}
	return len(file.Imports), nil
}

func Report(v []Violation) string {
	if len(v) == 0 {
		return "0 SOLID violations\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d SOLID violation(s):\n", len(v))
	for _, x := range v {
		fmt.Fprintf(&b, "  %s [%s] %d > %d\n", x.Path, x.Kind, x.Metric, x.Limit)
	}
	return b.String()
}
