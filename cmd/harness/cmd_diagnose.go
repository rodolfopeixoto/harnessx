// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/twoagent"
)

func newDiagnoseCmd() *cobra.Command {
	var stackTools []string
	c := &cobra.Command{
		Use:   "diagnose",
		Short: "Detect environment + project problems (paper §3.5.1)",
		Long: `Runs the bundled diagnosers (missing tools, dirty tree, unpinned plan)
and writes the diagnosis to .harness/artifacts/diagnoses/. Use
'harness fix' to apply registered remedies.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			diags := twoagent.DefaultDiagnosers(stackTools)
			d, err := twoagent.DiagnoseAll(cmd.Context(), dir, diags)
			if err != nil {
				return err
			}
			path, err := twoagent.SaveDiagnosis(dir, d)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), twoagent.FormatDiagnosis(d))
			fmt.Fprintf(cmd.OutOrStdout(), "\ndiagnosis saved: %s\n", path)
			return nil
		},
	}
	c.Flags().StringSliceVar(&stackTools, "tool", []string{"git"}, "tool name to probe (repeatable)")
	return c
}

func newFixCmd() *cobra.Command {
	var (
		diagPath string
		applyAll bool
		yes      bool
	)
	c := &cobra.Command{
		Use:   "fix [problem-id]",
		Short: "Apply registered remedies to problems surfaced by 'harness diagnose'",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			diag, err := loadDiagnosisFromFlag(dir, diagPath)
			if err != nil {
				return err
			}
			bin, _ := os.Executable()
			fixers := twoagent.DefaultFixers(bin)
			if len(args) > 0 {
				diag.Problems = filterByID(diag.Problems, args)
			} else if !applyAll && !yes {
				return errors.New("fix: pass --all or a problem id (or --yes to apply every fixable problem)")
			}
			res := twoagent.ApplyAll(cmd.Context(), dir, diag, fixers, cmd.OutOrStdout())
			ok := 0
			for _, r := range res {
				if r.Applied {
					ok++
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nfix summary: %d applied, %d skipped\n", ok, len(res)-ok)
			return nil
		},
	}
	c.Flags().StringVar(&diagPath, "diagnosis", "", "path to a previously saved diagnosis (defaults to the newest)")
	c.Flags().BoolVar(&applyAll, "all", false, "apply every fixer with a registered fix-id")
	c.Flags().BoolVar(&yes, "yes", false, "non-interactive (same as --all today)")
	return c
}

func loadDiagnosisFromFlag(dir, path string) (twoagent.Diagnosis, error) {
	if path != "" {
		return twoagent.LoadDiagnosis(path)
	}
	ctx := context.Background()
	d, err := twoagent.DiagnoseAll(ctx, dir, twoagent.DefaultDiagnosers([]string{"git"}))
	if err != nil {
		return twoagent.Diagnosis{}, err
	}
	return d, nil
}

func filterByID(ps []twoagent.Problem, ids []string) []twoagent.Problem {
	keep := map[string]bool{}
	for _, id := range ids {
		keep[id] = true
	}
	var out []twoagent.Problem
	for _, p := range ps {
		if keep[p.ID] {
			out = append(out, p)
		}
	}
	return out
}
