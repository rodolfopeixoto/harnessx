// SPDX-License-Identifier: MIT

package intentplan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type Goal string

const (
	GoalDev      Goal = "dev"
	GoalAds      Goal = "ads"
	GoalResearch Goal = "research"
	GoalOps      Goal = "ops"
)

func KnownGoals() []Goal {
	return []Goal{GoalDev, GoalAds, GoalResearch, GoalOps}
}

func validGoal(g Goal) bool {
	for _, k := range KnownGoals() {
		if k == g {
			return true
		}
	}
	return false
}

type StepKind string

const (
	StepHarness StepKind = "harness"
	StepShell   StepKind = "shell"
	StepWait    StepKind = "wait"
)

type Step struct {
	Kind  StepKind `json:"kind"`
	Title string   `json:"title,omitempty"`
	Cmd   []string `json:"cmd"`
}

type ExitCriteria struct {
	AllPass []string `json:"all_pass,omitempty"`
}

type Plan struct {
	Goal      Goal         `json:"goal"`
	Intent    string       `json:"intent"`
	Steps     []Step       `json:"steps"`
	ExitWhen  ExitCriteria `json:"exit_when,omitempty"`
	Generated time.Time    `json:"generated_at,omitempty"`
}

func GoalPalette(g Goal) []string {
	switch g {
	case GoalDev:
		return []string{"plan", "ship", "test", "lint", "ci", "check", "scaffold", "smoke", "memory", "evolve", "coverage"}
	case GoalAds:
		return []string{"agent", "do", "run", "explain", "ask"}
	case GoalResearch:
		return []string{"context", "ask", "explain", "memory", "audit"}
	case GoalOps:
		return []string{"doctor", "runtime", "containers", "backup", "audit", "metrics"}
	}
	return nil
}

func (p Plan) Validate() error {
	if !validGoal(p.Goal) {
		return fmt.Errorf("intentplan: unknown goal %q (want %v)", p.Goal, KnownGoals())
	}
	if strings.TrimSpace(p.Intent) == "" {
		return errors.New("intentplan: missing intent")
	}
	if len(p.Steps) == 0 {
		return errors.New("intentplan: at least one step required")
	}
	palette := GoalPalette(p.Goal)
	for i, s := range p.Steps {
		if err := s.validate(palette); err != nil {
			return fmt.Errorf("step %d: %w", i, err)
		}
	}
	return nil
}

func (s Step) validate(palette []string) error {
	switch s.Kind {
	case StepHarness:
		if len(s.Cmd) == 0 {
			return errors.New("harness step needs cmd")
		}
		if len(palette) > 0 && !contains(palette, s.Cmd[0]) {
			return fmt.Errorf("cmd %q not in goal palette %v", s.Cmd[0], palette)
		}
	case StepShell:
		if len(s.Cmd) == 0 {
			return errors.New("shell step needs cmd")
		}
	case StepWait:
	default:
		return fmt.Errorf("unknown step kind %q", s.Kind)
	}
	return nil
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func ParseJSON(r io.Reader) (Plan, error) {
	var p Plan
	dec := json.NewDecoder(r)
	if err := dec.Decode(&p); err != nil {
		return Plan{}, fmt.Errorf("intentplan: decode: %w", err)
	}
	if err := p.Validate(); err != nil {
		return Plan{}, err
	}
	return p, nil
}

func ParseString(s string) (Plan, error) { return ParseJSON(strings.NewReader(s)) }

func (p Plan) MarshalPretty() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

type StepResult struct {
	Step       int    `json:"step"`
	Title      string `json:"title,omitempty"`
	Kind       string `json:"kind"`
	Command    string `json:"command"`
	DurationMs int64  `json:"duration_ms"`
	ExitCode   int    `json:"exit_code"`
	Stdout     string `json:"stdout,omitempty"`
	Stderr     string `json:"stderr,omitempty"`
}

type ExecResult struct {
	OK    bool         `json:"ok"`
	Steps []StepResult `json:"steps"`
}

type ExecOptions struct {
	HarnessBin  string
	WorkingDir  string
	Out         io.Writer
	StepTimeout time.Duration
}

func Execute(ctx context.Context, p Plan, opts ExecOptions) (ExecResult, error) {
	if opts.HarnessBin == "" {
		exe, err := os.Executable()
		if err != nil {
			return ExecResult{}, err
		}
		opts.HarnessBin = exe
	}
	if opts.WorkingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return ExecResult{}, err
		}
		opts.WorkingDir = wd
	}
	if opts.Out == nil {
		opts.Out = io.Discard
	}
	timeout := opts.StepTimeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	res := ExecResult{OK: true}
	for i, step := range p.Steps {
		sr := runStep(ctx, i, step, opts.HarnessBin, opts.WorkingDir, timeout)
		res.Steps = append(res.Steps, sr)
		fmt.Fprintf(opts.Out, "[%d] %s — %s (%dms, exit=%d)\n",
			i, step.Kind, sr.Command, sr.DurationMs, sr.ExitCode)
		if sr.ExitCode != 0 {
			res.OK = false
			break
		}
	}
	return res, nil
}

func runStep(ctx context.Context, idx int, step Step, bin, dir string, timeout time.Duration) StepResult {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var c *exec.Cmd
	var commandLabel string
	switch step.Kind {
	case StepHarness:
		c = exec.CommandContext(cctx, bin, step.Cmd...)
		commandLabel = "harness " + strings.Join(step.Cmd, " ")
	case StepShell:
		c = exec.CommandContext(cctx, "/bin/sh", "-c", strings.Join(step.Cmd, " "))
		commandLabel = strings.Join(step.Cmd, " ")
	case StepWait:
		time.Sleep(parseDuration(step.Cmd))
		return StepResult{Step: idx, Kind: string(step.Kind), Command: "wait", Title: step.Title}
	default:
		return StepResult{Step: idx, Kind: string(step.Kind), ExitCode: -1, Stderr: "unknown step kind"}
	}
	c.Dir = dir
	c.Env = append(os.Environ(), "HARNESS_PLAIN=1", "NO_COLOR=1")
	start := time.Now()
	out, err := c.CombinedOutput()
	sr := StepResult{
		Step:       idx,
		Title:      step.Title,
		Kind:       string(step.Kind),
		Command:    commandLabel,
		DurationMs: time.Since(start).Milliseconds(),
		Stdout:     truncate(string(out), 4_000),
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			sr.ExitCode = ee.ExitCode()
		} else {
			sr.ExitCode = -1
			sr.Stderr = err.Error()
		}
	}
	return sr
}

func parseDuration(args []string) time.Duration {
	if len(args) == 0 {
		return 0
	}
	d, err := time.ParseDuration(args[0])
	if err != nil {
		return 0
	}
	return d
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}

func AllowedHarnessCmds() []string {
	seen := map[string]bool{}
	for _, g := range KnownGoals() {
		for _, c := range GoalPalette(g) {
			seen[c] = true
		}
	}
	out := make([]string, 0, len(seen))
	for c := range seen {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}
