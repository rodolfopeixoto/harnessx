// SPDX-License-Identifier: MIT

package repl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"

	"github.com/ropeixoto/harnessx/internal/agenthealth"
	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/intentplan"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/prompttpl"
	"github.com/ropeixoto/harnessx/internal/ui"
)

type Session struct {
	ID          string          `json:"id"`
	Goal        intentplan.Goal `json:"goal"`
	Started     time.Time       `json:"started"`
	Turns       []Turn          `json:"turns"`
	Root        string          `json:"root"`
	ContextMark int             `json:"context_mark,omitempty"`
	AutoGate    bool            `json:"auto_gate,omitempty"`
	BudgetUSD   float64         `json:"budget_usd,omitempty"`
	Label       string          `json:"label,omitempty"`
	ReadOnly    bool            `json:"-"`
}

type Turn struct {
	Time      time.Time              `json:"time"`
	Input     string                 `json:"input"`
	Action    string                 `json:"action"`
	AdapterID string                 `json:"adapter_id,omitempty"`
	TaskTag   string                 `json:"task_tag,omitempty"`
	Plan      *intentplan.Plan       `json:"plan,omitempty"`
	Result    *intentplan.ExecResult `json:"result,omitempty"`
	InTokens  int                    `json:"in_tokens,omitempty"`
	OutTokens int                    `json:"out_tokens,omitempty"`
	CostUSD   float64                `json:"cost_usd,omitempty"`
}

type Options struct {
	Root         string
	HarnessBin   string
	Goal         intentplan.Goal
	In           io.Reader
	Out          io.Writer
	Planner      Planner
	StepTimeout  time.Duration
	HealthProbe  *agenthealth.Probe
	Plain        bool
	Adapter      agents.AgentAdapter
	AdapterID    string
	Model        string
	Resume       *Session
	AutoGate     bool
	AdaptersList []string
	SwitchTo     func(id string) (agents.AgentAdapter, string, error)
	// Route selects an adapter from the registry for the given task tag
	// (planning, implementation, cheap_review, …). Lets `/plan` go to a
	// cheap model and plain text go to the implementation chain
	// without the REPL needing to import the router package directly.
	Route func(task string) (agents.AgentAdapter, string, error)
	// NoAdapter is true when the user passed --no-adapter; plain text
	// is refused with a clear message instead of falling into the
	// deterministic-planner harness do loop, which routinely takes
	// minutes against a fresh scratch project.
	NoAdapter     bool
	RouteEnabled  bool
	OneShot       bool
	Deterministic bool
	// Pipe is true for non-interactive runs (`harness chat --pipe`):
	// suppresses the greeting + session recap and bypasses the
	// readline TTY path, so stdin is consumed line-by-line and the
	// process exits cleanly on EOF.
	Pipe bool
	// OutputJSON emits one JSON envelope per completed turn on opts.Out
	// after the human-readable output, so wrapping tools (CI scripts,
	// IDE plugins) can parse the result without scraping ANSI text.
	OutputJSON bool
}

func levenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}
	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		curr[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(br)]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

type Planner func(ctx context.Context, goal intentplan.Goal, prompt string) (intentplan.Plan, error)

func NewDefaultPlanner() Planner {
	return func(ctx context.Context, goal intentplan.Goal, prompt string) (intentplan.Plan, error) {
		return DefaultPlan(goal, prompt), nil
	}
}

func DefaultPlan(goal intentplan.Goal, prompt string) intentplan.Plan {
	now := time.Now().UTC()
	switch goal {
	case intentplan.GoalDev:
		return intentplan.Plan{
			Goal: goal, Intent: prompt, Generated: now,
			Steps: []intentplan.Step{
				{Kind: intentplan.StepHarness, Title: "do (apply diff)", Cmd: []string{"do", prompt, "--yes", "--autonomy", "safe_execute"}},
				{Kind: intentplan.StepHarness, Title: "lint", Cmd: []string{"lint"}},
				{Kind: intentplan.StepHarness, Title: "test", Cmd: []string{"test"}},
				{Kind: intentplan.StepHarness, Title: "ci gate", Cmd: []string{"ci"}},
			},
			ExitWhen: intentplan.ExitCriteria{AllPass: []string{"ci"}},
		}
	case intentplan.GoalOps:
		return intentplan.Plan{
			Goal: goal, Intent: prompt, Generated: now,
			Steps: []intentplan.Step{
				{Kind: intentplan.StepHarness, Title: "doctor", Cmd: []string{"doctor"}},
			},
			ExitWhen: intentplan.ExitCriteria{AllPass: []string{"doctor"}},
		}
	case intentplan.GoalAds:
		return intentplan.Plan{
			Goal: goal, Intent: prompt, Generated: now,
			Steps: []intentplan.Step{
				{Kind: intentplan.StepHarness, Title: "explain prompt", Cmd: []string{"explain", prompt}},
			},
		}
	case intentplan.GoalResearch:
		return intentplan.Plan{
			Goal: goal, Intent: prompt, Generated: now,
			Steps: []intentplan.Step{
				{Kind: intentplan.StepHarness, Title: "context", Cmd: []string{"context"}},
			},
		}
	}
	return intentplan.Plan{Goal: goal, Intent: prompt, Generated: now}
}

