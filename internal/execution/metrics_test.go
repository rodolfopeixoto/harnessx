// SPDX-License-Identifier: MIT

package execution

import (
	"testing"
	"time"
)

func TestPopulateMetricsEmptyResult(t *testing.T) {
	res := Result{}
	populateMetrics(&res)
	if res.Trajectory.ToolCalls != 0 || res.Trajectory.EditCount != 0 {
		t.Errorf("empty: got %+v", res.Trajectory)
	}
	if res.Verification.OracleCount != 0 {
		t.Errorf("oracle: got %d", res.Verification.OracleCount)
	}
}

func TestPopulateMetricsWallMs(t *testing.T) {
	start := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	res := Result{StartedAt: start, FinishedAt: start.Add(450 * time.Millisecond)}
	populateMetrics(&res)
	if res.Trajectory.WallMs != 450 {
		t.Errorf("wall_ms: want 450, got %d", res.Trajectory.WallMs)
	}
}

func TestPopulateMetricsCountsHooksAndFiles(t *testing.T) {
	res := Result{
		Hooks:        []HookOutcome{{Name: "h1"}, {Name: "h2"}},
		ChangedFiles: []string{"a.go", "b.go", "c.go"},
	}
	populateMetrics(&res)
	if res.Trajectory.ToolCalls != 2 {
		t.Errorf("tool_calls: want 2, got %d", res.Trajectory.ToolCalls)
	}
	if res.Trajectory.EditCount != 3 {
		t.Errorf("edit_count: want 3, got %d", res.Trajectory.EditCount)
	}
}

func TestPopulateMetricsSensorsPassed(t *testing.T) {
	res := Result{Sensors: []SensorOutcome{
		{ID: "a", Status: "passed"},
		{ID: "b", Status: "failed"},
		{ID: "c", Status: "passed"},
	}}
	populateMetrics(&res)
	if res.Verification.SensorsRun != 3 {
		t.Errorf("sensors_run: want 3, got %d", res.Verification.SensorsRun)
	}
	if res.Verification.SensorsPassed != 2 {
		t.Errorf("sensors_passed: want 2, got %d", res.Verification.SensorsPassed)
	}
	if res.Verification.OracleCount != 3 {
		t.Errorf("oracle: want 3, got %d", res.Verification.OracleCount)
	}
}

func TestPopulateMetricsReplayability(t *testing.T) {
	cases := []struct {
		name string
		res  Result
		want bool
	}{
		{"nothing", Result{}, false},
		{"stdout", Result{StdoutPath: "/x"}, true},
		{"stderr", Result{StderrPath: "/x"}, true},
		{"jsonl", Result{JSONLPath: "/x"}, true},
	}
	for _, c := range cases {
		populateMetrics(&c.res)
		if c.res.Replayability.EventsComplete != c.want {
			t.Errorf("%s: want %v, got %v", c.name, c.want, c.res.Replayability.EventsComplete)
		}
	}
}

func TestPopulateMetricsZeroTimeNoWall(t *testing.T) {
	res := Result{}
	populateMetrics(&res)
	if res.Trajectory.WallMs != 0 {
		t.Errorf("zero start/finish should yield 0, got %d", res.Trajectory.WallMs)
	}
}
