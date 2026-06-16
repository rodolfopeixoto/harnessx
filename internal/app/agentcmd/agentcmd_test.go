// SPDX-License-Identifier: MIT

package agentcmd

import (
	"strings"
	"testing"
)

func TestListBundledReturnsKnownAdapters(t *testing.T) {
	got := listBundled()
	want := map[string]bool{
		"claude":             false,
		"claude-interactive": false,
		"codex":              false,
		"gemini":             false,
		"fake":               false,
		"kimi":               false,
		"anthropic-api":      false,
		"gemini-api":         false,
		"openai-api":         false,
		"moonshot-api":       false,
		"minimax-api":        false,
	}
	for _, id := range got {
		if _, ok := want[id]; ok {
			want[id] = true
		}
	}
	for id, ok := range want {
		if !ok {
			t.Errorf("bundled adapter missing: %s", id)
		}
	}
}

func TestListBundledSortedAscending(t *testing.T) {
	got := listBundled()
	for i := 1; i < len(got); i++ {
		if got[i-1] >= got[i] {
			t.Errorf("not sorted: %s >= %s", got[i-1], got[i])
		}
	}
}

func TestTruncatePassesThroughShort(t *testing.T) {
	if got := truncate("abc", 10); got != "abc" {
		t.Errorf("got %q", got)
	}
}

func TestTruncateAppendsEllipsis(t *testing.T) {
	got := truncate("abcdefghij", 5)
	if got != "abcd…" {
		t.Errorf("got %q", got)
	}
}

func TestTitleCaseUppercasesFirstRune(t *testing.T) {
	cases := map[string]string{
		"":       "",
		"a":      "A",
		"hello":  "Hello",
		"Hello":  "Hello",
		"1stage": "1stage",
	}
	for in, want := range cases {
		if got := titleCase(in); got != want {
			t.Errorf("titleCase(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestParseSpecAcceptsBundledShape(t *testing.T) {
	raw := []byte(`id: x
name: X
type: cli
command:
  binary: /bin/true
  check: /bin/true --version
`)
	s, err := parseSpec(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != "x" {
		t.Errorf("ID: got %q", s.ID)
	}
	if !strings.Contains(s.Command.Binary, "true") {
		t.Errorf("binary: got %q", s.Command.Binary)
	}
}
