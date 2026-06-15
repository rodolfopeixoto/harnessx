// SPDX-License-Identifier: MIT

// Package mcppkg bundles the canonical MCP server templates harness
// can install with one command. Templates carry the exact command +
// args + transport for popular servers so the operator does not have
// to look them up.
package mcppkg

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed templates/*.yaml
var bundled embed.FS

type Template struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Transport   string            `yaml:"transport"`
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args"`
	URL         string            `yaml:"url"`
	Env         map[string]string `yaml:"env"`
	Docs        string            `yaml:"docs"`
}

func Load(name string) (Template, error) {
	data, err := bundled.ReadFile("templates/" + name + ".yaml")
	if err != nil {
		return Template{}, fmt.Errorf("mcp: template %q not bundled: %w", name, err)
	}
	var t Template
	if err := yaml.Unmarshal(data, &t); err != nil {
		return Template{}, fmt.Errorf("mcp: parse %s: %w", name, err)
	}
	if t.Name == "" {
		t.Name = name
	}
	return t, nil
}

func List() ([]string, error) {
	var names []string
	err := fs.WalkDir(bundled, "templates", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		names = append(names, strings.TrimSuffix(strings.TrimPrefix(path, "templates/"), ".yaml"))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}
