// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestIndexNLFindsNewline(t *testing.T) {
	cases := map[string]int{
		"hello\nworld": 5,
		"single line":  -1,
		"win\r\nstyle": 3,
		"trailing\n":   8,
	}
	for in, want := range cases {
		if got := indexNL(in); got != want {
			t.Errorf("indexNL(%q)=%d want %d", in, got, want)
		}
	}
}

func TestTrimToWidthTruncatesWithEllipsis(t *testing.T) {
	if got := trimToWidth("hello world", 5); got != "hello…" {
		t.Errorf("trimToWidth wrong: %q", got)
	}
	if got := trimToWidth("short", 10); got != "short" {
		t.Errorf("short should pass through: %q", got)
	}
}

func TestRenderOnboardingPrintsAllSections(t *testing.T) {
	var buf bytes.Buffer
	r := onboardingResult{
		HarnessV:  "v0.143.0",
		Suggested: "claude",
		Tools: []checkedTool{
			{toolCheck: toolCheck{name: "git", purpose: "vcs", installHint: "brew install git"}, found: true, version: "git 2.45"},
			{toolCheck: toolCheck{name: "rg", purpose: "scan", installHint: "brew install ripgrep"}, found: false},
		},
		Adapters: []checkedTool{
			{toolCheck: toolCheck{name: "claude", purpose: "anthropic", installHint: "..."}, found: true, version: "2.1"},
		},
	}
	renderOnboarding(&buf, r)
	for _, want := range []string{
		"harness onboarding",
		"v0.143.0",
		"system tools",
		"git",
		"rg",
		"agent adapters",
		"claude",
		"next steps",
		"harness use claude",
	} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q\n%s", want, buf.String())
		}
	}
}

func TestRenderOnboardingFlagsNoAgentInstalled(t *testing.T) {
	var buf bytes.Buffer
	renderOnboarding(&buf, onboardingResult{HarnessV: "v0", Suggested: "", Adapters: nil})
	if !strings.Contains(buf.String(), "no agent CLI on PATH") {
		t.Errorf("missing no-agent warning: %s", buf.String())
	}
}

func TestPickSuggestedAdapterPrefersClaude(t *testing.T) {
	a := []checkedTool{
		{toolCheck: toolCheck{name: "gemini"}, found: true},
		{toolCheck: toolCheck{name: "claude"}, found: true},
	}
	if got := pickSuggestedAdapter(a, t.TempDir()); got != "claude" {
		t.Errorf("want claude, got %q", got)
	}
}

func TestPickSuggestedAdapterFallsBackToFirstFound(t *testing.T) {
	a := []checkedTool{
		{toolCheck: toolCheck{name: "unknown-x"}, found: true},
	}
	if got := pickSuggestedAdapter(a, t.TempDir()); got != "unknown-x" {
		t.Errorf("fallback wrong: %q", got)
	}
}

func TestPickSuggestedAdapterEmptyWhenNoneFound(t *testing.T) {
	a := []checkedTool{
		{toolCheck: toolCheck{name: "claude"}, found: false},
	}
	if got := pickSuggestedAdapter(a, t.TempDir()); got != "" {
		t.Errorf("want empty, got %q", got)
	}
}
