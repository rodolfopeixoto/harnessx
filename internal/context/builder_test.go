package context

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(body), 0o644))
}

func TestBuild_HashStable_AndCacheHit(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module sample\n\ngo 1.23\n")
	writeFile(t, root, "main.go", "package main\n\nfunc main() {}\n")

	opts := Options{
		Root: root, Task: "explain main entry point",
		Providers: []Provider{TestMapProvider{}}, // deterministic, no git/rg dependency
	}
	p1, err := Build(context.Background(), opts)
	require.NoError(t, err)
	require.NotEmpty(t, p1.Hash)
	require.False(t, p1.Stats.CacheHit)

	p2, err := Build(context.Background(), opts)
	require.NoError(t, err)
	require.Equal(t, p1.Hash, p2.Hash)
	require.True(t, p2.Stats.CacheHit)
}

func TestBuild_Force_BustsCache(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module sample\n\ngo 1.23\n")
	opts := Options{Root: root, Task: "x", Providers: []Provider{TestMapProvider{}}}
	_, err := Build(context.Background(), opts)
	require.NoError(t, err)
	opts.Force = true
	p, err := Build(context.Background(), opts)
	require.NoError(t, err)
	require.False(t, p.Stats.CacheHit)
}

func TestExtractKeywords_DropsStopWordsAndShortTokens(t *testing.T) {
	kw := extractKeywords("Please add a search filter to the product list", 10)
	require.Contains(t, kw, "search")
	require.Contains(t, kw, "filter")
	require.NotContains(t, kw, "a")
	require.NotContains(t, kw, "the")
	require.NotContains(t, kw, "please")
}

func TestEnrichRelevantFiles_PopulatesBytesAndHash(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "x.go", "package x\n")
	p := &Pack{RelevantFiles: []FileEntry{{Path: "x.go", Reason: "test"}}}
	enrichRelevantFiles(root, p, dummyEstimator{})
	require.Equal(t, 10, p.RelevantFiles[0].Bytes)
	require.NotEmpty(t, p.RelevantFiles[0].SHA256)
	require.Greater(t, p.RelevantFiles[0].EstimatedTokens, 0)
}

type dummyEstimator struct{}

func (dummyEstimator) Estimate(s string) int { return len(s) }
