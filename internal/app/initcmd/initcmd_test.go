package initcmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
)

func TestRun_CreatesLayoutAndRecordsBootstrap(t *testing.T) {
	root := t.TempDir()
	// Marker so FindProjectRoot returns this dir.
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x"), 0o644))

	var out bytes.Buffer
	res, err := Run(context.Background(), Options{StartDir: root}, &out)
	require.NoError(t, err)
	require.Equal(t, root, res.Root)

	for _, sub := range []string{"config", "db", "logs", "cache", "artifacts", "product", "project"} {
		_, err := os.Stat(filepath.Join(res.HarnessDir, sub))
		require.NoErrorf(t, err, "missing %s", sub)
	}
	require.FileExists(t, res.ConfigPath)
	require.FileExists(t, res.DBPath)
	require.FileExists(t, res.LogPath)

	repo, err := sqlite.Open(res.DBPath)
	require.NoError(t, err)
	defer repo.Close()

	sessions, err := repo.ListRecentSessions(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	require.Equal(t, root, sessions[0].ProjectPath)
}

func TestRun_IdempotentWithoutForce(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x"), 0o644))

	var out bytes.Buffer
	_, err := Run(context.Background(), Options{StartDir: root}, &out)
	require.NoError(t, err)
	res2, err := Run(context.Background(), Options{StartDir: root}, &out)
	require.NoError(t, err)
	require.True(t, res2.AlreadyInit)

	repo, err := sqlite.Open(res2.DBPath)
	require.NoError(t, err)
	defer repo.Close()

	sessions, err := repo.ListRecentSessions(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, sessions, 2, "second init should record a new bootstrap session")
}
