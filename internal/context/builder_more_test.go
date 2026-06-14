// SPDX-License-Identifier: MIT

package context

import (
	stdctx "context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitProvider_NoBinary_Skips(t *testing.T) {
	pack := &Pack{}
	require.NoError(t, GitProvider{}.Apply(stdctx.Background(), t.TempDir(), pack))
}

func TestRipgrepProvider_NoBinary_Skips(t *testing.T) {
	pack := &Pack{Task: "find foo"}
	require.NoError(t, RipgrepProvider{}.Apply(stdctx.Background(), t.TempDir(), pack))
}

func TestAutoLSP_NoBinary_FallsBackToDefaults(t *testing.T) {
	got := AutoLSP(t.TempDir())
	require.Equal(t, len(DefaultProviders()), len(got))
}

func TestPack_StatsCarriesBuildDuration(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n\ngo 1.23\n"), 0o644))
	p, err := Build(stdctx.Background(), Options{
		Root: root, Task: "describe",
		Providers: []Provider{TestMapProvider{}},
	})
	require.NoError(t, err)
	require.NotEmpty(t, p.Hash)
	require.GreaterOrEqual(t, p.Stats.BuildDurationMs, int64(0))
}

func TestReadCache_Missing(t *testing.T) {
	_, ok := readCache(filepath.Join(t.TempDir(), "missing.json"))
	require.False(t, ok)
}
