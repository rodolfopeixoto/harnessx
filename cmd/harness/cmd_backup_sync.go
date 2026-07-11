// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/backup"
)

func newBackupListCmd() *cobra.Command {
	var (
		remote  string
		jsonOut bool
	)
	c := &cobra.Command{
		Use:   "list",
		Short: "List snapshots stored on the remote",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			cfg, err := backup.LoadConfig(root)
			if err != nil {
				return err
			}
			pickedRemote, err := pickRemote(remote, cfg)
			if err != nil {
				return err
			}
			rc, err := backup.NewRclone()
			if err != nil {
				return err
			}
			items, err := rc.Ls(cmd.Context(), pickedRemote+":harness-backups/")
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return json.NewEncoder(out).Encode(items)
			}
			for _, name := range items {
				fmt.Fprintln(out, name)
			}
			return nil
		},
	}
	c.Flags().StringVar(&remote, "remote", "", "rclone remote name")
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	return c
}

func newBackupSyncCmd() *cobra.Command {
	var (
		remote string
		dryRun bool
	)
	c := &cobra.Command{
		Use:   "sync push|pull",
		Short: "Mirror .harness/config + specs to/from the remote",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			cfg, err := backup.LoadConfig(root)
			if err != nil {
				return err
			}
			pickedRemote, err := pickRemote(remote, cfg)
			if err != nil {
				return err
			}
			rc, err := backup.NewRclone()
			if err != nil {
				return err
			}
			local := filepath.Join(root, ".harness", "config")
			remotePath := pickedRemote + ":harness-sync/" + filepath.Base(root) + "/config"
			switch args[0] {
			case "push":
				return rc.Sync(cmd.Context(), local, remotePath, dryRun)
			case "pull":
				return rc.Sync(cmd.Context(), remotePath, local, dryRun)
			default:
				return fmt.Errorf("sync direction must be push or pull (got %q)", args[0])
			}
		},
	}
	c.Flags().StringVar(&remote, "remote", "", "rclone remote name")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "do not write")
	return c
}
