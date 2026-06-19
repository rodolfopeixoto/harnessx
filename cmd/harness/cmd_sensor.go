// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"

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

	var rootDir string
	runC := &cobra.Command{
		Use:   "run <id> [<id>...]",
		Short: "Run one or more sensors by ID",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := resolveSensorRoot(rootDir)
			if err != nil {
				return err
			}
			_, err = sensorcmd.Run(cmd.Context(), sensorcmd.RunOptions{
				StartDir: dir, IDs: args, FailOnError: false,
			}, cmd.OutOrStdout())
			return err
		},
	}
	runC.Flags().StringVar(&rootDir, "root", "", "project root (defaults to cwd)")

	c.AddCommand(listC, runC)
	return c
}

func resolveSensorRoot(rootDir string) (string, error) {
	if rootDir == "" {
		return cwd()
	}
	if filepath.IsAbs(rootDir) {
		return rootDir, nil
	}
	return filepath.Abs(rootDir)
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
	var fast bool
	c := &cobra.Command{
		Use:   "ci",
		Short: "Run every applicable sensor; exit non-zero on any failure",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			if _, err := sensorcmd.Run(cmd.Context(), sensorcmd.RunOptions{
				StartDir: dir, FailOnError: true, Fast: fast,
			}, cmd.OutOrStdout()); err != nil {
				os.Exit(1)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&fast, "fast", false, "skip slow sensors (today: secrets_scan)")
	return c
}
