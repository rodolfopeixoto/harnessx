// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/vcr"
	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/router"
	"github.com/ropeixoto/harnessx/internal/ui"
)

func newDriveCmd() *cobra.Command {
	var (
		slug           string
		autonomy       string
		maxAttempts    int
		skipCommit     bool
		featuresPath   string
		continueOnFail bool
		vcrDir         string
		vcrModeStr     string
	)
	c := &cobra.Command{
		Use:   "drive [<prompt>]",
		Short: "Spec → failing tests → impl → ci loop (paper §3.4 PEV)",
		Long: `Deterministic test-first cycle:

  1. harness feature  → writes .harness/artifacts/specs/<id>.md
  2. test-emit         → cheap LLM writes tests/test_<slug>.py
  3. harness test      → asserts the new tests fail (red bar)
  4. harness do        → implementation LLM fills them in
  5. harness ci        → gate
  6. conventional commit on green (unless --skip-commit)

Pass --features <file.md> to drive a backlog: one prompt per
non-empty, non-comment line (bullet "- " prefix optional).`,
		Args: cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			bin, err := os.Executable()
			if err != nil {
				return err
			}
			if featuresPath != "" {
				prompts, err := loadFeatureFile(featuresPath)
				if err != nil {
					return err
				}
				if len(prompts) == 0 {
					return fmt.Errorf("drive: --features %s yielded zero prompts", featuresPath)
				}
				return runDriveBatch(cmd.Context(), cmd.OutOrStdout(), driveOpts{
					root: dir, bin: bin, autonomy: autonomy,
					maxAttempts: maxAttempts, skipCommit: skipCommit,
					vcrDir: vcrDir, vcrMode: vcrModeStr,
				}, prompts, continueOnFail)
			}
			if len(args) == 0 {
				return fmt.Errorf("drive: pass a prompt or --features <file>")
			}
			prompt := strings.Join(args, " ")
			if slug == "" {
				slug = slugify(prompt)
			}
			return runDrive(cmd.Context(), cmd.OutOrStdout(), driveOpts{
				root: dir, bin: bin, prompt: prompt, slug: slug,
				autonomy: autonomy, maxAttempts: maxAttempts, skipCommit: skipCommit,
				vcrDir: vcrDir, vcrMode: vcrModeStr,
			})
		},
	}
	c.Flags().StringVar(&slug, "slug", "", "override the test file slug (default: derived from prompt)")
	c.Flags().StringVar(&autonomy, "autonomy", constants.DriveAutonomyDefault, "autonomy forwarded to harness do")
	c.Flags().IntVar(&maxAttempts, "max-attempts", constants.DriveDefaultMaxAttempt, "max impl→ci attempts")
	c.Flags().BoolVar(&skipCommit, "skip-commit", false, "do not commit on green")
	c.Flags().StringVar(&featuresPath, "features", "", "path to a markdown file; one prompt per line")
	c.Flags().BoolVar(&continueOnFail, "continue-on-fail", false, "with --features: keep going after a failed feature")
	c.Flags().StringVar(&vcrDir, "vcr", "", "wrap the test-emit adapter with VCR (record/replay) at <dir>")
	c.Flags().StringVar(&vcrModeStr, "vcr-mode", "auto", "vcr mode: auto|replay|record (only with --vcr)")
	return c
}

func parseVCRMode(s string) (vcr.Mode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "auto":
		return vcr.ModeAuto, nil
	case "replay":
		return vcr.ModeReplay, nil
	case "record":
		return vcr.ModeRecord, nil
	}
	return vcr.ModeAuto, fmt.Errorf("drive: unknown --vcr-mode %q (auto|replay|record)", s)
}

func wrapWithVCR(inner agents.AgentAdapter, dir, modeStr string) (agents.AgentAdapter, error) {
	if dir == "" {
		return inner, nil
	}
	mode, err := parseVCRMode(modeStr)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return vcr.New(vcr.Options{Inner: inner, Dir: dir, Mode: mode}), nil
}

