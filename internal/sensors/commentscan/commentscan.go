// SPDX-License-Identifier: MIT

package commentscan

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

type Finding struct {
	File    string
	Line    int
	Snippet string
	Reason  string
}

type Options struct {
	Roots []string
	Skip  func(path string) bool
}

func Scan(files []string, allow Allowlist) []Finding {
	var out []Finding
	fset := token.NewFileSet()
	for _, f := range files {
		findings := scanFile(fset, f, allow)
		out = append(out, findings...)
	}
	return out
}

type Allowlist struct {
	SPDX    bool
	Package bool
	Godoc   bool
}

func DefaultAllowlist() Allowlist {
	return Allowlist{SPDX: true, Package: true, Godoc: true}
}

func scanFile(fset *token.FileSet, path string, allow Allowlist) []Finding {
	if filepath.Ext(path) != ".go" {
		return nil
	}
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil
	}
	exported := exportedNamesInFile(file)
	var findings []Finding
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			pos := fset.Position(c.Slash)
			text := normalize(c.Text)
			if allow.SPDX && strings.HasPrefix(text, "SPDX-License-Identifier") {
				continue
			}
			if allow.Package && strings.HasPrefix(text, "Package ") {
				continue
			}
			if allow.Godoc && startsWithExportedName(text, exported) {
				continue
			}
			findings = append(findings, Finding{
				File: path, Line: pos.Line,
				Snippet: truncate(text, 80),
				Reason:  "comment is narrative; remove unless documenting an exported symbol",
			})
		}
	}
	return findings
}

func normalize(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "//") {
		s = strings.TrimPrefix(s, "//")
	} else if strings.HasPrefix(s, "/*") && strings.HasSuffix(s, "*/") {
		s = s[2 : len(s)-2]
	}
	return strings.TrimSpace(s)
}

func exportedNamesInFile(f *ast.File) map[string]bool {
	out := map[string]bool{}
	for _, d := range f.Decls {
		switch v := d.(type) {
		case *ast.FuncDecl:
			if v.Name.IsExported() {
				out[v.Name.Name] = true
			}
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						out[s.Name.Name] = true
					}
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n.IsExported() {
							out[n.Name] = true
						}
					}
				}
			}
		}
	}
	return out
}

func startsWithExportedName(text string, exported map[string]bool) bool {
	for name := range exported {
		if strings.HasPrefix(text, name+" ") || text == name {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
