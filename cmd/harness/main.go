// SPDX-License-Identifier: MIT

// Command harness is the HarnessX CLI entrypoint. Subcommand wiring lives
// in cmd_*.go files within this package; each file owns one Cobra group
// to keep main.go thin and reviewable.
package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/workflow"
	"github.com/ropeixoto/harnessx/internal/ui"
)

func main() {
	if err := newRoot().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRoot() *cobra.Command {
	var plain bool
	root := &cobra.Command{
		Use:           "harness",
		Short:         "HarnessX — adaptive runtime for agentic software engineering",
		SilenceUsage:  true,
		SilenceErrors: false,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			ui.SetPlain(plain || os.Getenv("HARNESS_PLAIN") != "")
		},
	}
	root.PersistentFlags().BoolVar(&plain, "plain", false, "disable ANSI styling")

	root.AddCommand(
		newVersionCmd(),
		newInitCmd(),
		newDoctorCmd(),
		newLogsCmd(),
		newProjectCmd(),
		newCatalogCmd(),
		newCleanupCmd(),
		newPaletteCmd(),
		newAutonomyCmd(),
		newHealthCmd(),
		newStackCmd(),
		newMCPCmd(),
		newHookCmd(),
		newAgentCmd(),
		newSensorCmd(),
		newCheckCmd(),
		newCICmd(),
		newContextCmd(),
		newAskCmd(),
		newPlanCmd(),
		newRunCmd(),
		newRunsCmd(),
		newExecuteCmd(),
		newMetricsCmd(),
		newAuditCmd(),
		newUpdateCmd(),
		newHelpCmd(),
		newFeatureCmd(),
		newBugfixCmd(),
		newReportCmd(),
		newDesignToProductCmd(),
		newDashboardCmd(),
		newOptimizeCmd(),
		newPerfSnapshotCmd(),
		newPerfCompareCmd(),
		newImageAuditCmd(),
		newDependencyAuditCmd(),
		newLogAuditCmd(),
		newSecurityAuditCmd(),
		newMemoryCmd(),
		newRoutesCmd(),
		newExplainCmd(),
		newSessionCmd(),
		newArtifactCmd(),
		newSkillCmd(),
		newSpecCmd(),
		newCompletionCmd(),
	)
	root.Args = cobra.ArbitraryArgs
	root.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		_, err = workflow.Run(cmd.Context(), workflow.Options{
			StartDir: cwd, Prompt: strings.Join(args, " "), BudgetUSD: 1.0,
		}, cmd.OutOrStdout())
		return err
	}
	return root
}