//nolint:gocognit,gocyclo // REPL main loop: setup → prompt → dispatch lives here intentionally
func Run(ctx context.Context, opts Options) error {
	if !inKnownGoals(opts.Goal) {
		return fmt.Errorf("repl: unknown goal %q", opts.Goal)
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.In == nil {
		opts.In = os.Stdin
	}
	if opts.Planner == nil {
		opts.Planner = NewDefaultPlanner()
	}
	if opts.Root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		opts.Root = wd
	}
	sess := Session{
		ID: ids.New(), Goal: opts.Goal,
		Started: time.Now().UTC(), Root: opts.Root,
	}
	if opts.Resume != nil {
		sess.ID = opts.Resume.ID
		if opts.Resume.Goal != "" {
			sess.Goal = opts.Resume.Goal
		}
		sess.Turns = append(sess.Turns, opts.Resume.Turns...)
		sess.AutoGate = opts.Resume.AutoGate
		sess.ContextMark = opts.Resume.ContextMark
		sess.BudgetUSD = opts.Resume.BudgetUSD
		sess.Label = opts.Resume.Label
		sess.ReadOnly = opts.Resume.ReadOnly
	}
	if opts.AutoGate {
		sess.AutoGate = true
	}
	if !opts.Pipe {
		greet(opts.Out, sess)
	}
	labels := append([]string{}, opts.AdaptersList...)
	if rows, err := ListSessions(opts.Root); err == nil {
		for _, r := range rows {
			if r.Label != "" {
				labels = append(labels, r.Label)
			}
		}
	}
	historyPath := filepath.Join(opts.Root, ".harness", "sessions", sess.ID+".history")
	var prompter promptReader
	if opts.Pipe {
		prompter = &bufioPromptReader{r: bufio.NewReader(opts.In), w: io.Discard}
	} else {
		prompter = newPromptReader(opts.In, opts.Out, historyPath, chatCompleter(opts.AdaptersList, labels))
	}
	defer func() { _ = prompter.Close() }()
	for {
		badge := ""
		if opts.HealthProbe != nil {
			badge = agenthealth.Badge(opts.HealthProbe.Snapshot(), opts.Plain)
		}
		prompt := fmt.Sprintf("[%s%s]> ", sess.Goal, badge)
		continuation := fmt.Sprintf("[%s]… ", sess.Goal)
		input, err := prompter.ReadInput(prompt, continuation)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if input == "" {
			continue
		}
		if shouldExit(input) {
			if !opts.Pipe {
				summariseSession(opts.Out, &sess)
				fmt.Fprintln(opts.Out, "bye")
			}
			break
		}
		turnCtx, cancel := signalAwareCtx(ctx, opts.Out)
		turn := handleInput(turnCtx, &sess, &opts, input)
		cancel()
		sess.Turns = append(sess.Turns, turn)
		if err := persist(opts.Root, sess); err != nil {
			fmt.Fprintf(opts.Out, "warn: persist: %v\n", err)
		}
		if opts.OutputJSON {
			emitTurnJSON(opts.Out, sess.ID, turn)
		}
		if opts.OneShot {
			break
		}
	}
	return persist(opts.Root, sess)
}

//nolint:gocognit,gocyclo // one slash-command dispatch table; splitting fragments break audit
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
	}
	return turn
}

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

