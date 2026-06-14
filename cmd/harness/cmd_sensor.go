// SPDX-License-Identifier: MIT

package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/sensorcmd"
)

func newSensorCmd() *cobra.Command {
	c := &cobra.Command{Use: "sensor", Short: "Sensor commands"}

	listC := &cobra.Command{
		Use:   "list",
		Short: "List sensors applicable to the current project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return sensorcmd.List(cmd.OutOrStdout(), dir)
		},
	}

	runC := &cobra.Command{
		Use:   "run <id> [<id>...]",
		Short: "Run one or more sensors by ID",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = sensorcmd.Run(cmd.Context(), sensorcmd.RunOptions{
				StartDir: dir, IDs: args, FailOnError: false,
			}, cmd.OutOrStdout())
			return err
		},
	}

	c.AddCommand(listC, runC)
	return c
}

func newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Run every applicable sensor",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = sensorcmd.Run(cmd.Context(), sensorcmd.RunOptions{
				StartDir: dir, FailOnError: false,
			}, cmd.OutOrStdout())
			return err
		},
	}
}

func newCICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ci",
		Short: "Run every applicable sensor; exit non-zero on any failure",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			if _, err := sensorcmd.Run(cmd.Context(), sensorcmd.RunOptions{
				StartDir: dir, FailOnError: true,
			}, cmd.OutOrStdout()); err != nil {
				os.Exit(1)
			}
			return nil
		},
	}
}
