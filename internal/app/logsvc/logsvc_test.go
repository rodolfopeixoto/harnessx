package logsvc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrint_Missing(t *testing.T) {
	var out bytes.Buffer
	require.NoError(t, Print(Options{Path: filepath.Join(t.TempDir(), "missing.jsonl")}, &out))
	require.Contains(t, out.String(), "no log file")
}

func TestPrint_TailLimits(t *testing.T) {
	p := filepath.Join(t.TempDir(), "events.jsonl")
	require.NoError(t, os.WriteFile(p, []byte("a\nb\nc\nd\ne\n"), 0o644))

	var out bytes.Buffer
	require.NoError(t, Print(Options{Path: p, Tail: 2}, &out))
	got := strings.TrimSpace(out.String())
	require.Equal(t, "d\ne", got)
}
