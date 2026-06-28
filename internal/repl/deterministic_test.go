package repl

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestDeterministicRefusesAgentCalls(t *testing.T) {
	sess := &Session{ID: "d"}
	var buf bytes.Buffer
	opts := &Options{Root: t.TempDir(), Out: &buf, Plain: true, Pipe: true, Deterministic: true}
	cases := []string{
		"plain text would call adapter",
		"/exec build",
		"/do refactor",
		"/ship feature",
		"/drive auth fix",
		"/recap",
		"/btw quick q",
	}
	for _, in := range cases {
		buf.Reset()
		turn := handleInput(context.Background(), sess, opts, in)
		if turn.Action != "deterministic-block" {
			t.Errorf("deterministic mode must refuse %q, got action=%s\noutput:\n%s", in, turn.Action, buf.String())
		}
		if !strings.Contains(buf.String(), "deterministic mode") {
			t.Errorf("%q output should mention deterministic mode, got:\n%s", in, buf.String())
		}
	}
}

func TestDeterministicAllowsLocalSlashes(t *testing.T) {
	sess := &Session{ID: "d"}
	var buf bytes.Buffer
	opts := &Options{Root: t.TempDir(), Out: &buf, Plain: true, Pipe: true, Deterministic: true}
	allowed := []string{"/history", "/agents", "/cost", "/diff", "/help", "/timeline", "/clear", "/save mylabel", "/budget 1.0", "/goal research"}
	for _, in := range allowed {
		buf.Reset()
		turn := handleInput(context.Background(), sess, opts, in)
		if turn.Action == "deterministic-block" {
			t.Errorf("local slash %q must be allowed in deterministic, got blocked\noutput:\n%s", in, buf.String())
		}
	}
}

func TestDeterministicAllowsShellBangs(t *testing.T) {
	sess := &Session{ID: "d"}
	var buf bytes.Buffer
	opts := &Options{Root: t.TempDir(), Out: &buf, Plain: true, Pipe: true, Deterministic: true}
	turn := handleInput(context.Background(), sess, opts, "!ls /tmp")
	if turn.Action == "deterministic-block" {
		t.Fatalf("shell bangs must be allowed in deterministic mode (no LLM), got blocked\noutput:\n%s", buf.String())
	}
}

func TestCallsLLMTable(t *testing.T) {
	cases := map[string]bool{
		"plain text":    true,
		"/exec foo":     true,
		"/do x":         true,
		"/drive y":      true,
		"/recap":        true,
		"/btw question": true,
		"/history":      false,
		"/cost":         false,
		"/help":         false,
		"!ls":           false,
		"/use claude":   false,
		"/model haiku":  false,
		"":              false,
	}
	for in, want := range cases {
		if got := callsLLM(in); got != want {
			t.Errorf("callsLLM(%q) want %v got %v", in, want, got)
		}
	}
}
