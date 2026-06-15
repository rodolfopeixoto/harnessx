// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/hookscan"
	"github.com/ropeixoto/harnessx/internal/mcpscan"
)

func newMCPCmd() *cobra.Command {
	c := &cobra.Command{Use: "mcp", Short: "MCP server discovery + listing"}
	c.AddCommand(mcpListCmd(), mcpScanCmd())
	return c
}

func newHookCmd() *cobra.Command {
	c := &cobra.Command{Use: "hook", Short: "Hook discovery + listing"}
	c.AddCommand(hookListCmd(), hookScanCmd())
	return c
}

func mcpScanCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "scan [root]",
		Short: "Deterministically scan filesystem for MCP server configs",
		Args:  cobra.MaximumNArgs(1),
		RunE:  mcpRun,
	}
	c.Flags().Bool("json", false, "emit JSON")
	return c
}

func mcpListCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "list [root]",
		Short: "Alias for `mcp scan`",
		Args:  cobra.MaximumNArgs(1),
		RunE:  mcpRun,
	}
	c.Flags().Bool("json", false, "emit JSON")
	return c
}

func hookScanCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "scan [root]",
		Short: "Deterministically scan filesystem for hooks",
		Args:  cobra.MaximumNArgs(1),
		RunE:  hookRun,
	}
	c.Flags().Bool("json", false, "emit JSON")
	return c
}

func hookListCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "list [root]",
		Short: "Alias for `hook scan`",
		Args:  cobra.MaximumNArgs(1),
		RunE:  hookRun,
	}
	c.Flags().Bool("json", false, "emit JSON")
	return c
}

func mcpRun(cmd *cobra.Command, args []string) error {
	root, err := rootFromArgs(args)
	if err != nil {
		return err
	}
	asJSON, _ := cmd.Flags().GetBool("json")
	servers, err := mcpscan.Scan(root)
	if err != nil {
		return err
	}
	if asJSON {
		return writeJSON(cmd, servers)
	}
	if len(servers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no MCP servers detected")
		return nil
	}
	tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SOURCE\tNAME\tTRANSPORT\tRISK\tPATH")
	for _, s := range servers {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", s.Source, s.Name, s.Transport, s.Risk, s.ConfigPath)
	}
	return tw.Flush()
}

func hookRun(cmd *cobra.Command, args []string) error {
	root, err := rootFromArgs(args)
	if err != nil {
		return err
	}
	asJSON, _ := cmd.Flags().GetBool("json")
	hooks, err := hookscan.Scan(root)
	if err != nil {
		return err
	}
	if asJSON {
		return writeJSON(cmd, hooks)
	}
	if len(hooks) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no hooks detected")
		return nil
	}
	tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SOURCE\tNAME\tEVENT\tSCOPE\tSTATUS\tRISK\tPATH")
	for _, h := range hooks {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", h.Source, h.Name, h.Event, h.Scope, h.Status, h.Risk, h.ConfigPath)
	}
	return tw.Flush()
}

func writeJSON(cmd *cobra.Command, payload any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}
