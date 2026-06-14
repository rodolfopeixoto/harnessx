// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/dashboardcmd"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func newDashboardCmd() *cobra.Command {
	var addr string
	var openBrowser bool
	c := &cobra.Command{
		Use:   "dashboard",
		Short: "Serve the local read-only dashboard (REST + React UI)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return dashboardcmd.Run(cmd.Context(), dashboardcmd.Options{
				StartDir: dir, Addr: addr, Open: openBrowser,
			}, cmd.OutOrStdout())
		},
	}
	c.Flags().StringVar(&addr, "addr", constants.DefaultDashboardAddr, "listen address (host:port)")
	c.Flags().BoolVar(&openBrowser, "open", false, "open the dashboard URL in the default browser")
	return c
}
