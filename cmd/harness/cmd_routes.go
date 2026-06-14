// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/routescmd"
)

func newRoutesCmd() *cobra.Command {
	var task string
	c := &cobra.Command{
		Use:   "routes [task]",
		Short: "Show task → agent-chain resolution (bundled + user routes.yaml)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			t := task
			if len(args) == 1 && t == "" {
				t = args[0]
			}
			return routescmd.Run(cmd.OutOrStdout(), routescmd.Options{StartDir: dir, Task: t})
		},
	}
	c.Flags().StringVar(&task, "task", "", "show only this task (e.g. implementation)")
	return c
}
