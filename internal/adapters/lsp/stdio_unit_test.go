// SPDX-License-Identifier: MIT

package lsp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathToURI_AndBack(t *testing.T) {
	u := pathToURI("/tmp/foo.go")
	require.True(t, strings.HasPrefix(u, "file://"))
	require.Contains(t, u, "/tmp/foo.go")
	require.Equal(t, "/tmp/foo.go", uriToPath(u))
}

func TestURIToPath_InvalidReturnsInput(t *testing.T) {
	require.Equal(t, ":::not-a-uri:::", uriToPath(":::not-a-uri:::"))
}

func TestSeverityName(t *testing.T) {
	cases := map[int]string{1: "error", 2: "warning", 3: "information", 4: "hint", 99: "info"}
	for k, v := range cases {
		require.Equal(t, v, severityName(k))
	}
}

func TestRepoAndQueryHash_Stable(t *testing.T) {
	require.Equal(t, repoHash("/x"), repoHash("/x"))
	require.NotEqual(t, repoHash("/x"), repoHash("/y"))
	require.Equal(t, queryHash("m", "p"), queryHash("m", "p"))
	require.NotEqual(t, queryHash("m", "a"), queryHash("m", "b"))
}

func TestCacheDir_AndKey(t *testing.T) {
	dir := CacheDir("/root", "abc", "go")
	require.Equal(t, filepath.Join("/root", ".harness", "cache", "lsp", "abc", "go"), dir)
	key := CacheKey("/root", "abc", "go", "q1")
	require.Equal(t, filepath.Join(dir, "q1.json"), key)
}

func TestCacheRoundtrip(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "deep", "nested", "cache.json")
	want := cachedResult{Symbols: []Symbol{{Name: "X", Path: "f.go", Line: 7}}}
	require.NoError(t, writeCache(tmp, want))
	got, ok := readCache(tmp)
	require.True(t, ok)
	require.Equal(t, want.Symbols, got.Symbols)
}

func TestReadCache_Missing(t *testing.T) {
	_, ok := readCache(filepath.Join(t.TempDir(), "nope.json"))
	require.False(t, ok)
}

func TestReadCache_Malformed(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.json")
	require.NoError(t, os.WriteFile(p, []byte("not-json"), 0o644))
	_, ok := readCache(p)
	require.False(t, ok)
}

func TestReadMessage_OK(t *testing.T) {
	payload := []byte(`{"jsonrpc":"2.0"}`)
	buf := &bytes.Buffer{}
	buf.WriteString("Content-Length: ")
	buf.WriteString("17\r\n\r\n")
	buf.Write(payload)
	msg, err := readMessage(bufio.NewReader(buf))
	require.NoError(t, err)
	require.Equal(t, payload, msg)
}

func TestReadMessage_MissingHeader(t *testing.T) {
	_, err := readMessage(bufio.NewReader(strings.NewReader("\r\n")))
	require.Error(t, err)
}

func TestParseDocumentSymbols_FlatRaw(t *testing.T) {
	raw := json.RawMessage(`[{"name":"Foo","location":{"range":{"start":{"line":4}}}}]`)
	syms := parseDocumentSymbols(raw, "x.go")
	require.Len(t, syms, 1)
	require.Equal(t, "Foo", syms[0].Name)
	require.Equal(t, 5, syms[0].Line)
}

func TestParseDocumentSymbols_HierarchicalRaw(t *testing.T) {
	raw := json.RawMessage(`[{"name":"Outer","range":{"start":{"line":1}},"children":[{"name":"Inner","range":{"start":{"line":2}}}]}]`)
	syms := parseDocumentSymbols(raw, "x.go")
	require.Len(t, syms, 2)
	require.Equal(t, "Outer", syms[0].Name)
	require.Equal(t, "Outer.Inner", syms[1].Name)
}

func TestParseDocumentSymbols_Empty(t *testing.T) {
	require.Nil(t, parseDocumentSymbols(json.RawMessage(`[]`), "x"))
	require.Nil(t, parseDocumentSymbols(json.RawMessage(`not-json`), "x"))
}

func TestStdio_LanguageField(t *testing.T) {
	c := NewStdio("/nonexistent-binary-xyz", nil, "go", "go", "/tmp")
	require.Equal(t, "go", c.Language())
}

func TestStdio_StartMissingBinary(t *testing.T) {
	c := NewStdio("/nonexistent-binary-xyz", nil, "go", "go", "/tmp")
	err := c.Start(context.Background())
	require.Error(t, err)
}

func TestStdio_CloseBeforeStartIsNoop(t *testing.T) {
	c := NewStdio("noop", nil, "go", "go", "/tmp")
	require.NoError(t, c.Close())
	// Idempotent.
	require.NoError(t, c.Close())
}

func TestStdio_StubbedLifecycle(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses /bin/cat")
	}
	c := NewStdio("/bin/cat", nil, "stub", "plaintext", "/tmp")
	// Skip actual Start (would block on initialize handshake); exercise
	// Definitions/References which are documented no-ops.
	defs, hit, err := c.Definitions(context.Background(), "/tmp", "x", 1, 1)
	require.Nil(t, defs)
	require.False(t, hit)
	require.NoError(t, err)
	refs, hit, err := c.References(context.Background(), "/tmp", "x", 1, 1)
	require.Nil(t, refs)
	require.False(t, hit)
	require.NoError(t, err)
}

func TestStdio_CachedRead(t *testing.T) {
	root := t.TempDir()
	c := NewStdio("noop", nil, "go", "go", root)
	cached := cachedResult{Symbols: []Symbol{{Name: "Cached", Path: "x", Line: 1}}}
	key := queryHash("documentSymbol", "x")
	require.NoError(t, writeCache(CacheKey(root, repoHash(root), "go", key), cached))
	got, hit, err := c.DocumentSymbols(context.Background(), root, "x")
	require.NoError(t, err)
	require.True(t, hit)
	require.Len(t, got, 1)
}

func TestStdio_CachedDiagnostics(t *testing.T) {
	root := t.TempDir()
	c := NewStdio("noop", nil, "go", "go", root)
	cached := cachedResult{Diagnostics: []Diagnostic{{Path: "x", Line: 2, Severity: "error", Message: "boom"}}}
	key := queryHash("diagnostics", "x")
	require.NoError(t, writeCache(CacheKey(root, repoHash(root), "go", key), cached))
	got, hit, err := c.Diagnostics(context.Background(), root, "x")
	require.NoError(t, err)
	require.True(t, hit)
	require.Equal(t, "boom", got[0].Message)
}

func TestStdio_DocumentSymbols_MissingFile(t *testing.T) {
	root := t.TempDir()
	c := NewStdio("noop", nil, "go", "go", root)
	_, _, err := c.DocumentSymbols(context.Background(), root, "does-not-exist.go")
	require.Error(t, err)
}

func TestRPCError_String(t *testing.T) {
	err := &rpcError{Code: -32601, Message: "method not found"}
	require.Equal(t, "method not found", err.Message)
}

func TestSendOnClosedClient(t *testing.T) {
	c := NewStdio("noop", nil, "go", "go", "/tmp")
	c.closed = true
	err := c.send(map[string]any{"jsonrpc": "2.0"})
	require.True(t, errors.Is(err, errors.New("lsp: closed")) || err.Error() == "lsp: closed")
}
