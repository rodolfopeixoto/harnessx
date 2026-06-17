// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/orchestrate"
)

func newOrchestrateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "orchestrate",
		Short: "Multi-agent role orchestration (paper §4.1.1 + §4.3.1)",
	}
	c.AddCommand(orchestrateListCmd(), orchestrateRunCmd(), orchestrateShowCmd())
	return c
}

func orchestrateListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List orchestration flows under .harness/orchestrations/",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			names, err := orchestrate.List(dir)
			if err != nil {
				return err
			}
			if len(names) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no flows (drop one at .harness/orchestrations/<name>.yaml)")
				return nil
			}
			for _, n := range names {
				fmt.Fprintln(cmd.OutOrStdout(), n)
			}
			return nil
		},
	}
}

func orchestrateShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Print parsed orchestration flow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			f, err := orchestrate.Load(dir, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "name: %s\ntopology: %s\nsteps: %d\n", f.Name, f.Topology, len(f.Steps))
			for i, s := range f.Steps {
				fmt.Fprintf(cmd.OutOrStdout(), "  %d. role=%s cmd=%s\n", i, s.Role, strings.Join(s.Command, " "))
			}
			return nil
		},
	}
}

func orchestrateRunCmd() *cobra.Command {
	var (
		dry            bool
		adapterTimeout time.Duration
	)
	c := &cobra.Command{
		Use:   "run <name>",
		Short: "Run an orchestration flow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			f, err := orchestrate.Load(dir, args[0])
			if err != nil {
				return err
			}
			opts := orchestrate.RunOptions{Root: dir, Flow: f, DryRun: dry}
			if hasAdapterStep(f) && !dry {
				reg, _, err := agentcmd.LoadAll(dir)
				if err != nil {
					return err
				}
				opts.AdapterRunner = orchestrate.NewAdapterRunner(reg, dir, adapterTimeout)
			}
			res, err := orchestrate.Run(cmd.Context(), opts, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			if !res.OK {
				return fmt.Errorf("orchestrate: flow failed (run %s)", res.RunID)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&dry, "dry-run", false, "print steps without executing")
	c.Flags().DurationVar(&adapterTimeout, "adapter-timeout", 2*time.Minute, "per-adapter-call timeout")
	return c
}

func hasAdapterStep(f orchestrate.Flow) bool {
	for _, s := range f.Steps {
		if s.Adapter != "" && len(s.Command) == 0 {
			return true
		}
	}
	return false
}
