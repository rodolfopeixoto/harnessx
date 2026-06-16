// SPDX-License-Identifier: MIT

package auditrun

import (
	"testing"
)

func TestEnvOrDefaultFallback(t *testing.T) {
	if got := envOrDefault("HARNESS_TEST_DEFINITELY_UNSET", "fallback"); got != "fallback" {
		t.Errorf("missing env should yield fallback, got %q", got)
	}
}

func TestEnvOrDefaultSet(t *testing.T) {
	t.Setenv("HARNESS_TEST_KEY_SET", "value")
	if got := envOrDefault("HARNESS_TEST_KEY_SET", "fallback"); got != "value" {
		t.Errorf("set env should win, got %q", got)
	}
}

func TestFilterViewportMatch(t *testing.T) {
	in := []Viewport{{Name: "mobile", Width: 375}, {Name: "desktop", Width: 1440}}
	got := filterViewport(in, "desktop")
	if len(got) != 1 || got[0].Name != "desktop" {
		t.Errorf("got %+v", got)
	}
}

func TestFilterViewportPassthroughOnNoMatch(t *testing.T) {
	in := []Viewport{{Name: "mobile"}, {Name: "desktop"}}
	got := filterViewport(in, "ultrawide")
	if len(got) != 2 {
		t.Errorf("no match should passthrough, got %+v", got)
	}
}

func TestBaseAddrStripsScheme(t *testing.T) {
	cases := map[string]string{
		"http://localhost:7373":    "localhost:7373",
		"https://example.com/path": "example.com",
		"localhost:8080":           "localhost:8080",
	}
	for in, want := range cases {
		if got := baseAddr(in); got != want {
			t.Errorf("baseAddr(%q)=%q, want %q", in, got, want)
		}
	}
}
