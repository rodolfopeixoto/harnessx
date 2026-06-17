// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/initcmd"
	"github.com/ropeixoto/harnessx/internal/projectcfg"
	"github.com/ropeixoto/harnessx/internal/scaffoldpkg"
	"github.com/ropeixoto/harnessx/internal/scm"
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
	if !opts.yes {
		if opts.stack == "" {
			s, err := promptChoice(opts.stdin, out, "stack", langs)
			if err != nil {
				return err
			}
			opts.stack = s
		}
		if opts.target == "" {
			s, err := promptString(opts.stdin, out, "target dir", "./"+opts.stack+"-app")
			if err != nil {
				return err
			}
			opts.target = s
		}
	}
	if opts.stack == "" || opts.target == "" {
		return errors.New("new: --stack and --target required with --yes")
	}
	if !contains(langs, opts.stack) {
		return fmt.Errorf("new: unknown stack %q (have %v)", opts.stack, langs)
	}
	abs, err := filepath.Abs(opts.target)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return err
	}
	fmt.Fprintf(out, "new: target %s, stack %s\n", abs, opts.stack)

	if !scm.HasRepo(abs) {
		if err := scm.Init(ctx, abs, opts.gitBranch); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
		fmt.Fprintf(out, "new: git initialised on %s\n", opts.gitBranch)
	}

	if _, err := initcmd.Run(ctx, initcmd.Options{StartDir: abs}, out); err != nil {
		return err
	}

	m, err := scaffoldpkg.Load(opts.stack)
	if err != nil {
		return err
	}
	name := opts.name
	if name == "" {
		name = filepath.Base(abs)
	}
	res, err := scaffoldpkg.Apply(m, scaffoldpkg.ApplyOptions{Root: abs, Name: name})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "new: scaffold %s applied — %d files\n", opts.stack, len(res.Created))
	cfg := projectcfg.FromMeta(m.Language, map[string]string{
		"lint": m.LintCommand,
		"test": m.TestCommand,
		"run":  m.RunCommand,
		"dev":  m.RunCommand,
	})
	if err := projectcfg.Save(abs, cfg); err != nil {
		fmt.Fprintf(out, "new: warning project.yaml: %v\n", err)
	}

	if opts.withHooks {
		if _, err := InstallPrePushHook(abs, false); err != nil {
			fmt.Fprintf(out, "new: hook skipped (%v)\n", err)
		} else {
			fmt.Fprintf(out, "new: pre-push hook installed\n")
		}
	}

	fmt.Fprintf(out, "\n✓ project ready at %s\n", abs)
	fmt.Fprintf(out, "  cd %s\n  harness lint && harness test && harness dev\n", opts.target)
	fmt.Fprintf(out, "  harness ship \"<your first feature>\"\n")
	return nil
}

func promptChoice(in io.Reader, out io.Writer, label string, options []string) (string, error) {
	fmt.Fprintf(out, "%s? (%s)\n> ", label, strings.Join(options, "|"))
	r := bufio.NewReader(in)
	line, err := r.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	line = strings.TrimSpace(line)
	if !contains(options, line) {
		return "", fmt.Errorf("invalid choice %q", line)
	}
	return line, nil
}

func promptString(in io.Reader, out io.Writer, label, fallback string) (string, error) {
	fmt.Fprintf(out, "%s? [%s]\n> ", label, fallback)
	r := bufio.NewReader(in)
	line, err := r.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return fallback, nil
	}
	return line, nil
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
