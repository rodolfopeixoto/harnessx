// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/ui"
)

// handleAuthFailure surfaces a clear next step when an adapter's
// healthcheck reports auth trouble. When the adapter exposes a
// LoginCommand, we prompt the user to run it from the current REPL
// instead of dumping a cryptic error.
func handleAuthFailure(ctx context.Context, out io.Writer, in io.Reader, adapter agents.AgentAdapter, adapterID string, h agents.HealthcheckResult) {
	caps := adapter.Capabilities()
	if !authy(h) {
		fmt.Fprintf(out, "chat: %s healthcheck warn: %s\n", adapterID, h.Err)
		return
	}
	fmt.Fprintf(out, "%s %s needs auth (%s)\n", ui.MarkWarn(), adapterID, oneLine(h.Err))
	if caps.LoginCommand == "" {
		if caps.AuthDocURL != "" {
			fmt.Fprintf(out, "  docs: %s\n", caps.AuthDocURL)
		}
		fmt.Fprintf(out, "  fix: log in via the %s CLI, then restart `harness chat`.\n", adapterID)
		return
	}
	fmt.Fprintf(out, "  fix: %s\n", ui.Accent.Render(caps.LoginCommand))
	r := bufio.NewReader(in)
	fmt.Fprintf(out, "  run it now? [y/N]: ")
	line, _ := r.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(line)) != "y" && strings.TrimSpace(strings.ToLower(line)) != "yes" {
		fmt.Fprintln(out, "  skipped — run the command yourself, then restart `harness chat`.")
		return
	}
	parts := strings.Fields(caps.LoginCommand)
	if len(parts) == 0 {
		fmt.Fprintln(out, "  ✗ adapter login command is empty")
		return
	}
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(out, "  ✗ login command exited: %v\n", err)
		return
	}
	if hh := adapter.Healthcheck(ctx); !hh.OK {
		fmt.Fprintf(out, "  %s still failing after login: %s\n", ui.MarkWarn(), hh.Err)
		return
	}
	fmt.Fprintf(out, "  %s %s authenticated\n", ui.MarkSuccess(), adapterID)
}

func authy(h agents.HealthcheckResult) bool {
	low := strings.ToLower(h.Err + " " + h.Detail)
	for _, hint := range []string{"unauthorized", "not logged in", "401", "invalid api key", "authentication failed", "auth required", "please log in", "login required"} {
		if strings.Contains(low, hint) {
			return true
		}
	}
	return false
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > 120 {
		return s[:120] + "…"
	}
	return s
}
