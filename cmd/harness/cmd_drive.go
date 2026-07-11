// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/vcr"
	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/execution"
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
	logDriveImplChain(out, opts.root)
	for attempt := 1; attempt <= opts.maxAttempts; attempt++ {
		driveHeader(out, "4/5", "impl",
			fmt.Sprintf("— harness do attempt %d/%d (implementation chain)", attempt, opts.maxAttempts))
		doArgs := []string{"do", opts.prompt, "--yes", "--autonomy", opts.autonomy}
		if pin := os.Getenv("HARNESS_DRIVE_IMPL_ADAPTER"); pin != "" {
			doArgs = append(doArgs, "--agent", pin)
			fmt.Fprintln(out, "  "+ui.Info.Render("HARNESS_DRIVE_IMPL_ADAPTER="+pin+" pinning implementation adapter"))
		}
		if err := runHarnessChild(ctx, opts.bin, opts.root, out, doArgs); err != nil {
			fmt.Fprintln(out, "  "+ui.MarkWarn()+" "+ui.Warn.Render(fmt.Sprintf("harness do failed (%v); retrying", err)))
			continue
		}
		if gitTreeSnapshot(ctx, opts.root) == preSnapshot {
			if isLatestRunWaitingApproval(opts.root) {
				fmt.Fprintln(out, "  "+ui.MarkSuccess()+" "+ui.Success.Render("agent produced a diff (waiting_approval); deferring apply per autonomy gate"))
				if opts.skipCommit {
					return nil
				}
				return driveCommit(ctx, out, opts)
			}
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

func logDriveImplChain(out io.Writer, root string) {
	reg, _, err := agentcmd.LoadAll(root)
	if err != nil || reg == nil {
		return
	}
	rtr := router.New(reg, router.Defaults(reg))
	dec, derr := rtr.Select(constants.DriveTaskImplementation)
	if derr != nil || len(dec.Chain) == 0 {
		return
	}
	ids := make([]string, 0, len(dec.Chain))
	for _, a := range dec.Chain {
		ids = append(ids, a.ID())
	}
	fmt.Fprintln(out, "  "+ui.Info.Render("implementation chain (from config): "+strings.Join(ids, " → ")))
}

func isLatestRunWaitingApproval(root string) bool {
	runs, err := execution.ListRuns(root)
	if err != nil || len(runs) == 0 {
		return false
	}
	return runs[0].Status == execution.StatusWaitingApproval && len(runs[0].ChangedFiles) > 0
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
