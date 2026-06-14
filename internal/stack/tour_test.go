// SPDX-License-Identifier: MIT

package stack

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/runtime/containers"
)

func TestTour_HappyPath(t *testing.T) {
	repoRoot, err := filepath.Abs("../../")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tour := &Tour{
		Root:           t.TempDir(),
		TemplatesSrc:   filepath.Join(repoRoot, "templates"),
		RegistryPath:   filepath.Join(t.TempDir(), "registry.sqlite"),
		DashboardProbe: srv.URL,
		Probe: containers.HealthProbe{
			URL:     srv.URL,
			Client:  &http.Client{Timeout: time.Second},
			Timeout: time.Second,
			Backoff: 25 * time.Millisecond,
		},
	}
	var buf bytes.Buffer
	results, err := tour.Run(context.Background(), &buf)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	for _, step := range results {
		require.NoError(t, step.Err, step.Name)
	}
	require.Contains(t, buf.String(), "ok")
}

func TestTour_DashboardProbeFails(t *testing.T) {
	repoRoot, err := filepath.Abs("../../")
	require.NoError(t, err)
	tour := &Tour{
		Root:           t.TempDir(),
		TemplatesSrc:   filepath.Join(repoRoot, "templates"),
		RegistryPath:   filepath.Join(t.TempDir(), "registry.sqlite"),
		DashboardProbe: "http://127.0.0.1:1",
		Probe: containers.HealthProbe{
			URL:     "http://127.0.0.1:1",
			Client:  &http.Client{Timeout: 100 * time.Millisecond},
			Timeout: 200 * time.Millisecond,
			Backoff: 25 * time.Millisecond,
		},
	}
	var buf bytes.Buffer
	_, err = tour.Run(context.Background(), &buf)
	require.Error(t, err)
}

func TestTour_RunsWithoutProbe(t *testing.T) {
	repoRoot, err := filepath.Abs("../../")
	require.NoError(t, err)
	tour := &Tour{
		Root:         t.TempDir(),
		TemplatesSrc: filepath.Join(repoRoot, "templates"),
		RegistryPath: filepath.Join(t.TempDir(), "registry.sqlite"),
	}
	results, err := tour.Run(context.Background(), &bytes.Buffer{})
	require.NoError(t, err)
	require.NotEmpty(t, results)
}

func TestCopyTree_NoOpWhenEmpty(t *testing.T) {
	require.NoError(t, copyTree("", t.TempDir()))
}
