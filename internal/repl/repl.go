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
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/agenthealth"
	"github.com/ropeixoto/harnessx/internal/intentplan"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
)

type Session struct {
	ID      string          `json:"id"`
	Goal    intentplan.Goal `json:"goal"`
	Started time.Time       `json:"started"`
	Turns   []Turn          `json:"turns"`
	Root    string          `json:"root"`
}

type Turn struct {
	Time   time.Time              `json:"time"`
	Input  string                 `json:"input"`
	Action string                 `json:"action"`
	Plan   *intentplan.Plan       `json:"plan,omitempty"`
	Result *intentplan.ExecResult `json:"result,omitempty"`
}

type Options struct {
	Root        string
	HarnessBin  string
	Goal        intentplan.Goal
	In          io.Reader
	Out         io.Writer
	Planner     Planner
	StepTimeout time.Duration
	HealthProbe *agenthealth.Probe
	Plain       bool
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
	greet(opts.Out, sess)
	rd := bufio.NewReader(opts.In)
	for {
		badge := ""
		if opts.HealthProbe != nil {
			badge = agenthealth.Badge(opts.HealthProbe.Snapshot(), opts.Plain)
		}
		fmt.Fprintf(opts.Out, "\n[%s%s]> ", sess.Goal, badge)
		input, err := readMultilineInput(rd, opts.Out, sess.Goal)
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
		turn := handleInput(ctx, &sess, opts, input)
		sess.Turns = append(sess.Turns, turn)
		if err := persist(opts.Root, sess); err != nil {
			fmt.Fprintf(opts.Out, "warn: persist: %v\n", err)
		}
	}
	return persist(opts.Root, sess)
}

func handleInput(ctx context.Context, sess *Session, opts Options, input string) Turn {
	turn := Turn{Time: time.Now().UTC(), Input: input}
	switch {
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
		plan, err := opts.Planner(ctx, sess.Goal, input)
		if err != nil {
			turn.Action = "plan-error"
			fmt.Fprintf(opts.Out, "planner: %v\n", err)
			return turn
		}
		turn.Plan = &plan
		res, err := intentplan.Execute(ctx, plan, intentplan.ExecOptions{
			HarnessBin: opts.HarnessBin, WorkingDir: opts.Root,
			Out: opts.Out, StepTimeout: opts.StepTimeout,
		})
		if err != nil {
			turn.Action = "execute-error"
			fmt.Fprintf(opts.Out, "executor: %v\n", err)
			return turn
		}
		turn.Result = &res
		turn.Action = "executed"
		if res.OK {
			fmt.Fprintln(opts.Out, "✓ plan green")
		} else {
			fmt.Fprintln(opts.Out, "✗ plan red — inspect step outputs above")
		}
	}
	return turn
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
	fmt.Fprintln(out, "type '/help' for commands, '/exit' to leave")
}

func printHelp(out io.Writer) {
	fmt.Fprintln(out, "commands:")
	fmt.Fprintln(out, "  /goal <dev|ads|research|ops>   switch session goal")
	fmt.Fprintln(out, "  /plan <prompt>                 emit plan JSON without executing")
	fmt.Fprintln(out, "  /last                          replay the previous prompt")
	fmt.Fprintln(out, "  /history                       list previous prompts in this session")
	fmt.Fprintln(out, "  /help                          this message")
	fmt.Fprintln(out, "  /exit | /quit                  leave the session")
	fmt.Fprintln(out, "  end a line with \\ to continue prompt on next line")
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
	return nil
}
