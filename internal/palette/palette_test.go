// SPDX-License-Identifier: MIT

package palette

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/catalog"
	"github.com/ropeixoto/harnessx/internal/catalog/kinds"
	"github.com/ropeixoto/harnessx/internal/workspace"
)

func TestScore_RankingOrder(t *testing.T) {
	require.Equal(t, 100, Score("exact", "exact"))
	require.Equal(t, 80, Score("ex", "example"))
	require.Equal(t, 60, Score("amp", "example"))
	require.True(t, Score("xyz", "example") == 0)
}

func TestSearch_EmptyQueryReturnsNothing(t *testing.T) {
	p := New(stubSource{name: "x", hits: []Hit{{Title: "y", Score: 50}}})
	out, err := p.Search(context.Background(), "")
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestSearch_SortsByScoreDescending(t *testing.T) {
	p := New(
		stubSource{name: "a", hits: []Hit{{Source: "a", Title: "low", Score: 10}}},
		stubSource{name: "b", hits: []Hit{{Source: "b", Title: "high", Score: 90}}},
	)
	out, err := p.Search(context.Background(), "anything")
	require.NoError(t, err)
	require.Equal(t, "high", out[0].Title)
}

func TestSearch_RespectsLimit(t *testing.T) {
	src := stubSource{name: "x"}
	for i := 0; i < 20; i++ {
		src.hits = append(src.hits, Hit{Source: "x", Title: "i", Score: 50})
	}
	out, err := New(src).WithLimit(5).Search(context.Background(), "i")
	require.NoError(t, err)
	require.Len(t, out, 5)
}

func TestProjectsSource_FiltersByRegistry(t *testing.T) {
	reg, err := workspace.Open(filepath.Join(t.TempDir(), "reg.sqlite"))
	require.NoError(t, err)
	defer reg.Close()
	dir := t.TempDir()
	_, err = reg.Add(context.Background(), dir, "Aurora", "aurora")
	require.NoError(t, err)
	hits, err := ProjectsSource{Registry: reg}.Search(context.Background(), "aurora")
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.Equal(t, "/workspace/aurora", hits[0].RouterPath)
}

func TestCapabilitiesSource_FindsBundled(t *testing.T) {
	c := catalog.New()
	for _, k := range kinds.All() {
		c.Register(k)
	}
	root, err := workspace.Open(":memory:")
	require.NoError(t, err)
	defer root.Close()
	hits, err := CapabilitiesSource{Catalog: c, Root: repoRoot(t)}.Search(context.Background(), "filesystem")
	require.NoError(t, err)
	require.NotEmpty(t, hits)
}

func TestCommandsSource_BuiltinCovered(t *testing.T) {
	hits, err := CommandsSource{Commands: BuiltinCommands}.Search(context.Background(), "settings")
	require.NoError(t, err)
	require.NotEmpty(t, hits)
}

type stubSource struct {
	name string
	hits []Hit
	err  error
}

func (s stubSource) Name() string                                      { return s.name }
func (s stubSource) Search(_ context.Context, _ string) ([]Hit, error) { return s.hits, s.err }

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../../")
	require.NoError(t, err)
	return dir
}
