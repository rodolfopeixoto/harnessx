// SPDX-License-Identifier: MIT

package repl

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// runAdapterLogin prints the pinned adapter's LoginCommand and runs
// it inline when the user confirms. Lets the user re-auth without
// quitting the chat session.
func runAdapterLogin(ctx context.Context, opts *Options) {
	if opts.Adapter == nil {
		fmt.Fprintln(opts.Out, "  ✗ no adapter wired")
		return
	}
	caps := opts.Adapter.Capabilities()
	if caps.LoginCommand == "" {
		fmt.Fprintf(opts.Out, "  ✗ %s does not declare a login command. Docs: %s\n", opts.AdapterID, caps.AuthDocURL)
		return
	}
	fmt.Fprintf(opts.Out, "  about to run: %s\n", caps.LoginCommand)
	parts := strings.Fields(caps.LoginCommand)
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Stdin = opts.In
	cmd.Stdout = opts.Out
	cmd.Stderr = opts.Out
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(opts.Out, "  ✗ login: %v\n", err)
		return
	}
	if h := opts.Adapter.Healthcheck(ctx); !h.OK {
		fmt.Fprintf(opts.Out, "  ✗ still failing after login: %s\n", h.Err)
		return
	}
	fmt.Fprintf(opts.Out, "  ✓ %s authenticated\n", opts.AdapterID)
}

// cycleAdapter rotates opts.Adapter to the next registered adapter,
// printing the new pick + cost label. Used by `/cycle` and (when the
// readline Listener intercepts Shift+Tab) by the popup.
func cycleAdapter(opts *Options) {
	if len(opts.AdaptersList) == 0 {
		fmt.Fprintln(opts.Out, "  ✗ no other adapters registered")
		return
	}
	idx := -1
	for i, id := range opts.AdaptersList {
		if id == opts.AdapterID {
			idx = i
			break
		}
	}
	next := opts.AdaptersList[(idx+1)%len(opts.AdaptersList)]
	if opts.SwitchTo == nil {
		fmt.Fprintln(opts.Out, "  ✗ adapter switching not wired in this session")
		return
	}
	a, id, err := opts.SwitchTo(next)
	if err != nil {
		fmt.Fprintf(opts.Out, "  ✗ cycle: %v\n", err)
		return
	}
	opts.Adapter = a
	opts.AdapterID = id
	if opts.HealthProbe != nil {
		opts.HealthProbe.Swap(a)
	}
	fmt.Fprintf(opts.Out, "  [swap] now using %s (%s)\n", id, adapterBillingMode(id))
}

func switchModel(opts *Options, model string) {
	if model == "" {
		fmt.Fprintln(opts.Out, "  ✗ /model needs a name (e.g. /model claude-sonnet-4-6, /model gpt-5-mini)")
		return
	}
	opts.Model = model
	fmt.Fprintf(opts.Out, "  ✓ model switched to %s (next turn picks it up)\n", model)
}

func printModel(opts *Options) {
	current := opts.Model
	if current == "" {
		current = "(adapter default)"
	}
	fmt.Fprintf(opts.Out, "  current model: %s\n", current)
	fmt.Fprintln(opts.Out, "  swap with: /model <name>")
}

// switchAdapter performs /use mid-session. Delegates to the SwitchTo
// callback that cmd_chat wires in — keeps the repl package free of the
// agentcmd registry dependency.
func switchAdapter(opts *Options, id string) {
	if id == "" {
		fmt.Fprintln(opts.Out, "  ✗ /use needs an adapter id (try /agents)")
		return
	}
	if opts.SwitchTo == nil {
		fmt.Fprintln(opts.Out, "  ✗ /use unavailable: chat was started without an adapter")
		return
	}
	if n, err := strconv.Atoi(id); err == nil && n >= 1 && n <= len(opts.AdaptersList) {
		id = opts.AdaptersList[n-1]
	}
	adapter, canonical, err := opts.SwitchTo(id)
	if err != nil {
		fmt.Fprintf(opts.Out, "  ✗ /use %s: %v\n", id, err)
		return
	}
	opts.Adapter = adapter
	opts.AdapterID = canonical
	if opts.HealthProbe != nil {
		opts.HealthProbe.Swap(adapter)
	}
	fmt.Fprintf(opts.Out, "  ✓ switched to %s\n", canonical)
}
