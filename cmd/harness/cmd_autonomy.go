// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/autonomy"
	"github.com/ropeixoto/harnessx/internal/health"
)

func newAutonomyCmd() *cobra.Command {
	c := &cobra.Command{Use: "autonomy", Short: "Autonomy level inspection + gate preview"}
	get := &cobra.Command{
		Use:   "get",
		Short: "Print every autonomy level and its gate matrix",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ops := []autonomy.Operation{autonomy.OpRead, autonomy.OpPlan, autonomy.OpExecuteLowRisk, autonomy.OpExecuteHighRisk, autonomy.OpClean, autonomy.OpSchedule}
			fmt.Fprintf(cmd.OutOrStdout(), "%-22s | %s\n", "level", "decisions")
			for _, lvl := range autonomy.AllLevels() {
				line := ""
				for _, op := range ops {
					dec, _ := autonomy.Gate(lvl, op)
					line += fmt.Sprintf("%s=%s ", op, dec)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-22s | %s\n", lvl, line)
			}
			return nil
		},
	}
	set := &cobra.Command{
		Use:   "set <level>",
		Short: "Persist the active autonomy level to .harness/config/autonomy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			level := autonomy.Level(args[0])
			if err := autonomy.Save(root, level); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "active autonomy: %s (persisted at %s)\n", level, autonomy.ActiveFileRel)
			return nil
		},
	}
	active := &cobra.Command{
		Use:   "active",
		Short: "Print the persisted active level (defaults to manual)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			level, err := autonomy.Load(root)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), level)
			return nil
		},
	}
	c.AddCommand(get, set, active)
	return c
}

func newHealthCmd() *cobra.Command {
	c := &cobra.Command{Use: "health", Short: "Project health score"}
	show := &cobra.Command{
		Use:   "show",
		Short: "Compute health score from placeholder inputs (real measurement lands in a later phase)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := health.Inputs{
				TestsPassPct: 100, SensorsPassPct: 100, SecurityFindings: 0,
				PerfBudgetExceeded: false, OutdatedDeps: 1, DocsCoverage: 70,
				DesignParityPct: 80, RoadmapClearPct: 60, MemoryFreshDays: 10, InvalidConfigs: 0,
			}.Compute()
			fmt.Fprintln(cmd.OutOrStdout(), "(stub) values are placeholders; real measurement lands when health.Compute is wired to sensors")
			fmt.Fprintf(cmd.OutOrStdout(), "score: %d/100 (stub)\n", s.Total)
			for _, sub := range s.Subsystems {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-22s %3d  %s\n", sub.Name, sub.Score, sub.Reason)
			}
			return nil
		},
	}
	c.AddCommand(show)
	return c
}
