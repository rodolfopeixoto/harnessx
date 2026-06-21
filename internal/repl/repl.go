// SPDX-License-Identifier: MIT

package repl

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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
	NoAdapter bool
}

type SessionSummary struct {
	ID        string
	Label     string
	Goal      intentplan.Goal
	Turns     int
	LastInput string
}

// LoadSession rehydrates a prior chat from .harness/sessions/<id>.jsonl.
// The file is JSONL of Turn records (see persist) without the Session
// envelope, so we synthesise a Session shell and replay every turn into
// it. Used by `harness chat --resume <id>`.
func LoadSession(root, id string) (*Session, error) {
	p := sessionPath(root, id)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sess := Session{ID: id, Started: time.Now().UTC(), Root: root}
	dec := json.NewDecoder(f)
	for dec.More() {
		var t Turn
		if err := dec.Decode(&t); err != nil {
			return nil, err
		}
		sess.Turns = append(sess.Turns, t)
	}
	if meta, err := loadSessionMeta(root, id); err == nil {
		if meta.Goal != "" {
			sess.Goal = intentplan.Goal(meta.Goal)
		}
		sess.Label = meta.Label
		sess.ContextMark = meta.ContextMark
		sess.AutoGate = meta.AutoGate
		sess.BudgetUSD = meta.BudgetUSD
	}
	if sess.Goal == "" {
		for i := len(sess.Turns) - 1; i >= 0; i-- {
			if sess.Turns[i].Plan != nil {
				sess.Goal = sess.Turns[i].Plan.Goal
				break
			}
		}
	}
	if sess.Goal == "" {
		sess.Goal = intentplan.GoalDev
	}
	return &sess, nil
}

// ResolveSessionID converts either a ulid or a /save label into the
// canonical session id. Labels are matched against ListSessions; on
// ambiguity (two sessions sharing a label) the newest wins. Returns
// the input unchanged when no label matches so callers can still
// load by raw ulid.
func ResolveSessionID(root, arg string) string {
	if arg == "" {
		return arg
	}
	rows, err := ListSessions(root)
	if err != nil {
		return arg
	}
	for _, r := range rows {
		if r.ID == arg {
			return arg
		}
	}
	for _, r := range rows {
		if r.Label == arg {
			return r.ID
		}
	}
	return arg
}

// SuggestSession returns the closest known label or id to arg via
// Levenshtein distance, with a max distance of 3. Empty when nothing
// is close enough — callers use the empty case to fall through to
// the canonical "session not found" error so we never auto-resolve
// to the wrong session silently.
func SuggestSession(root, arg string) string {
	if arg == "" {
		return ""
	}
	rows, err := ListSessions(root)
	if err != nil || len(rows) == 0 {
		return ""
	}
	candidates := make([]string, 0, 2*len(rows))
	for _, r := range rows {
		if r.Label != "" {
			candidates = append(candidates, r.Label)
		}
		candidates = append(candidates, r.ID)
	}
	best := ""
	bestDist := 4
	for _, c := range candidates {
		d := levenshtein(arg, c)
		if d < bestDist {
			best = c
			bestDist = d
		}
	}
	return best
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

// ListSessions returns one summary per .harness/sessions/*.jsonl file,
// sorted newest first by file mtime.
func ListSessions(root string) ([]SessionSummary, error) {
	dir := filepath.Join(root, ".harness", "sessions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	type entryWithStat struct {
		name  string
		mtime time.Time
	}
	var stats []entryWithStat
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		stats = append(stats, entryWithStat{name: e.Name(), mtime: info.ModTime()})
	}
	for i := range stats {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].mtime.After(stats[i].mtime) {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}
	out := make([]SessionSummary, 0, len(stats))
	for _, s := range stats {
		id := strings.TrimSuffix(s.name, ".jsonl")
		sess, err := LoadSession(root, id)
		if err != nil {
			continue
		}
		last := ""
		for i := len(sess.Turns) - 1; i >= 0; i-- {
			if sess.Turns[i].Input != "" {
				last = sess.Turns[i].Input
				if len(last) > 60 {
					last = last[:60] + "…"
				}
				break
			}
		}
		out = append(out, SessionSummary{
			ID: sess.ID, Label: sess.Label,
			Goal: sess.Goal, Turns: len(sess.Turns), LastInput: last,
		})
	}
	return out, nil
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
	greet(opts.Out, sess)
	labels := []string{}
	for _, s := range opts.AdaptersList {
		labels = append(labels, s)
	}
	if rows, err := ListSessions(opts.Root); err == nil {
		for _, r := range rows {
			if r.Label != "" {
				labels = append(labels, r.Label)
			}
		}
	}
	historyPath := filepath.Join(opts.Root, ".harness", "sessions", sess.ID+".history")
	prompter := newPromptReader(opts.In, opts.Out, historyPath, chatCompleter(opts.AdaptersList, labels))
	defer prompter.Close()
	for {
		badge := ""
		if opts.HealthProbe != nil {
			badge = agenthealth.Badge(opts.HealthProbe.Snapshot(), opts.Plain)
		}
		prompt := fmt.Sprintf("\n[%s%s]> ", sess.Goal, badge)
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
			fmt.Fprintln(opts.Out, "bye")
			break
		}
		turnCtx, cancel := signalAwareCtx(ctx, opts.Out)
		turn := handleInput(turnCtx, &sess, &opts, input)
		cancel()
		sess.Turns = append(sess.Turns, turn)
		if err := persist(opts.Root, sess); err != nil {
			fmt.Fprintf(opts.Out, "warn: persist: %v\n", err)
		}
	}
	return persist(opts.Root, sess)
}

