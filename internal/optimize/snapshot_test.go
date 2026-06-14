// SPDX-License-Identifier: MIT

package optimize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCapture_NoProject_ReturnsDefaults(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".harness", "artifacts", "perf"), 0o755))
	s, path, err := Capture(SnapshotOptions{Root: root, Label: "baseline"})
	require.NoError(t, err)
	require.NotEmpty(t, path)
	require.Equal(t, "baseline", s.Label)
	require.Equal(t, root, s.Root)
	require.NotZero(t, s.Runtime.ProcessNumGoroutines)
}

func TestLatestTwo_NeedsTwo(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".harness", "artifacts", "perf"), 0o755))
	_, _, err := Capture(SnapshotOptions{Root: root, Label: "first"})
	require.NoError(t, err)
	_, _, _, _, err = LatestTwo(root)
	require.Error(t, err)

	_, _, err = Capture(SnapshotOptions{Root: root, Label: "second"})
	require.NoError(t, err)
	first, second, fname, sname, err := LatestTwo(root)
	require.NoError(t, err)
	require.Equal(t, "first", first.Label)
	require.Equal(t, "second", second.Label)
	require.NotEmpty(t, fname)
	require.NotEmpty(t, sname)
}

func TestWriteSnapshotReport_HasMetricsBlock(t *testing.T) {
	root := t.TempDir()
	s := Snapshot{Root: root, Label: "x"}
	path, err := WriteSnapshotReport(root, s)
	require.NoError(t, err)
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	body := string(b)
	for _, want := range []string{"# Executive Summary", "# Metrics", "# Risks", "# Rollback Plan"} {
		require.Containsf(t, body, want, "missing %q", want)
	}
}

func TestCompare_FlagsRegressions(t *testing.T) {
	a := Snapshot{Deps: DepsMetrics{Total: 5}, Logs: LogsMetrics{TotalCallSites: 0}}
	b := Snapshot{Deps: DepsMetrics{Total: 8}, Logs: LogsMetrics{TotalCallSites: 4}}
	d := Compare(a, b)
	statuses := map[string]string{}
	for _, r := range d.Rows {
		statuses[r.Metric] = r.Status
	}
	require.Equal(t, "regressed", statuses["deps_total"])
	require.Equal(t, "regressed", statuses["noisy_log_call_sites"])
}
