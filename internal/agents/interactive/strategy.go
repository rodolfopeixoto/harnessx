// SPDX-License-Identifier: MIT

// Package interactive drives an interactive Claude Code REPL (or any
// other interactive LLM CLI) so HarnessX runs draw from the operator's
// subscription bucket rather than the Agent SDK monthly credit. Three
// orchestration strategies ship: PTY (default), tmux session, iTerm2
// session (macOS). All implement the same Strategy contract.
package interactive

import (
	"context"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type Strategy interface {
	ID() string
	Run(ctx context.Context, cfg Config, req agents.AgentRequest) (agents.AgentResult, error)
}

type Config struct {
	Binary             string
	Args               []string
	IdleMs             int
	HardTimeoutSeconds int
	BannerPattern      string
	TmuxSessionName    string
	ITermProfile       string
}
