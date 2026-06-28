package repl

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func mockSession(t *testing.T, root, id string, turns []Turn) {
	t.Helper()
	dir := filepath.Join(root, ".harness", "sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, id+".jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for _, tr := range turns {
		body, err := json.Marshal(tr)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(append(body, '\n')); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReplaySessionLoadsTurns(t *testing.T) {
	root := t.TempDir()
	prior := []Turn{
		{Time: time.Now().Add(-2 * time.Minute), Action: "chat", Input: "add /healthz", CostUSD: 0.012},
		{Time: time.Now().Add(-1 * time.Minute), Action: "chat", Input: "add tests", CostUSD: 0.024},
	}
	mockSession(t, root, "01abc", prior)
	sess, err := LoadSession(root, "01abc")
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if len(sess.Turns) != 2 {
		t.Fatalf("want 2 prior turns, got %d", len(sess.Turns))
	}
}

func TestReplayReadOnlyRefusesMutators(t *testing.T) {
	root := t.TempDir()
	sess := &Session{ID: "rs", ReadOnly: true}
	opts := &Options{Root: root, Out: &bytes.Buffer{}, Plain: true, Pipe: true}
	cases := []string{
		"/exec build orders",
		"plain text would call adapter",
		"/use claude",
		"/model claude-opus-4-7",
		"/route on",
		"/once",
		"/clear",
		"/budget 1.0",
	}
	for _, in := range cases {
		buf := opts.Out.(*bytes.Buffer)
		buf.Reset()
		turn := handleInput(context.Background(), sess, opts, in)
		if turn.Action != "read-only-block" {
			t.Errorf("read-only mode must refuse %q, got action=%s\noutput:\n%s", in, turn.Action, buf.String())
		}
		if !strings.Contains(buf.String(), "read-only") {
			t.Errorf("%q output should mention read-only, got:\n%s", in, buf.String())
		}
	}
}

func TestReplayReadOnlyAllowsInspectionSlashes(t *testing.T) {
	root := t.TempDir()
	prior := []Turn{{Action: "chat", Input: "hello", CostUSD: 0.01}}
	sess := &Session{ID: "rs", ReadOnly: true, Turns: prior}
	var buf bytes.Buffer
	opts := &Options{Root: root, Out: &buf, Plain: true, Pipe: true}
	inspections := []string{"/history", "/agents", "/cost", "/diff", "/help", "/timeline"}
	for _, in := range inspections {
		buf.Reset()
		turn := handleInput(context.Background(), sess, opts, in)
		if turn.Action == "read-only-block" {
			t.Errorf("inspection %q must be allowed in read-only, got blocked\noutput:\n%s", in, buf.String())
		}
	}
}

func TestReplayReadOnlyBlocksShellBangs(t *testing.T) {
	sess := &Session{ID: "rs", ReadOnly: true}
	var buf bytes.Buffer
	opts := &Options{Root: t.TempDir(), Out: &buf, Plain: true, Pipe: true}
	turn := handleInput(context.Background(), sess, opts, "!ls /tmp")
	if turn.Action != "read-only-block" {
		t.Fatalf("read-only must refuse shell bangs, got action=%s\noutput:\n%s", turn.Action, buf.String())
	}
}

func TestReplayMockConversationFullFlow(t *testing.T) {
	root := t.TempDir()
	prior := []Turn{
		{Time: time.Now().Add(-5 * time.Minute), Action: "chat", Input: "investigate broken auth", CostUSD: 0.05},
		{Time: time.Now().Add(-4 * time.Minute), Action: "ci", Input: "/ci", CostUSD: 0},
		{Time: time.Now().Add(-3 * time.Minute), Action: "chat", Input: "fix token expiry check", CostUSD: 0.08},
		{Time: time.Now().Add(-2 * time.Minute), Action: "exec", Input: "/exec apply fix", CostUSD: 0.12},
		{Time: time.Now().Add(-1 * time.Minute), Action: "ci", Input: "/ci", CostUSD: 0},
	}
	mockSession(t, root, "01full", prior)

	loaded, err := LoadSession(root, "01full")
	if err != nil {
		t.Fatal(err)
	}
	loaded.ReadOnly = true
	sess := loaded
	var buf bytes.Buffer
	opts := &Options{Root: root, Out: &buf, Plain: true, Pipe: true, Resume: loaded}

	buf.Reset()
	handleInput(context.Background(), sess, opts, "/history")
	historyOut := buf.String()
	for _, want := range []string{"investigate broken auth", "fix token expiry check"} {
		if !strings.Contains(historyOut, want) {
			t.Errorf("/history in replay must list %q, got:\n%s", want, historyOut)
		}
	}

	buf.Reset()
	handleInput(context.Background(), sess, opts, "/cost")
	costOut := buf.String()
	if !strings.Contains(costOut, "$") {
		t.Errorf("/cost in replay should print spend, got:\n%s", costOut)
	}

	buf.Reset()
	handleInput(context.Background(), sess, opts, "/timeline")
	timelineOut := buf.String()
	if !strings.Contains(timelineOut, "investigate broken auth") {
		t.Errorf("/timeline in replay should include prior turn input, got:\n%s", timelineOut)
	}

	buf.Reset()
	turn := handleInput(context.Background(), sess, opts, "/exec retry the fix")
	if turn.Action != "read-only-block" {
		t.Errorf("/exec must be refused in replay, got action=%s\noutput:\n%s", turn.Action, buf.String())
	}
}
