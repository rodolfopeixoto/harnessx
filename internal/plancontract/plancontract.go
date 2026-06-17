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
	if strings.HasPrefix(id, "PLAN-") || strings.HasSuffix(id, ".md") {
		return filepath.Join(root, ".harness", "artifacts", "plans", id), nil
	}
	return filepath.Join(root, ".harness", "artifacts", "plans", "PLAN-"+id+".md"), nil
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
		if ok, _ := filepath.Match(f, path); ok {
			return true
		}
		if f == path {
			return true
		}
	}
	return false
}
