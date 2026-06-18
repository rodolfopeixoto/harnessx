// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/audit"
	"github.com/ropeixoto/harnessx/internal/execution"
)

func newMetricsCmd() *cobra.Command {
	var (
		since    string
		jsonOut  bool
		withTraj bool
	)
	c := &cobra.Command{
		Use:   "metrics",
		Short: "Aggregate cost/tokens/sensor pass-rate (+ trajectory with --trajectory)",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			cutoff := parseSince(since)
			runs, err := execution.ListRuns(root)
			if err != nil {
				return err
			}
			agg := aggregateMetrics(runs, cutoff)
			if withTraj {
				agg.Trajectory = aggregateTrajectory(runs, cutoff)
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(agg)
			}
			return renderMetrics(cmd.OutOrStdout(), agg)
		},
	}
	c.Flags().StringVar(&since, "since", "7d", "1d|7d|30d|all")
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	c.Flags().BoolVar(&withTraj, "trajectory", false, "include process-effort metrics (tool calls, edits, wall ms, sensors, events)")
	return c
}

type Aggregate struct {
	Window         string         `json:"window"`
	Runs           int            `json:"runs"`
	Applied        int            `json:"applied"`
	WaitingApprove int            `json:"waiting_approval"`
	Denied         int            `json:"autonomy_denied"`
	NoChanges      int            `json:"no_changes"`
	AgentFailed    int            `json:"agent_failed"`
	SensorFailed   int            `json:"sensor_failed"`
	InputTokens    int            `json:"input_tokens"`
	OutputTokens   int            `json:"output_tokens"`
	TotalCostUSD   float64        `json:"total_cost_usd"`
	ByAgent        map[string]int `json:"by_agent"`
	Trajectory     TrajectoryAgg  `json:"trajectory,omitempty"`
}

// TrajectoryAgg aggregates per-run trajectory metrics so operators
// can answer "how efficient was this window of work?" (paper § 5.2.1).
type TrajectoryAgg struct {
	ToolCalls      int   `json:"tool_calls"`
	EditCount      int   `json:"edit_count"`
	WallMs         int64 `json:"wall_ms"`
	SensorsRun     int   `json:"sensors_run"`
	SensorsPassed  int   `json:"sensors_passed"`
	EventsComplete int   `json:"events_complete"`
}

func aggregateTrajectory(runs []execution.Result, cutoff time.Time) TrajectoryAgg {
	var agg TrajectoryAgg
	for _, r := range runs {
		if !cutoff.IsZero() && r.StartedAt.Before(cutoff) {
			continue
		}
		agg.ToolCalls += r.Trajectory.ToolCalls
		agg.EditCount += r.Trajectory.EditCount
		agg.WallMs += r.Trajectory.WallMs
		agg.SensorsRun += r.Verification.SensorsRun
		agg.SensorsPassed += r.Verification.SensorsPassed
		if r.Replayability.EventsComplete {
			agg.EventsComplete++
		}
	}
	return agg
}

func aggregateMetrics(runs []execution.Result, cutoff time.Time) Aggregate {
	a := Aggregate{ByAgent: map[string]int{}}
	for _, r := range runs {
		if !cutoff.IsZero() && r.StartedAt.Before(cutoff) {
			continue
		}
		a.Runs++
		switch r.Status {
		case execution.StatusApplied:
			a.Applied++
		case execution.StatusWaitingApproval:
			a.WaitingApprove++
		case execution.StatusAutonomyDenied:
			a.Denied++
		case execution.StatusNoChanges:
			a.NoChanges++
		case execution.StatusAgentFailed:
			a.AgentFailed++
		case execution.StatusSensorFailed:
			a.SensorFailed++
		}
		a.InputTokens += r.InputTokens
		a.OutputTokens += r.OutputTokens
		a.TotalCostUSD += r.EstimatedCostUSD
		a.ByAgent[r.AgentID]++
	}
	a.Window = formatWindow(cutoff)
	return a
}

