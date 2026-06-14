// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/indexcmd"
)

func newProjectCmd() *cobra.Command {
	c := &cobra.Command{Use: "project", Short: "Project index commands"}

	var force bool
	indexC := &cobra.Command{
		Use:   "index",
		Short: "Build or refresh .harness/project/*.json maps",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = indexcmd.RunIndex(cmd.Context(), indexcmd.IndexOptions{StartDir: dir, Force: force}, cmd.OutOrStdout())
			return err
		},
	}
	indexC.Flags().BoolVar(&force, "force", false, "rebuild every map even when inputs are unchanged")

	var mapName string
	inspectC := &cobra.Command{
		Use:   "inspect [map]",
		Short: "List project maps or pretty-print one",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			name := mapName
			if len(args) == 1 && name == "" {
				name = args[0]
			}
			return indexcmd.RunInspect(indexcmd.InspectOptions{StartDir: dir, Map: name}, cmd.OutOrStdout())
		},
	}
	inspectC.Flags().StringVar(&mapName, "map", "", "map name (e.g. profile, commands, dependencies)")

	c.AddCommand(indexC, inspectC)
	return c
}