func runDriveBatch(ctx context.Context, out io.Writer, base driveOpts, prompts []string, continueOnFail bool) error {
	fmt.Fprintln(out, ui.Heading.Render(fmt.Sprintf("drive: %d feature(s) queued", len(prompts))))
	var failed []string
	for i, p := range prompts {
		fmt.Fprintln(out, ui.Accent.Render(fmt.Sprintf("\n=== feature %d/%d ===", i+1, len(prompts))))
		opts := base
		opts.prompt = p
		opts.slug = slugify(p)
		if err := runDrive(ctx, out, opts); err != nil {
			fmt.Fprintln(out, "  "+ui.MarkFail()+" "+ui.Error.Render(err.Error()))
			failed = append(failed, p)
			if !continueOnFail {
				return fmt.Errorf("drive batch aborted on feature %d/%d: %w", i+1, len(prompts), err)
			}
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("drive batch: %d of %d features failed", len(failed), len(prompts))
	}
	fmt.Fprintln(out, ui.Success.Render(fmt.Sprintf("drive batch: all %d features green", len(prompts))))
	return nil
}

func loadFeatureFile(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("drive: read features %s: %w", path, err)
	}
	var prompts []string
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		trimmed = strings.TrimPrefix(trimmed, "- ")
		trimmed = strings.TrimPrefix(trimmed, "* ")
		if trimmed == "" {
			continue
		}
		prompts = append(prompts, trimmed)
	}
	return prompts, nil
}

type driveOpts struct {
	root        string
	bin         string
	prompt      string
	slug        string
	autonomy    string
	maxAttempts int
	skipCommit  bool
	vcrDir      string
	vcrMode     string
}

func runDrive(ctx context.Context, out io.Writer, opts driveOpts) error {
	fmt.Fprintf(out, "drive: %q (slug=%s)\n", opts.prompt, opts.slug)

	if err := driveSpec(ctx, out, opts); err != nil {
		return err
	}

	testPath, err := driveTestEmit(ctx, out, opts)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "drive: tests written at %s\n", testPath)

	if !driveExpectRedTests(ctx, out, opts) {
		fmt.Fprintln(out, "drive: warning — tests already green, nothing to implement")
		return nil
	}

	return driveImplLoop(ctx, out, opts)
}

func driveHeader(out io.Writer, step, title, hint string) {
	fmt.Fprintln(out, ui.Accent.Render("drive "+step)+"  "+
		ui.Heading.Render(title)+"  "+ui.Muted.Render(hint))
}

func driveSpec(ctx context.Context, out io.Writer, opts driveOpts) error {
	driveHeader(out, "1/5", "spec", "— harness feature")
	if err := runHarnessChild(ctx, opts.bin, opts.root, out,
		[]string{"feature", opts.prompt, "--yes", "--plan-only"}); err != nil {
		return fmt.Errorf("drive: spec step: %w", err)
	}
	return nil
}

func driveTestEmit(ctx context.Context, out io.Writer, opts driveOpts) (string, error) {
	driveHeader(out, "2/5", "test-emit", "— cheap chain writes failing tests")
	reg, _, err := agentcmd.LoadAll(opts.root)
	if err != nil {
		fmt.Fprintln(out, "  "+ui.MarkWarn()+" "+ui.Warn.Render("no adapter registry: "+err.Error()+" — falling back to placeholder"))
		return writePlaceholderTest(opts)
	}
	rtr := router.New(reg, router.Defaults(reg))
	dec, derr := rtr.Select(constants.DriveTaskTestEmit)
	if derr != nil || len(dec.Chain) == 0 {
		fmt.Fprintln(out, "  "+ui.MarkWarn()+" "+ui.Warn.Render("no cheap_review chain — placeholder test"))
		return writePlaceholderTest(opts)
	}
	adapter := dec.Chain[0]
	if wrapped, werr := wrapWithVCR(adapter, opts.vcrDir, opts.vcrMode); werr != nil {
		fmt.Fprintln(out, "  "+ui.MarkWarn()+" "+ui.Warn.Render("vcr disabled: "+werr.Error()))
	} else {
		adapter = wrapped
	}
	fmt.Fprintln(out, "  "+ui.Info.Render("routing through "+adapter.ID())+
		ui.Muted.Render(" ("+constants.DriveTaskTestEmit+")"))
	path, body, err := emitTestsViaAdapter(ctx, adapter, opts)
	if err != nil {
		fmt.Fprintln(out, "  "+ui.MarkWarn()+" "+ui.Warn.Render("adapter test-emit failed: "+err.Error()+" — placeholder"))
		return writePlaceholderTest(opts)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", err
	}
	fmt.Fprintln(out, "  "+ui.MarkSuccess()+" "+ui.Muted.Render(fmt.Sprintf("wrote %d bytes via %s", len(body), adapter.ID())))
	return path, nil
}

