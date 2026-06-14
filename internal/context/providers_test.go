package context

import (
	stdctx "context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeIn(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(body), 0o644))
}

func TestAppendFile_Dedupes(t *testing.T) {
	got := appendFile(nil, FileEntry{Path: "a.go", Reason: "first"})
	got = appendFile(got, FileEntry{Path: "a.go", Reason: "second"})
	require.Len(t, got, 1)
	require.Equal(t, "first", got[0].Reason)
}

func TestExtractKeywords_AlphaCappedAndStable(t *testing.T) {
	got := extractKeywords("Build a Product Search WITH FILTERS for the Marketplace", 3)
	require.Len(t, got, 3)
	for _, w := range got {
		require.Greater(t, len(w), 2)
	}
}

func TestTestMapProvider_LinksByBaseName(t *testing.T) {
	root := t.TempDir()
	writeIn(t, root, ".harness/project/test-map.json", `{
	  "suites": [{"framework": "go-test", "files": ["main_test.go"]}],
	  "total_files": 1
	}`)
	pack := &Pack{RelevantFiles: []FileEntry{{Path: "main.go"}}}
	require.NoError(t, TestMapProvider{}.Apply(stdctx.Background(), root, pack))
	require.Contains(t, pack.RelatedTests, "main_test.go")
}

func TestMemoryProvider_NoDB_SkipsCleanly(t *testing.T) {
	root := t.TempDir()
	pack := &Pack{}
	require.NoError(t, MemoryProvider{}.Apply(stdctx.Background(), root, pack))
	require.Equal(t, 1, pack.Stats.ProvidersSkipped)
}

func TestLSPProvider_NoClients_Skips(t *testing.T) {
	pack := &Pack{}
	require.NoError(t, LSPProvider{}.Apply(stdctx.Background(), "/tmp", pack))
	require.Equal(t, 1, pack.Stats.ProvidersSkipped)
}
