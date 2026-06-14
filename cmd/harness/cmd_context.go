// SPDX-License-Identifier: MIT

package main

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/contextcmd"
)

func newContextCmd() *cobra.Command {
	c := &cobra.Command{Use: "context", Short: "Context pack commands"}

	var force bool
	buildC := &cobra.Command{
		Use:   "build <task>",
		Short: "Build (or return cached) context pack for a task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = contextcmd.Build(cmd.Context(), contextcmd.BuildOptions{
				StartDir: dir, Task: strings.Join(args, " "), Force: force,
			}, cmd.OutOrStdout())
			return err
		},
	}
	buildC.Flags().BoolVar(&force, "force", false, "rebuild even if cache exists")

	var hash string
	inspectC := &cobra.Command{
		Use:   "inspect [hash]",
		Short: "Pretty-print a cached context pack (default: newest)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			h := hash
			if len(args) == 1 && h == "" {
				h = args[0]
			}
			return contextcmd.Inspect(contextcmd.InspectOptions{StartDir: dir, Hash: h}, cmd.OutOrStdout())
		},
	}
	inspectC.Flags().StringVar(&hash, "hash", "", "context pack hash (default: newest)")

	c.AddCommand(buildC, inspectC)
	return c
}
