// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os/exec"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/installcmd"
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
	c.AddCommand(newInstallListCmd(), newInstallShowCmd(), newInstallLSPCmd(), newInstallToolsCmd())
	return c
}

func newInstallToolsCmd() *cobra.Command {
	var dryRun bool
	c := &cobra.Command{
		Use:   "tools <stack>",
		Short: "Install the recommended lint/format/test tools for a stack",
		Long: `Resolve the tool pack for the given stack and install each missing
binary in turn. Supported stacks: go, python, node, react, ruby, rust,
java, kotlin, swift, elixir, php, dotnet, dart, flutter.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return installcmd.Install(cmd.Context(), cmd.OutOrStdout(), installcmd.InstallOptions{
				Stack:  args[0],
				DryRun: dryRun,
				Probe:  isInstalled,
			})
		},
	}
	c.Flags().BoolVar(&dryRun, "dry-run", false, "print commands without executing")
	return c
}

func toolsForStack(stack string) []string {
	return installcmd.ToolsForStack(stack)
}

// newInstallLSPCmd narrows install down to LSP servers. Without an arg
// it prints the table of every bundled LSP (installed marker + install
// hint). With an arg, it forwards to the regular install pipeline.
func newInstallLSPCmd() *cobra.Command {
	var (
		all     bool
		dryRun  bool
		upgrade bool
	)
	c := &cobra.Command{
		Use:   "lsp [name]",
		Short: "List or install Language Server Protocol servers",
		Long: `Examples:
  harness install lsp                # show every LSP and its install state
  harness install lsp gopls          # install one
  harness install lsp --all          # install everything missing
  harness install lsp --all --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if all {
				return installAllMissingLSPs(cmd, dryRun, upgrade)
			}
			if len(args) == 0 {
				return printLSPTable(cmd)
			}
			return runInstall(cmd, args[0], dryRun, upgrade)
		},
	}
	c.Flags().BoolVar(&all, "all", false, "install every LSP that is not present yet")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "print commands without executing")
	c.Flags().BoolVar(&upgrade, "upgrade", false, "reinstall even when probe already passes")
	return c
}

func lspManifests() ([]install.Manifest, error) {
	names, err := install.ListBundled()
	if err != nil {
		return nil, err
	}
	var out []install.Manifest
	for _, n := range names {
		m, err := install.LoadBundled(n)
		if err != nil || m.Category != "lsp" {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func printLSPTable(cmd *cobra.Command) error {
	manifests, err := lspManifests()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tINSTALLED\tINSTALL\tDESCRIPTION")
	for _, m := range manifests {
		state := "—"
		if isInstalled(m.Probe.Binary) {
			state = "✓"
		}
		fmt.Fprintf(w, "%s\t%s\tharness install %s\t%s\n", m.Name, state, m.Name, m.Description)
	}
	return w.Flush()
}

func installAllMissingLSPs(cmd *cobra.Command, dryRun, upgrade bool) error {
	manifests, err := lspManifests()
	if err != nil {
		return err
	}
	for _, m := range manifests {
		if !upgrade && isInstalled(m.Probe.Binary) {
			fmt.Fprintf(cmd.OutOrStdout(), "[skip] %s already installed\n", m.Name)
			continue
		}
		if err := runInstall(cmd, m.Name, dryRun, upgrade); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[fail] %s: %v\n", m.Name, err)
		}
	}
	return nil
}

func runInstall(cmd *cobra.Command, name string, dryRun, upgrade bool) error {
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
