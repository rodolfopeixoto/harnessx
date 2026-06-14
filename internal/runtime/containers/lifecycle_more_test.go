// SPDX-License-Identifier: MIT

package containers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func TestNewCompose_DefaultBinary(t *testing.T) {
	c := NewCompose("/tmp/compose.yaml")
	require.Equal(t, constants.DefaultDockerBinary, c.Binary)
	require.Equal(t, "/tmp/compose.yaml", c.File)
}

func TestCompose_BinaryFallback(t *testing.T) {
	require.Equal(t, constants.DefaultDockerBinary, Compose{File: "x"}.binary())
	require.Equal(t, "podman", Compose{Binary: "podman", File: "x"}.binary())
}

func TestCompose_DownPropagatesMissingBinary(t *testing.T) {
	c := Compose{Binary: "definitely-not-a-binary", File: "compose.yaml"}
	err := c.Down(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "containers compose")
}

func TestRealLister_MissingDockerReturnsNil(t *testing.T) {
	items, err := RealLister{Binary: "definitely-not-a-binary"}.List(context.Background())
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestRealLister_BinaryFallback(t *testing.T) {
	require.Equal(t, constants.DefaultDockerBinary, RealLister{}.binary())
	require.Equal(t, "podman", RealLister{Binary: "podman"}.binary())
}

func TestParseDockerTime_RFC3339(t *testing.T) {
	got := parseDockerTime("2026-06-14T10:00:00Z")
	require.False(t, got.IsZero())
}

func TestParseDockerTime_Invalid(t *testing.T) {
	require.True(t, parseDockerTime("not-a-date").IsZero())
}

func TestNewHealthProbe_Defaults(t *testing.T) {
	p := NewHealthProbe("http://example.test")
	require.Equal(t, "http://example.test", p.URL)
	require.NotNil(t, p.Client)
	require.Equal(t, 30*time.Second, p.Timeout)
}

func TestHealthProbe_SucceedsImmediately(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	probe := HealthProbe{URL: srv.URL, Client: &http.Client{Timeout: time.Second}, Timeout: 2 * time.Second, Backoff: 50 * time.Millisecond}
	require.NoError(t, probe.Wait(context.Background()))
}

func TestHealthProbe_BadURLReturnsError(t *testing.T) {
	probe := HealthProbe{URL: "::bad::", Client: &http.Client{Timeout: 200 * time.Millisecond}, Timeout: 200 * time.Millisecond, Backoff: 50 * time.Millisecond}
	err := probe.Wait(context.Background())
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "timeout")
}

func TestVerifyClean_ListsLeakedContainers(t *testing.T) {
	err := VerifyClean(context.Background(), stubLister{items: []Item{{Name: "leftover-a"}, {Name: "leftover-b"}}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "leftover-a")
}
