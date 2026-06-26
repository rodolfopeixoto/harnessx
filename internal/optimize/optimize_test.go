package optimize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanDockerfile_FindsCommonIssues(t *testing.T) {
	root := t.TempDir()
	body := `FROM ubuntu
RUN apt-get update && apt-get install -y curl
RUN apt-get install -y vim
RUN apt-get install -y git
COPY . /app
COPY scripts /app/scripts
COPY config /app/config
COPY assets /app/assets
COPY src /app/src
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "Dockerfile"), []byte(body), 0o644))

	d := scanDockerfile(root)
	require.NotNil(t, d)
	require.Equal(t, 1, d.Stages)
	require.Equal(t, 3, d.RunSteps)
	require.False(t, d.HasUSER)
	require.False(t, d.HasHealthcheck)
	require.True(t, d.UsesLatestTag)

	ids := map[string]bool{}
	for _, f := range d.Findings {
		ids[f.ID] = true
	}
	require.True(t, ids["docker.latest_tag"])
	require.True(t, ids["docker.no_user"])
	require.True(t, ids["docker.no_healthcheck"])
	require.True(t, ids["docker.no_cache_cleanup"])
}

func TestScanLogCallSites_FindsConsoleLog(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "app.ts"),
		[]byte("console.log('hi');\nconst x = 1;\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "app.test.ts"),
		[]byte("console.log('skipped');\n"), 0o644))
	hits := scanLogCallSites(root)
	require.Len(t, hits, 1)
	require.Equal(t, "app.ts", hits[0].Path)
	require.Equal(t, "console.log", hits[0].Kind)
}

func TestRemovalCandidate_Conservative(t *testing.T) {
	ok, _ := removalCandidate("node", "react")
	require.False(t, ok)
	ok, _ = removalCandidate("node", "eslint")
	require.True(t, ok)
}

func TestKeepReason_Observability(t *testing.T) {
	_, ok := keepReason("node", "@sentry/node")
	require.True(t, ok)
}

func TestCapture_WritesSnapshot(t *testing.T) {
	root := t.TempDir()
	s, path, err := Capture(SnapshotOptions{Root: root, Label: "baseline"})
	require.NoError(t, err)
	require.FileExists(t, path)
	require.Equal(t, "baseline", s.Label)
}

func TestCompare_DetectsRegression(t *testing.T) {
	a := Snapshot{Deps: DepsMetrics{Total: 10}, Logs: LogsMetrics{TotalCallSites: 0}}
	b := Snapshot{Deps: DepsMetrics{Total: 15}, Logs: LogsMetrics{TotalCallSites: 3}}
	d := Compare(a, b)
	status := map[string]string{}
	for _, r := range d.Rows {
		status[r.Metric] = r.Status
	}
	require.Equal(t, "regressed", status["deps_total"])
	require.Equal(t, "regressed", status["noisy_log_call_sites"])
}

func TestScanLogCallSites_ExcludesVenv_BUG23(t *testing.T) {
	root := t.TempDir()
	noisy := filepath.Join(root, ".venv", "lib", "python3.12", "site-packages", "noisy")
	if err := os.MkdirAll(noisy, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(noisy, "verbose.py"), []byte("print(\"oh\")\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hits := scanLogCallSites(root)
	if len(hits) != 0 {
		t.Fatalf("scanner should ignore .venv, got %d hits: %+v", len(hits), hits)
	}
}
