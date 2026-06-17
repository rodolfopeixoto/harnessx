// SPDX-License-Identifier: MIT

// Package smokecmd exercises HarnessX CLI surface against a fresh
// project for every bundled language scaffold. It exists to catch
// regressions where a command works in the dev repo but fails for
// downstream users (the install-git-hooks bug that motivated F0).
package smokecmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/scaffoldpkg"
)

// CommandKind tags whether a step exercises HarnessX itself or external
// tooling. Failures in tool-class steps do not fail the matrix because
// they reflect host-machine readiness, not a HarnessX bug.
type CommandKind string

const (
	KindCLI  CommandKind = "cli"
	KindTool CommandKind = "tool"
)

// Step is one command run inside a stack's temporary project.
type Step struct {
	Name string
	Args []string
	Kind CommandKind
}

// StepResult captures the outcome of a single Step.
type StepResult struct {
	Name       string      `json:"name"`
	Kind       CommandKind `json:"kind"`
	ExitCode   int         `json:"exit_code"`
	DurationMs int64       `json:"duration_ms"`
	Stdout     string      `json:"stdout,omitempty"`
	Stderr     string      `json:"stderr,omitempty"`
	Skipped    bool        `json:"skipped,omitempty"`
}

// StackResult aggregates every step for one language stack.
type StackResult struct {
	Stack   string       `json:"stack"`
	WorkDir string       `json:"work_dir"`
	Steps   []StepResult `json:"steps"`
	OK      bool         `json:"ok"`
}

// MatrixResult is what Run returns.
type MatrixResult struct {
	HarnessBin string        `json:"harness_bin"`
	Stacks     []StackResult `json:"stacks"`
	OK         bool          `json:"ok"`
}

// Options controls the matrix run.
type Options struct {
	HarnessBin  string   // path to the harness binary; defaults to os.Executable()
	Langs       []string // empty = every bundled scaffold
	Keep        bool     // do not delete temp dirs after run
	StepTimeout time.Duration
}

// DefaultSteps returns the canonical sequence executed in every fresh
// project. Order matters: init must precede every cmd that reads
// .harness/.
func DefaultSteps(lang string) []Step {
	return []Step{
		{Name: "git init", Args: []string{"_external_git", "init", "-q"}, Kind: KindTool},
		{Name: "harness init", Args: []string{"init"}, Kind: KindCLI},
		{Name: "harness install-git-hooks", Args: []string{"install-git-hooks"}, Kind: KindCLI},
		{Name: "harness scaffold apply", Args: []string{"scaffold", "apply", lang, "--apply"}, Kind: KindCLI},
		{Name: "harness doctor", Args: []string{"doctor"}, Kind: KindCLI},
		{Name: "harness sensor list", Args: []string{"sensor", "list"}, Kind: KindCLI},
		{Name: "harness check", Args: []string{"check"}, Kind: KindCLI},
		{Name: "harness memory list", Args: []string{"memory", "list"}, Kind: KindCLI},
		{Name: "harness flow list", Args: []string{"flow", "list"}, Kind: KindCLI},
		{Name: "harness routes", Args: []string{"routes"}, Kind: KindCLI},
	}
}

// Run executes the matrix and returns the aggregate result. A non-nil
// error is reserved for setup failures (cannot find binary, cannot
// resolve scaffolds). Per-step failures live in the result tree.
func Run(ctx context.Context, opts Options, out io.Writer) (MatrixResult, error) {
	bin := opts.HarnessBin
	if bin == "" {
		exe, err := os.Executable()
		if err != nil {
			return MatrixResult{}, fmt.Errorf("resolve harness binary: %w", err)
		}
		bin = exe
	}
	if _, err := os.Stat(bin); err != nil {
		return MatrixResult{}, fmt.Errorf("harness binary not found at %s", bin)
	}
	langs := opts.Langs
	if len(langs) == 0 {
		l, err := scaffoldpkg.List()
		if err != nil {
			return MatrixResult{}, fmt.Errorf("list scaffolds: %w", err)
		}
		langs = l
	}
	sort.Strings(langs)
	timeout := opts.StepTimeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	res := MatrixResult{HarnessBin: bin, OK: true}
	for _, lang := range langs {
		sr := runStack(ctx, bin, lang, timeout, opts.Keep, out)
		res.Stacks = append(res.Stacks, sr)
		if !sr.OK {
			res.OK = false
		}
	}
	return res, nil
}

