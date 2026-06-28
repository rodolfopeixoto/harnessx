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
		Binary           string   `yaml:"binary"`
		FallbackBinaries []string `yaml:"fallback_binaries"`
		Check            string   `yaml:"check"`
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
		Format                string   `yaml:"format"` // jsonl | text
		FinalMessageJSONPath  string   `yaml:"final_message_json_path"`
		FinalMessageJSONPaths []string `yaml:"final_message_json_paths"`
		ErrorJSONPath         string   `yaml:"error_json_path"`
		UsageJSONPath         string   `yaml:"usage_json_path"`
		UsageJSONPaths        []string `yaml:"usage_json_paths"`
	} `yaml:"output"`

	FailureDetection map[string][]string `yaml:"failure_detection"`

	Cost struct {
		Mode                  string  `yaml:"mode"`
		InputTokenPricePer1M  float64 `yaml:"input_token_price_per_1m"`
		OutputTokenPricePer1M float64 `yaml:"output_token_price_per_1m"`
	} `yaml:"cost"`

	// Auth describes how the user logs into this adapter's CLI. Harness
	// never wraps the login itself; it just prints the command + doc URL
	// when healthcheck / certify detects an auth failure.
	Auth struct {
		LoginCommand string `yaml:"login_command"`
		DocURL       string `yaml:"doc_url"`
		Check        string `yaml:"check"`
		EnvVar       string `yaml:"env_var"`
	} `yaml:"auth"`

	API APISpec `yaml:"api"`

	Interactive InteractiveSpec `yaml:"interactive"`

	Experimental bool `yaml:"experimental"`
}

type InteractiveSpec struct {
	Strategy           string   `yaml:"strategy"`
	Binary             string   `yaml:"binary"`
	Args               []string `yaml:"args"`
	IdleMs             int      `yaml:"idle_ms"`
	HardTimeoutSeconds int      `yaml:"hard_timeout_seconds"`
	BannerPattern      string   `yaml:"banner_pattern"`
	Tmux               struct {
		SessionName string `yaml:"session_name"`
	} `yaml:"tmux"`
	ITerm2 struct {
		Profile string `yaml:"profile"`
	} `yaml:"iterm2"`
}

type APISpec struct {
	Endpoint        string            `yaml:"endpoint"`
	Method          string            `yaml:"method"`
	Headers         map[string]string `yaml:"headers"`
	Auth            APIAuth           `yaml:"auth"`
	RequestTemplate string            `yaml:"request_template"`
	Response        APIResponse       `yaml:"response"`
	TimeoutSeconds  int               `yaml:"timeout_seconds"`
	Retry           APIRetry          `yaml:"retry"`
}

type APIAuth struct {
	Header     string `yaml:"header"`
	Scheme     string `yaml:"scheme"`
	SecretRef  string `yaml:"secret_ref"`
	QueryParam string `yaml:"query_param"`
}

type APIResponse struct {
	FinalMessage string `yaml:"final_message"`
	Usage        string `yaml:"usage"`
}

type APIRetry struct {
	Max       int `yaml:"max"`
	BackoffMs int `yaml:"backoff_ms"`
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
	if s.Capabilities.LoginCommand == "" {
		s.Capabilities.LoginCommand = s.Auth.LoginCommand
	}
	if s.Capabilities.AuthDocURL == "" {
		s.Capabilities.AuthDocURL = s.Auth.DocURL
	}
	if s.Capabilities.AuthEnvVar == "" {
		s.Capabilities.AuthEnvVar = s.Auth.EnvVar
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
	switch s.Type {
	case "cli":
		if s.Command.Binary == "" {
			return fmt.Errorf("type=cli: missing command.binary")
		}
	case "api":
		if s.API.Endpoint == "" {
			return fmt.Errorf("type=api: missing api.endpoint")
		}
	case "interactive":
		if s.Interactive.Binary == "" {
			return fmt.Errorf("type=interactive: missing interactive.binary")
		}
		switch s.Interactive.Strategy {
		case "", "pty", "tmux", "iterm2":
		default:
			return fmt.Errorf("type=interactive: unknown strategy %q (pty|tmux|iterm2)", s.Interactive.Strategy)
		}
	default:
		return fmt.Errorf("unknown type %q (cli|api|interactive)", s.Type)
	}
	return nil
}
