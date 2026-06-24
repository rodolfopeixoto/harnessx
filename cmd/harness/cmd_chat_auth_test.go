// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/fake"
)

func TestAuthyDetectsAuthHints(t *testing.T) {
	cases := map[string]bool{
		"unauthorized request":      true,
		"401 from upstream":         true,
		"please log in":             true,
		"invalid api key for codex": true,
		"network timeout":           false,
		"":                          false,
	}
	for in, want := range cases {
		got := authy(agents.HealthcheckResult{Err: in})
		if got != want {
			t.Errorf("authy(%q)=%v want %v", in, got, want)
		}
	}
}

func TestOneLineCollapsesWhitespaceAndCaps(t *testing.T) {
	got := oneLine("hello\nworld\n")
	if got != "hello world" {
		t.Errorf("got %q", got)
	}
	long := strings.Repeat("x", 200)
	if oneLine(long) != strings.Repeat("x", 120)+"…" {
		t.Errorf("long should cap at 120")
	}
}

func TestHandleAuthFailureNoLoginCmdJustWarns(t *testing.T) {
	a := fake.New("claude")
	a.CapsValue.LoginCommand = ""
	a.CapsValue.AuthDocURL = "https://docs.example.com/login"
	var out bytes.Buffer
	in := strings.NewReader("")
	handleAuthFailure(context.Background(), &out, in, a, "claude",
		agents.HealthcheckResult{Err: "401 unauthorized"})
	got := out.String()
	for _, w := range []string{"claude needs auth", "401 unauthorized", "https://docs.example.com/login"} {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q\n%s", w, got)
		}
	}
}

func TestHandleAuthFailureNonAuthErrorPassesThrough(t *testing.T) {
	a := fake.New("claude")
	var out bytes.Buffer
	in := strings.NewReader("")
	handleAuthFailure(context.Background(), &out, in, a, "claude",
		agents.HealthcheckResult{Err: "network timeout"})
	if !strings.Contains(out.String(), "healthcheck warn") {
		t.Errorf("non-auth error should fall through to generic warn: %s", out.String())
	}
}

func TestHandleAuthFailureSkipsLoginWhenUserSaysNo(t *testing.T) {
	a := fake.New("claude")
	a.CapsValue.LoginCommand = "echo not-going-to-run"
	var out bytes.Buffer
	in := strings.NewReader("n\n")
	handleAuthFailure(context.Background(), &out, in, a, "claude",
		agents.HealthcheckResult{Err: "401 not logged in"})
	got := out.String()
	if !strings.Contains(got, "skipped") {
		t.Errorf("expected skip message, got %s", got)
	}
	// "not-going-to-run" appears in the fix:-recipe line; check
	// instead that the command DIDN'T run by counting occurrences:
	// printed twice would mean both the prompt AND echo executed.
	if strings.Count(got, "not-going-to-run") > 1 {
		t.Errorf("login command must not have executed (multiple echoes): %s", got)
	}
}
