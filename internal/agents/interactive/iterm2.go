// SPDX-License-Identifier: MIT

package interactive

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type ITermStrategy struct{}

func (ITermStrategy) ID() string { return "iterm2" }

func (ITermStrategy) Run(ctx context.Context, cfg Config, req agents.AgentRequest) (agents.AgentResult, error) {
	if runtime.GOOS != "darwin" {
		return agents.AgentResult{}, errors.New("interactive: iterm2 strategy is macOS-only")
	}
	if _, err := exec.LookPath("osascript"); err != nil {
		return agents.AgentResult{}, errors.New("interactive: osascript not on PATH; pin strategy: pty")
	}
	if cfg.Binary == "" {
		return agents.AgentResult{}, errors.New("interactive: iterm2: binary required")
	}
	profile := cfg.ITermProfile
	if profile == "" {
		profile = "Default"
	}

	start := time.Now()
	prompt := promptFromRequest(req)
	script := buildITermScript(cfg.Binary, cfg.Args, profile, prompt)

	hard := durationOrDefault(cfg.HardTimeoutSeconds, defaultHardTimeout)
	rctx, cancel := context.WithTimeout(ctx, hard)
	defer cancel()

	out, err := exec.CommandContext(rctx, "osascript", "-e", script).CombinedOutput()
	if err != nil {
		return failedResult(err, start), fmt.Errorf("osascript: %w: %s", err, strings.TrimSpace(string(out)))
	}
	body := []byte(strings.TrimSpace(string(out)))
	return agents.AgentResult{
		Output:   agents.AgentOutput{Stdout: body, FinalMessage: string(body)},
		Duration: time.Since(start),
	}, nil
}

func buildITermScript(binary string, args []string, profile, prompt string) string {
	cmd := binary
	if len(args) > 0 {
		cmd = binary + " " + strings.Join(args, " ")
	}
	escapedPrompt := strings.ReplaceAll(prompt, `"`, `\"`)
	escapedCmd := strings.ReplaceAll(cmd, `"`, `\"`)
	return fmt.Sprintf(`tell application "iTerm2"
  set newWindow to create window with profile "%s"
  tell current session of newWindow
    write text "%s"
    delay 2
    write text "%s"
  end tell
end tell`, profile, escapedCmd, escapedPrompt)
}
