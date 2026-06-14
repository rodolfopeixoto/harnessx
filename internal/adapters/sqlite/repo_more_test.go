// SPDX-License-Identifier: MIT

package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/domain"
)

func openTemp(t *testing.T) *Repo {
	t.Helper()
	r, err := Open(filepath.Join(t.TempDir(), "h.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close() })
	return r
}

func TestOpen_MemoryDSN(t *testing.T) {
	r, err := Open(":memory:")
	require.NoError(t, err)
	require.NotNil(t, r.DB())
	require.NoError(t, r.Close())
}

func TestAgentCertification_Roundtrip(t *testing.T) {
	r := openTemp(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	cert := domain.AgentCertification{
		ID: "c1", AgentID: "claude", CLIVersion: "1.2.3", AdapterVersion: "0.1",
		Score: 95, Status: "certified", DetailsJSON: "{}", CertifiedAt: now,
	}
	require.NoError(t, r.WriteAgentCertification(ctx, cert))
	got, err := r.LatestAgentCertification(ctx, "claude")
	require.NoError(t, err)
	require.Equal(t, "c1", got.ID)
	require.Equal(t, 95, got.Score)
	require.Equal(t, "1.2.3", got.CLIVersion)
}

func TestAgentCertification_MissingAgent(t *testing.T) {
	r := openTemp(t)
	_, err := r.LatestAgentCertification(context.Background(), "nope")
	require.Error(t, err)
}

func TestSensorResults_Roundtrip(t *testing.T) {
	r := openTemp(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	require.NoError(t, r.CreateSession(ctx, domain.Session{
		ID: "s1", ProjectPath: "/x", Mode: domain.ModeBootstrap,
		Status: domain.StatusRunning, StartedAt: now,
	}))
	require.NoError(t, r.CreateRun(ctx, domain.Run{
		ID: "r1", SessionID: "s1", Stage: domain.StageSensors,
		Status: domain.StatusRunning, StartedAt: now,
	}))
	require.NoError(t, r.WriteSensorResult(ctx, "r1", "go_vet", "passed", 120, "/tmp/out.txt", now))
	require.NoError(t, r.WriteSensorResult(ctx, "r1", "go_test", "failed", 500, "", now))
	results, err := r.ListSensorResults(ctx, "r1")
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Equal(t, "go_vet", results[0].Sensor)
	require.Equal(t, int64(120), results[0].DurationMs)
}

func TestSensorResults_EmptyRun(t *testing.T) {
	r := openTemp(t)
	results, err := r.ListSensorResults(context.Background(), "nonexistent")
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestMetric_Insert(t *testing.T) {
	r := openTemp(t)
	ctx := context.Background()
	now := time.Now().UTC()
	require.NoError(t, r.CreateSession(ctx, domain.Session{
		ID: "s1", ProjectPath: "/x", Mode: domain.ModeBootstrap,
		Status: domain.StatusRunning, StartedAt: now,
	}))
	require.NoError(t, r.CreateRun(ctx, domain.Run{
		ID: "r1", SessionID: "s1", Stage: domain.StageInit,
		Status: domain.StatusRunning, StartedAt: now,
	}))
	require.NoError(t, r.WriteMetric(ctx, "r1", "latency_ms", 42.5, "ms", `{"agent":"x"}`, now))
	require.NoError(t, r.WriteMetric(ctx, "r1", "tokens", 100, "", "", now))

	var count int
	require.NoError(t, r.DB().QueryRow("select count(*) from metrics where run_id = 'r1'").Scan(&count))
	require.Equal(t, 2, count)
}

func TestUpdateRunCostAndTokens(t *testing.T) {
	r := openTemp(t)
	ctx := context.Background()
	now := time.Now().UTC()
	require.NoError(t, r.CreateSession(ctx, domain.Session{
		ID: "s1", ProjectPath: "/x", Mode: domain.ModeFeature,
		Status: domain.StatusRunning, StartedAt: now,
	}))
	require.NoError(t, r.CreateRun(ctx, domain.Run{
		ID: "r1", SessionID: "s1", Stage: domain.StageExecution,
		Status: domain.StatusRunning, StartedAt: now,
	}))
	require.NoError(t, r.UpdateRunCostAndTokens(ctx, "r1", 100, 50, 200, 10, 0.0123, "claude", "sonnet", "kimi", ""))

	var (
		in, out int
		cost    float64
		agent   string
	)
	require.NoError(t, r.DB().QueryRow(
		"select input_tokens, output_tokens, estimated_cost_usd, agent from runs where id = 'r1'").
		Scan(&in, &out, &cost, &agent))
	require.Equal(t, 100, in)
	require.Equal(t, 200, out)
	require.InDelta(t, 0.0123, cost, 0.0001)
	require.Equal(t, "claude", agent)
}

func TestListRecentSessions_HonoursLimit(t *testing.T) {
	r := openTemp(t)
	ctx := context.Background()
	base := time.Now().UTC().Truncate(time.Second)
	for i := 0; i < 5; i++ {
		require.NoError(t, r.CreateSession(ctx, domain.Session{
			ID: "s" + string(rune('a'+i)), ProjectPath: "/x", Mode: domain.ModeBootstrap,
			Status: domain.StatusSucceeded, StartedAt: base.Add(time.Duration(i) * time.Second),
		}))
	}
	got, err := r.ListRecentSessions(ctx, 3)
	require.NoError(t, err)
	require.Len(t, got, 3)
	// Newest first.
	require.Equal(t, "se", got[0].ID)
}

func TestListRecentSessions_DefaultLimit(t *testing.T) {
	r := openTemp(t)
	got, err := r.ListRecentSessions(context.Background(), 0)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestFinishRun_SetsLatency(t *testing.T) {
	r := openTemp(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	require.NoError(t, r.CreateSession(ctx, domain.Session{
		ID: "s1", ProjectPath: "/x", Mode: domain.ModeBootstrap,
		Status: domain.StatusRunning, StartedAt: now,
	}))
	require.NoError(t, r.CreateRun(ctx, domain.Run{
		ID: "r1", SessionID: "s1", Stage: domain.StageInit,
		Status: domain.StatusRunning, StartedAt: now,
	}))
	require.NoError(t, r.FinishRun(ctx, "r1", domain.StatusSucceeded, now.Add(250*time.Millisecond), 0))
	var latency int64
	require.NoError(t, r.DB().QueryRow("select latency_ms from runs where id='r1'").Scan(&latency))
	require.GreaterOrEqual(t, latency, int64(200))
}

func TestNullIfEmpty(t *testing.T) {
	require.Nil(t, nullIfEmpty(""))
	require.Equal(t, "x", nullIfEmpty("x"))
}
