// SPDX-License-Identifier: MIT

package repl

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/ropeixoto/harnessx/internal/intentplan"
	"github.com/ropeixoto/harnessx/internal/ui"
)

func shouldExit(line string) bool {
	switch line {
	case "/exit", "/quit", "exit", "quit", ":q", ":wq":
		return true
	}
	first := firstToken(line)
	switch first {
	case "/exit", "/quit":
		return true
	}
	return false
}

func inKnownGoals(g intentplan.Goal) bool {
	for _, k := range intentplan.KnownGoals() {
		if k == g {
			return true
		}
	}
	return false
}

func callsLLM(input string) bool {
	if input == "" {
		return false
	}
	llmSlashes := []string{
		"/exec ", "/do ", "/ship ", "/drive ", "/spec ", "/recap", "/btw ", "/cycle", "/plan ",
		"/login",
	}
	for _, p := range llmSlashes {
		if input == strings.TrimSpace(p) || strings.HasPrefix(input, p) {
			return true
		}
	}
	if strings.HasPrefix(input, "/") || strings.HasPrefix(input, "!") {
		return false
	}
	return true
}

// isMutatingInput reports whether an input would change state or
// call out to an agent. Used by --replay to refuse anything beyond
// inspection commands.
func isMutatingInput(input string) bool {
	if input == "" {
		return false
	}
	if strings.HasPrefix(input, "!") {
		return true
	}
	mutating := []string{
		"/exec ", "/do ", "/ship ", "/drive ", "/spec ", "/ci", "/test", "/lint",
		"/use ", "/budget ", "/auto-gate", "/autogate",
		"/clear", "/save ", "/recap", "/btw ", "/cycle", "/login", "/plan ", "/goal ",
		"/last", "/branch ", "/save-prompt ", "/prompt ",
		"/model ", "/route on", "/route off", "/routing on", "/routing off", "/once",
	}
	for _, p := range mutating {
		if input == strings.TrimSpace(p) || strings.HasPrefix(input, p) {
			return true
		}
	}
	if strings.HasPrefix(input, "/") {
		return false
	}
	return true // plain text → would call adapter
}

// looksLikeShellOrSlash detects inputs that look like CLI commands
// being typed into the chat by mistake — the user expected the REPL
// to parse `harness use 4` or `exec do thing` as a command but the
// REPL would otherwise bill them for a full agent turn. Returns the
// warning message when triggered, empty when input is legit chat.
func looksLikeShellOrSlash(input string) string {
	first := firstToken(input)
	if first == "" {
		return ""
	}
	if first == "harness" {
		return "  [guard] looks like a shell command. To run a shell command from chat prefix with '!' (e.g. !" + input + "). Slash equivalents: /use <id>, /ci, /test, /lint. Skipping the agent call to save tokens."
	}
	for _, slashish := range []string{"use", "exec", "do", "ship", "drive", "spec", "plan", "ci", "test", "lint", "agents", "cost", "diff", "save", "branch", "recap", "auto-gate", "budget", "goal", "help", "exit"} {
		if first == slashish {
			return "  [guard] '" + input + "' looks like a slash command — did you mean '/" + input + "'? Skipping the agent call. Prefix with '/' or '!' to run."
		}
	}
	return ""
}

// adapterSwitchPattern matches natural-language adapter-switch
// intents the user is likely to type without remembering the slash
// form. Examples: "use kimi", "harness use kimi", "switch to codex",
// "switch codex", "change adapter to gemini", "model kimi". On
// match we route through /use <id> so the user does not get billed
// for a paid agent turn just to swap models.
var adapterSwitchPattern = regexp.MustCompile(`(?i)^(?:harness\s+)?(?:use|switch(?:\s+to)?|change\s+adapter(?:\s+to)?|model)\s+([a-z0-9][a-z0-9_-]*)\s*$`)

// detectAdapterSwitch returns the adapter id captured from an
// English-y switch intent, or "" when the input is not a switch.
func detectAdapterSwitch(input string) string {
	m := adapterSwitchPattern.FindStringSubmatch(strings.TrimSpace(input))
	if len(m) != 2 {
		return ""
	}
	return strings.ToLower(m[1])
}

func printHelp(out io.Writer) {
	fmt.Fprintln(out, "commands:")
	fmt.Fprintln(out, "  plain text                     chat with pinned agent (streams)")
	fmt.Fprintln(out, "  !<shell cmd>                   run shell command")
	fmt.Fprintln(out, "  /exec <prompt>                 deterministic plan: do + lint + test + ci")
	fmt.Fprintln(out, "  /do <prompt>                   alias for /exec")
	fmt.Fprintln(out, "  /ship <prompt>                 harness ship (branch + spec + loop + commit)")
	fmt.Fprintln(out, "  /drive <prompt>                spec → failing tests → impl → ci (paper §3.4)")
	fmt.Fprintln(out, "  /spec <prompt>                 interactive editable spec author (Q&A + edit loop)")
	fmt.Fprintln(out, "  /ci | /test | /lint            run harness gate")
	fmt.Fprintln(out, "  /agents                        list registered adapters; mark active")
	fmt.Fprintln(out, "  /use <id>                      switch adapter mid-session")
	fmt.Fprintln(out, "  /model [name]                  swap model mid-session (no arg = print current)")
	fmt.Fprintln(out, "  /route [on|off]                toggle per-task multi-agent routing")
	fmt.Fprintln(out, "  /once                          exit after the next prompt (non-iterative)")
	fmt.Fprintln(out, "  /diff                          git diff --stat + full diff (project root)")
	fmt.Fprintln(out, "  /cost                          cumulative session token + USD spend")
	fmt.Fprintln(out, "  /timeline                      ASCII timeline of every turn in this session")
	fmt.Fprintln(out, "  /budget <usd|off>              cap session spend (refuses turns when exceeded)")
	fmt.Fprintln(out, "  /save <name>                   label this session for harness chat list")
	fmt.Fprintln(out, "  /branch <name>                 git checkout -B <name> + auto-label session")
	fmt.Fprintln(out, "  /save-prompt <name>            capture last plain text into a reusable template")
	fmt.Fprintln(out, "  /prompt <name>                 replay a saved prompt template")
	fmt.Fprintln(out, "  /prompts                       list saved prompt templates")
	fmt.Fprintln(out, "  /recap                         ask the agent to summarise the session so far")
	fmt.Fprintln(out, "  /clear                         drop conversation history from next prompt")
	fmt.Fprintln(out, "  /auto-gate on|off              toggle harness ci after each agent turn")
	fmt.Fprintln(out, "  /goal <dev|ads|research|ops>   switch session goal")
	fmt.Fprintln(out, "  /plan <prompt>                 emit plan JSON without executing")
	fmt.Fprintln(out, "  /last                          replay the previous prompt")
	fmt.Fprintln(out, "  /history                       list previous prompts")
	fmt.Fprintln(out, "  /help                          this message")
	fmt.Fprintln(out, "  /exit | /quit                  leave the session")
	fmt.Fprintln(out, "  end a line with \\ to continue prompt on next line")
}

