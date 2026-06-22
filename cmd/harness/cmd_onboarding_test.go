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

func TestTrimAndLowerStripsSpaceAndCase(t *testing.T) {
	cases := map[string]string{
		"  Y  ":  "y",
		"NO":     "no",
		"":       "",
		"  yes ": "yes",
	}
	for in, want := range cases {
		if got := trimAndLower(in); got != want {
			t.Errorf("trimAndLower(%q)=%q want %q", in, got, want)
		}
	}
}

func TestAskYesNoDefaults(t *testing.T) {
	out := &bytes.Buffer{}
	if !askYesNo(strings.NewReader("\n"), out, "yes default?", true) {
		t.Error("empty answer should accept default true")
	}
	if askYesNo(strings.NewReader("\n"), out, "no default?", false) {
		t.Error("empty answer should accept default false")
	}
}

func TestAskYesNoExplicit(t *testing.T) {
	out := &bytes.Buffer{}
	if !askYesNo(strings.NewReader("y\n"), out, "?", false) {
		t.Error("explicit y should be true")
	}
	if !askYesNo(strings.NewReader("YES\n"), out, "?", false) {
		t.Error("YES should be true")
	}
	if askYesNo(strings.NewReader("n\n"), out, "?", true) {
		t.Error("explicit n should be false")
	}
	if !askYesNo(strings.NewReader("1\n"), out, "?", false) {
		t.Error("1 should be true")
	}
	if askYesNo(strings.NewReader("0\n"), out, "?", true) {
		t.Error("0 should be false")
	}
	if !askYesNo(strings.NewReader("ok\n"), out, "?", false) {
		t.Error("ok should be true")
	}
}

func TestAskChoiceParsesNumber(t *testing.T) {
	out := &bytes.Buffer{}
	idx, err := askChoice(strings.NewReader("2\n"), out, "pick:", []string{"a", "b", "c"}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if idx != 1 {
		t.Errorf("want index 1 (b), got %d", idx)
	}
}

func TestAskChoiceEmptyUsesDefault(t *testing.T) {
	out := &bytes.Buffer{}
	idx, _ := askChoice(strings.NewReader("\n"), out, "pick:", []string{"a", "b", "c"}, 2)
	if idx != 2 {
		t.Errorf("want default 2, got %d", idx)
	}
}

func TestAskChoiceInvalidFallsBackToDefault(t *testing.T) {
	out := &bytes.Buffer{}
	idx, _ := askChoice(strings.NewReader("99\n"), out, "pick:", []string{"a", "b"}, 0)
	if idx != 0 {
		t.Errorf("want default 0, got %d", idx)
	}
	if !strings.Contains(out.String(), "invalid choice") {
		t.Errorf("missing invalid-choice warn: %s", out.String())
	}
}

func TestAskChoiceEmptyOptionsErrors(t *testing.T) {
	_, err := askChoice(strings.NewReader("1\n"), &bytes.Buffer{}, "x", nil, 0)
	if err == nil {
		t.Fatal("want error for empty options")
	}
}

func TestAskLineDefaultWhenBlank(t *testing.T) {
	got, ok := askLine(strings.NewReader("\n"), &bytes.Buffer{}, "dir", "./default")
	if !ok || got != "./default" {
		t.Errorf("default not used: %q ok=%v", got, ok)
	}
}

func TestAskLineUsesAnswer(t *testing.T) {
	got, _ := askLine(strings.NewReader("./custom\n"), &bytes.Buffer{}, "dir", "./default")
	if got != "./custom" {
		t.Errorf("want ./custom, got %q", got)
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
