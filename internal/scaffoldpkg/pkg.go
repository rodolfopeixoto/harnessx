// SPDX-License-Identifier: MIT

// Package scaffoldpkg loads and applies bundled language scaffolds.
// Templates are pure text (no LLM call) and produce byte-identical
// output for the same (lang, name) pair so the operation is
// deterministic and re-runnable.
package scaffoldpkg

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed all:templates
var bundled embed.FS

// Meta describes one bundled scaffold (one language).
type Meta struct {
	Language      string     `yaml:"language"`
	Description   string     `yaml:"description"`
	TestedAgainst string     `yaml:"tested_against"`
	RequiredTools []string   `yaml:"required_tools"`
	Files         []FileSpec `yaml:"files"`
	PostSteps     []PostStep `yaml:"post_steps"`
	LintCommand   string     `yaml:"lint_command"`
	TestCommand   string     `yaml:"test_command"`
	RunCommand    string     `yaml:"run_command"`
}

type FileSpec struct {
	Path     string `yaml:"path"`
	Template string `yaml:"template"`
	Mode     int    `yaml:"mode"`
}

type PostStep struct {
	Name string   `yaml:"name"`
	Cmd  []string `yaml:"cmd"`
}

// List returns the language ids of every bundled scaffold (sorted).
func List() ([]string, error) {
	entries, err := fs.ReadDir(bundled, "templates")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out, nil
}

// Load reads and parses scaffold.yaml for the named language.
func Load(lang string) (Meta, error) {
	raw, err := bundled.ReadFile(path.Join("templates", lang, "scaffold.yaml"))
	if err != nil {
		return Meta{}, fmt.Errorf("scaffoldpkg: unknown language %q (run 'harness scaffold list')", lang)
	}
	var m Meta
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return Meta{}, fmt.Errorf("scaffoldpkg: parse %s/scaffold.yaml: %w", lang, err)
	}
	if m.Language == "" {
		m.Language = lang
	}
	return m, nil
}

// ReadTemplate returns the body of one template file referenced by a
// FileSpec, with $NAME placeholders substituted.
func ReadTemplate(lang, template, name string) ([]byte, error) {
	raw, err := bundled.ReadFile(path.Join("templates", lang, template))
	if err != nil {
		return nil, fmt.Errorf("scaffoldpkg: missing template %s/%s", lang, template)
	}
	if name == "" {
		return raw, nil
	}
	return []byte(strings.ReplaceAll(string(raw), "$NAME", name)), nil
}

// ApplyOptions controls how Apply writes the scaffold to disk.
type ApplyOptions struct {
	Root  string
	Name  string
	Force bool
	Dry   bool
}

// ApplyResult lists what Apply did (or would do under Dry).
type ApplyResult struct {
	Created []string
	Skipped []string
}

// Apply writes every template file under Root according to the Meta.
// Returns the list of paths created and skipped (already exist without
// --force).
func Apply(m Meta, opts ApplyOptions) (ApplyResult, error) {
	if opts.Root == "" {
		return ApplyResult{}, errors.New("scaffoldpkg: missing Root")
	}
	var res ApplyResult
	for _, f := range m.Files {
		target := filepath.Join(opts.Root, f.Path)
		if _, err := os.Stat(target); err == nil && !opts.Force {
			res.Skipped = append(res.Skipped, f.Path)
			continue
		}
		body, err := ReadTemplate(m.Language, f.Template, opts.Name)
		if err != nil {
			return res, err
		}
		if opts.Dry {
			res.Created = append(res.Created, f.Path)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return res, err
		}
		mode := os.FileMode(f.Mode)
		if mode == 0 {
			mode = 0o644
		}
		if err := os.WriteFile(target, body, mode); err != nil {
			return res, err
		}
		res.Created = append(res.Created, f.Path)
	}
	return res, nil
}
