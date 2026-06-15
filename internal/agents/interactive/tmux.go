// SPDX-License-Identifier: MIT

package interactive

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type TmuxStrategy struct{}

func (TmuxStrategy) ID() string { return "tmux" }

func (TmuxStrategy) Run(ctx context.Context, cfg Config, req agents.AgentRequest) (agents.AgentResult, error) {
	if _, err := exec.LookPath("tmux"); err != nil {
		return agents.AgentResult{}, errors.New("interactive: tmux not on PATH; pin strategy: pty in adapter YAML")
	}
	if cfg.Binary == "" {
		return agents.AgentResult{}, errors.New("interactive: tmux: binary required")
	}
	session := cfg.TmuxSessionName
	if session == "" {
		session = "harness-claude-interactive"
	}
	if err := ensureTmuxSession(ctx, session, cfg.Binary, cfg.Args); err != nil {
		return agents.AgentResult{}, err
	}

	idle := idleThreshold(cfg.IdleMs)
	hard := durationOrDefault(cfg.HardTimeoutSeconds, defaultHardTimeout)
	rctx, cancel := context.WithTimeout(ctx, hard)
	defer cancel()

	start := time.Now()
	if _, err := captureUntilStable(rctx, session, idle); err != nil {
		return failedResult(err, start), err
	}
	before, err := capturePane(rctx, session)
	if err != nil {
		return failedResult(err, start), err
	}

	prompt := promptFromRequest(req)
	if err := sendKeys(rctx, session, prompt); err != nil {
		return failedResult(err, start), err
	}

	after, err := captureUntilStable(rctx, session, idle)
	if err != nil {
		return failedResult(err, start), err
	}
	body := diffAfter(before, after)
	return agents.AgentResult{
		Output:   agents.AgentOutput{Stdout: body, FinalMessage: string(body)},
		Duration: time.Since(start),
	}, nil
}

func ensureTmuxSession(ctx context.Context, session, binary string, args []string) error {
	if exec.CommandContext(ctx, "tmux", "has-session", "-t", session).Run() == nil {
		return nil
	}
	cmd := append([]string{"new-session", "-d", "-s", session, binary}, args...)
	out, err := exec.CommandContext(ctx, "tmux", cmd...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux new-session: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func sendKeys(ctx context.Context, session, prompt string) error {
	out, err := exec.CommandContext(ctx, "tmux", "send-keys", "-t", session, prompt, "Enter").CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux send-keys: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func capturePane(ctx context.Context, session string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, "tmux", "capture-pane", "-p", "-t", session, "-S", "-").Output()
	if err != nil {
		return nil, fmt.Errorf("tmux capture-pane: %w", err)
	}
	return out, nil
}

func captureUntilStable(ctx context.Context, session string, idle time.Duration) ([]byte, error) {
	var prev []byte
	stableSince := time.Time{}
	for {
		if ctx.Err() != nil {
			return prev, ctx.Err()
		}
		current, err := capturePane(ctx, session)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(prev, current) {
			if stableSince.IsZero() {
				stableSince = time.Now()
			}
			if time.Since(stableSince) >= idle {
				return current, nil
			}
		} else {
			stableSince = time.Time{}
			prev = current
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func diffAfter(before, after []byte) []byte {
	if len(before) == 0 {
		return after
	}
	if idx := bytes.Index(after, before); idx >= 0 {
		return after[idx+len(before):]
	}
	return after
}