// quickAnswer routes a `/btw <question>` through the cheapest
// adapter chain (task tag cheap_review) so users can fire a side
// question without burning opus tokens. Costs are recorded on the
// turn so /cost + analytics aggregate them under the btw tag.
func quickAnswer(ctx context.Context, sess *Session, opts Options, question string, turn *Turn) {
	if question == "" {
		fmt.Fprintln(opts.Out, "  /btw needs a question")
		return
	}
	if !checkBudget(sess, opts.Out) {
		return
	}
	prompt := "Answer this side question briefly (<=3 sentences). Do not propose code or actions; just a direct answer.\n\nQuestion: " + question
	chatTurnFor(ctx, sess, opts, prompt, "cheap_review", turn)
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

// writerIsTerminal answers true only when out is an *os.File pointing
// at a real TTY. Buffers, pipes, and tee'd log writers all return
// false so the spinner falls back to the static "agent: working…"
// line.
func writerIsTerminal(out io.Writer) bool {
	f, ok := out.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
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

// signalAwareCtx returns a derived context that cancels on SIGINT so
// Ctrl-C during a long agent call or `harness ci` aborts just that
// turn instead of killing the whole REPL. Printing the carriage-
// return + clear lets the next prompt land on a fresh line even when
// the spinner was mid-frame. The caller MUST invoke the returned
// cancel func to release the signal handler.
func signalAwareCtx(parent context.Context, out io.Writer) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		select {
		case <-sig:
			fmt.Fprint(out, "\r  \r✗ interrupted — back to prompt (Ctrl-D or /exit to leave)\n")
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, func() {
		signal.Stop(sig)
		cancel()
	}
}

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

// recapSession asks the pinned adapter for a short summary of the
// current session and prints the reply inline. Useful at the end of a
// long chat to capture intent + decisions into the persistence layer.
// Falls back to a deterministic bullet list when no adapter is wired.
// recapSession asks the cheapest review model to summarise the session
// — typically gemini/kimi according to the cheap_review route — so
// the recap does not burn opus tokens just to bullet-list intent.
func recapSession(ctx context.Context, sess *Session, opts Options, turn *Turn) {
	if sess == nil || len(sess.Turns) == 0 {
		fmt.Fprintln(opts.Out, "  no turns to recap yet")
		return
	}
	if opts.Adapter == nil && opts.Route == nil {
		fmt.Fprintln(opts.Out, "  [recap] (no adapter wired — listing inputs)")
		for i, t := range sess.Turns {
			if t.Input == "" {
				continue
			}
			fmt.Fprintf(opts.Out, "  %d. %s\n", i+1, truncateForContext(t.Input, 120))
		}
		return
	}
	prompt := buildRecapPrompt(sess)
	chatTurnFor(ctx, sess, opts, prompt, "cheap_review", turn)
}

func buildRecapPrompt(sess *Session) string {
	var b strings.Builder
	b.WriteString("Summarise this HarnessX chat session in <=8 bullet points. ")
	b.WriteString("Focus on: what was built, what tests/sensors ran, what is still open. ")
	b.WriteString("Do NOT make plans; this is a recap, not a new task.\n\n")
	b.WriteString("# Session turns\n\n")
	for i, t := range sess.Turns {
		in := strings.TrimSpace(t.Input)
		if in == "" {
			continue
		}
		fmt.Fprintf(&b, "%d. [%s] %s\n", i+1, t.Action, truncateForContext(in, 200))
	}
	return b.String()
}

func executePlan(ctx context.Context, sess *Session, opts Options, prompt string, turn *Turn) {
	if opts.Planner == nil {
		turn.Action = "no-planner"
		fmt.Fprintln(opts.Out, "  ✗ no planner wired — pin an adapter with `harness use <id>` or pass --no-adapter to fall back to deterministic mode")
		return
	}
	plan, err := opts.Planner(ctx, sess.Goal, prompt)
	if err != nil {
		turn.Action = "plan-error"
		fmt.Fprintf(opts.Out, "planner: %v\n", err)
		return
	}
	turn.Plan = &plan
	res, err := intentplan.Execute(ctx, plan, intentplan.ExecOptions{
		HarnessBin: opts.HarnessBin, WorkingDir: opts.Root,
		Out: opts.Out, StepTimeout: opts.StepTimeout,
	})
	if err != nil {
		turn.Action = "execute-error"
		fmt.Fprintf(opts.Out, "executor: %v\n", err)
		return
	}
	turn.Result = &res
	turn.Action = "executed"
	if res.OK {
		fmt.Fprintln(opts.Out, "✓ plan green")
	} else {
		fmt.Fprintln(opts.Out, "✗ plan red — inspect step outputs above")
	}
}

func chatTurn(ctx context.Context, sess *Session, opts Options, prompt string, turn *Turn) {
	chatTurnFor(ctx, sess, opts, prompt, "implementation", turn)
}

// chatTurnFor is the routed variant: when opts.Route is wired and the
// task tag resolves to a registered adapter, we hand the request to
// that adapter instead of opts.Adapter. Lets `/plan` use a cheap
// model while plain text keeps using the implementation chain.
func chatTurnFor(ctx context.Context, sess *Session, opts Options, prompt, task string, turn *Turn) {
	adapter := opts.Adapter
	adapterID := opts.AdapterID
	if opts.Route != nil && opts.RouteEnabled {
		if a, id, err := opts.Route(task); err == nil && a != nil {
			adapter = a
			adapterID = id
		}
	}
	if adapter == nil {
		fmt.Fprintln(opts.Out, "✗ no adapter wired")
		return
	}
	fmt.Fprintf(opts.Out, "  %s %s %s\n",
		ui.Accent.Render("[agent]"),
		ui.Info.Render("calling "+adapterID),
		ui.Muted.Render("("+task+", "+adapterBillingMode(adapterID)+")…"))
	live := &prefixWriter{w: opts.Out, prefix: "  │ "}
	timeout := opts.StepTimeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	spinStop := startSpinner(opts.Out, opts.Plain)
	defer spinStop()
	live.onFirstWrite = spinStop
	res := adapter.Run(rctx, agents.AgentRequest{
		Prompt:     withConversationContext(sess, opts.AdapterID, opts.Model, prompt),
		Model:      opts.Model,
		WorkingDir: opts.Root,
		Timeout:    timeout,
		LiveOut:    live,
		Extra:      map[string]string{"task": task},
	})
	spinStop()
	live.flush()
	if res.Err != nil {
		fmt.Fprintf(opts.Out, "✗ %s: %v\n", adapterID, res.Err)
		return
	}
	if msg := strings.TrimSpace(res.Output.FinalMessage); msg != "" {
		fmt.Fprintln(opts.Out, msg)
	}
	fmt.Fprintf(opts.Out, "  %s %s %s\n",
		ui.MarkSuccess(),
		ui.Accent.Render(adapterID),
		ui.Muted.Render(fmt.Sprintf("done in %s · in=%d out=%d · ~$%.4f",
			res.Duration.Round(time.Millisecond),
			res.Usage.InputTokens, res.Usage.OutputTokens, res.Usage.EstimatedCostUSD)))
	if turn != nil {
		turn.InTokens = res.Usage.InputTokens
		turn.OutTokens = res.Usage.OutputTokens
		turn.CostUSD = res.Usage.EstimatedCostUSD
		turn.AdapterID = adapterID
		turn.TaskTag = task
	}
}

// startSpinner returns a stop function that is idempotent and safe to
// call multiple times. The spinner prints a frame every ~120ms while
// the agent runs so the user can tell the terminal is alive — claude
// and codex CLIs hold their JSON output until the call completes,
// which used to make `harness chat` look frozen for tens of seconds.
func startSpinner(out io.Writer, plain bool) func() {
	if plain || !writerIsTerminal(out) {
		// Pipe / log file / non-TTY: emit a single "agent: working…"
		// line so the consumer still sees liveness without getting
		// CR-clobbered braille glyphs in the log file.
		fmt.Fprintln(out, "agent: working…")
		return func() {}
	}
	frames := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		t := time.NewTicker(120 * time.Millisecond)
		defer t.Stop()
		i := 0
		for {
			select {
			case <-stop:
				fmt.Fprint(out, "\r\033[2K")
				return
			case <-t.C:
				fmt.Fprintf(out, "\r\033[2K  %c thinking…", frames[i%len(frames)])
				i++
			}
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			close(stop)
			<-done
		})
	}
}

