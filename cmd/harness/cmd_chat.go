// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/activeagent"
	"github.com/ropeixoto/harnessx/internal/agenthealth"
	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/intentplan"
	"github.com/ropeixoto/harnessx/internal/repl"
	"github.com/ropeixoto/harnessx/internal/ui"
)

func newChatCmd() *cobra.Command {
	var (
		goal        string
		adapterID   string
		model       string
		stepTimeout time.Duration
	)
	c := &cobra.Command{
		Use:   "chat",
		Short: "Iterative REPL: prompt → plan JSON → deterministic execution (paper §3.1.4)",
		Long: `Drops into an interactive loop scoped to a goal:

  dev      — develop code
  ads      — agent workflows for advertising tasks
  research — context engineering, ask, explain
  ops      — doctor, runtime, containers, backup, audit, metrics

By default uses the deterministic planner. Pass --adapter to back the
planner with a configured agent (e.g. claude, kimi, ollama). The
adapter is called with the "planning" task tag so router cost rules
apply.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			bin, err := os.Executable()
			if err != nil {
				return err
			}
			opts := repl.Options{
				Root:        dir,
				HarnessBin:  bin,
				Goal:        intentplan.Goal(goal),
				In:          cmd.InOrStdin(),
				Out:         cmd.OutOrStdout(),
				StepTimeout: stepTimeout,
			}
			adapterID = activeagent.ResolveAgentID(dir, adapterID)
			if adapterID != "" {
				reg, _, err := agentcmd.LoadAll(dir)
				if err != nil {
					return err
				}
				adapter, ok := reg.Get(adapterID)
				if !ok {
					for _, fid := range []string{"claude", "codex", "gemini", "kimi"} {
						if a, found := reg.Get(fid); found && fid != adapterID {
							fmt.Fprintf(cmd.OutOrStdout(), "chat: adapter %q not registered, falling back to %q\n", adapterID, fid)
							adapter, adapterID, ok = a, fid, true
							break
						}
					}
					if !ok {
						return fmt.Errorf("chat: adapter %q not registered (no fallback found)", adapterID)
					}
				}
				if h := adapter.Healthcheck(cmd.Context()); !h.OK {
					fmt.Fprintf(cmd.OutOrStdout(), "chat: %s healthcheck warn: %s\n", adapterID, h.Err)
				}
				planner, err := repl.NewLLMPlanner(repl.LLMPlannerOptions{
					Adapter:    adapter,
					Model:      model,
					WorkingDir: dir,
				})
				if err != nil {
					return err
				}
				opts.Planner = planner
				opts.Adapter = adapter
				opts.AdapterID = adapterID
				opts.Model = model
				fmt.Fprintf(cmd.OutOrStdout(), "chat: %s wired — plain text streams to agent; /exec for harness plan\n", adapterID)
				probe := agenthealth.New(adapter, 30*time.Second)
				probe.Start(cmd.Context())
				defer probe.Stop()
				opts.HealthProbe = probe
				opts.Plain = ui.IsPlain()
			}
			return repl.Run(cmd.Context(), opts)
		},
	}
	c.Flags().StringVar(&goal, "goal", "dev", "session goal (dev|ads|research|ops)")
	c.Flags().StringVar(&adapterID, "adapter", "", "adapter id for LLM-backed planning (empty = deterministic)")
	c.Flags().StringVar(&model, "model", "", "model override for the adapter")
	c.Flags().DurationVar(&stepTimeout, "step-timeout", 5*time.Minute, "per-step timeout")
	return c
}
