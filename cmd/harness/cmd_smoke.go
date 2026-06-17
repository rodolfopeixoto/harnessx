// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/smokecmd"
)

func newSmokeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "smoke",
		Short: "Diagnostics that run HarnessX against a fresh project",
	}
	c.AddCommand(newSmokeMatrixCmd())
	return c
}

func newSmokeMatrixCmd() *cobra.Command {
	var (
		langs   string
		keep    bool
		jsonOut bool
		timeout time.Duration
		binPath string
	)
	c := &cobra.Command{
		Use:   "matrix",
		Short: "Exercise core HarnessX commands against every bundled scaffold",
		Long: `Creates a throw-away project for each language scaffold, runs the
canonical command sequence (init, install-git-hooks, scaffold apply,
doctor, sensor list, check, memory list, flow list, routes) and reports
which combinations pass.

A CLI-class failure exits non-zero; tool-class failures (missing host
binary) are reported but do not fail the matrix.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := smokecmd.Options{
				HarnessBin:  binPath,
				Keep:        keep,
				StepTimeout: timeout,
			}
			if s := strings.TrimSpace(langs); s != "" && s != "all" {
				for _, l := range strings.Split(s, ",") {
					if l = strings.TrimSpace(l); l != "" {
						opts.Langs = append(opts.Langs, l)
					}
				}
			}
			out := cmd.OutOrStdout()
			res, err := smokecmd.Run(cmd.Context(), opts, out)
			if err != nil {
				return err
			}
			if jsonOut {
				if err := smokecmd.FormatJSON(res, out); err != nil {
					return err
				}
			} else {
				smokecmd.FormatTable(res, out)
			}
			if !res.OK {
				return fmt.Errorf("smoke matrix failed: %v", smokecmd.FailedSteps(res))
			}
			return nil
		},
	}
	c.Flags().StringVar(&langs, "langs", "all", "comma-separated scaffolds to test, or 'all'")
	c.Flags().BoolVar(&keep, "keep", false, "keep temporary project directories on disk")
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON instead of the table")
	c.Flags().DurationVar(&timeout, "step-timeout", 60*time.Second, "per-step timeout")
	c.Flags().StringVar(&binPath, "bin", "", "override harness binary path (defaults to current executable)")
	return c
}