// withConversationContext threads the last few session turns into the
// prompt so multi-turn chat reads as one conversation instead of a
// series of isolated single-shot calls (paper §3.2.1 Working Memory).
// Truncates aggressively because adapter context budgets are bounded
// and the most recent few turns dominate the signal.
func withConversationContext(sess *Session, adapterID, model, prompt string) string {
	identity := buildIdentityPreface(adapterID, model)
	if sess == nil || len(sess.Turns) == 0 {
		return identity + prompt
	}
	var b strings.Builder
	b.WriteString(identity)
	b.WriteString("# Conversation so far\n\n")
	if !writeConversationTurns(&b, sess) {
		return identity + prompt
	}
	b.WriteString("# New user message\n\n")
	b.WriteString(prompt)
	return b.String()
}

// writeConversationTurns appends the trailing window of session turns to
// b and returns true when at least one turn was written. Split out to
// keep withConversationContext's cyclomatic complexity within budget.
func writeConversationTurns(b *strings.Builder, sess *Session) bool {
	const maxTurns = 5
	const maxOutputBytes = 1200
	start := sess.ContextMark
	if len(sess.Turns)-start > maxTurns {
		start = len(sess.Turns) - maxTurns
	}
	if start < 0 {
		start = 0
	}
	wrote := false
	for i := start; i < len(sess.Turns); i++ {
		t := sess.Turns[i]
		in := strings.TrimSpace(t.Input)
		if in == "" || strings.HasPrefix(in, "/") {
			continue
		}
		wrote = true
		fmt.Fprintf(b, "[turn %d] user: %s\n", i+1, in)
		if t.Result != nil && len(t.Result.Steps) > 0 {
			fmt.Fprintf(b, "[turn %d] harness ran %d steps (ok=%t)\n", i+1, len(t.Result.Steps), t.Result.OK)
		}
		if t.Action == "chat" {
			if out := truncateForContext(t.Input, maxOutputBytes); out != "" {
				fmt.Fprintf(b, "[turn %d] agent replied (truncated): %s\n", i+1, out)
			}
		}
		b.WriteByte('\n')
	}
	return wrote
}

