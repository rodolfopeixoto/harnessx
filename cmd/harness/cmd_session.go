// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/sessioncmd"
)

func newSessionCmd() *cobra.Command {
	c := &cobra.Command{Use: "session", Short: "Session commands"}
	c.AddCommand(&cobra.Command{
		Use:   "show <id>",
		Short: "Show one session's runs + sensors + cost from sqlite",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return sessioncmd.Show(cmd.OutOrStdout(), sessioncmd.ShowOptions{StartDir: dir, ID: args[0]})
		},
	})
	return c
}