func renderMetrics(out interface{ Write(p []byte) (int, error) }, a Aggregate) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "window\t%s\n", a.Window)
	fmt.Fprintf(w, "runs\t%d\n", a.Runs)
	fmt.Fprintf(w, "applied\t%d\n", a.Applied)
	fmt.Fprintf(w, "waiting_approval\t%d\n", a.WaitingApprove)
	fmt.Fprintf(w, "autonomy_denied\t%d\n", a.Denied)
	fmt.Fprintf(w, "no_changes\t%d\n", a.NoChanges)
	fmt.Fprintf(w, "agent_failed\t%d\n", a.AgentFailed)
	fmt.Fprintf(w, "sensor_failed\t%d\n", a.SensorFailed)
	fmt.Fprintf(w, "input_tokens\t%d\n", a.InputTokens)
	fmt.Fprintf(w, "output_tokens\t%d\n", a.OutputTokens)
	fmt.Fprintf(w, "total_cost_usd\t$%.4f\n", a.TotalCostUSD)
	if len(a.ByAgent) > 0 {
		fmt.Fprintln(w, "by_agent:")
		for k, v := range a.ByAgent {
			fmt.Fprintf(w, "  %s\t%d\n", k, v)
		}
	}
	if a.Trajectory.SensorsRun > 0 || a.Trajectory.ToolCalls > 0 || a.Trajectory.EditCount > 0 {
		fmt.Fprintln(w, "trajectory:")
		fmt.Fprintf(w, "  tool_calls\t%d\n", a.Trajectory.ToolCalls)
		fmt.Fprintf(w, "  edit_count\t%d\n", a.Trajectory.EditCount)
		fmt.Fprintf(w, "  wall_ms\t%d\n", a.Trajectory.WallMs)
		fmt.Fprintf(w, "  sensors_run\t%d\n", a.Trajectory.SensorsRun)
		fmt.Fprintf(w, "  sensors_passed\t%d\n", a.Trajectory.SensorsPassed)
		fmt.Fprintf(w, "  events_complete\t%d\n", a.Trajectory.EventsComplete)
	}
	return w.Flush()
}

func parseSince(s string) time.Time {
	switch s {
	case "", "all":
		return time.Time{}
	case "1d":
		return time.Now().Add(-24 * time.Hour)
	case "7d":
		return time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		return time.Now().Add(-30 * 24 * time.Hour)
	}
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(-d)
	}
	return time.Time{}
}

func formatWindow(t time.Time) string {
	if t.IsZero() {
		return "all"
	}
	return "since " + t.Format("2006-01-02 15:04")
}

func newAuditCmd() *cobra.Command {
	var (
		limit   int
		kind    string
		jsonOut bool
	)
	c := &cobra.Command{
		Use:   "audit",
		Short: "Cross-project event timeline from .harness/audit/events.jsonl",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			events, err := loadAuditEvents(root)
			if err != nil {
				return err
			}
			if kind != "" {
				filtered := events[:0]
				for _, e := range events {
					if e.Kind == kind {
						filtered = append(filtered, e)
					}
				}
				events = filtered
			}
			if limit > 0 && len(events) > limit {
				events = events[:limit]
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(events)
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "WHEN\tKIND\tSOURCE\tSUBJECT")
			for _, e := range events {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.OccurredAt.Format("01-02 15:04:05"), e.Kind, e.Source, e.Subject)
			}
			return w.Flush()
		},
	}
	c.Flags().IntVar(&limit, "limit", 50, "max rows")
	c.Flags().StringVar(&kind, "kind", "", "filter by kind (sensor|hook|agent|cleanup|...)")
	c.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	c.AddCommand(newAuditTailCmd())
	return c
}

func newAuditTailCmd() *cobra.Command {
	var limit int
	c := &cobra.Command{
		Use:   "tail",
		Short: "Tail the last N events from the project event log",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			events, err := loadAuditEvents(root)
			if err != nil {
				return err
			}
			if limit > 0 && len(events) > limit {
				events = events[len(events)-limit:]
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "WHEN\tKIND\tSOURCE\tSUBJECT")
			for _, e := range events {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.OccurredAt.Format("01-02 15:04:05"), e.Kind, e.Source, e.Subject)
			}
			return w.Flush()
		},
	}
	c.Flags().IntVar(&limit, "limit", 30, "max rows")
	return c
}

func loadAuditEvents(root string) ([]audit.Event, error) {
	candidates := []string{
		filepath.Join(root, ".harness", "audit", "events.jsonl"),
		filepath.Join(root, ".harness", "logs", "events.jsonl"),
	}
	merged := []audit.Event{}
	for _, p := range candidates {
		sink := &audit.FileSink{Path: p}
		evs, err := sink.List(context.Background())
		if err != nil {
			continue
		}
		merged = append(merged, evs...)
	}
	return merged, nil
}
