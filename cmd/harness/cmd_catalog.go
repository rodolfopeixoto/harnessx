// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/catalogcmd"
	"github.com/ropeixoto/harnessx/internal/domain"
)

func newCatalogCmd() *cobra.Command {
	c := &cobra.Command{Use: "catalog", Short: "Capabilities (agents · MCPs · hooks · sensors · skills · context · resources · plugins)"}

	list := &cobra.Command{
		Use:   "list",
		Short: "List discovered + installed capabilities",
		RunE: func(cmd *cobra.Command, _ []string) error {
			kind, _ := cmd.Flags().GetString("kind")
			root, err := cwd()
			if err != nil {
				return err
			}
			return catalogcmd.List(cmd.Context(), root, domain.CapabilityKind(kind), cmd.OutOrStdout())
		},
	}
	list.Flags().String("kind", "", "filter by kind (agent|mcp|hook|sensor|skill|context|resource|plugin)")

	show := &cobra.Command{
		Use:   "show <kind> <name>",
		Short: "Show one capability",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			return catalogcmd.Show(cmd.Context(), root, domain.CapabilityKind(args[0]), args[1], cmd.OutOrStdout())
		},
	}

	plan := &cobra.Command{
		Use:   "plan <kind> <name>",
		Short: "Render the diff that install would apply",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			return catalogcmd.Plan(cmd.Context(), root, domain.CapabilityKind(args[0]), args[1], cmd.OutOrStdout())
		},
	}

	install := &cobra.Command{
		Use:   "install <kind> <name>",
		Short: "Install a capability (requires approval unless --yes)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, _ := cmd.Flags().GetBool("yes")
			dry, _ := cmd.Flags().GetBool("dry-run")
			root, err := cwd()
			if err != nil {
				return err
			}
			return catalogcmd.Install(cmd.Context(), root, domain.CapabilityKind(args[0]), args[1], yes, dry, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	install.Flags().Bool("yes", false, "skip interactive approval")
	install.Flags().Bool("dry-run", false, "print the plan without writing")

	remove := &cobra.Command{
		Use:   "remove <kind> <name>",
		Short: "Remove an installed capability config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			return catalogcmd.Remove(cmd.Context(), root, domain.CapabilityKind(args[0]), args[1], cmd.OutOrStdout())
		},
	}

	c.AddCommand(list, show, plan, install, remove)
	return c
}
