// SPDX-License-Identifier: MIT

package detectors

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/cleanup"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/runtime/containers"
)

func TestWorktrees_DiscoversChildrenAndRisk(t *testing.T) {
	root := t.TempDir()
	wtDir := filepath.Join(root, ".git", "worktrees", "branch-a")
	require.NoError(t, os.MkdirAll(wtDir, 0o755))
	stale := filepath.Join(root, ".git", "worktrees", "stale")
	require.NoError(t, os.MkdirAll(stale, 0o755))
	old := time.Now().Add(-time.Duration(constants.CleanupStaleThresholdHours+1) * time.Hour)
	require.NoError(t, os.Chtimes(stale, old, old))

	out, err := Worktrees{}.Detect(context.Background(), root)
	require.NoError(t, err)
	require.Len(t, out, 2)
	risks := map[string]cleanup.Risk{}
	for _, f := range out {
		risks[filepath.Base(f.Path)] = f.Risk
	}
	require.Equal(t, cleanup.RiskMedium, risks["branch-a"])
	require.Equal(t, cleanup.RiskHigh, risks["stale"])
}

func TestCaches_FindsKnownPaths(t *testing.T) {
	root := t.TempDir()
	cache := filepath.Join(root, ".npm")
	require.NoError(t, os.MkdirAll(cache, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cache, "file"), []byte("x"), 0o644))
	out, err := Caches{}.Detect(context.Background(), root)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, constants.KindCleanupCache, out[0].Kind)
}

func TestAbandonedHarness_FlagsMissingDB(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, constants.HarnessDir), 0o755))
	out, err := AbandonedHarness{}.Detect(context.Background(), root)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, cleanup.RiskHigh, out[0].Risk)
}

func TestLargeFiles_RespectsThreshold(t *testing.T) {
	root := t.TempDir()
	huge := filepath.Join(root, "blob.bin")
	require.NoError(t, os.WriteFile(huge, make([]byte, 1024), 0o644))
	out, err := LargeFiles{Threshold: 256}.Detect(context.Background(), root)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, huge, out[0].Path)
}

func TestVMLeftovers_AndClaudeLeftovers(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".vagrant"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".claude"), 0o755))
	vm, err := VMLeftovers{}.Detect(context.Background(), root)
	require.NoError(t, err)
	require.Len(t, vm, 1)
	cl, err := ClaudeLeftovers{}.Detect(context.Background(), root)
	require.NoError(t, err)
	require.Len(t, cl, 1)
}

type stubLister struct {
	items []containers.Item
	err   error
}

func (s stubLister) List(_ context.Context) ([]containers.Item, error) {
	return s.items, s.err
}

func TestContainers_RiskBucketing(t *testing.T) {
	d := Containers{Lister: stubLister{items: []containers.Item{
		{ID: "abc", Name: "n1", Status: "Up", CreatedAt: time.Now()},
		{ID: "def", Name: "n2", Status: "Exited (0)", CreatedAt: time.Now()},
		{ID: "ghi", Name: "n3", Status: "Exited (137)", CreatedAt: time.Now().Add(-time.Duration(constants.CleanupStaleThresholdHours+1) * time.Hour)},
	}}}
	out, err := d.Detect(context.Background(), "/")
	require.NoError(t, err)
	require.Len(t, out, 3)
	risks := map[string]cleanup.Risk{}
	for _, f := range out {
		risks[f.Path] = f.Risk
	}
	require.Equal(t, cleanup.RiskLow, risks["abc"])
	require.Equal(t, cleanup.RiskMedium, risks["def"])
	require.Equal(t, cleanup.RiskHigh, risks["ghi"])
}

func TestContainers_ListerError(t *testing.T) {
	_, err := (Containers{Lister: stubLister{err: errors.New("boom")}}).Detect(context.Background(), "/")
	require.Error(t, err)
}
