// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os/exec"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/install"
)

func newInstallCmd() *cobra.Command {
	var (
		dryRun  bool
		upgrade bool
	)
	c := &cobra.Command{
		Use:   "install <tool>",
		Short: "Install a system tool, LSP server, or agent CLI from a bundled manifest",
		Long: `Resolve the manifest for <tool> and run the first viable install
strategy on the host (brew, apt, dnf, pacman, go install, npm -g,
cargo install, pip --user). Examples:

  harness install gopls            # go install golang.org/x/tools/gopls@latest
  harness install ripgrep          # brew install ripgrep (mac) | apt-get (linux)
  harness install --dry-run gemini
  harness install --upgrade gitleaks
  harness install list             # all bundled manifests
  harness install show gopls       # resolved plan`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			out := cmd.OutOrStdout()
			m, err := install.LoadBundled(name)
			if err != nil {
				return err
			}
			if !upgrade && isInstalled(m.Probe.Binary) {
				fmt.Fprintf(out, "%s already installed (pass --upgrade to reinstall)\n", name)
				return nil
			}
			plan, err := install.NewRegistry().Pick(m)
			if err != nil {
				return err
			}
			return install.Execute(cmd.Context(), plan, dryRun, out, out)
		},
	}
	c.Flags().BoolVar(&dryRun, "dry-run", false, "print command without executing")
	c.Flags().BoolVar(&upgrade, "upgrade", false, "reinstall even when probe already passes")
	c.AddCommand(newInstallListCmd(), newInstallShowCmd())
	return c
}

func newInstallListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List bundled install manifests",
		RunE: func(cmd *cobra.Command, _ []string) error {
			names, err := install.ListBundled()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCATEGORY\tINSTALLED\tDESCRIPTION")
			for _, n := range names {
				m, err := install.LoadBundled(n)
				if err != nil {
					continue
				}
				installed := "—"
				if isInstalled(m.Probe.Binary) {
					installed = "✓"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Name, m.Category, installed, m.Description)
			}
			return w.Flush()
		},
	}
}

func newInstallShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <tool>",
		Short: "Show the resolved install plan without executing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := install.LoadBundled(args[0])
			if err != nil {
				return err
			}
			plan, err := install.NewRegistry().Pick(m)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), plan.String())
			return nil
		},
	}
}

func isInstalled(binary string) bool {
	if binary == "" {
		return false
	}
	_, err := exec.LookPath(binary)
	return err == nil
}