func runStack(ctx context.Context, bin, lang string, stepTimeout time.Duration, keep bool, out io.Writer) StackResult {
	dir, err := os.MkdirTemp("", "harness-smoke-"+lang+"-")
	if err != nil {
		return StackResult{Stack: lang, OK: false, Steps: []StepResult{{Name: "mkdir", ExitCode: -1, Stderr: err.Error()}}}
	}
	if !keep {
		defer os.RemoveAll(dir)
	}
	sr := StackResult{Stack: lang, WorkDir: dir, OK: true}
	fmt.Fprintf(out, "==> %s [%s]\n", lang, dir)
	for _, step := range DefaultSteps(lang) {
		r := runStep(ctx, bin, dir, step, stepTimeout)
		sr.Steps = append(sr.Steps, r)
		mark := "✓"
		if r.ExitCode != 0 {
			mark = "✗"
			if step.Kind == KindCLI {
				sr.OK = false
			}
		}
		fmt.Fprintf(out, "  %s %s (%dms, exit=%d)\n", mark, r.Name, r.DurationMs, r.ExitCode)
	}
	return sr
}

func runStep(ctx context.Context, bin, dir string, step Step, timeout time.Duration) StepResult {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var cmd *exec.Cmd
	if len(step.Args) > 0 && step.Args[0] == "_external_git" {
		cmd = exec.CommandContext(cctx, "git", step.Args[1:]...)
	} else {
		cmd = exec.CommandContext(cctx, bin, step.Args...)
	}
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HARNESS_PLAIN=1", "NO_COLOR=1")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	start := time.Now()
	err := cmd.Run()
	dur := time.Since(start).Milliseconds()
	r := StepResult{
		Name:       step.Name,
		Kind:       step.Kind,
		DurationMs: dur,
		Stdout:     truncate(stdout.String(), 4_000),
		Stderr:     truncate(stderr.String(), 4_000),
	}
	if err == nil {
		return r
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		r.ExitCode = ee.ExitCode()
		return r
	}
	r.ExitCode = -1
	if r.Stderr == "" {
		r.Stderr = err.Error()
	}
	return r
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...[truncated]"
}

// FormatTable renders a human-readable summary suitable for terminal
// output.
func FormatTable(res MatrixResult, w io.Writer) {
	fmt.Fprintf(w, "\nHarness smoke matrix — binary: %s\n\n", res.HarnessBin)
	for _, s := range res.Stacks {
		fmt.Fprintf(w, "%s: ", s.Stack)
		if s.OK {
			fmt.Fprint(w, "OK\n")
		} else {
			fmt.Fprint(w, "FAIL\n")
		}
		for _, st := range s.Steps {
			mark := "✓"
			if st.ExitCode != 0 {
				mark = "✗"
			}
			fmt.Fprintf(w, "  %s %-32s %6dms\n", mark, st.Name, st.DurationMs)
		}
	}
	fmt.Fprintln(w)
	if res.OK {
		fmt.Fprintln(w, "matrix: PASS")
	} else {
		fmt.Fprintln(w, "matrix: FAIL")
	}
}

// FormatJSON writes the full result as JSON for machine consumers.
func FormatJSON(res MatrixResult, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

// FailedSteps returns a flat list of (stack, step) pairs for CLI-kind
// failures. Useful for shaping a "next bug to fix" list.
func FailedSteps(res MatrixResult) []string {
	var out []string
	for _, s := range res.Stacks {
		for _, st := range s.Steps {
			if st.Kind != KindCLI || st.ExitCode == 0 {
				continue
			}
			out = append(out, filepath.Join(s.Stack, st.Name))
		}
	}
	return out
}
