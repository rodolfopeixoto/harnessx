// SPDX-License-Identifier: MIT

// Package devloop wraps a workflow run in a deterministic
// verify-and-retry loop: agent runs ⇒ lint+test ⇒ on failure,
// canonicalised error is fed back to the agent as a follow-up prompt.
// Bounded by --max-attempts and --budget-usd.
package devloop

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/app/workflow"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
	"github.com/ropeixoto/harnessx/internal/scaffoldpkg"
)

type Options struct {
	StartDir    string
	Prompt      string
	AgentID     string
	Autonomy    string
	BudgetUSD   float64
	MaxAttempts int
	LintCmd     string
	TestCmd     string
	Apply       bool
}

type Attempt struct {
	Number      int
	WorkflowRes workflow.Result
	LintOK      bool
	LintOutput  string
	TestOK      bool
	TestOutput  string
	Elapsed     time.Duration
}

type Result struct {
	Attempts []Attempt
	Passed   bool
	Reason   string
}

// Run kicks off the loop. The first attempt uses opts.Prompt; later
// attempts prepend a canonicalised failure block from the previous
// attempt.
func Run(ctx context.Context, opts Options, out io.Writer) (Result, error) {
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 3
	}
	if opts.MaxAttempts > 10 {
		opts.MaxAttempts = 10
	}
	if err := resolveCommands(opts.StartDir, &opts); err != nil {
		return Result{}, err
	}
	var res Result
	prompt := opts.Prompt
	for i := 1; i <= opts.MaxAttempts; i++ {
		fmt.Fprintf(out, "\n=== loop attempt %d/%d ===\n", i, opts.MaxAttempts)
		start := time.Now()
		wfRes, err := workflow.Feature(ctx, workflow.Options{
			StartDir:  opts.StartDir,
			Prompt:    prompt,
			AgentID:   opts.AgentID,
			Apply:     opts.Apply,
			Execute:   true,
			AutoYes:   true,
			BudgetUSD: opts.BudgetUSD,
			Autonomy:  opts.Autonomy,
		}, out)
		if err != nil {
			return res, err
		}
		att := Attempt{Number: i, WorkflowRes: wfRes}
		att.LintOK, att.LintOutput = runShell(ctx, opts.StartDir, opts.LintCmd)
		att.TestOK, att.TestOutput = runShell(ctx, opts.StartDir, opts.TestCmd)
		att.Elapsed = time.Since(start)
		res.Attempts = append(res.Attempts, att)
		fmt.Fprintf(out, "  lint: %s  test: %s  elapsed %s\n", okOrFail(att.LintOK), okOrFail(att.TestOK), att.Elapsed.Round(time.Millisecond))
		if att.LintOK && att.TestOK {
			res.Passed = true
			res.Reason = "lint + test green"
			break
		}
		prompt = Canonicalise(opts.Prompt, att)
	}
	if !res.Passed {
		res.Reason = fmt.Sprintf("exhausted %d attempts", opts.MaxAttempts)
	}
	if path, err := writeReport(opts.StartDir, res); err == nil {
		fmt.Fprintf(out, "\nloop report: %s\n", path)
	}
	return res, nil
}

func resolveCommands(root string, opts *Options) error {
	if opts.LintCmd != "" && opts.TestCmd != "" {
		return nil
	}
	if cmd := detectScaffoldMeta(root); cmd != nil {
		if opts.LintCmd == "" {
			opts.LintCmd = cmd.LintCommand
		}
		if opts.TestCmd == "" {
			opts.TestCmd = cmd.TestCommand
		}
	}
	if opts.LintCmd == "" && opts.TestCmd == "" {
		return fmt.Errorf("devloop: no lint/test command (pass --lint and --test or scaffold a project first)")
	}
	return nil
}

// detectScaffoldMeta loads scaffold.yaml hints by sniffing top-level
// files. Best-effort; nil if no match.
func detectScaffoldMeta(root string) *scaffoldpkg.Meta {
	mapping := map[string]string{
		"requirements.txt": "python",
		"Cargo.toml":       "rust",
		"Gemfile":          "ruby",
		"package.json":     "react",
		"go.mod":           "go",
	}
	for marker, lang := range mapping {
		if exists(filepath.Join(root, marker)) {
			m, err := scaffoldpkg.Load(lang)
			if err == nil {
				return &m
			}
		}
	}
	return nil
}

func runShell(ctx context.Context, root, cmdline string) (bool, string) {
	if cmdline == "" {
		return true, ""
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdline)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	return err == nil, string(out)
}

func okOrFail(b bool) string {
	if b {
		return "✓"
	}
	return "✗"
}

// Canonicalise builds a follow-up prompt that prepends the previous
// attempt's lint/test failures in a structured block the LLM can
// reliably parse.
func Canonicalise(original string, a Attempt) string {
	var b strings.Builder
	b.WriteString("# Previous attempt failed verification.\n\n")
	fmt.Fprintf(&b, "Attempt #%d ran for %s.\n\n", a.Number, a.Elapsed.Round(time.Millisecond))
	if !a.LintOK {
		b.WriteString("## Lint failure\n\n```\n")
		b.WriteString(trimToLines(a.LintOutput, 80))
		b.WriteString("\n```\n\n")
	}
	if !a.TestOK {
		b.WriteString("## Test failure\n\n```\n")
		b.WriteString(trimToLines(a.TestOutput, 80))
		b.WriteString("\n```\n\n")
	}
	b.WriteString("# Original request\n\n")
	b.WriteString(original)
	b.WriteString("\n\nFix the lint/test failures above. Do not regress unrelated tests.")
	return b.String()
}

func trimToLines(s string, max int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= max {
		return s
	}
	return strings.Join(lines[len(lines)-max:], "\n")
}

func writeReport(root string, res Result) (string, error) {
	dir := filepath.Join(paths.HarnessDir(root), "runs", "_loop")
	if err := mkdirAll(dir); err != nil {
		return "", err
	}
	path := filepath.Join(dir, fmt.Sprintf("loop-%s.md", time.Now().UTC().Format("20060102-150405")))
	var b strings.Builder
	b.WriteString("# devloop report\n\n")
	fmt.Fprintf(&b, "passed: %v\nreason: %s\nattempts: %d\n\n", res.Passed, res.Reason, len(res.Attempts))
	for _, a := range res.Attempts {
		fmt.Fprintf(&b, "## attempt %d (elapsed %s)\n\n", a.Number, a.Elapsed.Round(time.Millisecond))
		fmt.Fprintf(&b, "- workflow run: %s status=%s\n", a.WorkflowRes.ExecutionRunID, a.WorkflowRes.ExecutionStatus)
		fmt.Fprintf(&b, "- lint: %s\n- test: %s\n\n", okOrFail(a.LintOK), okOrFail(a.TestOK))
	}
	return path, writeFile(path, []byte(b.String()))
}

func exists(p string) bool {
	_, err := statFn(p)
	return err == nil
}
