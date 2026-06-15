// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/skillcmd"
	"github.com/ropeixoto/harnessx/internal/skillpkg"
)

func newSkillCmd() *cobra.Command {
	c := &cobra.Command{Use: "skill", Short: "Skill commands"}
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List versioned skills",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return skillcmd.List(cmd.OutOrStdout(), dir)
		},
	})
	var name, file string
	promoteC := &cobra.Command{
		Use:   "promote --name <skill> --file <md>",
		Short: "Promote a skill version (benchmark-gated)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return skillcmd.Promote(cmd.OutOrStdout(), skillcmd.PromoteOptions{
				StartDir: dir, Name: name, File: file,
			})
		},
	}
	promoteC.Flags().StringVar(&name, "name", "", "skill name")
	promoteC.Flags().StringVar(&file, "file", "", "markdown body")
	c.AddCommand(promoteC, newSkillTemplatesCmd(), newSkillInstallCmd())
	return c
}

func newSkillTemplatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates",
		Short: "List bundled skill snippets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			all, err := skillpkg.List()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tMODE\tDESCRIPTION")
			for _, t := range all {
				fmt.Fprintf(w, "%s\t%s\t%s\n", t.Name, t.Mode, t.Description)
			}
			return w.Flush()
		},
	}
}

func newSkillInstallCmd() *cobra.Command {
	var yes bool
	c := &cobra.Command{
		Use:   "install <name>",
		Short: "Drop a bundled skill snippet into .harness/skills/<name>.md",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			t, err := skillpkg.Load(args[0])
			if err != nil {
				return err
			}
			dir := filepath.Join(root, ".harness", "skills")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			target := filepath.Join(dir, args[0]+".md")
			if _, err := os.Stat(target); err == nil && !yes {
				return fmt.Errorf("%s already exists (pass --yes to overwrite)", target)
			}
			if err := os.WriteFile(target, []byte(t.Body), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n  mode: %s\n", target, t.Mode)
			return nil
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "overwrite if exists")
	return c
}
