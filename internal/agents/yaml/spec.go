// SPDX-License-Identifier: MIT

// Package yaml loads YAML-defined agent adapters that wrap a CLI binary.
package yaml

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/ropeixoto/harnessx/internal/agents"
)

// Spec mirrors the YAML schema documented in docs/agents.md (Phase 3).
type Spec struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Enabled bool   `yaml:"enabled"`
	Type    string `yaml:"type"` // "cli"

	Command struct {
		Binary string `yaml:"binary"`
		Check  string `yaml:"check"`
	} `yaml:"command"`

	Capabilities agents.Capabilities `yaml:"capabilities"`
	Strengths    []string            `yaml:"strengths"`
	Models       map[string]string   `yaml:"models"`

	Execution struct {
		PromptMode       string `yaml:"prompt_mode"`       // stdin | arg
		WorkingDirectory string `yaml:"working_directory"` // project
		TimeoutSeconds   int    `yaml:"timeout_seconds"`
		MaxRetries       int    `yaml:"max_retries"`
	} `yaml:"execution"`

	Run struct {
		Args []string `yaml:"args"`
	} `yaml:"run"`

	Output struct {
		Format               string `yaml:"format"` // jsonl | text
		FinalMessageJSONPath string `yaml:"final_message_json_path"`
		ErrorJSONPath        string `yaml:"error_json_path"`
		UsageJSONPath        string `yaml:"usage_json_path"`
	} `yaml:"output"`

	FailureDetection map[string][]string `yaml:"failure_detection"`

	Cost struct {
		Mode                  string  `yaml:"mode"`
		InputTokenPricePer1M  float64 `yaml:"input_token_price_per_1m"`
		OutputTokenPricePer1M float64 `yaml:"output_token_price_per_1m"`
	} `yaml:"cost"`
}

// Load reads a YAML file and validates it well enough to fail fast on
// missing required fields. It does not check whether the binary exists —
// that's healthcheck/certification territory.
func Load(path string) (Spec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, fmt.Errorf("yaml: read %s: %w", path, err)
	}
	var s Spec
	if err := yaml.Unmarshal(b, &s); err != nil {
		return Spec{}, fmt.Errorf("yaml: parse %s: %w", path, err)
	}
	if err := s.validate(); err != nil {
		return Spec{}, fmt.Errorf("yaml: %s: %w", path, err)
	}
	if s.Models == nil {
		s.Models = map[string]string{}
	}
	if s.Capabilities.Models == nil && len(s.Models) > 0 {
		s.Capabilities.Models = s.Models
	}
	if len(s.Capabilities.Strengths) == 0 {
		s.Capabilities.Strengths = s.Strengths
	}
	return s, nil
}

func (s Spec) validate() error {
	if s.ID == "" {
		return fmt.Errorf("missing id")
	}
	if s.Name == "" {
		return fmt.Errorf("missing name")
	}
	if s.Type == "" {
		s.Type = "cli"
	}
	if s.Type != "cli" {
		return fmt.Errorf("only type=cli is supported (got %q)", s.Type)
	}
	if s.Command.Binary == "" {
		return fmt.Errorf("missing command.binary")
	}
	return nil
}
