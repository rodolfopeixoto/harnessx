// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/initcmd"
	"github.com/ropeixoto/harnessx/internal/scaffoldpkg"
	"github.com/ropeixoto/harnessx/internal/scm"
	"github.com/ropeixoto/harnessx/internal/ui"
)

type newOptions struct {
	stack     string
	target    string
	name      string
	withDeps  bool
	withHooks bool
	yes       bool
	gitBranch string
	stdin     io.Reader
}

func newNewCmd() *cobra.Command {
	opts := newOptions{
		gitBranch: "main",
		withHooks: true,
	}
	c := &cobra.Command{
		Use:   "new [stack] [path]",
		Short: "Bootstrap a new project: git init + harness init + scaffold + hooks",
		Long: `Single-command project bootstrap. Equivalent to:

  mkdir <path> && cd <path>
  git init -b <git-branch>
  harness init
  harness scaffold apply <stack> --apply [--with-deps]
  harness install-git-hooks

Without --yes, prompts for stack and path. With --yes, requires both
positional arguments (or --stack / --target).`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) >= 1 {
				opts.stack = args[0]
			}
			if len(args) >= 2 {
				opts.target = args[1]
			}
			opts.stdin = cmd.InOrStdin()
			return runNewProject(cmd.Context(), cmd.OutOrStdout(), opts)
		},
	}
	c.Flags().StringVar(&opts.stack, "stack", "", "scaffold to apply (go|python|rails|react|ruby|rust)")
	c.Flags().StringVar(&opts.target, "target", "", "destination directory")
	c.Flags().StringVar(&opts.name, "name", "", "project name (default: dirname)")
	c.Flags().BoolVar(&opts.withDeps, "with-deps", false, "run scaffold post_steps (deps install)")
	c.Flags().BoolVar(&opts.withHooks, "with-hooks", true, "install git pre-push hook")
	c.Flags().BoolVar(&opts.yes, "yes", false, "non-interactive (requires stack + target)")
	c.Flags().StringVar(&opts.gitBranch, "git-branch", "main", "initial git branch")
	return c
}

func runNewProject(ctx context.Context, out io.Writer, opts newOptions) error {
	langs, err := scaffoldpkg.List()
	if err != nil {
		return err
	}
	if err := resolveNewInputs(&opts, langs, out); err != nil {
		return err
	}
	abs, err := prepareNewTarget(ctx, &opts, out)
	if err != nil {
		return err
	}
	if err := applyNewScaffold(abs, &opts, out); err != nil {
		return err
	}
	installNewHooks(abs, opts, out)
	commitScaffoldBaseline(ctx, abs, out)
	fmt.Fprintf(out, "\n%s %s\n", ui.MarkSuccess(), ui.Accent.Render("project ready at "+abs))
	fmt.Fprintf(out, "  %s cd %s\n", ui.Muted.Render("→"), opts.target)
	fmt.Fprintf(out, "  %s harness lint %s harness test %s harness dev\n", ui.Muted.Render("→"), ui.Muted.Render("&&"), ui.Muted.Render("&&"))
	fmt.Fprintf(out, "  %s harness ship \"<your first feature>\"\n", ui.Muted.Render("→"))
	return nil
}

//nolint:gocognit // interactive wizard branches one screen per option
func resolveNewInputs(opts *newOptions, langs []string, out io.Writer) error {
	if !opts.yes {
		if opts.stack == "" {
			s, err := promptChoice(opts.stdin, out, "stack", langs)
			if err != nil {
				return err
			}
			opts.stack = s
		}
		if opts.target == "" {
			cwd, _ := os.Getwd()
			fallback := "./" + opts.stack + "-app"
			fmt.Fprintf(out, "\ncurrent dir: %s\n", cwd)
			fmt.Fprintln(out, "tip: use an absolute path (/Users/...) or relative (./, ../) — the project will land at <resolved> below.")
			for attempt := 0; attempt < 3; attempt++ {
				s, err := promptString(opts.stdin, out, "target dir", fallback)
				if err != nil {
					return err
				}
				abs, _ := filepath.Abs(s)
				fmt.Fprintf(out, "  → resolves to %s\n", abs)
				confirm, err := promptString(opts.stdin, out, "ok? (yes / new path)", "yes")
				if err != nil {
					return err
				}
				if strings.EqualFold(strings.TrimSpace(confirm), "yes") || confirm == "" {
					opts.target = s
					break
				}
				fallback = strings.TrimSpace(confirm)
			}
			if opts.target == "" {
				return fmt.Errorf("new: target dir not confirmed after 3 attempts")
			}
		}
	}
	if opts.stack == "" || opts.target == "" {
		return errors.New("new: --stack and --target required with --yes")
	}
	if !contains(langs, opts.stack) {
		return fmt.Errorf("new: unknown stack %q (have %v)", opts.stack, langs)
	}
	return nil
}

func prepareNewTarget(ctx context.Context, opts *newOptions, out io.Writer) (string, error) {
	abs, err := filepath.Abs(opts.target)
	if err != nil {
		return "", err
	}
	if err := guardNestedTarget(abs); err != nil {
		return "", err
	}
	if err := guardNewTarget(abs); err != nil {
		return "", err
	}
	fmt.Fprintf(out, "new: creating at %s (stack %s)\n", abs, opts.stack)
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", err
	}
	if !scm.HasRepo(abs) {
		if err := scm.Init(ctx, abs, opts.gitBranch); err != nil {
			return "", fmt.Errorf("git init: %w", err)
		}
		fmt.Fprintf(out, "new: git initialised on %s\n", opts.gitBranch)
	}
	if _, err := initcmd.Run(ctx, initcmd.Options{StartDir: abs}, out); err != nil {
		return "", err
	}
	return abs, nil
}

func guardNestedTarget(abs string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	cwdAbs, err := filepath.Abs(cwd)
	if err != nil {
		return nil
	}
	if filepath.Base(cwdAbs) != filepath.Base(abs) {
		return nil
	}
	rel, err := filepath.Rel(cwdAbs, abs)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return nil
	}
	return fmt.Errorf("new: refusing nested target %s — you are already inside a directory named %q. cd .. first, or pick a different path",
		abs, filepath.Base(abs))
}

func guardNewTarget(abs string) error {
	entries, err := os.ReadDir(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if name == ".git" || name == ".harness" {
			continue
		}
		return fmt.Errorf("new: target %s is not empty (found %q); pick a fresh path or rerun inside the existing project", abs, name)
	}
	return nil
}

func commitScaffoldBaseline(ctx context.Context, root string, out io.Writer) {
	for _, args := range [][]string{
		{"add", "-A"},
		{"-c", "user.email=harness@local", "-c", "user.name=harness new", "commit", "-q", "-m", "chore: scaffold baseline"},
	} {
		cmd := osexec.CommandContext(ctx, "git", args...)
		cmd.Dir = root
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(out, "new: baseline commit step failed (%v) — commit manually before harness ship\n", err)
			return
		}
	}
	fmt.Fprintln(out, "new: scaffold committed as baseline (chore: scaffold baseline)")
}

func installNewHooks(abs string, opts newOptions, out io.Writer) {
	if !opts.withHooks {
		return
	}
	if _, err := InstallPrePushHook(abs, false); err != nil {
		fmt.Fprintf(out, "new: hook skipped (%v)\n", err)
		return
	}
	fmt.Fprintf(out, "new: pre-push hook installed\n")
}
