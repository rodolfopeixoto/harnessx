// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/flowpkg"
)

func newFlowCmd() *cobra.Command {
	c := &cobra.Command{Use: "flow", Short: "Domain-agnostic deterministic flow registry"}
	c.AddCommand(flowListCmd(), flowShowCmd(), flowApplyCmd(), flowInitCmd())
	return c
}

func flowInitCmd() *cobra.Command {
	var yes bool
	c := &cobra.Command{
		Use:   "init <name>",
		Short: "One-shot dry-run plan (pass --yes to execute)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			f, err := flowpkg.Load(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "→ flow init: %s (%d phases)\n", f.Name, len(f.Phases))
			if !yes {
				fmt.Fprintln(out, "  dry-run only; pass --yes to execute")
			}
			res, err := flowpkg.Apply(cmd.Context(), f, flowpkg.ApplyOptions{Root: dir, Dry: !yes}, out)
			if err != nil {
				return err
			}
			for _, r := range res {
				icon := "✓"
				switch {
				case r.Err != nil:
					icon = "✗"
				case r.Skipped:
					icon = "·"
				}
				fmt.Fprintf(out, "  %s %s\n", icon, r.Name)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "execute phases (default dry-run)")
	return c
}

func flowListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List bundled flows",
		RunE: func(cmd *cobra.Command, _ []string) error {
			names, err := flowpkg.List()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(names) == 0 {
				fmt.Fprintln(out, "no bundled flows yet (v0.93+ ships rails-api, python-fastapi, meta-ads-campaign, etc.)")
				return nil
			}
			for _, n := range names {
				fmt.Fprintln(out, n)
			}
			return nil
		},
	}
}

func flowShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Dump a flow's YAML definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := flowpkg.Load(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "name: %s\ndomain: %s\nphases: %d\n", f.Name, f.Domain, len(f.Phases))
			for _, p := range f.Phases {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s (%s)\n", p.Name, p.Kind)
			}
			return nil
		},
	}
}

func flowApplyCmd() *cobra.Command {
	var dry bool
	c := &cobra.Command{
		Use:   "apply <name>",
		Short: "Apply a bundled flow to the current directory (default dry-run)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			f, err := flowpkg.Load(args[0])
			if err != nil {
				return err
			}
			res, err := flowpkg.Apply(cmd.Context(), f, flowpkg.ApplyOptions{Root: dir, Dry: dry}, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			for _, r := range res {
				icon := "✓"
				if r.Err != nil {
					icon = "✗"
				} else if r.Skipped {
					icon = "·"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %s %s\n", icon, r.Name)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&dry, "dry-run", true, "do not execute phases; just print the plan")
	return c
}
