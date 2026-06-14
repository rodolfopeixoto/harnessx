// SPDX-License-Identifier: MIT

package main

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/logsvc"
	"github.com/ropeixoto/harnessx/internal/app/watchcmd"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

func newLogsCmd() *cobra.Command {
	var tail int
	var follow bool
	c := &cobra.Command{
		Use:   "logs",
		Short: "Print (or follow) recent JSONL events from .harness/logs/events.jsonl",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			root, err := paths.FindProjectRoot(dir)
			if err != nil {
				return err
			}
			cfg, err := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
			if err != nil {
				return err
			}
			logPath := config.Resolve(root, cfg.Logging.Path)
			if follow {
				return watchcmd.Run(cmd.Context(), watchcmd.Options{Path: logPath, Tail: tail}, cmd.OutOrStdout())
			}
			return logsvc.Print(logsvc.Options{Path: logPath, Tail: tail}, cmd.OutOrStdout())
		},
	}
	c.Flags().IntVar(&tail, "tail", 50, "number of trailing entries to show (0 = all)")
	c.Flags().BoolVarP(&follow, "follow", "f", false, "stream new entries (TUI; q to quit)")
	return c
}
