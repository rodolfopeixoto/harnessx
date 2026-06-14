// SPDX-License-Identifier: MIT

package auditrun

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func TestRunner_GeneratesArtifactsInSkipMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	root := t.TempDir()
	var buf bytes.Buffer
	r := New(Options{
		RepoRoot:       root,
		BaseURL:        srv.URL,
		Timestamp:      "20260614T120000Z",
		Out:            &buf,
		PlaywrightSkip: true,
	})
	summary, err := r.Run(context.Background())
	require.NoError(t, err)
	require.Greater(t, summary.TotalResults, 0)

	layout := r.Layout()
	for _, must := range []string{
		filepath.Join(layout.JSON, constants.AuditFeatureMapFile),
		filepath.Join(layout.JSON, constants.AuditResultsFile),
		filepath.Join(layout.JSON, constants.AuditSummaryFile),
		filepath.Join(layout.JSON, constants.AuditConsoleFile),
		filepath.Join(layout.JSON, constants.AuditNetworkFile),
		filepath.Join(layout.JSON, constants.AuditSelectorsFile),
		filepath.Join(layout.JSON, constants.AuditVisualDiffFile),
		filepath.Join(layout.JSON, constants.AuditLayoutFile),
		filepath.Join(layout.ReportDir, constants.AuditBacklogFile),
		filepath.Join(layout.ReportDir, constants.AuditHTMLFile),
		layout.RunLog,
	} {
		_, err := os.Stat(must)
		require.NoErrorf(t, err, "missing artifact: %s", must)
	}

	body, err := os.ReadFile(filepath.Join(layout.JSON, constants.AuditSummaryFile))
	require.NoError(t, err)
	var sum Summary
	require.NoError(t, json.Unmarshal(body, &sum))
	require.Equal(t, srv.URL, sum.BaseURL)
	require.Equal(t, summary.TotalResults, sum.Counts[constants.AuditStatusNotImplemented])

	require.Contains(t, buf.String(), "Audit finished")
}

func TestRunner_FilterByFeatureLimitsScope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := New(Options{
		RepoRoot:       t.TempDir(),
		BaseURL:        srv.URL,
		Timestamp:      "20260614T120100Z",
		OnlyFeature:    "p01-public-landing",
		PlaywrightSkip: true,
	})
	summary, err := r.Run(context.Background())
	require.NoError(t, err)
	require.Equal(t, 3, summary.TotalResults)
}

func TestSynthesiseSkipped_RespectsFeatureViewports(t *testing.T) {
	results := synthesiseSkipped([]Feature{{ID: "x", Viewports: []string{constants.AuditViewportDesk}}}, DefaultViewports(), "skip")
	require.Len(t, results, 1)
	require.Equal(t, constants.AuditViewportDesk, results[0].Viewport)
}

func TestBaseAddr_StripsScheme(t *testing.T) {
	require.Equal(t, "127.0.0.1:7373", baseAddr("http://127.0.0.1:7373"))
	require.Equal(t, "127.0.0.1:7373", baseAddr("https://127.0.0.1:7373/api/health"))
}
