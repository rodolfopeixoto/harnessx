// SPDX-License-Identifier: MIT

package main

import (
	"testing"
	"time"

	"github.com/ropeixoto/harnessx/internal/execution"
)

func TestAggregateTrajectoryEmpty(t *testing.T) {
	got := aggregateTrajectory(nil, time.Time{})
	if got.ToolCalls != 0 || got.EditCount != 0 || got.WallMs != 0 {
		t.Errorf("empty: got %+v", got)
	}
}

func TestAggregateTrajectorySums(t *testing.T) {
	runs := []execution.Result{
		{
			Trajectory:    execution.Trajectory{ToolCalls: 3, EditCount: 5, WallMs: 1200},
			Verification:  execution.Verification{SensorsRun: 4, SensorsPassed: 3},
			Replayability: execution.Replayability{EventsComplete: true},
		},
		{
			Trajectory:    execution.Trajectory{ToolCalls: 2, EditCount: 1, WallMs: 300},
			Verification:  execution.Verification{SensorsRun: 2, SensorsPassed: 2},
			Replayability: execution.Replayability{EventsComplete: false},
		},
	}
	got := aggregateTrajectory(runs, time.Time{})
	if got.ToolCalls != 5 {
		t.Errorf("tool_calls: want 5, got %d", got.ToolCalls)
	}
	if got.EditCount != 6 {
		t.Errorf("edit_count: want 6, got %d", got.EditCount)
	}
	if got.WallMs != 1500 {
		t.Errorf("wall_ms: want 1500, got %d", got.WallMs)
	}
	if got.SensorsRun != 6 {
		t.Errorf("sensors_run: want 6, got %d", got.SensorsRun)
	}
	if got.SensorsPassed != 5 {
		t.Errorf("sensors_passed: want 5, got %d", got.SensorsPassed)
	}
	if got.EventsComplete != 1 {
		t.Errorf("events_complete: want 1, got %d", got.EventsComplete)
	}
}

func TestAggregateTrajectoryRespectsCutoff(t *testing.T) {
	cutoff := time.Now().Add(-1 * time.Hour)
	runs := []execution.Result{
		{StartedAt: time.Now().Add(-2 * time.Hour), Trajectory: execution.Trajectory{ToolCalls: 10}},
		{StartedAt: time.Now().Add(-30 * time.Minute), Trajectory: execution.Trajectory{ToolCalls: 3}},
	}
	got := aggregateTrajectory(runs, cutoff)
	if got.ToolCalls != 3 {
		t.Errorf("tool_calls after cutoff: want 3, got %d", got.ToolCalls)
	}
}
