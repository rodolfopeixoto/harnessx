// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/hookscan"
)

// HookEvent is one phase of the agent lifecycle. Hooks file by event
// name matching the harness convention: pre-tool-use, post-tool-use,
// session-start, session-end.
type HookEvent string

const (
	HookPreToolUse  HookEvent = "pre-tool-use"
	HookPostToolUse HookEvent = "post-tool-use"
)

// HookOutcome records one hook invocation for the report + audit trail.
type HookOutcome struct {
	Name       string `json:"name"`
	Event      string `json:"event"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
	Skipped    bool   `json:"skipped,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// DispatchHooks runs every executable hook script registered under the
// project for the given event. Returns the per-hook outcome list. The
// caller decides whether a non-zero exit blocks the run.
func DispatchHooks(ctx context.Context, projectRoot string, event HookEvent, env []string) ([]HookOutcome, error) {
	all, err := hookscan.Scan(projectRoot)
	if err != nil {
		return nil, err
	}
	var outs []HookOutcome
	for _, h := range all {
		if !strings.EqualFold(h.Event, string(event)) {
			continue
		}
		out := HookOutcome{Name: h.Name, Event: h.Event}
		if !looksExecutable(h.ConfigPath) {
			out.Skipped = true
			out.Reason = "config not executable (chmod +x " + h.ConfigPath + ")"
			outs = append(outs, out)
			continue
		}
		out.Reason = h.ConfigPath
		start := time.Now()
		cmd := exec.CommandContext(ctx, h.ConfigPath)
		cmd.Dir = projectRoot
		cmd.Env = append(os.Environ(), env...)
		if err := cmd.Run(); err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				out.ExitCode = ee.ExitCode()
			} else {
				out.ExitCode = -1
				out.Reason = err.Error()
			}
		}
		out.DurationMs = time.Since(start).Milliseconds()
		outs = append(outs, out)
	}
	return outs, nil
}

func looksExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().Perm()&0o111 != 0
}

// FormatHookFailures returns a printable summary of blocking hook
// failures (non-zero exit, not skipped).
func FormatHookFailures(outs []HookOutcome) string {
	var parts []string
	for _, o := range outs {
		if o.Skipped {
			continue
		}
		if o.ExitCode != 0 {
			script := o.Reason
			if script == "" {
				script = o.Name
			}
			parts = append(parts,
				fmt.Sprintf("%s blocked by %s (exit %d)\n  → make the script exit 0 to allow, or remove %s",
					o.Event, script, o.ExitCode, script))
		}
	}
	return strings.Join(parts, "; ")
}

// HookOutputDir returns the per-run directory where hook logs land.
func HookOutputDir(runDir string) string {
	return filepath.Join(runDir, "hooks")
}
