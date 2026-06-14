package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrite_LineIsValidJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "events.jsonl")
	l, err := Open(p, 0)
	require.NoError(t, err)
	defer l.Close()

	require.NoError(t, l.Write("info", map[string]any{"stage": "init", "ok": true}))

	b, err := os.ReadFile(p)
	require.NoError(t, err)
	line := strings.TrimRight(string(b), "\n")
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &m))
	require.Equal(t, "info", m["level"])
	require.Equal(t, "init", m["stage"])
	require.Equal(t, true, m["ok"])
	require.NotEmpty(t, m["ts"])
}

func TestRotation(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "events.jsonl")
	l, err := Open(p, 128)
	require.NoError(t, err)
	defer l.Close()

	for i := 0; i < 20; i++ {
		require.NoError(t, l.Write("info", map[string]any{"i": i, "pad": strings.Repeat("x", 32)}))
	}

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	// at least the active file + 1 rotated archive
	var rotated int
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "events.jsonl.") {
			rotated++
		}
	}
	require.GreaterOrEqual(t, rotated, 1)
}

func TestWrite_AfterClose_Errors(t *testing.T) {
	p := filepath.Join(t.TempDir(), "events.jsonl")
	l, err := Open(p, 0)
	require.NoError(t, err)
	require.NoError(t, l.Close())
	require.Error(t, l.Write("info", nil))
}

// silence unused import warning when this file is the only user of bufio
var _ = bufio.ScanLines
