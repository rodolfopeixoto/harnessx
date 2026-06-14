// SPDX-License-Identifier: MIT

package audit

import (
	"context"
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

func TestFileSink_ListMissingReturnsEmpty(t *testing.T) {
	out, err := (&FileSink{Path: filepath.Join(t.TempDir(), "missing.jsonl")}).List(context.Background())
	require.NoError(t, err)
	require.Empty(t, out)
}
