// SPDX-License-Identifier: MIT

// Package install resolves tool manifests + per-platform installation
// strategies. Operators run `harness install <name>`; this package picks
// the right strategy for the host (brew on mac, apt on debian, go install
// for Go-based tools, etc.) and records the install audit event.
package install

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed manifests/*.yaml
var bundled embed.FS

type Manifest struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Category    string             `yaml:"category"`
	Probe       ProbeRef           `yaml:"probe"`
	Strategies  []StrategyManifest `yaml:"strategies"`
}

type ProbeRef struct {
	Binary       string   `yaml:"binary"`
	VersionArgs  []string `yaml:"version_args"`
	VersionRegex string   `yaml:"version_regex,omitempty"`
}

type StrategyManifest struct {
	Kind     string            `yaml:"kind"`
	Platform PlatformMatch     `yaml:"platform,omitempty"`
	Args     map[string]string `yaml:"args,omitempty"`
}

type PlatformMatch struct {
	OS   []string `yaml:"os,omitempty"`
	Arch []string `yaml:"arch,omitempty"`
}

func (p PlatformMatch) Matches(os, arch string) bool {
	if len(p.OS) > 0 && !contains(p.OS, os) {
		return false
	}
	if len(p.Arch) > 0 && !contains(p.Arch, arch) {
		return false
	}
	return true
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func LoadBundled(name string) (Manifest, error) {
	data, err := bundled.ReadFile("manifests/" + name + ".yaml")
	if err != nil {
		return Manifest{}, fmt.Errorf("install: manifest %q not bundled: %w", name, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("install: parse %s: %w", name, err)
	}
	if m.Name == "" {
		m.Name = name
	}
	return m, nil
}

func ListBundled() ([]string, error) {
	var names []string
	err := fs.WalkDir(bundled, "manifests", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		base := strings.TrimSuffix(strings.TrimPrefix(path, "manifests/"), ".yaml")
		names = append(names, base)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}
