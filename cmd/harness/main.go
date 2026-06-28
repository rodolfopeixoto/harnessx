// SPDX-License-Identifier: MIT

// Command harness is the HarnessX CLI entrypoint. Subcommand wiring lives
// in cmd_*.go files within this package; each file owns one Cobra group
// to keep main.go thin and reviewable.
package main

import (
	"os"

	"github.com/spf13/cobra"

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
		newAuditSolidCmd(),
		newUpdateCmd(),
		newHelpCmd(),
		newInstallCmd(),
		newRuntimeCmd(),
		newContainersCmd(),
		newSecretCmd(),
		newImagesCmd(),
		newBackupCmd(),
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
		newListCmd(),
		newFlowCmd(),
		newScaffoldCmd(),
		newInstallGitHooksCmd(),
		newLoopCmd(),
		newDoCmd(),
		newRouteCmd(),
		newUninstallCmd(),
		newCompletionCmd(),
		newSmokeCmd(),
		newTestCmd(),
		newLintCmd(),
		newDevCmd(),
		newBenchCmd(),
		newProfileCmd(),
		newShipCmd(),
		newDriveCmd(),
		newAutoCmd(),
		newOnboardingCmd(),
		newAnalyticsCmd(),
		newNewCmd(),
		newOrchestrateCmd(),
		newEvolveCmd(),
		newConfigCmd(),
		newChatCmd(),
		newCoverageCmd(),
		newDiagnoseCmd(),
		newFixCmd(),
		newUseCmd(),
		newCostCompareCmd(),
		newFixEnvCmd(),
	)
	root.SuggestionsMinimumDistance = 2
	root.Args = cobra.NoArgs
	root.RunE = func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	}
	return root
}