var knownSlashes = []string{
	"/help", "/exit", "/quit", "/history", "/last",
	"/agents", "/cost", "/diff", "/clear", "/timeline",
	"/auto-gate", "/autogate",
	"/save", "/branch", "/recap",
	"/save-prompt", "/prompt", "/prompts",
	"/exec", "/do", "/ship", "/drive", "/spec", "/ci", "/test", "/lint",
	"/use", "/model", "/route", "/routing", "/once",
	"/budget", "/goal", "/plan",
}

func firstToken(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexAny(s, " \t"); idx > 0 {
		return s[:idx]
	}
	return s
}

func suggestSlash(out io.Writer, attempt string) {
	if attempt == "" {
		return
	}
	best := ""
	bestScore := 999
	for _, c := range knownSlashes {
		score := slashScore(attempt, c)
		if score < bestScore {
			bestScore = score
			best = c
		}
	}
	if best == "" || bestScore > 6 {
		fmt.Fprintf(out, "  %s unknown slash %q. try / for the menu or /help for long-form\n", ui.MarkWarn(), attempt)
		return
	}
	fmt.Fprintf(out, "  %s unknown slash %q — did you mean %s? (/ for menu, /help for long-form)\n",
		ui.MarkWarn(), attempt, ui.Accent.Render(best))
}

func slashScore(attempt, candidate string) int {
	dist := levenshtein(attempt, candidate)
	prefix := commonPrefixLen(attempt, candidate)
	return dist*2 - prefix
}

func commonPrefixLen(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func printSlashMenu(out io.Writer) {
	groups := []struct {
		title string
		items []slashEntry
	}{
		{"chat", []slashEntry{
			{"plain text", "talk to pinned agent (implementation chain)"},
			{"!<cmd>", "run shell command in project root"},
			{"/exec <p>", "deterministic plan: do + lint + test + ci"},
			{"/ship <p>", "branch + spec + loop + commit"},
			{"/drive <p>", "spec → failing tests → impl → ci"},
			{"/spec <p>", "interactive editable spec author (Q&A + edit)"},
			{"/recap", "cheap chain summary of session so far"},
			{"/btw <q>", "side question on the cheapest chain (no impact on current thread)"},
			{"/cycle", "rotate to the next registered adapter (cheap → expensive)"},
			{"/login", "run the pinned adapter's auth/login command"},
		}},
		{"gate", []slashEntry{
			{"/ci", "harness ci"},
			{"/test", "harness test"},
			{"/lint", "harness lint"},
			{"/auto-gate on|off", "run ci after each agent turn"},
		}},
		{"agents + cost", []slashEntry{
			{"/agents", "list registered adapters"},
			{"/use <id>", "switch active adapter"},
			{"/model [name]", "swap model mid-session"},
			{"/route [on|off]", "toggle per-task multi-agent routing"},
			{"/once", "exit after the next prompt"},
			{"/cost", "per-adapter token + USD spend"},
			{"/budget <usd|off>", "cap session spend"},
			{"/timeline", "ASCII turn timeline"},
			{"/diff", "git diff in project root"},
		}},
		{"memory", []slashEntry{
			{"/clear", "drop conversation history"},
			{"/history", "list previous prompts"},
			{"/last", "replay previous prompt"},
		}},
		{"session", []slashEntry{
			{"/save <name>", "label session for chat list"},
			{"/branch <name>", "git checkout -B + auto-label"},
			{"/save-prompt <n>", "capture last plain text template"},
			{"/prompt <n>", "replay saved template"},
			{"/prompts", "list saved templates"},
			{"/goal <id>", "switch session goal"},
			{"/plan <p>", "emit plan JSON without executing"},
		}},
		{"exit", []slashEntry{
			{"/exit · /quit", "leave the session"},
			{"/help", "long-form help"},
		}},
	}
	fmt.Fprintln(out, ui.Heading.Render("slash commands"))
	for _, g := range groups {
		fmt.Fprintln(out, "  "+ui.Accent.Render(g.title))
		for _, it := range g.items {
			fmt.Fprintf(out, "    %s  %s\n",
				ui.Info.Render(padRight(it.name, 22)),
				ui.Muted.Render(it.desc))
		}
	}
	fmt.Fprintln(out, ui.Muted.Render("  tip: ↑/↓ scrolls history, TAB completes a slash, /<TAB> shows all"))
}

type slashEntry struct {
	name, desc string
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}
