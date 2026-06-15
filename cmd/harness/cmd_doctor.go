// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/adapters/execprobe"
	"github.com/ropeixoto/harnessx/internal/app/doctor"
	"github.com/ropeixoto/harnessx/internal/install"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
	"github.com/ropeixoto/harnessx/internal/ui"
)

func newDoctorCmd() *cobra.Command {
	var (
		fix    bool
		dryRun bool
	)
	c := &cobra.Command{
		Use:   "doctor",
		Short: "Probe toolchain and agent CLIs",
		Long: `Probe every required + recommended binary; render a table grouped
by Tools / LSP / Quality / Agents. Exits non-zero when a required
binary is missing.

  --fix       run 'harness install <name>' for every missing tool that
              has a bundled manifest
  --dry-run   with --fix, print the install plan without executing`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			root, err := paths.FindProjectRoot(dir)
			if err != nil {
				return err
			}
			report := doctor.Run(cmd.Context(), execprobe.Default(),
				doctor.DefaultProbes(), doctor.DetectProject(root), 0)
			ui.RenderDoctor(cmd.OutOrStdout(), report)
			if fix {
				return runDoctorFix(cmd, report, dryRun)
			}
			if !report.AllRequiredPresent() {
				os.Exit(1)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&fix, "fix", false, "auto-install every missing tool with a bundled manifest")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "with --fix, print the install plan without executing")
	return c
}

func runDoctorFix(cmd *cobra.Command, report doctor.Report, dryRun bool) error {
	out := cmd.OutOrStdout()
	missing := collectFixableInstallIDs(report)
	if len(missing) == 0 {
		fmt.Fprintln(out, "→ nothing to fix")
		return nil
	}
	fmt.Fprintf(out, "→ fix plan: %d tool(s)\n", len(missing))
	registry := install.NewRegistry()
	for _, id := range missing {
		m, err := install.LoadBundled(id)
		if err != nil {
			fmt.Fprintf(out, "  ✗ %s: %v\n", id, err)
			continue
		}
		plan, err := registry.Pick(m)
		if err != nil {
			fmt.Fprintf(out, "  ✗ %s: %v\n", id, err)
			continue
		}
		if err := install.Execute(cmd.Context(), plan, dryRun, out, out); err != nil {
			fmt.Fprintf(out, "  ✗ %s: %v\n", id, err)
			continue
		}
	}
	return nil
}

func collectFixableInstallIDs(report doctor.Report) []string {
	seen := map[string]bool{}
	var ids []string
	add := func(entries []doctor.Entry) {
		for _, e := range entries {
			if e.Spec.InstallID == "" {
				continue
			}
			if e.Result.Present && e.Result.Err == nil {
				continue
			}
			if seen[e.Spec.InstallID] {
				continue
			}
			seen[e.Spec.InstallID] = true
			ids = append(ids, e.Spec.InstallID)
		}
	}
	add(report.Tools)
	add(report.LSPs)
	add(report.Quality)
	add(report.Agents)
	return ids
}