func emitTestsViaAdapter(ctx context.Context, adapter agents.AgentAdapter, opts driveOpts) (string, string, error) {
	path := filepath.Join(opts.root, "tests", constants.DriveTestFilePrefix+opts.slug+constants.DriveTestFileSuffix)
	prompt := renderTestEmitPrompt(opts.prompt, opts.slug, opts.root)
	timeout := 2 * time.Minute
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	res := adapter.Run(rctx, agents.AgentRequest{
		Prompt:     prompt,
		WorkingDir: opts.root,
		Timeout:    timeout,
		Extra:      map[string]string{"task": constants.DriveTaskTestEmit},
	})
	if res.Err != nil {
		return path, "", res.Err
	}
	body := extractPythonBody(res.Output.FinalMessage)
	if body == "" {
		body = extractPythonBody(string(res.Output.Stdout))
	}
	if body == "" {
		return path, "", fmt.Errorf("no python block found in adapter output")
	}
	if !strings.Contains(body, "def test_") {
		return path, "", fmt.Errorf("emitted file has no pytest test function")
	}
	return path, body, nil
}

func renderTestEmitPrompt(featurePrompt, slug, root string) string {
	return fmt.Sprintf(`You are HarnessX in spec-driven test-first mode.

A new feature is about to be implemented:

  %s

Write a pytest module to live at tests/test_drive_%s.py inside the
project at %s. The tests MUST fail today because the implementation
is not yet written. Cover the happy path plus one error case.

Constraints:
  - Output ONLY the Python file body inside a single
    triple-backtick code fence (with or without the "python" tag).
  - Do not include any prose outside the fence.
  - Do not edit any other file.
  - Use pytest function-style tests.
  - If the project uses FastAPI, assume a client fixture is already
    provided in tests/conftest.py (TestClient(app)).

Return only the file body.`, featurePrompt, slug, root)
}

func extractPythonBody(s string) string {
	if s == "" {
		return ""
	}
	for _, fence := range []string{"```python", "```py", "```"} {
		if idx := strings.Index(s, fence); idx >= 0 {
			rest := s[idx+len(fence):]
			if end := strings.Index(rest, "```"); end > 0 {
				body := strings.TrimSpace(rest[:end])
				body = strings.TrimPrefix(body, "\n")
				return body + "\n"
			}
		}
	}
	if strings.Contains(s, "def test_") {
		return strings.TrimSpace(s) + "\n"
	}
	return ""
}

func driveExpectRedTests(ctx context.Context, out io.Writer, opts driveOpts) bool {
	driveHeader(out, "3/5", "test", "— expecting red bar")
	if err := runHarnessChild(ctx, opts.bin, opts.root, out, []string{"test"}); err == nil {
		return false
	}
	fmt.Fprintln(out, "  "+ui.MarkSuccess()+" "+ui.Muted.Render("tests red as expected"))
	return true
}

