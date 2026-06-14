package router

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyStats_ReordersByOnSuccessRate(t *testing.T) {
	cfg := RouteConfig{
		Primary:  "codex",
		Fallback: []string{"claude", "gemini", "kimi"},
	}
	stats := map[string]AgentStats{
		"gemini": {AgentID: "gemini", SuccessRate: 0.95, TotalRuns: 20},
		"claude": {AgentID: "claude", SuccessRate: 0.70, TotalRuns: 10},
		"kimi":   {AgentID: "kimi", SuccessRate: 0.40, TotalRuns: 5},
	}
	got := ApplyStats(cfg, stats)
	require.Equal(t, "codex", got.Primary)
	require.Equal(t, []string{"gemini", "claude", "kimi"}, got.Fallback)
}

func TestApplyStats_NoHistory_KeepsOrder(t *testing.T) {
	cfg := RouteConfig{Primary: "a", Fallback: []string{"b", "c", "d"}}
	got := ApplyStats(cfg, nil)
	require.Equal(t, []string{"b", "c", "d"}, got.Fallback)
}

func TestApplyStats_NoHistorySortsLast(t *testing.T) {
	cfg := RouteConfig{Primary: "x", Fallback: []string{"untried", "experienced"}}
	stats := map[string]AgentStats{"experienced": {SuccessRate: 0.9, TotalRuns: 5}}
	got := ApplyStats(cfg, stats)
	require.Equal(t, []string{"experienced", "untried"}, got.Fallback)
}
