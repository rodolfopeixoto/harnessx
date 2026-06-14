// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/catalogcmd"
	"github.com/ropeixoto/harnessx/internal/palette"
	"github.com/ropeixoto/harnessx/internal/workspace"
)

func newPaletteCmd() *cobra.Command {
	c := &cobra.Command{Use: "palette", Short: "Search across projects, capabilities and commands"}
	search := &cobra.Command{
		Use:   "search <q>",
		Short: "Run the command palette search from the terminal",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			q := args[0]
			for _, more := range args[1:] {
				q += " " + more
			}
			reg, _ := workspace.Open("")
			if reg != nil {
				defer reg.Close()
			}
			cat := catalogcmd.New()
			root, err := cwd()
			if err != nil {
				return err
			}
			p := palette.New(
				palette.ProjectsSource{Registry: reg},
				palette.CapabilitiesSource{Catalog: cat, Root: root},
				palette.CommandsSource{Commands: palette.BuiltinCommands},
			)
			hits, err := p.Search(cmd.Context(), q)
			if err != nil {
				return err
			}
			if len(hits) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no hits")
				return nil
			}
			for _, h := range hits {
				fmt.Fprintf(cmd.OutOrStdout(), "%-12s %-14s %-40s %s\n", h.Source, h.Kind, h.Title, h.RouterPath)
			}
			return nil
		},
	}
	c.AddCommand(search)
	return c
}
