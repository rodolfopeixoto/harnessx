// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/configwiz"
	"github.com/ropeixoto/harnessx/internal/router"
)

func newConfigCmd() *cobra.Command {
	c := &cobra.Command{Use: "config", Short: "Inspect and edit .harness/config (paper §3.5.3)"}
	c.AddCommand(configShowCmd(), configSetCmd(), configUnsetCmd(), configWizardCmd(), newConfigSourcesCmd())
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
			out := cmd.OutOrStdout()
			source := "override"
			routes := snap.Routes
			if len(routes) == 0 {
				fmt.Fprintln(out, "no overrides — showing effective bundled defaults:")
				reg, _, regErr := agentcmd.LoadAll(dir)
				if regErr != nil {
					reg = agents.NewRegistry()
				}
				routes = router.Defaults(reg)
				source = "bundled"
			}
			for task, r := range routes {
				fmt.Fprintf(out, "%-22s [%s] primary=%s fallback=%s budget=$%.2f model=%s\n",
					task, source, r.Primary, strings.Join(r.Fallback, ","), r.BudgetUSD, r.Model)
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

func newConfigSourcesCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "sources",
		Short: "Manage upstream source URLs (update repo, adapter index)",
	}
	c.AddCommand(configSourcesGetCmd(), configSourcesSetCmd(), configSourcesResetCmd())
	return c
}

func configSourcesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Print one source key or the whole file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := readSourcesFile()
			if err != nil {
				return err
			}
			if len(args) == 0 {
				if len(data) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "(no overrides — using bundled defaults)")
					return nil
				}
				_, err := cmd.OutOrStdout().Write(data)
				return err
			}
			key := args[0]
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, key+":") {
					fmt.Fprintln(cmd.OutOrStdout(), strings.TrimSpace(strings.TrimPrefix(line, key+":")))
					return nil
				}
			}
			return fmt.Errorf("source key %q not set (try `harness config sources set %s <value>`)", key, key)
		},
	}
}

func configSourcesSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Persist a source override (key in: update_repo, adapter_index_url, install_index_url)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := writeSourceKey(args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "set %s = %s\n", args[0], args[1])
			return nil
		},
	}
}

func configSourcesResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset [key]",
		Short: "Remove one source override (or all when no key passed)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return removeSourcesFile()
			}
			return removeSourceKey(args[0])
		},
	}
}
