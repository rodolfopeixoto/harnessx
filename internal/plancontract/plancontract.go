// SPDX-License-Identifier: MIT

package plancontract

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Contract struct {
	ID         string
	Path       string
	Intent     string
	Files      []string
	Invariants []string
	Validation []string
	Rollback   string
	Risk       string
}

func Resolve(root, id string) (string, error) {
	if filepath.IsAbs(id) {
		return id, nil
	}
	name := id
	if !strings.HasPrefix(name, "PLAN-") {
		name = "PLAN-" + name
	}
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}
	return filepath.Join(root, ".harness", "artifacts", "plans", name), nil
}

func Load(root, id string) (Contract, error) {
	path, err := Resolve(root, id)
	if err != nil {
		return Contract{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return Contract{}, fmt.Errorf("plancontract: open %s: %w", path, err)
	}
	defer f.Close()
	c := Contract{ID: planID(path), Path: path}
	section := ""
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "## ") {
			section = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			continue
		}
		applySection(&c, section, line)
	}
	if err := sc.Err(); err != nil {
		return Contract{}, err
	}
	if c.Intent == "" {
		return Contract{}, errors.New("plancontract: missing intent")
	}
	return c, nil
}

func applySection(c *Contract, section, line string) {
	switch strings.ToLower(section) {
	case "intent":
		appendIntent(c, line)
	case "files in scope":
		appendListItem(line, &c.Files, "_unconstrained_")
	case "invariants":
		appendListItem(line, &c.Invariants, "_none declared_")
	case "validation":
		if cmd := codeLine(line); cmd != "" {
			c.Validation = append(c.Validation, cmd)
		}
	case "rollback":
		setFirstNonEmpty(&c.Rollback, codeLine(line))
	case "risk tier":
		setFirstNonEmpty(&c.Risk, strings.TrimSpace(line))
	}
}

func appendListItem(line string, dst *[]string, sentinel string) {
	item := bullet(line)
	if item == "" || item == sentinel {
		return
	}
	*dst = append(*dst, item)
}

func setFirstNonEmpty(dst *string, v string) {
	if *dst != "" || v == "" {
		return
	}
	*dst = v
}

func appendIntent(c *Contract, line string) {
	t := strings.TrimSpace(line)
	if t == "" {
		return
	}
	if c.Intent == "" {
		c.Intent = t
		return
	}
	c.Intent += " " + t
}

func planID(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return strings.TrimPrefix(base, "PLAN-")
}

func bullet(line string) string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "- ") {
		return ""
	}
	item := strings.TrimSpace(strings.TrimPrefix(line, "- "))
	item = strings.Trim(item, "`")
	return item
}

func codeLine(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "```") {
		return ""
	}
	return line
}

func (c Contract) InScope(path string) bool {
	if len(c.Files) == 0 {
		return true
	}
	for _, f := range c.Files {
		if f == "_unconstrained_" {
			return true
		}
		if ok, _ := filepath.Match(f, path); ok {
			return true
		}
		if f == path {
			return true
		}
		// Allow `tests/` to mean `tests/**` so users do not have to write
		// the glob form by hand. Same for any path ending in a slash.
		if strings.HasSuffix(f, "/") && strings.HasPrefix(path, f) {
			return true
		}
		// Treat `app/**` and `app/...` as recursive matches. filepath.Match
		// only handles single segments, so we expand the leading dir match
		// manually.
		if rec, dir := recursiveGlob(f); rec && strings.HasPrefix(path, dir+"/") {
			return true
		}
	}
	return alwaysInScope(path)
}

// alwaysInScope lists project metadata files that PLAN scope refuses to
// fence off. Scaffold-driven flows (`harness new`, `harness scaffold
// apply`) routinely touch these even when the user's plan only names the
// feature files; failing the gate on them was creating false negatives
// during real walks of the e-commerce tutorial.
func alwaysInScope(path string) bool {
	switch filepath.Base(path) {
	case ".gitignore",
		".gitattributes",
		"README.md",
		"CHANGELOG.md",
		"Makefile",
		"pyproject.toml",
		"requirements.txt",
		"requirements-dev.txt",
		"ruff.toml",
		".ruff.toml",
		"mypy.ini",
		"pytest.ini",
		"package.json",
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
		"Cargo.toml",
		"Cargo.lock",
		"Gemfile",
		"Gemfile.lock",
		"go.mod",
		"go.sum":
		return true
	}
	// Anything under the harness state dir is always in scope — sensors,
	// plans, sessions, etc. — because the contract files themselves live
	// there and committing the plan/spec is part of the flow.
	if strings.HasPrefix(path, ".harness/") || path == ".harness" {
		return true
	}
	return false
}

func recursiveGlob(pattern string) (bool, string) {
	for _, suf := range []string{"/**", "/..."} {
		if strings.HasSuffix(pattern, suf) {
			return true, strings.TrimSuffix(pattern, suf)
		}
	}
	return false, ""
}
