// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

func newUninstallCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove HarnessX state (project, global, or everything)",
	}
	c.AddCommand(uninstallProjectCmd(), uninstallGlobalCmd(), uninstallAllCmd())
	return c
}

func uninstallProjectCmd() *cobra.Command {
	var yes bool
	c := &cobra.Command{
		Use:   "project",
		Short: "Delete .harness/ in the current directory",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			target := filepath.Join(dir, ".harness")
			return wipePath(cmd.OutOrStdout(), target, yes)
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "skip confirmation")
	return c
}

func uninstallGlobalCmd() *cobra.Command {
	var yes bool
	c := &cobra.Command{
		Use:   "global",
		Short: "Delete the cross-project registry + cached state under the global HarnessX home",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return wipePath(cmd.OutOrStdout(), paths.GlobalHarnessDir(), yes)
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "skip confirmation")
	return c
}

func uninstallAllCmd() *cobra.Command {
	var yes bool
	c := &cobra.Command{
		Use:   "all",
		Short: "Wipe project .harness, global state, the binary, and the brew install",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			dir, err := cwd()
			if err != nil {
				return err
			}
			if err := wipePath(out, filepath.Join(dir, ".harness"), yes); err != nil {
				return err
			}
			if err := wipePath(out, paths.GlobalHarnessDir(), yes); err != nil {
				return err
			}
			if err := wipeBinary(out, yes); err != nil {
				return err
			}
			fmt.Fprintln(out, "✓ harness fully removed. shell session may still cache the old binary; open a new terminal.")
			return nil
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "skip confirmation prompts")
	return c
}

func wipePath(out io.Writer, target string, yes bool) error {
	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(out, "· %s (already absent)\n", target)
			return nil
		}
		return err
	}
	if !yes && !confirm(out, fmt.Sprintf("delete %s (%s)?", target, sizeHint(info))) {
		fmt.Fprintln(out, "· cancelled")
		return nil
	}
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	fmt.Fprintf(out, "✓ removed %s\n", target)
	return nil
}

func wipeBinary(out io.Writer, yes bool) error {
	if path, err := exec.LookPath("harness"); err == nil {
		fmt.Fprintf(out, "· found binary at %s\n", path)
		if !yes && !confirm(out, fmt.Sprintf("delete %s?", path)) {
			fmt.Fprintln(out, "· skipped binary removal")
			return nil
		}
		if err := os.Remove(path); err != nil {
			fmt.Fprintf(out, "  → could not remove (likely needs sudo): %v\n", err)
			fmt.Fprintf(out, "  → manual: sudo rm %s\n", path)
		} else {
			fmt.Fprintf(out, "✓ removed %s\n", path)
		}
	}
	if _, err := exec.LookPath("brew"); err == nil {
		fmt.Fprintln(out, "· brew detected — run: brew uninstall harness && brew untap rodolfopeixoto/harnessx")
	}
	return nil
}

func confirm(out io.Writer, prompt string) bool {
	fmt.Fprintf(out, "%s [y/N] ", prompt)
	var raw string
	_, _ = fmt.Scanln(&raw)
	return raw == "y" || raw == "Y" || raw == "yes"
}

func sizeHint(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}
	return fmt.Sprintf("%d bytes", info.Size())
}
