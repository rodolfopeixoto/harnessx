package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/config"
)

func bootstrap(t *testing.T) (root string) {
	t.Helper()
	root = t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".harness", "config"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".harness", "db"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".harness", "logs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".harness", "config", "harness.yaml"),
		[]byte("version: 1\ndatabase:\n  path: .harness/db/harness.sqlite\nlogging:\n  path: .harness/logs/events.jsonl\n  rotate_max_bytes: 10485760\n"), 0o644))
	return
}

func TestHealth(t *testing.T) {
	root := bootstrap(t)
	srv, _ := New(Options{Root: root})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	srv.Handler().ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, true, body["ok"])
}

func TestListSessions_WithRow(t *testing.T) {
	root := bootstrap(t)
	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	repo, err := sqlite.Open(config.Resolve(root, cfg.Database.Path))
	require.NoError(t, err)
	defer repo.Close()
	ctx := context.Background()
	require.NoError(t, repo.CreateSession(ctx, domain.Session{
		ID: "s1", ProjectPath: root, Mode: domain.ModeBootstrap,
		Status: domain.StatusSucceeded, StartedAt: time.Now().UTC(),
	}))

	srv, _ := New(Options{Root: root})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	srv.Handler().ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "s1")
}

func TestCost_Empty(t *testing.T) {
	root := bootstrap(t)
	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	repo, err := sqlite.Open(config.Resolve(root, cfg.Database.Path))
	require.NoError(t, err)
	require.NoError(t, repo.Close())

	srv, _ := New(Options{Root: root})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/cost", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "total_usd")
}

func TestStatic_BuiltinFallback(t *testing.T) {
	root := bootstrap(t)
	srv, _ := New(Options{Root: root})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, strings.Contains(rec.Body.String(), "HarnessX"))
}

func TestStart_BindsAndShutsDown(t *testing.T) {
	root := bootstrap(t)
	srv, _ := New(Options{Root: root, Addr: "127.0.0.1:0"})
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start(ctx) }()
	// Allow the listener to bind.
	time.Sleep(100 * time.Millisecond)
	require.NotEmpty(t, srv.Addr())
	cancel()
	// Server shutdown emits http.ErrServerClosed; treat as expected.
	select {
	case err := <-errCh:
		require.True(t, err == nil || err == http.ErrServerClosed)
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down")
	}
}
