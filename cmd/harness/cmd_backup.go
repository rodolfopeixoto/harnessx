// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/backup"
)

const secretBackupEnv = "HARNESS_BACKUP_I_UNDERSTAND_SECRETS"

func newBackupCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "backup",
		Short: "Portable snapshot / restore / sync via rclone (drive, s3, dropbox, onedrive, r2, ...)",
		Long: `Wraps the rclone CLI to push and pull harness configuration
and run artifacts to any rclone-supported remote. Provider credentials
live in your rclone config; harness never touches them directly.

Default snapshot excludes secrets. To include them, pass
--include-secrets AND set HARNESS_BACKUP_I_UNDERSTAND_SECRETS=1.
Recommended: route the bucket through an rclone crypt overlay.`,
	}
	c.AddCommand(
		newBackupSnapshotCmd(), newBackupRestoreCmd(), newBackupListCmd(),
		newBackupSyncCmd(), newBackupRemotesCmd(), newBackupRemoteAddCmd(),
		newBackupConfigCmd(),
	)
	return c
}

func newBackupConfigCmd() *cobra.Command {
	c := &cobra.Command{Use: "config", Short: "Show or set backup config"}
	c.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Print the resolved .harness/config/backup.yaml",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			cfg, err := backup.LoadConfig(root)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "default_remote: %q\n", cfg.DefaultRemote)
			fmt.Fprintf(out, "compression:    %q\n", cfg.Compression)
			fmt.Fprintln(out, "include:")
			for _, i := range cfg.Include {
				fmt.Fprintln(out, "  -", i)
			}
			fmt.Fprintln(out, "exclude:")
			for _, e := range cfg.Exclude {
				fmt.Fprintln(out, "  -", e)
			}
			return nil
		},
	}, &cobra.Command{
		Use:   "set-default-remote <name>",
		Short: "Pin the default rclone remote without editing YAML",
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
			cfg.DefaultRemote = args[0]
			if err := backup.SaveConfig(root, cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "default_remote: %q\n", args[0])
			return nil
		},
	})
	return c
}

func newBackupSnapshotCmd() *cobra.Command {
	var (
		remote         string
		tag            string
		includeSecrets bool
		dryRun         bool
	)
	c := &cobra.Command{
		Use:   "snapshot [path]",
		Short: "Snapshot a project's .harness state to the configured remote",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				root = args[0]
			}
			cfg, err := backup.LoadConfig(root)
			if err != nil {
				return err
			}
			pickedRemote, err := pickRemote(remote, cfg)
			if err != nil {
				return err
			}
			if includeSecrets && os.Getenv(secretBackupEnv) != "1" {
				return fmt.Errorf("--include-secrets requires %s=1", secretBackupEnv)
			}
			rc, err := backup.NewRclone()
			if err != nil {
				return err
			}
			tmpDir, err := os.MkdirTemp("", "harness-backup-")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpDir)
			name := backup.PackName(tag)
			tarPath := filepath.Join(tmpDir, name)
			fmt.Fprintf(cmd.OutOrStdout(), "→ packing %s\n", tarPath)
			manifest, err := backup.Pack(root, cfg, tag, tarPath, includeSecrets)
			if err != nil {
				return err
			}
			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "→ dry-run: would copy %s -> %s:harness-backups/%s\n", tarPath, pickedRemote, name)
				return printManifest(cmd.OutOrStdout(), manifest)
			}
			dst := pickedRemote + ":harness-backups/"
			fmt.Fprintf(cmd.OutOrStdout(), "→ rclone copy → %s%s\n", dst, name)
			if err := rc.Copy(cmd.Context(), tarPath, dst); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "✓ snapshot uploaded")
			return printManifest(cmd.OutOrStdout(), manifest)
		},
	}
	c.Flags().StringVar(&remote, "remote", "", "rclone remote name (overrides default_remote)")
	c.Flags().StringVar(&tag, "tag", "", "label appended to the snapshot filename")
	c.Flags().BoolVar(&includeSecrets, "include-secrets", false, "include secrets.enc + secret-seed (DANGEROUS)")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "pack locally; skip upload")
	return c
}

func newBackupRestoreCmd() *cobra.Command {
	var (
		remote string
		target string
		force  bool
	)
	c := &cobra.Command{
		Use:   "restore <snapshot>",
		Short: "Download and unpack a snapshot",
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
			if target == "" {
				target = root
			}
			tmpDir, err := os.MkdirTemp("", "harness-restore-")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpDir)
			src := pickedRemote + ":harness-backups/" + args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "→ rclone copy ← %s\n", src)
			if err := rc.Copy(cmd.Context(), src, tmpDir); err != nil {
				return err
			}
			tarPath := filepath.Join(tmpDir, args[0])
			fmt.Fprintf(cmd.OutOrStdout(), "→ unpacking into %s\n", target)
			manifest, err := backup.Unpack(tarPath, target, force)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "✓ restore complete")
			return printManifest(cmd.OutOrStdout(), manifest)
		},
	}
	c.Flags().StringVar(&remote, "remote", "", "rclone remote name")
	c.Flags().StringVar(&target, "target", "", "where to unpack (defaults to current project)")
	c.Flags().BoolVar(&force, "force", false, "allow restoring into a non-empty directory")
	return c
}

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

func pickRemote(flag string, cfg backup.Config) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if cfg.DefaultRemote != "" {
		return cfg.DefaultRemote, nil
	}
	return "", fmt.Errorf("backup: no remote chosen\n  fix: harness backup remotes                              # list existing\n       harness backup remote add gdrive --provider drive --interactive\n       harness backup config set-default-remote gdrive\n  or pass --remote <name> per call")
}

func printManifest(out io.Writer, m backup.Manifest) error {
	fmt.Fprintf(out, "manifest: %d files, harness=%s, %s/%s\n", len(m.IncludedFiles), m.HarnessVersion, m.OS, m.Arch)
	return nil
}

var _ = strings.Builder{}
var _ = context.Background
