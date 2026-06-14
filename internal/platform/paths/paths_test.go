package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindProjectRoot_GoMod(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x"), 0o644))

	sub := filepath.Join(root, "a", "b", "c")
	require.NoError(t, os.MkdirAll(sub, 0o755))

	got, err := FindProjectRoot(sub)
	require.NoError(t, err)
	require.Equal(t, root, got)
}

func TestFindProjectRoot_HarnessDirWins(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".harness"), 0o755))

	got, err := FindProjectRoot(root)
	require.NoError(t, err)
	require.Equal(t, root, got)
}

func TestFindProjectRoot_NoMarker(t *testing.T) {
	dir := t.TempDir()
	got, err := FindProjectRoot(dir)
	require.NoError(t, err)
	require.Equal(t, dir, got)
}

func TestFindProjectRoot_RelativeErrors(t *testing.T) {
	_, err := FindProjectRoot("relative/path")
	require.Error(t, err)
}
