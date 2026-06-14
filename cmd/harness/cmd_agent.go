// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
)

func newAgentCmd() *cobra.Command {
	c := &cobra.Command{Use: "agent", Short: "Agent adapter commands"}

	listC := &cobra.Command{
		Use:   "list",
		Short: "List registered agent adapters",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return agentcmd.List(cmd.OutOrStdout(), dir)
		},
	}

	addC := &cobra.Command{
		Use:   "add <id>",
		Short: "Copy a bundled adapter YAML into .harness/config/agents/",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return agentcmd.Add(cmd.OutOrStdout(), dir, args[0])
		},
	}

	discoverC := &cobra.Command{
		Use:   "discover <binary>",
		Short: "Print a YAML scaffold for a CLI binary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return agentcmd.Discover(cmd.OutOrStdout(), args[0])
		},
	}

	var skipRun bool
	certifyC := &cobra.Command{
		Use:   "certify <id>",
		Short: "Run the certification suite against an adapter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			_, err = agentcmd.Certify(cmd.Context(), cmd.OutOrStdout(), agentcmd.CertifyOptions{
				ID: args[0], StartDir: dir, SkipRun: skipRun,
			})
			return err
		},
	}
	certifyC.Flags().BoolVar(&skipRun, "skip-run", false, "skip checks that execute the CLI binary")

	c.AddCommand(listC, addC, discoverC, certifyC)
	return c
}
