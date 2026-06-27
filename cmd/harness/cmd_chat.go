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
	"github.com/ropeixoto/harnessx/internal/router"
	"github.com/ropeixoto/harnessx/internal/ui"
)

//nolint:gocognit,gocyclo // cobra chat setup wires every flag inline; splitting hurts readability
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
		pipeMode    bool
		outputJSON  bool
		oneShot     bool
		routeOn     bool
	)
	c := &cobra.Command{
		Use:   "chat [<session-id|label>]",
		Args:  cobra.MaximumNArgs(1),
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
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			bin, err := os.Executable()
			if err != nil {
				return err
			}
			if len(args) > 0 && resumeID == "" && replayID == "" {
				resumeID = repl.ResolveSessionID(dir, args[0])
			}
			if resumeID != "" {
				resumeID = repl.ResolveSessionID(dir, resumeID)
			}
			if replayID != "" {
				replayID = repl.ResolveSessionID(dir, replayID)
			}
			opts := repl.Options{
				Root:         dir,
				HarnessBin:   bin,
				Goal:         intentplan.Goal(goal),
				In:           cmd.InOrStdin(),
				Out:          cmd.OutOrStdout(),
				StepTimeout:  stepTimeout,
				AutoGate:     autoGate,
				NoAdapter:    noAdapter,
				Pipe:         pipeMode,
				OneShot:      oneShot,
				RouteEnabled: routeOn,
			}
			if pipeMode || outputJSON {
				opts.Plain = true
				opts.Pipe = true
			}
			opts.OutputJSON = outputJSON
			if resumeID != "" && replayID != "" {
				return fmt.Errorf("chat: --resume and --replay are mutually exclusive")
			}
			if resumeID != "" {
				sess, err := repl.LoadSession(dir, resumeID)
				if err != nil {
					if hint := repl.SuggestSession(dir, resumeID); hint != "" {
						return fmt.Errorf("chat: resume %q: %w (did you mean %q?)", resumeID, err, hint)
					}
					return fmt.Errorf("chat: resume %q: %w", resumeID, err)
				}
				opts.Resume = sess
				fmt.Fprintf(cmd.OutOrStdout(), "chat: resuming session %s (%d prior turns)\n", sess.ID, len(sess.Turns))
			}
			if replayID != "" {
				sess, err := repl.LoadSession(dir, replayID)
				if err != nil {
					if hint := repl.SuggestSession(dir, replayID); hint != "" {
						return fmt.Errorf("chat: replay %q: %w (did you mean %q?)", replayID, err, hint)
					}
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
					handleAuthFailure(cmd.Context(), cmd.OutOrStdout(), cmd.InOrStdin(), adapter, adapterID, h)
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
				opts.AdaptersList = append(opts.AdaptersList, reg.IDs()...)
				registry := reg
				opts.SwitchTo = func(req string) (agents.AgentAdapter, string, error) {
					a, ok := registry.Get(req)
					if !ok {
						return nil, "", fmt.Errorf("adapter %q not registered", req)
					}
					if perr := activeagent.Save(opts.Root, activeagent.Pin{AgentID: req}); perr != nil {
						fmt.Fprintf(opts.Out, "chat: warning — could not persist /use %s pin: %v\n", req, perr)
					}
					return a, req, nil
				}
				// Multi-agent routing: route plain text and slash
				// commands through the router so /plan reaches the
				// cheap chain and plain text reaches the implementation
				// chain instead of always hitting the pinned adapter.
				rtr := router.New(registry, router.Defaults(registry))
				opts.Route = func(task string) (agents.AgentAdapter, string, error) {
					dec, err := rtr.Select(task)
					if err != nil || len(dec.Chain) == 0 {
						return adapter, adapterID, nil
					}
					return dec.Chain[0], dec.Chain[0].ID(), nil
				}
				mode := "iterative"
				if oneShot {
					mode = "one-shot"
				}
				routing := "off"
				if routeOn {
					routing = "on"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "chat: %s wired — mode=%s routing=%s; /model swaps model mid-session, /route on|off toggles per-task routing, /once exits after next turn\n", adapterID, mode, routing)
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
	c.Flags().BoolVar(&pipeMode, "pipe", false, "non-interactive mode for scripts/CI: plain output, no greet/recap, exit on stdin EOF")
	c.Flags().BoolVar(&outputJSON, "output-json", false, "emit one JSON envelope per turn on stdout (forces --pipe + --no-adapter unless --adapter set)")
	c.Flags().BoolVar(&oneShot, "once", false, "non-iterative: exit after the first prompt (one-shot mode)")
	c.Flags().BoolVar(&routeOn, "route", false, "enable per-task multi-agent routing from the start (toggle in-chat with /route on|off)")
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
	available := []string{}
	for _, fid := range []string{"ollama", "kimi", "gemini", "codex", "claude"} {
		if _, ok := reg.Get(fid); ok {
			available = append(available, fid)
		}
	}
	if len(available) > 0 {
		fmt.Fprintln(out, "chat: no pin yet — available adapters:")
		for i, fid := range available {
			fmt.Fprintf(out, "  %d. %s\n", i+1, fid)
		}
		fmt.Fprintln(out, "  tip: run 'harness use <id>' to make a pin permanent")
	}
	for _, fid := range []string{"ollama", "kimi", "gemini", "codex", "claude"} {
		if _, ok := reg.Get(fid); ok {
			fmt.Fprintf(out, "chat: no active pin — auto-selected %s (cheapest available; run 'harness use <id>' to pin a stronger model)\n", fid)
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
