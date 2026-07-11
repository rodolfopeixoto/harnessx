// SPDX-License-Identifier: MIT

package repl

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ropeixoto/harnessx/internal/prompttpl"
)

// savePromptTemplate writes the most recent plain-text user input
// into .harness/prompts/<name>.md so the user can replay it later
// with /prompt. The "most recent plain-text" rule means slash and
// shell-escape turns are skipped; otherwise the slash command would
// save itself.
func savePromptTemplate(sess *Session, opts Options, name string) {
	if name == "" {
		fmt.Fprintln(opts.Out, "  ✗ /save-prompt needs a name (e.g. /save-prompt add-endpoint)")
		return
	}
	if !prompttpl.ValidName(name) {
		fmt.Fprintf(opts.Out, "  ✗ /save-prompt: %q is not a valid name (lowercase alnum, _ or -, ≤40 chars)\n", name)
		return
	}
	body := ""
	for i := len(sess.Turns) - 1; i >= 0; i-- {
		in := sess.Turns[i].Input
		if in == "" || strings.HasPrefix(in, "/") || strings.HasPrefix(in, "!") {
			continue
		}
		body = in
		break
	}
	if body == "" {
		fmt.Fprintln(opts.Out, "  ✗ /save-prompt: no plain-text turn to capture yet")
		return
	}
	if err := prompttpl.Save(opts.Root, name, body); err != nil {
		fmt.Fprintf(opts.Out, "  ✗ /save-prompt: %v\n", err)
		return
	}
	fmt.Fprintf(opts.Out, "  ✓ prompt %q saved (%d chars)\n", name, len(body))
}

func loadPromptTemplate(opts Options, name string) string {
	if name == "" {
		fmt.Fprintln(opts.Out, "  ✗ /prompt needs a name (try /prompts to list)")
		return ""
	}
	body, err := prompttpl.Load(opts.Root, name)
	if err != nil {
		fmt.Fprintf(opts.Out, "  ✗ /prompt %s: %v\n", name, err)
		return ""
	}
	return strings.TrimSpace(body)
}

func listPromptTemplates(opts Options) {
	names, err := prompttpl.List(opts.Root)
	if err != nil {
		fmt.Fprintf(opts.Out, "  ✗ /prompts: %v\n", err)
		return
	}
	if len(names) == 0 {
		fmt.Fprintln(opts.Out, "  no saved prompts (capture one with /save-prompt <name>)")
		return
	}
	for _, n := range names {
		fmt.Fprintf(opts.Out, "  %s\n", n)
	}
}

// runBranch creates or switches to a git branch in one step and
// labels the session with the same name so /save + git stay in
// sync. A nested-slash branch like "feature/cart" turns into the
// label "feature-cart" so it remains a single token in chat list
// output.
func runBranch(ctx context.Context, sess *Session, opts Options, name string) {
	if name == "" {
		fmt.Fprintln(opts.Out, "  ✗ /branch needs a name (try /branch feature/cart)")
		return
	}
	fmt.Fprintf(opts.Out, "  $ git checkout -B %s\n", name)
	c := exec.CommandContext(ctx, "git", "checkout", "-B", name)
	c.Dir = opts.Root
	c.Stdout = opts.Out
	c.Stderr = opts.Out
	if err := c.Run(); err != nil {
		fmt.Fprintf(opts.Out, "  ✗ /branch %s: %v\n", name, err)
		return
	}
	label := strings.ReplaceAll(name, "/", "-")
	if sess.Label == "" {
		sess.Label = label
		fmt.Fprintf(opts.Out, "  ✓ session labelled %q to match the branch\n", label)
	}
}
