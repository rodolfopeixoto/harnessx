// SPDX-License-Identifier: MIT

package update

import (
	"errors"
	"testing"
	"time"
)

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %s: %v", s, err)
	}
	return v
}

func TestPickLatest_StableSkipsPrerelease(t *testing.T) {
	rs := []Release{
		{TagName: "v0.5.0-beta1", Prerelease: true, PublishedAt: mustTime(t, "2026-07-01T00:00:00Z")},
		{TagName: "v0.4.0", PublishedAt: mustTime(t, "2026-06-15T00:00:00Z")},
		{TagName: "v0.3.0", PublishedAt: mustTime(t, "2026-06-01T00:00:00Z")},
	}
	got, err := PickLatest(ChannelStable, rs)
	if err != nil {
		t.Fatal(err)
	}
	if got.TagName != "v0.4.0" {
		t.Fatalf("expected v0.4.0, got %s", got.TagName)
	}
}

func TestPickLatest_BetaIncludesPrerelease(t *testing.T) {
	rs := []Release{
		{TagName: "v0.5.0-beta1", Prerelease: true, PublishedAt: mustTime(t, "2026-07-01T00:00:00Z")},
		{TagName: "v0.4.0", PublishedAt: mustTime(t, "2026-06-15T00:00:00Z")},
	}
	got, err := PickLatest(ChannelBeta, rs)
	if err != nil {
		t.Fatal(err)
	}
	if got.TagName != "v0.5.0-beta1" {
		t.Fatalf("expected v0.5.0-beta1, got %s", got.TagName)
	}
}

func TestPickLatest_SkipsDraft(t *testing.T) {
	rs := []Release{
		{TagName: "v0.6.0", Draft: true, PublishedAt: mustTime(t, "2026-08-01T00:00:00Z")},
		{TagName: "v0.4.0", PublishedAt: mustTime(t, "2026-06-15T00:00:00Z")},
	}
	got, err := PickLatest(ChannelStable, rs)
	if err != nil {
		t.Fatal(err)
	}
	if got.TagName != "v0.4.0" {
		t.Fatalf("expected v0.4.0, got %s", got.TagName)
	}
}

func TestPickLatest_DevelopReturnsSentinel(t *testing.T) {
	_, err := PickLatest(ChannelDevelop, nil)
	if !errors.Is(err, ErrSourceChannel) {
		t.Fatalf("expected ErrSourceChannel, got %v", err)
	}
}

func TestPickLatest_UnknownChannel(t *testing.T) {
	_, err := PickLatest(Channel("bogus"), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFilterChannel_NewestFirst(t *testing.T) {
	rs := []Release{
		{TagName: "v0.3.0", PublishedAt: mustTime(t, "2026-06-01T00:00:00Z")},
		{TagName: "v0.4.0", PublishedAt: mustTime(t, "2026-06-15T00:00:00Z")},
	}
	got := FilterChannel(ChannelStable, rs)
	if len(got) != 2 || got[0].TagName != "v0.4.0" {
		t.Fatalf("expected newest first, got %+v", got)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v0.4.0", "v0.3.0", 1},
		{"v0.3.0", "v0.4.0", -1},
		{"v0.4.0", "v0.4.0", 0},
		{"v0.4.0", "v0.4.0-beta1", 1},
		{"v0.4.0-beta2", "v0.4.0-beta1", 1},
		{"0.4.0", "v0.4.0", 0},
		{"v1.0.0", "v0.99.99", 1},
	}
	for _, c := range cases {
		if got := CompareVersions(c.a, c.b); got != c.want {
			t.Errorf("Compare(%s,%s) = %d want %d", c.a, c.b, got, c.want)
		}
	}
}
