// SPDX-License-Identifier: MIT

package stale

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetect_FirstTime(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "package.json"), []byte(`{}`), 0o644))
	out, err := Detect(root)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "dependencies", out[0].Kind)
	require.Empty(t, out[0].HashOld)
}

func TestRecordAndStable(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "Dockerfile"), []byte("FROM scratch"), 0o644))
	_, err := Record(root)
	require.NoError(t, err)
	out, err := Detect(root)
	require.NoError(t, err)
	require.Empty(t, out, "no changes after Record")
}

func TestDetect_ContentChange(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "go.mod")
	require.NoError(t, os.WriteFile(target, []byte("module x"), 0o644))
	_, err := Record(root)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(target, []byte("module y"), 0o644))
	out, err := Detect(root)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "content changed since last index", out[0].Reason)
}

func TestDetect_KindCategorisation(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "docker-compose.yaml"), []byte("services: {}"), 0o644))
	out, err := Detect(root)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "container", out[0].Kind)
}

func TestLoad_MissingReturnsEmpty(t *testing.T) {
	fp, err := Load(t.TempDir())
	require.NoError(t, err)
	require.Empty(t, fp.Files)
}
