package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/domain"
)

func TestOpenAndMigrate_FileDB(t *testing.T) {
	dir := t.TempDir()
	repo, err := Open(filepath.Join(dir, "h.sqlite"))
	require.NoError(t, err)
	defer repo.Close()

	// All Phase 1 tables exist and are queryable.
	for _, table := range []string{
		"sessions", "runs", "sensor_results", "metrics",
		"memories", "skill_versions", "agent_certifications", "artifacts",
	} {
		_, err := repo.DB().Exec("select count(*) from " + table)
		require.NoErrorf(t, err, "table %s missing", table)
	}
}

func TestSessionAndRun_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	repo, err := Open(filepath.Join(dir, "h.sqlite"))
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	require.NoError(t, repo.CreateSession(ctx, domain.Session{
		ID:          "s1",
		ProjectPath: "/tmp/x",
		Mode:        domain.ModeBootstrap,
		Status:      domain.StatusRunning,
		StartedAt:   now,
	}))
	require.NoError(t, repo.CreateRun(ctx, domain.Run{
		ID:        "r1",
		SessionID: "s1",
		Stage:     domain.StageInit,
		Status:    domain.StatusRunning,
		StartedAt: now,
	}))
	require.NoError(t, repo.FinishRun(ctx, "r1", domain.StatusSucceeded, now.Add(50*time.Millisecond), 0))
	require.NoError(t, repo.FinishSession(ctx, "s1", domain.StatusSucceeded, now.Add(60*time.Millisecond)))

	sessions, err := repo.ListRecentSessions(ctx, 10)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	require.Equal(t, "s1", sessions[0].ID)
	require.Equal(t, domain.StatusSucceeded, sessions[0].Status)
	require.NotNil(t, sessions[0].FinishedAt)
}

func TestMigrateIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	repo, err := Open(filepath.Join(dir, "h.sqlite"))
	require.NoError(t, err)
	require.NoError(t, repo.migrate())
	require.NoError(t, repo.Close())
}
