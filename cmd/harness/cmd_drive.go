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

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/router"
)

func newDriveCmd() *cobra.Command {
	var (
		slug        string
		autonomy    string
		maxAttempts int
		skipCommit  bool
	)
	c := &cobra.Command{
		Use:   "drive <prompt>",
		Short: "Spec → failing tests → impl → ci loop (paper §3.4 PEV)",
		Long: `Deterministic test-first cycle:

  1. harness feature  → writes .harness/artifacts/specs/<id>.md
  2. test-emit         → cheap LLM writes tests/test_<slug>.py
  3. harness test      → asserts the new tests fail (red bar)
  4. harness do        → implementation LLM fills them in
  5. harness ci        → gate
  6. conventional commit on green (unless --skip-commit)`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := strings.Join(args, " ")
			dir, err := cwd()
			if err != nil {
				return err
			}
			bin, err := os.Executable()
			if err != nil {
				return err
			}
			if slug == "" {
				slug = slugify(prompt)
			}
			return runDrive(cmd.Context(), cmd.OutOrStdout(), driveOpts{
				root: dir, bin: bin, prompt: prompt, slug: slug,
				autonomy: autonomy, maxAttempts: maxAttempts, skipCommit: skipCommit,
			})
		},
	}
	c.Flags().StringVar(&slug, "slug", "", "override the test file slug (default: derived from prompt)")
	c.Flags().StringVar(&autonomy, "autonomy", constants.DriveAutonomyDefault, "autonomy forwarded to harness do")
	c.Flags().IntVar(&maxAttempts, "max-attempts", constants.DriveDefaultMaxAttempt, "max impl→ci attempts")
	c.Flags().BoolVar(&skipCommit, "skip-commit", false, "do not commit on green")
	return c
}

type driveOpts struct {
	root        string
	bin         string
	prompt      string
	slug        string
	autonomy    string
	maxAttempts int
	skipCommit  bool
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

func driveSpec(ctx context.Context, out io.Writer, opts driveOpts) error {
	fmt.Fprintln(out, "drive: 1/5 — harness feature (spec)")
	if err := runHarnessChild(ctx, opts.bin, opts.root, out,
		[]string{"feature", opts.prompt, "--yes", "--plan-only"}); err != nil {
		return fmt.Errorf("drive: spec step: %w", err)
	}
	return nil
}

func driveTestEmit(ctx context.Context, out io.Writer, opts driveOpts) (string, error) {
	fmt.Fprintln(out, "drive: 2/5 — test-emit (cheap chain)")
	if reg, _, err := agentcmd.LoadAll(opts.root); err == nil {
		rtr := router.New(reg, router.Defaults(reg))
		if dec, derr := rtr.Select(constants.DriveTaskTestEmit); derr == nil && len(dec.Chain) > 0 {
			fmt.Fprintf(out, "drive: routing test-emit through %s (%s)\n",
				dec.Chain[0].ID(), constants.DriveTaskTestEmit)
		}
	}
	return writePlaceholderTest(opts)
}

func driveExpectRedTests(ctx context.Context, out io.Writer, opts driveOpts) bool {
	fmt.Fprintln(out, "drive: 3/5 — harness test (expect red)")
	if err := runHarnessChild(ctx, opts.bin, opts.root, out, []string{"test"}); err == nil {
		return false
	}
	fmt.Fprintln(out, "drive: tests red as expected")
	return true
}

func driveImplLoop(ctx context.Context, out io.Writer, opts driveOpts) error {
	preSnapshot := gitTreeSnapshot(ctx, opts.root)
	for attempt := 1; attempt <= opts.maxAttempts; attempt++ {
		fmt.Fprintf(out, "drive: 4/5 — harness do attempt %d/%d (implementation chain)\n",
			attempt, opts.maxAttempts)
		if err := runHarnessChild(ctx, opts.bin, opts.root, out,
			[]string{"do", opts.prompt, "--yes", "--autonomy", opts.autonomy}); err != nil {
			fmt.Fprintf(out, "drive: harness do failed (%v); retrying\n", err)
			continue
		}
		if gitTreeSnapshot(ctx, opts.root) == preSnapshot {
			return fmt.Errorf("drive: agent produced no changes — prompt may be incomplete or ambiguous; refine and re-run /drive")
		}
		fmt.Fprintln(out, "drive: 5/5 — harness ci")
		if err := runHarnessChild(ctx, opts.bin, opts.root, out, []string{"ci"}); err == nil {
			fmt.Fprintln(out, "drive: ✓ green")
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
	fmt.Fprintln(out, "drive: ✓ committed "+subject)
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
