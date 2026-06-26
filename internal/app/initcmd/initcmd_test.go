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

func TestRun_AddsWorktreeIgnoreToRootGitignore_BUG16(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x"), 0o644))
	// Pre-existing .gitignore with unrelated entry must be preserved.
	require.NoError(t, os.WriteFile(filepath.Join(root, ".gitignore"), []byte("dist/\n"), 0o644))

	var out bytes.Buffer
	_, err := Run(context.Background(), Options{StartDir: root}, &out)
	require.NoError(t, err)

	contents, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	require.NoError(t, err)
	require.Contains(t, string(contents), ".harness/worktrees/")
	require.Contains(t, string(contents), "dist/")

	// Idempotent: second init must not duplicate the line.
	_, err = Run(context.Background(), Options{StartDir: root, Force: true}, &out)
	require.NoError(t, err)
	contents2, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	require.NoError(t, err)
	count := 0
	for _, line := range bytesSplitLines(contents2) {
		if line == ".harness/worktrees/" {
			count++
		}
	}
	require.Equal(t, 1, count, "worktree ignore should appear exactly once")
}

func bytesSplitLines(b []byte) []string {
	var out []string
	cur := ""
	for _, ch := range string(b) {
		if ch == '\n' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(ch)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
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
