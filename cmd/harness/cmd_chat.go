// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/activeagent"
	"github.com/ropeixoto/harnessx/internal/agenthealth"
	"github.com/ropeixoto/harnessx/internal/agents"
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
		noAdapter   bool
		resumeID    string
		replayID    string
		autoGate    bool
	)
	c := &cobra.Command{
		Use:   "chat",
		Short: "Iterative REPL: prompt → plan JSON → deterministic execution (paper §3.1.4)",
		Long: `Drops into an interactive loop scoped to a goal:

  dev      — develop code
  ads      — agent workflows for advertising tasks
  research — context engineering, ask, explain
  ops      — doctor, runtime, containers, backup, audit, metrics

Without --adapter, harness chat auto-pins the agent recorded in
.harness/config/active.yaml (set via 'harness use <id>'). If no pin
exists it falls back to the first registered adapter in
claude / codex / gemini / kimi / ollama. Pass --no-adapter to force
the deterministic planner.`,
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
				AutoGate:    autoGate,
			}
			if resumeID != "" && replayID != "" {
				return fmt.Errorf("chat: --resume and --replay are mutually exclusive")
			}
			if resumeID != "" {
				sess, err := repl.LoadSession(dir, resumeID)
				if err != nil {
					return fmt.Errorf("chat: resume %q: %w", resumeID, err)
				}
				opts.Resume = sess
				fmt.Fprintf(cmd.OutOrStdout(), "chat: resuming session %s (%d prior turns)\n", sess.ID, len(sess.Turns))
			}
			if replayID != "" {
				sess, err := repl.LoadSession(dir, replayID)
				if err != nil {
					return fmt.Errorf("chat: replay %q: %w", replayID, err)
				}
				sess.ReadOnly = true
				opts.Resume = sess
				fmt.Fprintf(cmd.OutOrStdout(), "chat: replay %s (read-only, %d prior turns) — only /history /agents /cost /diff /help allowed\n", sess.ID, len(sess.Turns))
			}
			if !noAdapter {
				adapterID = resolveChatAdapter(dir, adapterID, cmd.OutOrStdout())
			}
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
				for _, id := range reg.IDs() {
					opts.AdaptersList = append(opts.AdaptersList, id)
				}
				registry := reg
				opts.SwitchTo = func(req string) (agents.AgentAdapter, string, error) {
					if a, ok := registry.Get(req); ok {
						return a, req, nil
					}
					return nil, "", fmt.Errorf("adapter %q not registered", req)
				}
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
	c.Flags().StringVar(&adapterID, "adapter", "", "adapter id (default: active.yaml pin, then first registered)")
	c.Flags().StringVar(&model, "model", "", "model override for the adapter")
	c.Flags().DurationVar(&stepTimeout, "step-timeout", 5*time.Minute, "per-step timeout")
	c.Flags().BoolVar(&noAdapter, "no-adapter", false, "force deterministic planner; skip adapter auto-pin")
	c.Flags().StringVar(&resumeID, "resume", "", "resume a prior session id from .harness/sessions/")
	c.Flags().StringVar(&replayID, "replay", "", "open a prior session in read-only mode (inspection slashes only)")
	c.Flags().BoolVar(&autoGate, "auto-gate", false, "run harness ci after every agent turn (toggle in-chat with /auto-gate)")
	c.AddCommand(newChatListCmd())
	return c
}

// resolveChatAdapter picks the adapter for `harness chat` when the user
// does not pass --adapter. Precedence:
//  1. the --adapter flag value (already wired by the caller),
//  2. the active.yaml pin (`harness use <id>`),
//  3. the first registered adapter from a short fallback chain so the
//     fresh-install case ("just typed harness chat") is not silently
//     degraded to the deterministic planner.
func resolveChatAdapter(dir, override string, out io.Writer) string {
	resolved := activeagent.ResolveAgentID(dir, override)
	if resolved != "" {
		if override == "" {
			fmt.Fprintf(out, "chat: auto-pinned %s from .harness/config/active.yaml\n", resolved)
		}
		return resolved
	}
	reg, _, err := agentcmd.LoadAll(dir)
	if err != nil {
		return ""
	}
	for _, fid := range []string{"claude", "codex", "gemini", "kimi", "ollama"} {
		if _, ok := reg.Get(fid); ok {
			fmt.Fprintf(out, "chat: no active pin — auto-selected %s (run 'harness use <id>' to set a default)\n", fid)
			return fid
		}
	}
	return ""
}

func newChatListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List prior harness chat sessions persisted under .harness/sessions/",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			rows, err := repl.ListSessions(dir)
			if err != nil {
				return err
			}
			if len(rows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no sessions yet")
				return nil
			}
			for _, r := range rows {
				label := r.Label
				if label == "" {
					label = "—"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %-20s  %s  turns=%d  last=%s\n", r.ID, label, r.Goal, r.Turns, r.LastInput)
			}
			return nil
		},
	}
}
