// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/initcmd"
	"github.com/ropeixoto/harnessx/internal/app/projectcmd"
	"github.com/ropeixoto/harnessx/internal/scm"
)

func newInitCmd() *cobra.Command {
	var (
		force     bool
		withGit   bool
		all       bool
		slug      string
		gitBranch string
	)
	c := &cobra.Command{
		Use:   "init",
		Short: "Initialise .harness/ in the project root",
		Long: `Bootstraps .harness/ with config, db, hooks, and gitignore.

With --git, also runs 'git init -b <branch>' when no .git/ is present.
With --all, runs --git AND registers the project in the cross-project
workspace registry (equivalent to running 'harness project add . --slug
<basename>').`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if all {
				withGit = true
			}
			if withGit {
				if err := bootstrapGit(cmd.Context(), out, dir, gitBranch); err != nil {
					return err
				}
			}
			if _, err := initcmd.Run(cmd.Context(), initcmd.Options{StartDir: dir, Force: force}, out); err != nil {
				return err
			}
			if scm.HasRepo(dir) {
				if path, err := InstallPrePushHook(dir, false); err == nil {
					fmt.Fprintf(out, "git: pre-push hook installed at %s\n", path)
				} else {
					fmt.Fprintf(out, "git: pre-push hook skipped (%v)\n", err)
				}
			}
			if all {
				if slug == "" {
					slug = filepath.Base(dir)
				}
				if err := projectcmd.Add(cmd.Context(), projectcmd.Options{}, dir, "", slug, out); err != nil {
					return err
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&force, "force", false, "overwrite existing config")
	c.Flags().BoolVar(&withGit, "git", false, "git init when .git/ is absent")
	c.Flags().BoolVar(&all, "all", false, "implies --git + 'project add' with derived slug")
	c.Flags().StringVar(&gitBranch, "git-branch", "main", "branch name when --git creates the repo")
	c.Flags().StringVar(&slug, "slug", "", "slug used when --all registers the project")
	return c
}

func bootstrapGit(ctx context.Context, out io.Writer, root, branch string) error {
	if scm.HasRepo(root) {
		fmt.Fprintf(out, "git: .git already present, skipping\n")
		return nil
	}
	if err := scm.Init(ctx, root, branch); err != nil {
		return err
	}
	fmt.Fprintf(out, "git: initialised on %s\n", branch)
	return nil
}
