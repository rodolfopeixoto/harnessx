// SPDX-License-Identifier: MIT

package repl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/intentplan"
)

//nolint:gocyclo // one slash-command dispatch table; splitting fragments break audit
func handleInput(ctx context.Context, sess *Session, opts *Options, input string) Turn {
	turn := Turn{Time: time.Now().UTC(), Input: input}
	if opts.Deterministic && callsLLM(input) {
		fmt.Fprintln(opts.Out, "  ✗ deterministic mode (chat --deterministic): refuses any agent call. /ci /test /lint /history /agents /cost /diff /timeline /help /clear /budget /goal /save /branch are allowed.")
		turn.Action = "deterministic-block"
		return turn
	}
	if sess.ReadOnly && isMutatingInput(input) {
		fmt.Fprintf(opts.Out, "  ✗ session is read-only (--replay); /history /agents /cost /diff /help are still available\n")
		turn.Action = "read-only-block"
		return turn
	}
	switch {
	case strings.HasPrefix(input, "!"):
		turn.Action = "shell"
		runShell(ctx, *opts, strings.TrimSpace(input[1:]))
	case strings.HasPrefix(input, "/exec "), strings.HasPrefix(input, "/do "):
		prompt := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(input, "/exec"), "/do"))
		executePlan(ctx, sess, *opts, prompt, &turn)
	case strings.HasPrefix(input, "/ship "):
		args := []string{"ship", strings.TrimSpace(strings.TrimPrefix(input, "/ship ")), "--yes", "--allow-dirty"}
		runHarnessCmd(ctx, *opts, args)
		turn.Action = "ship"
	case strings.HasPrefix(input, "/drive "):
		args := []string{"drive", strings.TrimSpace(strings.TrimPrefix(input, "/drive "))}
		runHarnessCmd(ctx, *opts, args)
		turn.Action = "drive"
	case strings.HasPrefix(input, "/spec "):
		args := []string{"spec", "author", strings.TrimSpace(strings.TrimPrefix(input, "/spec "))}
		runHarnessCmd(ctx, *opts, args)
		turn.Action = "spec"
	case strings.HasPrefix(input, "/ci"):
		runHarnessCmd(ctx, *opts, []string{"ci"})
		turn.Action = "ci"
	case strings.HasPrefix(input, "/test"):
		runHarnessCmd(ctx, *opts, []string{"test"})
		turn.Action = "test"
	case strings.HasPrefix(input, "/lint"):
		runHarnessCmd(ctx, *opts, []string{"lint"})
		turn.Action = "lint"
	case input == "/agents":
		turn.Action = "agents"
		printAgents(*opts)
	case input == "/cost":
		turn.Action = "cost"
		printCost(opts.Out, sess)
	case input == "/diff":
		turn.Action = "diff"
		runShell(ctx, *opts, "git diff --stat HEAD && echo '---' && git diff HEAD")
	case strings.HasPrefix(input, "/use "):
		turn.Action = "use"
		switchAdapter(opts, strings.TrimSpace(strings.TrimPrefix(input, "/use ")))
	case strings.HasPrefix(input, "/model "):
		turn.Action = "model"
		switchModel(opts, strings.TrimSpace(strings.TrimPrefix(input, "/model ")))
	case input == "/model":
		turn.Action = "model"
		printModel(opts)
	case input == "/route on", input == "/routing on":
		turn.Action = "route"
		opts.RouteEnabled = true
		fmt.Fprintln(opts.Out, "  ✓ multi-agent routing ON — text is classified and dispatched per task type")
	case input == "/route off", input == "/routing off":
		turn.Action = "route"
		opts.RouteEnabled = false
		fmt.Fprintln(opts.Out, "  ✓ multi-agent routing OFF — text always hits the pinned adapter")
	case input == "/route", input == "/routing":
		turn.Action = "route"
		state := "off"
		if opts.RouteEnabled {
			state = "on"
		}
		fmt.Fprintf(opts.Out, "  routing: %s (toggle: /route on|off)\n", state)
	case input == "/once":
		turn.Action = "once"
		opts.OneShot = true
		fmt.Fprintln(opts.Out, "  ✓ one-shot mode — REPL exits after the next prompt")
	case strings.HasPrefix(input, "/budget "):
		turn.Action = "budget"
		setBudget(sess, opts.Out, strings.TrimSpace(strings.TrimPrefix(input, "/budget ")))
	case strings.HasPrefix(input, "/save "):
		turn.Action = "save"
		name := strings.TrimSpace(strings.TrimPrefix(input, "/save "))
		setSessionLabel(sess, opts.Out, name)
	case strings.HasPrefix(input, "/branch "):
		turn.Action = "branch"
		name := strings.TrimSpace(strings.TrimPrefix(input, "/branch "))
		runBranch(ctx, sess, *opts, name)
	case strings.HasPrefix(input, "/save-prompt "):
		turn.Action = "save-prompt"
		name := strings.TrimSpace(strings.TrimPrefix(input, "/save-prompt "))
		savePromptTemplate(sess, *opts, name)
	case strings.HasPrefix(input, "/prompt "):
		turn.Action = "prompt"
		name := strings.TrimSpace(strings.TrimPrefix(input, "/prompt "))
		expanded := loadPromptTemplate(*opts, name)
		if expanded == "" {
			return turn
		}
		fmt.Fprintf(opts.Out, "↻ replaying prompt %q\n", name)
		return handleInput(ctx, sess, opts, expanded)
	case input == "/prompts":
		turn.Action = "prompts"
		listPromptTemplates(*opts)
	case input == "/timeline":
		turn.Action = "timeline"
		printTimeline(opts.Out, sess)
	case input == "/recap":
		turn.Action = "recap"
		recapSession(ctx, sess, *opts, &turn)
	case input == "/btw":
		turn.Action = "btw-help"
		fmt.Fprintln(opts.Out, "  /btw needs a question (e.g. /btw is gofmt enabled in this repo?)")
	case strings.HasPrefix(input, "/btw "):
		turn.Action = "btw"
		quickAnswer(ctx, sess, *opts, strings.TrimSpace(strings.TrimPrefix(input, "/btw ")), &turn)
	case input == "/cycle":
		turn.Action = "cycle"
		cycleAdapter(opts)
	case input == "/login":
		turn.Action = "login"
		runAdapterLogin(ctx, opts)
	case input == "/clear":
		turn.Action = "clear-context"
		sess.ContextMark = len(sess.Turns)
		fmt.Fprintln(opts.Out, "  ✓ working memory cleared — next agent turn starts fresh")
	case input == "/auto-gate on", input == "/autogate on":
		turn.Action = "auto-gate-on"
		sess.AutoGate = true
		fmt.Fprintln(opts.Out, "  ✓ auto-gate ON — harness ci will run after every agent turn")
	case input == "/auto-gate off", input == "/autogate off":
		turn.Action = "auto-gate-off"
		sess.AutoGate = false
		fmt.Fprintln(opts.Out, "  ✓ auto-gate OFF")
	case strings.HasPrefix(input, "/goal "):
		newGoal := intentplan.Goal(strings.TrimSpace(strings.TrimPrefix(input, "/goal ")))
		if inKnownGoals(newGoal) {
			sess.Goal = newGoal
			turn.Action = "goal-switch"
			fmt.Fprintf(opts.Out, "goal → %s\n", newGoal)
		} else {
			turn.Action = "goal-reject"
			fmt.Fprintf(opts.Out, "unknown goal %q; have %v\n", newGoal, intentplan.KnownGoals())
		}
	case strings.HasPrefix(input, "/plan "):
		prompt := strings.TrimSpace(strings.TrimPrefix(input, "/plan "))
		plan, err := opts.Planner(ctx, sess.Goal, prompt)
		if err != nil {
			turn.Action = "plan-error"
			fmt.Fprintf(opts.Out, "plan error: %v\n", err)
			return turn
		}
		turn.Plan = &plan
		turn.Action = "plan"
		body, _ := plan.MarshalPretty()
		fmt.Fprintln(opts.Out, string(body))
	case input == "/", input == "/?":
		turn.Action = "slash-menu"
		printSlashMenu(opts.Out)
	case input == "/help":
		turn.Action = "help"
		printHelp(opts.Out)
	case input == "/history":
		turn.Action = "history"
		printHistory(opts.Out, sess)
	case input == "/last":
		last := lastPromptInput(sess)
		if last == "" {
			fmt.Fprintln(opts.Out, "no previous prompt yet")
			turn.Action = "no-history"
			return turn
		}
		fmt.Fprintf(opts.Out, "↻ replaying: %s\n", last)
		return handleInput(ctx, sess, opts, last)
	default:
		return handleDefault(ctx, sess, opts, input, turn)
	}
	return turn
}