func driveImplLoop(ctx context.Context, out io.Writer, opts driveOpts) error {
	preSnapshot := gitTreeSnapshot(ctx, opts.root)
	for attempt := 1; attempt <= opts.maxAttempts; attempt++ {
		driveHeader(out, "4/5", "impl",
			fmt.Sprintf("— harness do attempt %d/%d (implementation chain)", attempt, opts.maxAttempts))
		if err := runHarnessChild(ctx, opts.bin, opts.root, out,
			[]string{"do", opts.prompt, "--yes", "--autonomy", opts.autonomy}); err != nil {
			fmt.Fprintln(out, "  "+ui.MarkWarn()+" "+ui.Warn.Render(fmt.Sprintf("harness do failed (%v); retrying", err)))
			continue
		}
		if gitTreeSnapshot(ctx, opts.root) == preSnapshot {
			return fmt.Errorf("drive: agent produced no changes — prompt may be incomplete or ambiguous; refine and re-run /drive")
		}
		driveHeader(out, "5/5", "gate", "— harness ci")
		if err := runHarnessChild(ctx, opts.bin, opts.root, out, []string{"ci"}); err == nil {
			fmt.Fprintln(out, "  "+ui.MarkSuccess()+" "+ui.Success.Render("green"))
			if opts.skipCommit {
				return nil
			}
			return driveCommit(ctx, out, opts)
		}
		fmt.Fprintf(out, "drive: ci red on attempt %d; retrying\n", attempt)
	}
	return fmt.Errorf("drive: ci still red after %d attempts", opts.maxAttempts)
}

func gitTreeSnapshot(ctx context.Context, root string) string {
	c := exec.CommandContext(ctx, "git", "status", "--porcelain")
	c.Dir = root
	out, err := c.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func writePlaceholderTest(opts driveOpts) (string, error) {
	testDir := filepath.Join(opts.root, "tests")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(testDir, constants.DriveTestFilePrefix+opts.slug+constants.DriveTestFileSuffix)
	body := renderPlaceholderTest(opts.prompt, opts.slug)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func renderPlaceholderTest(prompt, slug string) string {
	return fmt.Sprintf(`# harness drive — failing placeholder test for: %s
# Replaced once the implementation lands and the gate passes.

import pytest


def test_drive_placeholder_for_%s() -> None:
    pytest.fail("harness drive scaffold — implementation pending")
`, prompt, sanitisePyIdent(slug))
}

func sanitisePyIdent(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + 32)
		default:
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" || (out[0] >= '0' && out[0] <= '9') {
		out = "x_" + out
	}
	return out
}

func driveCommit(ctx context.Context, out io.Writer, opts driveOpts) error {
	subject := conventionalDriveSubject(opts.prompt)
	body := fmt.Sprintf("Generated by `harness drive`.\n\nPrompt: %s", opts.prompt)
	steps := [][]string{
		{"add", "-A"},
		{"commit", "-m", subject, "-m", body},
	}
	for _, args := range steps {
		if err := runGitInDir(ctx, opts.root, args...); err != nil {
			return err
		}
	}
	fmt.Fprintln(out, "  "+ui.MarkSuccess()+" "+ui.Success.Render("committed ")+ui.Heading.Render(subject))
	return nil
}

func conventionalDriveSubject(prompt string) string {
	prefix := constants.DriveCommitTypeFeat + ": "
	return prefix + truncSubject(prompt, constants.DriveCommitSubjectMax-len(prefix))
}

func truncSubject(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func runGitInDir(ctx context.Context, dir string, args ...string) error {
	c := exec.CommandContext(ctx, "git", args...)
	c.Dir = dir
	var buf bytes.Buffer
	c.Stdout = &buf
	c.Stderr = &buf
	if err := c.Run(); err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, buf.String())
	}
	return nil
}

func runHarnessChild(ctx context.Context, bin, dir string, out io.Writer, args []string) error {
	c := exec.CommandContext(ctx, bin, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "HARNESS_PLAIN=1", "NO_COLOR=1")
	c.Stdout = out
	c.Stderr = out
	if err := c.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return fmt.Errorf("exit %d", ee.ExitCode())
		}
		return err
	}
	return nil
}
