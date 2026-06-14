package lsp

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// goplsAvailable is true when the binary is on PATH; tests that need to
// actually exercise the server are skipped otherwise (CI minimal install).
func goplsAvailable() bool {
	_, err := exec.LookPath("gopls")
	return err == nil
}

func TestPathToURI_AbsoluteFile(t *testing.T) {
	uri := pathToURI("/tmp/x.go")
	require.True(t, strings.HasPrefix(uri, "file://"))
	require.Contains(t, uri, "/tmp/x.go")
}

func TestParseDocumentSymbols_Hierarchical(t *testing.T) {
	raw := json.RawMessage(`[
	  {"name":"Foo","range":{"start":{"line":0}},"children":[
	    {"name":"Method","range":{"start":{"line":2}}}
	  ]}
	]`)
	syms := parseDocumentSymbols(raw, "x.go")
	require.GreaterOrEqual(t, len(syms), 2)
	require.Equal(t, "Foo", syms[0].Name)
	require.Equal(t, 1, syms[0].Line)
	require.Equal(t, "Foo.Method", syms[1].Name)
}

func TestParseDocumentSymbols_Flat(t *testing.T) {
	raw := json.RawMessage(`[
	  {"name":"Bar","location":{"range":{"start":{"line":4}}}}
	]`)
	syms := parseDocumentSymbols(raw, "y.go")
	require.Len(t, syms, 1)
	require.Equal(t, "Bar", syms[0].Name)
	require.Equal(t, 5, syms[0].Line)
}

func TestCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "cache.json")
	require.NoError(t, writeCache(p, cachedResult{
		Symbols: []Symbol{{Name: "x", Path: "a.go", Line: 1}},
	}))
	got, ok := readCache(p)
	require.True(t, ok)
	require.Len(t, got.Symbols, 1)
}

func TestStart_MissingGopls(t *testing.T) {
	if goplsAvailable() {
		t.Skip("gopls present — this test only meaningful when missing")
	}
	g := NewGopls(t.TempDir())
	err := g.Start(context.Background())
	require.Error(t, err)
}

func TestGopls_LiveDocumentSymbols(t *testing.T) {
	if !goplsAvailable() {
		t.Skip("gopls not installed")
	}
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module sample\n\ngo 1.21\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "main.go"), []byte(`package main

func Greet(name string) string { return "hi " + name }

func main() {}
`), 0o644))

	g := NewGopls(root)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	require.NoError(t, g.Start(ctx))
	defer g.Close()

	syms, _, err := g.DocumentSymbols(ctx, root, "main.go")
	require.NoError(t, err)
	names := map[string]bool{}
	for _, s := range syms {
		names[s.Name] = true
	}
	require.True(t, names["Greet"] || names["main"], "got %v", syms)

	// Cache hit on second call.
	_, hit, err := g.DocumentSymbols(ctx, root, "main.go")
	require.NoError(t, err)
	require.True(t, hit)
}

// guard against unused-import warnings if the file shrinks later.
var _ = io.EOF
