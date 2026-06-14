// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/cleanupcmd"
)

func newCleanupCmd() *cobra.Command {
	c := &cobra.Command{Use: "cleanup", Short: "Scan + apply cleanup policies (worktrees, caches, containers, leftovers)"}

	scan := &cobra.Command{
		Use:   "scan [root]",
		Short: "List cleanup candidates (report-only)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := rootFromArgs(args)
			if err != nil {
				return err
			}
			asJSON, _ := cmd.Flags().GetBool("json")
			return cleanupcmd.Scan(cmd.Context(), root, asJSON, cmd.OutOrStdout())
		},
	}
	scan.Flags().Bool("json", false, "emit JSON instead of table")

	apply := &cobra.Command{
		Use:   "apply [root]",
		Short: "Apply cleanup (requires policy match OR interactive y OR --yes)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := rootFromArgs(args)
			if err != nil {
				return err
			}
			policy, _ := cmd.Flags().GetString("policy")
			yes, _ := cmd.Flags().GetBool("yes")
			return cleanupcmd.Apply(cmd.Context(), root, policy, yes, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	apply.Flags().String("policy", "", "policy file path (default: <root>/.harness/cleanup/policy.yaml)")
	apply.Flags().Bool("yes", false, "auto-approve every finding (CI use; honours HARNESS_CLEANUP_I_UNDERSTAND=1)")

	policy := &cobra.Command{Use: "policy", Short: "Inspect or scaffold cleanup policy"}
	policyInit := &cobra.Command{
		Use:   "init [root]",
		Short: "Create a default policy file under .harness/cleanup/",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := rootFromArgs(args)
			if err != nil {
				return err
			}
			return cleanupcmd.PolicyInit(root, cmd.OutOrStdout())
		},
	}
	policy.AddCommand(policyInit)

	c.AddCommand(scan, apply, policy)
	return c
}

func rootFromArgs(args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	return cwd()
}
