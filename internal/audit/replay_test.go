// SPDX-License-Identifier: MIT

package audit

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func writeEvents(t *testing.T, path string, evs []Event) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range evs {
		require.NoError(t, enc.Encode(e))
	}
}

func TestReplay_HappyPath_ListsEventsForRun(t *testing.T) {
	root := t.TempDir()
	runID := "run-123"
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	writeEvents(t, filepath.Join(root, ".harness", "audit", "events.jsonl"), []Event{
		{ID: runID, Kind: "hook", Source: "pre", Subject: "pre-run", OccurredAt: base},
		{ID: "other-run", Kind: "hook", Source: "pre", Subject: "unrelated", OccurredAt: base},
		{ID: runID, Kind: "sensor", Source: "tests", Subject: "sensor=passed", OccurredAt: base.Add(2 * time.Second)},
	})

	rep, err := Replay(context.Background(), root, runID, ReplayOptions{DryRun: true, Now: func() time.Time { return base }})
	require.NoError(t, err)
	require.Equal(t, runID, rep.RunID)
	require.True(t, rep.DryRun)
	require.Equal(t, 2, rep.Events)
	require.Len(t, rep.Steps, 2)
	// Chronological order enforced by replay.
	require.Equal(t, "hook", rep.Steps[0].Kind)
	require.Equal(t, "sensor", rep.Steps[1].Kind)
	require.Contains(t, rep.Steps[0].Action, "would replay")
	// TmpDir must exist and live under the OS temp root.
	st, err := os.Stat(rep.TmpDir)
	require.NoError(t, err)
	require.True(t, st.IsDir())
	require.True(t, strings.HasPrefix(rep.TmpDir, os.TempDir()) || strings.HasPrefix(rep.TmpDir, "/private"+os.TempDir()))
}

func TestReplay_RunNotFound(t *testing.T) {
	root := t.TempDir()
	_, err := Replay(context.Background(), root, "no-such-run", ReplayOptions{DryRun: true})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrRunNotFound))
}

func TestReplay_DryRunVsMutatingSameShape(t *testing.T) {
	root := t.TempDir()
	runID := "run-abc"
	base := time.Date(2026, 3, 4, 10, 0, 0, 0, time.UTC)
	writeEvents(t, filepath.Join(root, ".harness", "logs", "events.jsonl"), []Event{
		{ID: runID, Kind: "agent", Source: "claude", Subject: "invoked", OccurredAt: base},
	})

	dry, err := Replay(context.Background(), root, runID, ReplayOptions{DryRun: true, Now: func() time.Time { return base }})
	require.NoError(t, err)
	wet, err := Replay(context.Background(), root, runID, ReplayOptions{DryRun: false, Now: func() time.Time { return base }})
	require.NoError(t, err)

	require.Equal(t, dry.Events, wet.Events)
	require.Equal(t, len(dry.Steps), len(wet.Steps))
	require.Equal(t, dry.Steps[0].Kind, wet.Steps[0].Kind)
	require.Equal(t, dry.Steps[0].Subject, wet.Steps[0].Subject)
	// The verb differs; that's the only visible surface change.
	require.Contains(t, dry.Steps[0].Action, "would replay")
	require.Contains(t, wet.Steps[0].Action, "replayed")

	// Mutating run writes a manifest; dry run must not.
	_, err = os.Stat(filepath.Join(wet.TmpDir, "manifest.txt"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(dry.TmpDir, "manifest.txt"))
	require.True(t, os.IsNotExist(err))
}

func TestReplay_JSONShapeStable(t *testing.T) {
	root := t.TempDir()
	runID := "run-json"
	base := time.Date(2026, 5, 6, 8, 0, 0, 0, time.UTC)
	writeEvents(t, filepath.Join(root, ".harness", "audit", "events.jsonl"), []Event{
		{ID: runID, Kind: "hook", Source: "pre", Subject: "start", OccurredAt: base},
	})
	rep, err := Replay(context.Background(), root, runID, ReplayOptions{DryRun: true, TmpDir: filepath.Join(root, "tmp"), Now: func() time.Time { return base }})
	require.NoError(t, err)

	b, err := json.Marshal(rep)
	require.NoError(t, err)
	// Assert stable field names so downstream consumers can rely on
	// them.
	for _, key := range []string{`"run_id"`, `"dry_run"`, `"tmp_dir"`, `"started_at"`, `"steps"`, `"events"`, `"index"`, `"occurred_at"`, `"kind"`, `"source"`, `"subject"`, `"action"`} {
		require.Contains(t, string(b), key, "expected JSON key %s in %s", key, string(b))
	}
}

func TestReplay_RunDirWithoutEventsIsNotError(t *testing.T) {
	root := t.TempDir()
	runID := "run-empty"
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".harness", "runs", runID), 0o755))

	rep, err := Replay(context.Background(), root, runID, ReplayOptions{DryRun: true})
	require.NoError(t, err)
	require.Equal(t, 0, rep.Events)
	require.NotEmpty(t, rep.Notes)
}
