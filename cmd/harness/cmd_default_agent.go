// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ropeixoto/harnessx/internal/activeagent"
	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"golang.org/x/term"
)

// resolveDefaultAgent picks the agent id when a command did not get an
// explicit --agent / --adapter. Order:
//
//  1. explicit flag value passed in (non-empty wins)
//  2. .harness/config/active.yaml pin (`harness use <id>`)
//  3. HARNESS_DEFAULT_AGENT env var
//  4. interactive picker when stdin is a TTY
//  5. hard error when none of the above apply
//
// The hard error is intentional: a stale "claude" / "fake-real"
// literal default silently overrode the active.yaml pin for years and
// users got billed against the wrong adapter. Empty default + explicit
// resolution chain makes the choice visible and auditable.
func resolveDefaultAgent(explicit, dir string, out io.Writer, in io.Reader) (string, error) {
	if id := strings.TrimSpace(explicit); id != "" {
		return id, nil
	}
	if id := activeagent.ResolveAgentID(dir, ""); id != "" {
		return id, nil
	}
	if id := strings.TrimSpace(os.Getenv("HARNESS_DEFAULT_AGENT")); id != "" {
		return id, nil
	}
	available := registeredAgentIDs(dir)
	if len(available) == 0 {
		return "", fmt.Errorf("no adapter pinned: run `harness use <id>` (no candidate registered either; see `harness agent list`)")
	}
	if isTerminalReader(in) {
		fmt.Fprintln(out, "no adapter pinned. available:")
		for i, id := range available {
			fmt.Fprintf(out, "  %d) %s\n", i+1, id)
		}
		fmt.Fprintf(out, "pick [1-%d, default 1]: ", len(available))
		line, _ := bufio.NewReader(in).ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			return available[0], nil
		}
		var pick int
		if _, err := fmt.Sscanf(line, "%d", &pick); err != nil || pick < 1 || pick > len(available) {
			return available[0], nil
		}
		return available[pick-1], nil
	}
	return "", fmt.Errorf("no adapter pinned: run `harness use <id>` (registered: %s)", strings.Join(available, ", "))
}

func registeredAgentIDs(dir string) []string {
	reg, _, err := agentcmd.LoadAll(dir)
	if err != nil {
		return nil
	}
	return reg.IDs()
}

func isTerminalReader(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