func handleInput(ctx context.Context, sess *Session, opts *Options, input string) Turn {
	turn := Turn{Time: time.Now().UTC(), Input: input}
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
		if opts.Adapter != nil {
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
	adapter, canonical, err := opts.SwitchTo(id)
	if err != nil {
		fmt.Fprintf(opts.Out, "  ✗ /use %s: %v\n", id, err)
		return
	}
	opts.Adapter = adapter
	opts.AdapterID = canonical
	fmt.Fprintf(opts.Out, "  ✓ switched to %s\n", canonical)
}

// setBudget parses a USD value from /budget and stores it on the
// session. Cumulative spend is enforced by checkBudget before each
// chat turn.
func setBudget(sess *Session, out io.Writer, raw string) {
	raw = strings.TrimPrefix(raw, "$")
	if raw == "" || raw == "off" || raw == "0" {
		sess.BudgetUSD = 0
		fmt.Fprintln(out, "  ✓ budget cleared")
		return
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		fmt.Fprintf(out, "  ✗ /budget needs a non-negative USD number (e.g. /budget 0.50)\n")
		return
	}
	sess.BudgetUSD = v
	fmt.Fprintf(out, "  ✓ budget set to $%.4f for this session\n", v)
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
		"/exec ", "/do ", "/ship ", "/drive ", "/ci", "/test", "/lint",
		"/use ", "/budget ", "/auto-gate", "/autogate",
		"/clear", "/save ", "/recap", "/plan ", "/goal ",
		"/last", "/branch ", "/save-prompt ", "/prompt ",
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

// setSessionLabel tags the session with a human-readable alias so
// `harness chat list` and exports show "shop-api-checkout" instead of
// an opaque ulid. Refuses obviously broken inputs (slashes, dot
// prefix) so it stays safe to interpolate into paths and CLI output.
func setSessionLabel(sess *Session, out io.Writer, label string) {
	if label == "" {
		fmt.Fprintln(out, "  ✗ /save needs a name (try /save my-feature)")
		return
	}
	if strings.ContainsAny(label, "/\\\n\t") || strings.HasPrefix(label, ".") {
		fmt.Fprintf(out, "  ✗ /save: %q contains an unsupported character\n", label)
		return
	}
	if len(label) > 80 {
		label = label[:80]
	}
	sess.Label = label
	fmt.Fprintf(out, "  ✓ session labelled %q\n", label)
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

// checkBudget returns false when running another chat turn would push
// cumulative spend over the configured cap. Prints why and refuses.
func checkBudget(sess *Session, out io.Writer) bool {
	if sess == nil || sess.BudgetUSD <= 0 {
		return true
	}
	var spent float64
	for _, t := range sess.Turns {
		spent += t.CostUSD
	}
	if spent >= sess.BudgetUSD {
		fmt.Fprintf(out, "  ✗ budget exhausted: $%.4f spent of $%.4f cap. /budget off to reset.\n",
			spent, sess.BudgetUSD)
		return false
	}
	return true
}

func executePlan(ctx context.Context, sess *Session, opts Options, prompt string, turn *Turn) {
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
	if opts.Route != nil {
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
	res := adapter.Run(rctx, agents.AgentRequest{
		Prompt:     withConversationContext(sess, prompt),
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
	if plain {
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
				fmt.Fprint(out, "\r  \r")
				return
			case <-t.C:
				fmt.Fprintf(out, "\r  %c thinking…", frames[i%len(frames)])
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
func withConversationContext(sess *Session, prompt string) string {
	if sess == nil || len(sess.Turns) == 0 {
		return prompt
	}
	const maxTurns = 5
	const maxOutputBytes = 1200
	start := sess.ContextMark
	if len(sess.Turns)-start > maxTurns {
		start = len(sess.Turns) - maxTurns
	}
	if start < 0 {
		start = 0
	}
	var b strings.Builder
	b.WriteString("# Conversation so far\n\n")
	wrote := false
	for i := start; i < len(sess.Turns); i++ {
		t := sess.Turns[i]
		in := strings.TrimSpace(t.Input)
		if in == "" || strings.HasPrefix(in, "/") {
			continue
		}
		wrote = true
		fmt.Fprintf(&b, "[turn %d] user: %s\n", i+1, in)
		if t.Result != nil && len(t.Result.Steps) > 0 {
			fmt.Fprintf(&b, "[turn %d] harness ran %d steps (ok=%t)\n", i+1, len(t.Result.Steps), t.Result.OK)
		}
		if t.Action == "chat" {
			out := truncateForContext(t.Input, maxOutputBytes)
			if out != "" {
				fmt.Fprintf(&b, "[turn %d] agent replied (truncated): %s\n", i+1, out)
			}
		}
		b.WriteByte('\n')
	}
	if !wrote {
		return prompt
	}
	b.WriteString("# New user message\n\n")
	b.WriteString(prompt)
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
	w      io.Writer
	prefix string
	buf    []byte
	atBOL  bool
}

func (p *prefixWriter) Write(b []byte) (int, error) {
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

func readMultilineInput(rd *bufio.Reader, out io.Writer, goal intentplan.Goal) (string, error) {
	var parts []string
	for {
		line, err := rd.ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		trimmed := strings.TrimRight(line, "\n")
		if strings.HasSuffix(trimmed, "\\") {
			parts = append(parts, strings.TrimSuffix(trimmed, "\\"))
			fmt.Fprintf(out, "[%s]… ", goal)
			continue
		}
		parts = append(parts, trimmed)
		break
	}
	return strings.TrimSpace(strings.Join(parts, "\n")), nil
}

func shouldExit(line string) bool {
	switch line {
	case "/exit", "/quit", "exit", "quit":
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

func greet(out io.Writer, s Session) {
	fmt.Fprintf(out, "harness chat — session %s, goal=%s\n", s.ID, s.Goal)
	fmt.Fprintln(out, `plain text → talk to agent · /exec → plan+run · !<cmd> → shell · /help · /exit`)
	fmt.Fprintln(out, ui.Muted.Render(`multi-line: end line with \  ·  or wrap with """ … """  ·  / lists slashes`))
}

func printHelp(out io.Writer) {
	fmt.Fprintln(out, "commands:")
	fmt.Fprintln(out, "  plain text                     chat with pinned agent (streams)")
	fmt.Fprintln(out, "  !<shell cmd>                   run shell command")
	fmt.Fprintln(out, "  /exec <prompt>                 deterministic plan: do + lint + test + ci")
	fmt.Fprintln(out, "  /do <prompt>                   alias for /exec")
	fmt.Fprintln(out, "  /ship <prompt>                 harness ship (branch + spec + loop + commit)")
	fmt.Fprintln(out, "  /drive <prompt>                spec → failing tests → impl → ci (paper §3.4)")
	fmt.Fprintln(out, "  /ci | /test | /lint            run harness gate")
	fmt.Fprintln(out, "  /agents                        list registered adapters; mark active")
	fmt.Fprintln(out, "  /use <id>                      switch adapter mid-session")
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

// printTimeline renders an at-a-glance ASCII view of every turn in
// the session: clock, action label, truncated input, and cost in
// USD. Designed for the "what happened today?" lookback after a long
// chat. Cumulative cost is printed at the foot for a one-line sanity
// check.
func printTimeline(out io.Writer, sess *Session) {
	if sess == nil || len(sess.Turns) == 0 {
		fmt.Fprintln(out, "  no turns yet")
		return
	}
	var total float64
	for i, t := range sess.Turns {
		clock := t.Time.Local().Format("15:04:05")
		input := truncateForContext(t.Input, 60)
		cost := ""
		if t.CostUSD > 0 {
			cost = fmt.Sprintf("  $%.4f", t.CostUSD)
			total += t.CostUSD
		}
		fmt.Fprintf(out, "  %3d  %s  [%-15s] %s%s\n", i+1, clock, t.Action, input, cost)
	}
	fmt.Fprintf(out, "\n  total: %d turns, ~$%.4f\n", len(sess.Turns), total)
}

// adapterBillingMode tells the user whether the adapter is going to
// charge against their API key (oneshot CLI: claude --print, codex
// exec, gemini -p) or against a logged-in plan/subscription
// (interactive CLI: claude-interactive, kimi chat). The distinction
// matters because oneshot calls show up on a separate invoice while
// interactive ones drain the user's chat-mode token quota.
func adapterBillingMode(id string) string {
	switch id {
	case "claude", "codex", "gemini", "anthropic-api", "openai-api",
		"gemini-api", "moonshot-api", "minimax-api":
		return "oneshot · API-billed"
	case "claude-interactive", "kimi", "ollama":
		return "interactive · plan/local"
	case "fake":
		return "fake · free"
	}
	return "unknown billing"
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
			{"/recap", "cheap chain summary of session so far"},
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

func printAgents(opts Options) {
	if len(opts.AdaptersList) == 0 {
		fmt.Fprintln(opts.Out, "no adapters registered (run 'harness agent list' for details)")
		return
	}
	fmt.Fprintf(opts.Out, "active: %s\n", opts.AdapterID)
	for _, id := range opts.AdaptersList {
		marker := "  "
		if id == opts.AdapterID {
			marker = "→ "
		}
		fmt.Fprintf(opts.Out, "%s%s\n", marker, id)
	}
}

func printCost(out io.Writer, sess *Session) {
	if sess == nil || len(sess.Turns) == 0 {
		fmt.Fprintln(out, "no agent turns yet")
		return
	}
	totals := aggregateCost(sess.Turns)
	renderCostReport(out, sess.ID, totals)
}

type costRow struct {
	AdapterID string
	Task      string
	Turns     int
	InTokens  int
	OutTokens int
	CostUSD   float64
}

type costTotals struct {
	ChatTurns int
	Total     costRow
	PerAgent  []costRow
}

func aggregateCost(turns []Turn) costTotals {
	byKey := map[string]*costRow{}
	keys := []string{}
	total := costRow{AdapterID: "TOTAL"}
	chatTurns := 0
	for _, t := range turns {
		if t.Action == "chat" {
			chatTurns++
		}
		if t.InTokens == 0 && t.OutTokens == 0 && t.CostUSD == 0 {
			continue
		}
		id := t.AdapterID
		if id == "" {
			id = "unknown"
		}
		key := id + "|" + t.TaskTag
		row, ok := byKey[key]
		if !ok {
			row = &costRow{AdapterID: id, Task: t.TaskTag}
			byKey[key] = row
			keys = append(keys, key)
		}
		row.Turns++
		row.InTokens += t.InTokens
		row.OutTokens += t.OutTokens
		row.CostUSD += t.CostUSD
		total.Turns++
		total.InTokens += t.InTokens
		total.OutTokens += t.OutTokens
		total.CostUSD += t.CostUSD
	}
	rows := make([]costRow, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, *byKey[k])
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].CostUSD > rows[j].CostUSD })
	return costTotals{ChatTurns: chatTurns, Total: total, PerAgent: rows}
}

func renderCostReport(out io.Writer, sessionID string, t costTotals) {
	fmt.Fprintf(out, "session %s: %d chat turns\n", sessionID, t.ChatTurns)
	if len(t.PerAgent) == 0 {
		fmt.Fprintln(out, "  (no recorded usage)")
		return
	}
	fmt.Fprintf(out, "  %-12s %-16s %5s %8s %8s %10s\n", "ADAPTER", "TASK", "TURNS", "IN", "OUT", "COST")
	for _, r := range t.PerAgent {
		task := r.Task
		if task == "" {
			task = "-"
		}
		fmt.Fprintf(out, "  %-12s %-16s %5d %8d %8d $%9.4f\n",
			r.AdapterID, task, r.Turns, r.InTokens, r.OutTokens, r.CostUSD)
	}
	fmt.Fprintln(out, "  "+strings.Repeat("─", 64))
	fmt.Fprintf(out, "  %-12s %-16s %5d %8d %8d $%9.4f\n",
		"TOTAL", "", t.Total.Turns, t.Total.InTokens, t.Total.OutTokens, t.Total.CostUSD)
}

func printHistory(out io.Writer, sess *Session) {
	if sess == nil || len(sess.Turns) == 0 {
		fmt.Fprintln(out, "history empty")
		return
	}
	start := 0
	if len(sess.Turns) > 20 {
		start = len(sess.Turns) - 20
	}
	for i := start; i < len(sess.Turns); i++ {
		fmt.Fprintf(out, "%3d  %s\n", i+1, sess.Turns[i].Input)
	}
}

func lastPromptInput(sess *Session) string {
	if sess == nil {
		return ""
	}
	for i := len(sess.Turns) - 1; i >= 0; i-- {
		in := sess.Turns[i].Input
		if in == "" || strings.HasPrefix(in, "/") {
			continue
		}
		return in
	}
	return ""
}

func sessionPath(root, id string) string {
	return filepath.Join(root, ".harness", "sessions", id+".jsonl")
}

func sessionMetaPath(root, id string) string {
	return filepath.Join(root, ".harness", "sessions", id+".meta.json")
}

// sessionMeta carries the Session fields that do not fit in the
// per-turn JSONL stream. Persisted as a sidecar alongside the JSONL so
// older readers (which ignore it) keep working.
type sessionMeta struct {
	ID          string  `json:"id"`
	Goal        string  `json:"goal"`
	Label       string  `json:"label,omitempty"`
	ContextMark int     `json:"context_mark,omitempty"`
	AutoGate    bool    `json:"auto_gate,omitempty"`
	BudgetUSD   float64 `json:"budget_usd,omitempty"`
}

func persist(root string, s Session) error {
	p := sessionPath(root, s.ID)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, t := range s.Turns {
		if err := enc.Encode(t); err != nil {
			return err
		}
	}
	meta := sessionMeta{
		ID:          s.ID,
		Goal:        string(s.Goal),
		Label:       s.Label,
		ContextMark: s.ContextMark,
		AutoGate:    s.AutoGate,
		BudgetUSD:   s.BudgetUSD,
	}
	body, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionMetaPath(root, s.ID), body, 0o644)
}

func loadSessionMeta(root, id string) (sessionMeta, error) {
	body, err := os.ReadFile(sessionMetaPath(root, id))
	if err != nil {
		return sessionMeta{}, err
	}
	var m sessionMeta
	if err := json.Unmarshal(body, &m); err != nil {
		return sessionMeta{}, err
	}
	return m, nil
}
