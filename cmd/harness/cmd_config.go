// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/configwiz"
)

func newConfigCmd() *cobra.Command {
	c := &cobra.Command{Use: "config", Short: "Inspect and edit .harness/config (paper §3.5.3)"}
	c.AddCommand(configShowCmd(), configSetCmd(), configUnsetCmd(), configWizardCmd())
	return c
}

func configShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print current task → adapter routing",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			snap, err := configwiz.Load(dir)
			if err != nil {
				return err
			}
			if len(snap.Routes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no overrides (using bundled defaults)")
				return nil
			}
			out := cmd.OutOrStdout()
			for task, r := range snap.Routes {
				fmt.Fprintf(out, "%-22s primary=%s fallback=%s budget=$%.2f model=%s\n",
					task, r.Primary, strings.Join(r.Fallback, ","), r.BudgetUSD, r.Model)
			}
			return nil
		},
	}
}

func configSetCmd() *cobra.Command {
	var task, primary, fallback, model string
	var budget float64
	c := &cobra.Command{
		Use:   "set",
		Short: "Set or override a task's route",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if task == "" || primary == "" {
				return errors.New("config set: --task and --primary required")
			}
			dir, err := cwd()
			if err != nil {
				return err
			}
			fb := configwiz.SplitCSVForCLI(fallback)
			return configwiz.SetTaskPrimary(dir, task, primary, fb, budget, model)
		},
	}
	c.Flags().StringVar(&task, "task", "", "task id (planning|implementation|security_review|...)")
	c.Flags().StringVar(&primary, "primary", "", "primary adapter id")
	c.Flags().StringVar(&fallback, "fallback", "", "csv fallback chain")
	c.Flags().Float64Var(&budget, "budget", 0, "budget cap USD (0 = none)")
	c.Flags().StringVar(&model, "model", "", "explicit model id (overrides adapter default)")
	return c
}

func configUnsetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unset <task>",
		Short: "Remove a task override and fall back to bundled default",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return configwiz.DeleteTask(dir, args[0])
		},
	}
}

func configWizardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wizard",
		Short: "Interactive routing wizard (paper §3.5.3 governed mutation)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			ids, err := agentcmd.AvailableAdapterIDs()
			if err != nil {
				return err
			}
			return configwiz.RunWizard(configwiz.WizardOptions{
				Root:         dir,
				AvailableIDs: ids,
				In:           cmd.InOrStdin(),
				Out:          cmd.OutOrStdout(),
			})
		},
	}
}
