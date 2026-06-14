// SPDX-License-Identifier: MIT

package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/adapters/execprobe"
	"github.com/ropeixoto/harnessx/internal/app/doctor"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
	"github.com/ropeixoto/harnessx/internal/ui"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Probe toolchain and agent CLIs",
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
			if !report.AllRequiredPresent() {
				os.Exit(1)
			}
			return nil
		},
	}
}
