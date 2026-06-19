// SPDX-License-Identifier: MIT

package audit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemorySink_WriteList(t *testing.T) {
	sink := NewMemorySink()
	require.NoError(t, sink.Write(context.Background(), Event{Kind: "k", Subject: "a"}))
	require.NoError(t, sink.Write(context.Background(), Event{Kind: "k", Subject: "b", OccurredAt: time.Now().Add(time.Minute)}))
	out, err := sink.List(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.Equal(t, "b", out[0].Subject)
}

func TestFileSink_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	sink := &FileSink{Path: path}
	require.NoError(t, sink.Write(context.Background(), Event{Kind: "settings_changed", Subject: "autonomy"}))
	require.NoError(t, sink.Write(context.Background(), Event{Kind: "plan_approved", Subject: "feature-x"}))
	out, err := sink.List(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 2)
}

func TestEvent_UnmarshalRunLogShape(t *testing.T) {
	line := `{"level":"info","root":"/p","run_id":"R1","sensor":"py_pytest","stage":"sensor","status":"passed","ts":"2026-06-19T02:20:26.421366Z"}`
	var ev Event
	require.NoError(t, json.Unmarshal([]byte(line), &ev))
	require.Equal(t, "R1", ev.ID)
	require.Equal(t, "sensor", ev.Kind)
	require.Equal(t, "py_pytest", ev.Source)
	require.Equal(t, "py_pytest=passed", ev.Subject)
	require.False(t, ev.OccurredAt.IsZero())
	require.Equal(t, 2026, ev.OccurredAt.Year())
}

func TestFileSink_LoadsRunLogTimestamps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	body := `{"level":"info","stage":"init","ts":"2026-06-19T02:18:34.442122Z"}
{"level":"info","stage":"sensor","sensor":"changed_files","status":"passed","ts":"2026-06-19T02:20:26.421366Z"}
`
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	out, err := (&FileSink{Path: path}).List(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 2)
	for _, e := range out {
		require.False(t, e.OccurredAt.IsZero(), "row %+v lost its timestamp", e)
	}
}

func TestFileSink_ListMissingReturnsEmpty(t *testing.T) {
	out, err := (&FileSink{Path: filepath.Join(t.TempDir(), "missing.jsonl")}).List(context.Background())
	require.NoError(t, err)
	require.Empty(t, out)
}
