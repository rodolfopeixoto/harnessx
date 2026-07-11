// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/backup"
)

func newBackupRemotesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remotes",
		Short: "List configured rclone remotes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rc, err := backup.NewRclone()
			if err != nil {
				return err
			}
			names, err := rc.Listremotes(cmd.Context())
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME")
			for _, n := range names {
				fmt.Fprintln(w, n)
			}
			return w.Flush()
		},
	}
}

func newBackupRemoteAddCmd() *cobra.Command {
	var (
		provider    string
		interactive bool
	)
	c := &cobra.Command{
		Use:   "remote add <name>",
		Short: "Add an rclone remote (--provider drive|s3|dropbox|onedrive|r2|webdav|crypt)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rc, err := backup.NewRclone()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if interactive {
				fmt.Fprintln(cmd.OutOrStdout(), "→ delegating to: rclone config create "+args[0]+" "+provider)
				return rc.ConfigInteractive(ctx, args[0], provider)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "→ creating remote %s (%s)\n", args[0], provider)
			return rc.ConfigCreate(ctx, args[0], provider)
		},
	}
	c.Flags().StringVar(&provider, "provider", "drive", "drive|s3|dropbox|onedrive|r2|webdav|crypt")
	c.Flags().BoolVar(&interactive, "interactive", false, "run rclone config interactively (TTY required)")
	return c
}
