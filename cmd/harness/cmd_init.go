// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/initcmd"
)

func newInitCmd() *cobra.Command {
	var force bool
	c := &cobra.Command{
		Use:   "init",
		Short: "Initialise .harness/ in the project root",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = initcmd.Run(cmd.Context(), initcmd.Options{StartDir: dir, Force: force}, cmd.OutOrStdout())
			return err
		},
	}
	c.Flags().BoolVar(&force, "force", false, "overwrite existing config")
	return c
}
