// SPDX-License-Identifier: MIT

package containers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseDockerPS_HappyPath(t *testing.T) {
	in := `{"ID":"abc","Names":"web","Image":"nginx","Status":"Up 2 minutes","CreatedAt":"2026-06-14 10:00:00 -0300 -03"}
{"ID":"def","Names":"db","Image":"postgres","Status":"Exited (0) 1 hour ago","CreatedAt":"2026-06-14 09:00:00 -0300 -03"}`
	items, err := parseDockerPS(in)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "web", items[0].Name)
}

func TestParseDockerPS_SkipsMalformed(t *testing.T) {
	items, err := parseDockerPS("not-json\n{\"ID\":\"x\",\"Names\":\"y\"}")
	require.NoError(t, err)
	require.Len(t, items, 1)
}

type stubLister struct {
	items []Item
	err   error
}

func (s stubLister) List(_ context.Context) ([]Item, error) {
	return s.items, s.err
}

func TestVerifyClean_OkWhenEmpty(t *testing.T) {
	require.NoError(t, VerifyClean(context.Background(), stubLister{}))
}

func TestVerifyClean_ErrorsWhenContainersPresent(t *testing.T) {
	err := VerifyClean(context.Background(), stubLister{items: []Item{{ID: "x", Name: "leftover"}}})
	require.Error(t, err)
}

func TestVerifyClean_PropagatesListerError(t *testing.T) {
	err := VerifyClean(context.Background(), stubLister{err: errors.New("boom")})
	require.Error(t, err)
}

func TestHealthProbe_SuccessAfterDelay(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		if hits < 2 {
			http.Error(w, "starting", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	probe := HealthProbe{
		URL:     srv.URL,
		Client:  &http.Client{Timeout: time.Second},
		Timeout: 3 * time.Second,
		Backoff: 50 * time.Millisecond,
	}
	require.NoError(t, probe.Wait(context.Background()))
	require.GreaterOrEqual(t, hits, 2)
}

func TestHealthProbe_TimesOutOnFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "down", http.StatusInternalServerError)
	}))
	defer srv.Close()
	probe := HealthProbe{
		URL:     srv.URL,
		Client:  &http.Client{Timeout: 200 * time.Millisecond},
		Timeout: 300 * time.Millisecond,
		Backoff: 50 * time.Millisecond,
	}
	err := probe.Wait(context.Background())
	require.Error(t, err)
}

func TestCompose_FailsWithMissingBinary(t *testing.T) {
	c := Compose{Binary: "definitely-not-a-binary", File: "compose.yaml"}
	err := c.Up(context.Background())
	require.Error(t, err)
	var msg string
	if err != nil {
		msg = fmt.Sprintf("%v", err)
	}
	require.Contains(t, msg, "containers compose")
}
