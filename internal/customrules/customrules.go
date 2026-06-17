// SPDX-License-Identifier: MIT

// Package customrules loads user-defined sensor rules from
// .harness/rules/*.yaml so projects can declare structure-grounded
// invariants (paper §3.1.2) without modifying core sensors.
package customrules

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const rulesDirRel = ".harness/rules"

type Rule struct {
	ID          string   `yaml:"id"`
	Description string   `yaml:"description"`
	Severity    string   `yaml:"severity"`
	When        When     `yaml:"when"`
	Forbid      []string `yaml:"forbid"`
	Require     []string `yaml:"require"`
}

type When struct {
	Stacks    []string `yaml:"stacks"`
	PathGlobs []string `yaml:"path_globs"`
}

func (r Rule) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return errors.New("rule: missing id")
	}
	if r.Severity == "" {
		r.Severity = "warn"
	}
	switch r.Severity {
	case "info", "warn", "error":
	default:
		return fmt.Errorf("rule %s: invalid severity %q", r.ID, r.Severity)
	}
	if len(r.Forbid) == 0 && len(r.Require) == 0 {
		return fmt.Errorf("rule %s: must declare at least one forbid or require pattern", r.ID)
	}
	return nil
}

func dir(root string) string { return filepath.Join(root, rulesDirRel) }

// Load returns every valid rule found in <root>/.harness/rules/. The
// list is sorted by rule ID. Invalid rules are surfaced via the error
// (other rules still load).
func Load(root string) ([]Rule, error) {
	d := dir(root)
	entries, err := os.ReadDir(d)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("customrules: read dir: %w", err)
	}
	var rules []Rule
	var errs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if ext := filepath.Ext(e.Name()); ext != ".yaml" && ext != ".yml" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(d, e.Name()))
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", e.Name(), err))
			continue
		}
		var r Rule
		if err := yaml.Unmarshal(raw, &r); err != nil {
			errs = append(errs, fmt.Sprintf("%s: yaml: %v", e.Name(), err))
			continue
		}
		if err := r.Validate(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", e.Name(), err))
			continue
		}
		rules = append(rules, r)
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	if len(errs) > 0 {
		return rules, fmt.Errorf("customrules: %d invalid file(s): %s", len(errs), strings.Join(errs, "; "))
	}
	return rules, nil
}

// AppliesTo reports whether the rule's When clause selects the given
// stack and file path.
func (r Rule) AppliesTo(stack, path string) bool {
	if len(r.When.Stacks) > 0 {
		match := false
		for _, s := range r.When.Stacks {
			if s == stack || s == "*" {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	if len(r.When.PathGlobs) > 0 {
		for _, g := range r.When.PathGlobs {
			if ok, _ := filepath.Match(g, path); ok {
				return true
			}
		}
		return false
	}
	return true
}
