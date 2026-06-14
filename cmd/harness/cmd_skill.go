// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/skillcmd"
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
	c.AddCommand(promoteC)
	return c
}
