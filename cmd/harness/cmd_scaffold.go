// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/ropeixoto/harnessx/internal/projectcfg"
	"github.com/ropeixoto/harnessx/internal/scaffoldpkg"
	"github.com/ropeixoto/harnessx/internal/scm"
)

func newScaffoldCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "scaffold",
		Short: "Drop a deterministic language scaffold (no LLM call)",
	}
	c.AddCommand(scaffoldListCmd(), scaffoldShowCmd(), scaffoldApplyCmd())
	return c
}

func scaffoldListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List bundled language scaffolds",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			langs, err := scaffoldpkg.List()
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "%-10s %-50s %s\n", "LANG", "DESCRIPTION", "TESTED AGAINST")
			for _, l := range langs {
				m, err := scaffoldpkg.Load(l)
				if err != nil {
					continue
				}
				fmt.Fprintf(out, "%-10s %-50s %s\n", l, truncate(m.Description, 50), m.TestedAgainst)
			}
			return nil
		},
	}
}

func scaffoldShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <lang>",
		Short: "Dump the scaffold.yaml for a language",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := scaffoldpkg.Load(args[0])
			if err != nil {
				return err
			}
			enc := yaml.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent(2)
			return enc.Encode(m)
		},
	}
}

func scaffoldApplyCmd() *cobra.Command {
	var (
		name      string
		apply     bool
		withGit   bool
		withDeps  bool
		force     bool
		gitBranch string
	)
	c := &cobra.Command{
		Use:   "apply <lang>",
		Short: "Apply a bundled scaffold to the current directory",
		Long: `Default is dry-run (prints files that would be written). Pass --apply
to actually write them. Use --with-git to 'git init' on a fresh
project and --with-deps to run the scaffold's post_steps (venv +
pip install, go mod tidy, npm install, bundle install, cargo build).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScaffoldApply(cmd.Context(), cmd.OutOrStdout(), args[0], scaffoldRunOpts{
				name:      name,
				apply:     apply,
				withGit:   withGit,
				withDeps:  withDeps,
				force:     force,
				gitBranch: gitBranch,
			})
		},
	}
	c.Flags().StringVar(&name, "name", "", "project name (default: dirname)")
	c.Flags().BoolVar(&apply, "apply", false, "write files (default is dry-run)")
	c.Flags().BoolVar(&withGit, "with-git", false, "git init when .git/ is absent")
	c.Flags().BoolVar(&withDeps, "with-deps", false, "run scaffold post_steps (deps install)")
	c.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	c.Flags().StringVar(&gitBranch, "git-branch", "main", "branch name for --with-git")
	return c
}

type scaffoldRunOpts struct {
	name      string
	apply     bool
	withGit   bool
	withDeps  bool
	force     bool
	gitBranch string
}

func runScaffoldApply(ctx context.Context, out io.Writer, lang string, opts scaffoldRunOpts) error {
	root, err := cwd()
	if err != nil {
		return err
	}
	if opts.name == "" {
		opts.name = filepath.Base(root)
	}
	m, err := scaffoldpkg.Load(lang)
	if err != nil {
		return err
	}
	if err := ensureGitForScaffold(ctx, root, opts, out); err != nil {
		return err
	}
	res, err := scaffoldpkg.Apply(m, scaffoldpkg.ApplyOptions{
		Root: root, Name: opts.name, Force: opts.force, Dry: !opts.apply,
	})
	if err != nil {
		return err
	}
	printScaffoldSummary(out, lang, opts.apply, res)
	if opts.withDeps && opts.apply {
		runPostSteps(ctx, out, root, m)
	}
	if opts.apply {
		writeProjectCfgFromMeta(root, m, out)
	}
	fmt.Fprintf(out, "\nnext:\n  harness lint\n  harness test\n  harness dev\n")
	return nil
}

func ensureGitForScaffold(ctx context.Context, root string, opts scaffoldRunOpts, out io.Writer) error {
	if !opts.withGit || !opts.apply {
		return nil
	}
	if scm.HasRepo(root) {
		return nil
	}
	if err := scm.Init(ctx, root, opts.gitBranch); err != nil {
		return err
	}
	fmt.Fprintf(out, "git: initialised on %s\n", opts.gitBranch)
	return nil
}

func printScaffoldSummary(out io.Writer, lang string, apply bool, res scaffoldpkg.ApplyResult) {
	mode := "dry-run"
	if apply {
		mode = "applied"
	}
	fmt.Fprintf(out, "scaffold: %s (%s) — %d files\n", lang, mode, len(res.Created))
	for _, p := range res.Created {
		fmt.Fprintf(out, "  + %s\n", p)
	}
	for _, p := range res.Skipped {
		fmt.Fprintf(out, "  · skip %s (exists; pass --force to overwrite)\n", p)
	}
}

func writeProjectCfgFromMeta(root string, m scaffoldpkg.Meta, out io.Writer) {
	cfg := projectcfg.FromMeta(m.Language, map[string]string{
		"lint": m.LintCommand,
		"test": m.TestCommand,
		"run":  m.RunCommand,
		"dev":  m.RunCommand,
	})
	if err := projectcfg.Save(root, cfg); err != nil {
		fmt.Fprintf(out, "warning: could not write project.yaml: %v\n", err)
	}
}

func runPostSteps(ctx context.Context, out io.Writer, root string, m scaffoldpkg.Meta) {
	for _, step := range m.PostSteps {
		fmt.Fprintf(out, "post: %s — %v\n", step.Name, step.Cmd)
		cmd := exec.CommandContext(ctx, step.Cmd[0], step.Cmd[1:]...)
		cmd.Dir = root
		cmd.Stdout = out
		cmd.Stderr = out
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(out, "  ✗ %s failed: %v (continuing)\n", step.Name, err)
		}
	}
}

var _ = os.Stat

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