// buildIdentityPreface tells the underlying agent which adapter and
// model it is running as. Without it the CLI cannot answer "qual modelo
// está usando?" — most adapters (kimi-cli, antigravity, codex) have no
// introspection of the harness routing layer that picked them.
func buildIdentityPreface(adapterID, model string) string {
	if adapterID == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("# Runtime identity\n\n")
	fmt.Fprintf(&b, "You are running inside harness chat as the `%s` adapter.\n", adapterID)
	if model != "" {
		fmt.Fprintf(&b, "Active model override: `%s`.\n", model)
	} else {
		b.WriteString("Active model: adapter default (no /model override set).\n")
	}
	b.WriteString("If the user asks which model or adapter is in use, answer with these values.\n\n")
	return b.String()
}

func truncateForContext(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func runShell(ctx context.Context, opts Options, line string) {
	if line == "" {
		fmt.Fprintln(opts.Out, "shell: empty")
		return
	}
	fmt.Fprintf(opts.Out, "  $ %s\n", line)
	c := exec.CommandContext(ctx, "sh", "-c", line)
	c.Dir = opts.Root
	c.Stdout = opts.Out
	c.Stderr = opts.Out
	if err := c.Run(); err != nil {
		fmt.Fprintf(opts.Out, "  ✗ exit %v\n", err)
	}
}

func runHarnessCmd(ctx context.Context, opts Options, args []string) {
	fmt.Fprintf(opts.Out, "  $ harness %s\n", strings.Join(args, " "))
	c := exec.CommandContext(ctx, opts.HarnessBin, args...)
	c.Dir = opts.Root
	c.Stdout = opts.Out
	c.Stderr = opts.Out
	if err := c.Run(); err != nil {
		fmt.Fprintf(opts.Out, "  ✗ exit %v\n", err)
	}
}

type prefixWriter struct {
	w            io.Writer
	prefix       string
	buf          []byte
	atBOL        bool
	onFirstWrite func()
	firstWritten bool
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	if !p.firstWritten && len(b) > 0 {
		p.firstWritten = true
		if p.onFirstWrite != nil {
			p.onFirstWrite()
		}
	}
	if p.buf == nil {
		p.atBOL = true
	}
	for _, c := range b {
		if p.atBOL {
			p.buf = append(p.buf, []byte(p.prefix)...)
			p.atBOL = false
		}
		p.buf = append(p.buf, c)
		if c == '\n' {
			p.atBOL = true
		}
	}
	if p.atBOL {
		_, err := p.w.Write(p.buf)
		p.buf = p.buf[:0]
		if err != nil {
			return len(b), err
		}
	}
	return len(b), nil
}

func (p *prefixWriter) flush() {
	if len(p.buf) > 0 {
		_, _ = p.w.Write(p.buf)
		_, _ = p.w.Write([]byte{'\n'})
		p.buf = p.buf[:0]
	}
}
