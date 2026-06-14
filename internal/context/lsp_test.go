// SPDX-License-Identifier: MIT

package context

import (
	stdctx "context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/adapters/lsp"
)

type fakeLSP struct {
	lang   string
	syms   []lsp.Symbol
	diags  []lsp.Diagnostic
	hit    bool
	errSym error
	errDia error
}

func (f fakeLSP) Language() string { return f.lang }
func (f fakeLSP) DocumentSymbols(_ stdctx.Context, _, _ string) ([]lsp.Symbol, bool, error) {
	return f.syms, f.hit, f.errSym
}
func (f fakeLSP) Diagnostics(_ stdctx.Context, _, _ string) ([]lsp.Diagnostic, bool, error) {
	return f.diags, f.hit, f.errDia
}
func (f fakeLSP) Definitions(_ stdctx.Context, _, _ string, _, _ int) ([]lsp.Symbol, bool, error) {
	return nil, false, nil
}
func (f fakeLSP) References(_ stdctx.Context, _, _ string, _, _ int) ([]lsp.Symbol, bool, error) {
	return nil, false, nil
}

func TestLSPProvider_PopulatesSymbolsAndDiags(t *testing.T) {
	pack := &Pack{RelevantFiles: []FileEntry{{Path: "a.go"}}}
	c := fakeLSP{
		lang:  "go",
		syms:  []lsp.Symbol{{Name: "Foo", Path: "a.go", Line: 1}},
		diags: []lsp.Diagnostic{{Path: "a.go", Line: 2, Severity: "error", Message: "boom"}},
		hit:   true,
	}
	require.NoError(t, LSPProvider{Clients: []lsp.Client{c}}.Apply(stdctx.Background(), "/x", pack))
	require.Len(t, pack.LSPSymbols, 1)
	require.Equal(t, "Foo", pack.LSPSymbols[0].Name)
	require.Len(t, pack.LSPDiagnostics, 1)
	require.Equal(t, 2, pack.Stats.LSPQueries)
	require.Equal(t, 2, pack.Stats.LSPCacheHits)
}

func TestLSPProvider_SymbolError_SkipsClient(t *testing.T) {
	pack := &Pack{RelevantFiles: []FileEntry{{Path: "a.go"}}}
	c := fakeLSP{lang: "go", errSym: errors.New("boom")}
	require.NoError(t, LSPProvider{Clients: []lsp.Client{c}}.Apply(stdctx.Background(), "/x", pack))
	require.Empty(t, pack.LSPSymbols)
}

func TestAutoLSP_NoBinaries_ReturnsDefaults(t *testing.T) {
	chain := AutoLSP(t.TempDir())
	// Without LSP binaries on PATH this should equal the default chain.
	require.Equal(t, len(DefaultProviders()), len(chain))
}

func TestAnyFile_Hit(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module x")
	require.True(t, anyFile(root, []string{"package.json", "go.mod"}))
	require.False(t, anyFile(root, []string{"package.json", "Cargo.toml"}))
}

func TestAnyPath_Negative(t *testing.T) {
	require.False(t, anyPath([]string{"absolutely-not-a-real-binary-xyz"}))
}

func TestStartable_NilClient(t *testing.T) {
	// non-Stdio fake: startable should return nil
	require.NoError(t, startable(fakeLSP{}))
}

func TestBuild_WithLSP_Provider(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "x.go", "package x\n")
	c := fakeLSP{lang: "go", syms: []lsp.Symbol{{Name: "X", Path: "x.go", Line: 1}}}
	opts := Options{
		Root: root, Task: "anything",
		Providers: []Provider{TestMapProvider{}, LSPProvider{Clients: []lsp.Client{c}}},
	}
	p, err := Build(stdctx.Background(), opts)
	require.NoError(t, err)
	require.NotEmpty(t, p.Hash)
}

func TestCanonicalHash_ChangesWithTask(t *testing.T) {
	h1 := canonicalHash("a", nil, []Provider{TestMapProvider{}}, "/r")
	h2 := canonicalHash("b", nil, []Provider{TestMapProvider{}}, "/r")
	require.NotEqual(t, h1, h2)
}

func TestGitHead_MissingReturnsEmpty(t *testing.T) {
	require.Equal(t, "", gitHead(t.TempDir()))
}

func TestGitHead_Present(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, filepath.Join(".git", "HEAD"), "ref: refs/heads/main\n")
	require.NotEmpty(t, gitHead(root))
}
