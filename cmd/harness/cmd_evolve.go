// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/evolve"
)

func newEvolveCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "evolve",
		Short: "Telemetry-driven harness mutation (paper §3.5 AHE)",
	}
	c.AddCommand(evolveDiagnoseCmd(), evolveReplayCmd(), evolveProposeCmd(), evolvePromoteCmd(), evolveSandboxCmd())
	return c
}

func evolveSandboxCmd() *cobra.Command {
	var (
		candidate string
		timeout   time.Duration
	)
	c := &cobra.Command{
		Use:   "sandbox <trace-file>",
		Short: "Replay a trace against baseline + candidate harness binaries (paper §3.5.2)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bin, err := os.Executable()
			if err != nil {
				return err
			}
			res, err := evolve.RunSandbox(cmd.Context(), evolve.SandboxOptions{
				HarnessBin:   bin,
				CandidateBin: candidate,
				TraceFile:    args[0],
				Timeout:      timeout,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "baseline failures=%d candidate failures=%d delta=%d improved=%t\n",
				res.Baseline.Failures, res.Candidate.Failures, res.Improvement.FailuresDelta, res.Improvement.Improved)
			return evolve.WriteJSON(cmd.OutOrStdout(), res)
		},
	}
	c.Flags().StringVar(&candidate, "candidate", "", "candidate harness binary (defaults to baseline)")
	c.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "per-replay timeout")
	return c
}

func evolveDiagnoseCmd() *cobra.Command {
	var jsonOut bool
	c := &cobra.Command{
		Use:   "diagnose",
		Short: "Scan .harness/logs/events.jsonl and surface failure clusters",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			d, err := evolve.Diagnose(dir)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return evolve.WriteJSON(out, d)
			}
			fmt.Fprintf(out, "events scanned: %d   failures: %d   clusters: %d\n", d.Events, d.Failures, len(d.Clusters))
			for _, c := range d.Clusters {
				fmt.Fprintf(out, "  %4d × %s\n", c.Count, c.Signature)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON diagnosis")
	return c
}

func evolveReplayCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "replay <trace-file>",
		Short: "Score a candidate trace against current failure clusters",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			d, err := evolve.Diagnose(dir)
			if err != nil {
				return err
			}
			res, err := evolve.Replay(dir, args[0], d)
			if err != nil {
				return err
			}
			return evolve.WriteJSON(cmd.OutOrStdout(), res)
		},
	}
	return c
}

func evolveProposeCmd() *cobra.Command {
	var m evolve.Mutation
	c := &cobra.Command{
		Use:   "propose",
		Short: "Record a candidate harness mutation (logged, not applied)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			id, err := evolve.Propose(dir, m)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "proposed mutation %s\n", id)
			return nil
		},
	}
	c.Flags().StringVar(&m.Component, "component", "", "harness component (router|sensor|memory|orchestrator|...)")
	c.Flags().StringVar(&m.Description, "description", "", "what changes")
	c.Flags().StringVar(&m.Rationale, "rationale", "", "why")
	c.Flags().StringVar(&m.Risk, "risk", "medium", "low|medium|high")
	return c
}

func evolvePromoteCmd() *cobra.Command {
	var opts evolve.PromoteOptions
	c := &cobra.Command{
		Use:   "promote <mutation-id>",
		Short: "HITL-gated promotion of a previously proposed mutation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			opts.MutationID = args[0]
			if err := evolve.Promote(dir, opts); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "promoted mutation %s\n", opts.MutationID)
			return nil
		},
	}
	c.Flags().BoolVar(&opts.HITL, "hitl", false, "human-in-the-loop approval (required by paper §3.5.3)")
	c.Flags().StringVar(&opts.Reason, "reason", "", "approval reason recorded in mutations.jsonl")
	return c
}
