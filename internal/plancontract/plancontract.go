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
		switch strings.ToLower(section) {
		case "intent":
			if t := strings.TrimSpace(line); t != "" {
				if c.Intent == "" {
					c.Intent = t
				} else {
					c.Intent += " " + t
				}
			}
		case "files in scope":
			if item := bullet(line); item != "" && item != "_unconstrained_" {
				c.Files = append(c.Files, item)
			}
		case "invariants":
			if item := bullet(line); item != "" && item != "_none declared_" {
				c.Invariants = append(c.Invariants, item)
			}
		case "validation":
			if cmd := codeLine(line); cmd != "" {
				c.Validation = append(c.Validation, cmd)
			}
		case "rollback":
			if cmd := codeLine(line); cmd != "" && c.Rollback == "" {
				c.Rollback = cmd
			}
		case "risk tier":
			if t := strings.TrimSpace(line); t != "" && c.Risk == "" {
				c.Risk = t
			}
		}
	}
	if err := sc.Err(); err != nil {
		return Contract{}, err
	}
	if c.Intent == "" {
		return Contract{}, errors.New("plancontract: missing intent")
	}
	return c, nil
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
