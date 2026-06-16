// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newInstallGitHooksCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install-git-hooks",
		Short: "Install scripts/git/pre-push.sh as .git/hooks/pre-push",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			src := filepath.Join(dir, "scripts", "git", "pre-push.sh")
			if _, err := os.Stat(src); err != nil {
				return fmt.Errorf("source hook missing: %w", err)
			}
			dst := filepath.Join(dir, ".git", "hooks", "pre-push")
			if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
				return fmt.Errorf("not a git repo: %s", dir)
			}
			if err := writeSymlinkOrCopy(src, dst); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ installed %s\n", dst)
			return nil
		},
	}
}

func writeSymlinkOrCopy(src, dst string) error {
	_ = os.Remove(dst)
	if err := os.Symlink(src, dst); err == nil {
		return nil
	}
	body, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, body, 0o755)
}
