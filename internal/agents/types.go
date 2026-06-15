// SPDX-License-Identifier: MIT

// Package agents defines the adapter contract for coding-CLI plug-ins.
// Concrete adapters live under internal/agents/{fake,yaml,…}. The core
// never imports Claude/Codex/Gemini/Kimi-specific code.
package agents

import (
	"context"
	"time"
)

type AgentAdapter interface {
	ID() string
	Name() string
	Capabilities() Capabilities
	Healthcheck(ctx context.Context) HealthcheckResult
	Run(ctx context.Context, req AgentRequest) AgentResult
	ParseUsage(output AgentOutput) Usage
	ClassifyFailure(output AgentOutput, exitCode int, runErr error) FailureType
}

type Capabilities struct {
	Text             bool              `json:"text"               yaml:"text"`
	Vision           bool              `json:"vision"             yaml:"vision"`
	Files            bool              `json:"files"              yaml:"files"`
	Folders          bool              `json:"folders"            yaml:"folders"`
	Diff             bool              `json:"diff"               yaml:"diff"`
	JSONOutput       bool              `json:"json_output"        yaml:"json_output"`
	StreamOutput     bool              `json:"stream_output"      yaml:"stream_output"`
	ToolUse          bool              `json:"tool_use"           yaml:"tool_use"`
	MCP              bool              `json:"mcp"                yaml:"mcp"`
	MaxContextTokens int               `json:"max_context_tokens" yaml:"max_context_tokens"`
	Strengths        []string          `json:"strengths,omitempty" yaml:"strengths,omitempty"`
	Models           map[string]string `json:"models,omitempty"    yaml:"models,omitempty"`
}

type HealthcheckResult struct {
	OK         bool   `json:"ok"`
	Detail     string `json:"detail,omitempty"`
	CLIVersion string `json:"cli_version,omitempty"`
	Err        string `json:"err,omitempty"`
}

type AgentRequest struct {
	Prompt     string
	Model      string
	WorkingDir string
	Timeout    time.Duration
	Files      []string
	Stdin      string
	// Extra is an opaque per-task hint bag (e.g. {"task":"implementation"}).
	Extra map[string]string
	// ExtraArgs are appended verbatim to the adapter's invocation. Used
	// by the executor to inject capabilities the YAML spec opts in to
	// (e.g. --mcp-config <path>) without templating the spec itself.
	ExtraArgs []string
}

type AgentResult struct {
	Output   AgentOutput
	Usage    Usage
	Err      error
	ExitCode int
	Duration time.Duration
	Failure  FailureType
}

type AgentOutput struct {
	Stdout       []byte
	Stderr       []byte
	FinalMessage string
}

type Usage struct {
	InputTokens       int     `json:"input_tokens"`
	CachedInputTokens int     `json:"cached_input_tokens"`
	OutputTokens      int     `json:"output_tokens"`
	ReasoningTokens   int     `json:"reasoning_tokens"`
	EstimatedCostUSD  float64 `json:"estimated_cost_usd"`
	// Mode is "reported" when usage came from the CLI's own metadata,
	// "estimated" when computed locally from prompt+output lengths.
	Mode string `json:"mode"`
}

type FailureType string

const (
	FailureNone         FailureType = ""
	FailureRateLimit    FailureType = "rate_limit"
	FailureContextLimit FailureType = "context_limit"
	FailureAuth         FailureType = "auth"
	FailureTransient    FailureType = "transient"
	FailurePermanent    FailureType = "permanent"
	FailureTimeout      FailureType = "timeout"
)

// IsRecoverable reports whether a failure justifies trying the next agent
// in the fallback chain. Auth + permanent failures abort the chain since
// the next agent will most likely hit the same wall (or worse, succeed
// hiding a real config issue).
func (f FailureType) IsRecoverable() bool {
	switch f {
	case FailureRateLimit, FailureContextLimit, FailureTransient, FailureTimeout:
		return true
	default:
		return false
	}
}