// handleDefault handles the fallthrough case in handleInput: unknown
// slashes, natural-language adapter switches, and plain text that
// should either hit the agent or the deterministic planner.
func handleDefault(ctx context.Context, sess *Session, opts *Options, input string, turn Turn) Turn {
	if strings.HasPrefix(input, "/") {
		turn.Action = "slash-unknown"
		suggestSlash(opts.Out, firstToken(input))
		return turn
	}
	if opts.Adapter != nil {
		if id := detectAdapterSwitch(input); id != "" {
			return handleInput(ctx, sess, opts, "/use "+id)
		}
		if hint := looksLikeShellOrSlash(input); hint != "" {
			fmt.Fprintln(opts.Out, hint)
			turn.Action = "intent_redirect"
			turn.CostUSD = 0
			return turn
		}
		if !checkBudget(sess, opts.Out) {
			turn.Action = "budget-exceeded"
			return turn
		}
		turn.Action = "chat"
		chatTurn(ctx, sess, *opts, input, &turn)
		if sess.AutoGate || opts.AutoGate {
			fmt.Fprintln(opts.Out, "  [auto-gate] running harness ci…")
			runHarnessCmd(ctx, *opts, []string{"ci"})
		}
		return turn
	}
	if opts.NoAdapter {
		fmt.Fprintf(opts.Out, "  ✗ no adapter wired (chat --no-adapter). use /exec %s or pin one with 'harness use <id>'\n", truncateForContext(input, 60))
		turn.Action = "no-adapter-block"
		return turn
	}
	executePlan(ctx, sess, *opts, input, &turn)
	return turn
}
