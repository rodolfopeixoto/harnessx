// SPDX-License-Identifier: MIT

package flowpkg

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"

	"gopkg.in/yaml.v3"
)

//go:embed all:templates
var bundled embed.FS

type PhaseKind string

const (
	PhaseDeterministic PhaseKind = "deterministic"
	PhaseLLM           PhaseKind = "llm"
	PhaseSensor        PhaseKind = "sensor"
)

type Phase struct {
	Name           string    `yaml:"name"`
	Kind           PhaseKind `yaml:"kind"`
	Cmd            []string  `yaml:"cmd,omitempty"`
	Prompt         string    `yaml:"prompt,omitempty"`
	SensorID       string    `yaml:"sensor_id,omitempty"`
	Gates          []string  `yaml:"gates,omitempty"`
	TimeoutSeconds int       `yaml:"timeout_seconds,omitempty"`
}

type Flow struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Domain      string  `yaml:"domain"`
	Phases      []Phase `yaml:"phases"`
}

func List() ([]string, error) {
	entries, err := fs.ReadDir(bundled, "templates")
	if err != nil {
		return nil, nil
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

func Load(name string) (Flow, error) {
	raw, err := bundled.ReadFile(path.Join("templates", name, "flow.yaml"))
	if err != nil {
		return Flow{}, fmt.Errorf("flowpkg: unknown flow %q", name)
	}
	var f Flow
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return Flow{}, fmt.Errorf("flowpkg: parse %s: %w", name, err)
	}
	if err := f.Validate(); err != nil {
		return Flow{}, err
	}
	return f, nil
}

func (f Flow) Validate() error {
	if f.Name == "" {
		return errors.New("flowpkg: missing name")
	}
	if len(f.Phases) == 0 {
		return errors.New("flowpkg: at least one phase required")
	}
	for i, p := range f.Phases {
		if p.Name == "" {
			return fmt.Errorf("flowpkg: phase %d missing name", i)
		}
		switch p.Kind {
		case PhaseDeterministic, PhaseLLM, PhaseSensor:
		default:
			return fmt.Errorf("flowpkg: phase %q unknown kind %q", p.Name, p.Kind)
		}
	}
	return nil
}
