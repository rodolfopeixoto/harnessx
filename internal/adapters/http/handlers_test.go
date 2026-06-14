// SPDX-License-Identifier: MIT

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/config"
)

func openRepo(t *testing.T, root string) *sqlite.Repo {
	t.Helper()
	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	repo, err := sqlite.Open(config.Resolve(root, cfg.Database.Path))
	require.NoError(t, err)
	return repo
}

func do(t *testing.T, root, path string) *httptest.ResponseRecorder {
	t.Helper()
	srv, _ := New(Options{Root: root})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestSessionDetail_NotFound(t *testing.T) {
	root := bootstrap(t)
	rec := do(t, root, "/api/sessions/")
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSessionDetail_WithRuns(t *testing.T) {
	root := bootstrap(t)
	repo := openRepo(t, root)
	defer repo.Close()
	ctx := context.Background()
	require.NoError(t, repo.CreateSession(ctx, domain.Session{
		ID: "sess-1", ProjectPath: root, Mode: domain.ModeBootstrap,
		Status: domain.StatusSucceeded, StartedAt: time.Now().UTC(),
	}))
	require.NoError(t, repo.CreateRun(ctx, domain.Run{
		ID: "run-1", SessionID: "sess-1", Stage: domain.StageInit,
		Status: domain.StatusSucceeded, StartedAt: time.Now().UTC(),
	}))
	rec := do(t, root, "/api/sessions/sess-1")
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "sess-1", body["session_id"])
	runs := body["runs"].([]any)
	require.Len(t, runs, 1)
}

func TestRunDetail_NotFound(t *testing.T) {
	root := bootstrap(t)
	rec := do(t, root, "/api/runs/")
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRunDetail_Empty(t *testing.T) {
	root := bootstrap(t)
	repo := openRepo(t, root)
	require.NoError(t, repo.Close())
	rec := do(t, root, "/api/runs/nonexistent")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "nonexistent")
}

func TestSensors_Empty(t *testing.T) {
	root := bootstrap(t)
	repo := openRepo(t, root)
	require.NoError(t, repo.Close())
	rec := do(t, root, "/api/sensors")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAgents_Empty(t *testing.T) {
	root := bootstrap(t)
	repo := openRepo(t, root)
	require.NoError(t, repo.Close())
	rec := do(t, root, "/api/agents")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestMemory_Empty(t *testing.T) {
	root := bootstrap(t)
	repo := openRepo(t, root)
	require.NoError(t, repo.Close())
	rec := do(t, root, "/api/memory")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestLogsTail_Empty(t *testing.T) {
	root := bootstrap(t)
	rec := do(t, root, "/api/logs?tail=10")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "lines")
}

func TestLogsTail_WithContent(t *testing.T) {
	root := bootstrap(t)
	logPath := filepath.Join(root, ".harness", "logs", "events.jsonl")
	require.NoError(t, os.WriteFile(logPath, []byte("a\nb\nc\n"), 0o644))
	rec := do(t, root, "/api/logs?tail=2")
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	lines := body["lines"].([]any)
	require.Len(t, lines, 2)
}

func TestProductJSON_Endpoints(t *testing.T) {
	root := bootstrap(t)
	prodDir := filepath.Join(root, ".harness", "product")
	require.NoError(t, os.MkdirAll(prodDir, 0o755))
	for _, f := range []string{"design-manifest.json", "roadmap.json", "toggle-map.json", "feature-map.json"} {
		require.NoError(t, os.WriteFile(filepath.Join(prodDir, f), []byte(`{"ok":true}`), 0o644))
	}
	for _, path := range []string{"/api/design", "/api/roadmap", "/api/toggles", "/api/features"} {
		rec := do(t, root, path)
		require.Equal(t, http.StatusOK, rec.Code, path)
		require.Contains(t, rec.Body.String(), "ok", path)
	}
}

func TestProfile_MissingReturnsEmptyEnvelope(t *testing.T) {
	root := bootstrap(t)
	rec := do(t, root, "/api/profile")
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, []any{}, body["stacks"])
	require.Nil(t, body["generated_at"])
}

func TestProfile_Endpoint(t *testing.T) {
	root := bootstrap(t)
	projDir := filepath.Join(root, ".harness", "project")
	require.NoError(t, os.MkdirAll(projDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(projDir, "profile.json"), []byte(`{"stacks":[]}`), 0o644))
	rec := do(t, root, "/api/profile")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "stacks")
}

func TestCost_WithRows(t *testing.T) {
	root := bootstrap(t)
	repo := openRepo(t, root)
	defer repo.Close()
	ctx := context.Background()
	require.NoError(t, repo.CreateSession(ctx, domain.Session{
		ID: "s1", ProjectPath: root, Mode: domain.ModeFeature,
		Status: domain.StatusSucceeded, StartedAt: time.Now().UTC(),
	}))
	require.NoError(t, repo.CreateRun(ctx, domain.Run{
		ID: "r1", SessionID: "s1", Stage: domain.StageExecution,
		Status: domain.StatusSucceeded, StartedAt: time.Now().UTC(),
		Agent: "claude",
	}))
	require.NoError(t, repo.UpdateRunCostAndTokens(ctx, "r1", 100, 0, 200, 0, 0.5, "claude", "sonnet", "", ""))
	rec := do(t, root, "/api/cost")
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.InDelta(t, 0.5, body["total_usd"].(float64), 0.001)
}
