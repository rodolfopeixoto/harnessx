package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_Missing_ReturnsDefaults(t *testing.T) {
	root := t.TempDir()
	cfg, err := Load(filepath.Join(root, "absent.yaml"), root)
	require.NoError(t, err)
	require.Equal(t, 1, cfg.Version)
	require.Equal(t, filepath.Base(root), cfg.Project.Name)
	require.NotEmpty(t, cfg.Database.Path)
	require.EqualValues(t, 10*1024*1024, cfg.Logging.RotateMaxBytes)
}

func TestLoad_PartialOverride_MergesDefaults(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "harness.yaml")
	require.NoError(t, os.WriteFile(p, []byte("project:\n  name: custom\n"), 0o644))

	cfg, err := Load(p, root)
	require.NoError(t, err)
	require.Equal(t, "custom", cfg.Project.Name)
	require.Equal(t, root, cfg.Project.Root) // default kept
	require.NotEmpty(t, cfg.Database.Path)   // default kept
	require.EqualValues(t, 10*1024*1024, cfg.Logging.RotateMaxBytes)
}

func TestResolve(t *testing.T) {
	require.Equal(t, "/abs/x", Resolve("/root", "/abs/x"))
	require.Equal(t, filepath.Join("/root", "rel/x"), Resolve("/root", "rel/x"))
}
