// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/artifactcmd"
)

func newArtifactCmd() *cobra.Command {
	c := &cobra.Command{Use: "artifact", Short: "Artifact commands"}
	var kind string
	listC := &cobra.Command{
		Use:   "ls",
		Short: "List artifacts under .harness/artifacts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return artifactcmd.List(cmd.OutOrStdout(), artifactcmd.ListOptions{StartDir: dir, Kind: kind})
		},
	}
	listC.Flags().StringVar(&kind, "kind", "", "filter by kind (specs|plans|reports|sensors|perf)")
	catC := &cobra.Command{
		Use:   "cat <path>",
		Short: "Print an artifact (path relative to .harness/artifacts)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return artifactcmd.Cat(cmd.OutOrStdout(), artifactcmd.CatOptions{StartDir: dir, Path: args[0]})
		},
	}
	c.AddCommand(listC, catC)
	return c
}
