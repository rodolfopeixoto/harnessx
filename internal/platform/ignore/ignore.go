// SPDX-License-Identifier: MIT

// Package ignore parses .harnessignore files. Pattern grammar is a thin
// subset of gitignore: glob-style filename matches, leading "/" anchors
// to the project root, trailing "/" matches directories only, "#" starts
// a comment. Patterns are evaluated in declaration order; later patterns
// override earlier ones (no "!" inversion yet — add when needed).
package ignore

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Matcher struct {
	root     string
	patterns []pattern
}

type pattern struct {
	raw      string
	anchored bool
	dirOnly  bool
}

// Load reads <root>/.harnessignore. Returns an empty Matcher when the
// file is missing (matches nothing).
func Load(root string) (*Matcher, error) {
	m := &Matcher{root: root}
	f, err := os.Open(filepath.Join(root, ".harnessignore"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return m, nil
		}
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := pattern{raw: line}
		if strings.HasPrefix(line, "/") {
			p.anchored = true
			line = line[1:]
		}
		if strings.HasSuffix(line, "/") {
			p.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}
		p.raw = line
		m.patterns = append(m.patterns, p)
	}
	return m, sc.Err()
}

// Match reports whether the given project-relative path is ignored.
// dir indicates the candidate is a directory entry (so trailing-slash
// rules can apply).
func (m *Matcher) Match(rel string, dir bool) bool {
	if m == nil || len(m.patterns) == 0 {
		return false
	}
	rel = filepath.ToSlash(rel)
	base := filepath.Base(rel)
	for _, p := range m.patterns {
		if p.dirOnly && !dir {
			continue
		}
		target := base
		if p.anchored || strings.Contains(p.raw, "/") {
			target = rel
		}
		ok, err := filepath.Match(p.raw, target)
		if err == nil && ok {
			return true
		}
	}
	return false
}

// Root returns the matcher's project root.
func (m *Matcher) Root() string {
	if m == nil {
		return ""
	}
	return m.root
}
