// SPDX-License-Identifier: MIT

package repl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/agenthealth"
	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/intentplan"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
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
