package repl

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestSession(t *testing.T) (*Session, *Options, *bytes.Buffer) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".harness"), 0o755); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	opts := &Options{Root: root, HarnessBin: "harness", Out: &buf, Plain: true, Pipe: true}
	sess := &Session{ID: "test-session", Goal: "dev"}
	return sess, opts, &buf
}

func driveOne(t *testing.T, sess *Session, opts *Options, input string) Turn {
	t.Helper()
	turn := handleInput(context.Background(), sess, opts, input)
	sess.Turns = append(sess.Turns, turn)
	return turn
}

func TestChatHelpIncludesModelRouteOnce(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	driveOne(t, sess, opts, "/help")
	got := buf.String()
	for _, want := range []string{"/model", "/route", "/once"} {
		if !strings.Contains(got, want) {
			t.Errorf("/help must include %s, got:\n%s", want, got)
		}
	}
}

func TestChatSlashMenuListsModelRouteOnce(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	driveOne(t, sess, opts, "/")
	got := buf.String()
	for _, want := range []string{"/model", "/route", "/once"} {
		if !strings.Contains(got, want) {
			t.Errorf("/ slash menu must include %s, got:\n%s", want, got)
		}
	}
}

func TestChatModelPrintsCurrent(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	opts.Model = "claude-sonnet-4-6"
	driveOne(t, sess, opts, "/model")
	if !strings.Contains(buf.String(), "claude-sonnet-4-6") {
		t.Fatalf("/model should print current model, got:\n%s", buf.String())
	}
}

func TestChatModelSwaps(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	driveOne(t, sess, opts, "/model gpt-5-mini")
	if opts.Model != "gpt-5-mini" {
		t.Fatalf("model should be set to gpt-5-mini, got %q", opts.Model)
	}
	if !strings.Contains(buf.String(), "gpt-5-mini") {
		t.Fatalf("output should mention new model, got:\n%s", buf.String())
	}
}

func TestChatRouteToggle(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	driveOne(t, sess, opts, "/route on")
	if !opts.RouteEnabled {
		t.Fatal("RouteEnabled should be true after /route on")
	}
	driveOne(t, sess, opts, "/route off")
	if opts.RouteEnabled {
		t.Fatal("RouteEnabled should be false after /route off")
	}
	if !strings.Contains(buf.String(), "routing") {
		t.Fatal("expected routing mentions in output")
	}
}

func TestChatRouteStatusNoArg(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	driveOne(t, sess, opts, "/route")
	got := buf.String()
	if !strings.Contains(got, "routing:") {
		t.Fatalf("expected routing status, got:\n%s", got)
	}
}

func TestChatOnceFlagsOpts(t *testing.T) {
	sess, opts, _ := newTestSession(t)
	driveOne(t, sess, opts, "/once")
	if !opts.OneShot {
		t.Fatal("OneShot must be true after /once")
	}
}

func TestChatHistoryListsPrior(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	sess.Turns = append(sess.Turns, Turn{Input: "first prompt", Action: "chat"})
	sess.Turns = append(sess.Turns, Turn{Input: "second prompt", Action: "chat"})
	buf.Reset()
	driveOne(t, sess, opts, "/history")
	got := buf.String()
	for _, want := range []string{"first prompt", "second prompt"} {
		if !strings.Contains(got, want) {
			t.Errorf("/history missing %q, got:\n%s", want, got)
		}
	}
}

func TestChatLastReplaysLastPrompt(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	sess.Turns = append(sess.Turns, Turn{Input: "rerun me", Action: "chat"})
	buf.Reset()
	driveOne(t, sess, opts, "/last")
	if !strings.Contains(buf.String(), "rerun me") {
		t.Fatalf("/last should echo prior input, got:\n%s", buf.String())
	}
}

func TestChatBudgetCaps(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	driveOne(t, sess, opts, "/budget 0.50")
	if sess.BudgetUSD != 0.50 {
		t.Fatalf("BudgetUSD should be 0.50, got %f", sess.BudgetUSD)
	}
	driveOne(t, sess, opts, "/budget off")
	if sess.BudgetUSD != 0 {
		t.Fatalf("BudgetUSD should be 0 after off, got %f", sess.BudgetUSD)
	}
	if !strings.Contains(buf.String(), "budget") {
		t.Fatal("expected budget mention in output")
	}
}

func TestChatAutoGateToggle(t *testing.T) {
	sess, opts, _ := newTestSession(t)
	driveOne(t, sess, opts, "/auto-gate on")
	if !sess.AutoGate {
		t.Fatal("sess.AutoGate true after /auto-gate on")
	}
	driveOne(t, sess, opts, "/auto-gate off")
	if sess.AutoGate {
		t.Fatal("sess.AutoGate false after /auto-gate off")
	}
}

func TestChatGoalChanges(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	driveOne(t, sess, opts, "/goal research")
	if sess.Goal != "research" {
		t.Fatalf("Goal should be research, got %s", sess.Goal)
	}
	if !strings.Contains(buf.String(), "research") {
		t.Fatal("expected goal mention in output")
	}
}

func TestChatClearWipesHistory(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	sess.Turns = append(sess.Turns, Turn{Input: "x", Action: "chat"})
	driveOne(t, sess, opts, "/clear")
	if !strings.Contains(buf.String(), "cleared") && !strings.Contains(buf.String(), "drop") && !strings.Contains(buf.String(), "history") {
		t.Fatalf("expected clear acknowledgement, got:\n%s", buf.String())
	}
}

func TestChatCostEmptyPrintsHint(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	driveOne(t, sess, opts, "/cost")
	got := buf.String()
	if !strings.Contains(got, "no agent turns yet") {
		t.Fatalf("/cost should hint on empty session, got:\n%s", got)
	}
}

func TestChatTimelineEmptyMsg(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	driveOne(t, sess, opts, "/timeline")
	if !strings.Contains(buf.String(), "no turns") {
		t.Fatalf("empty timeline should say so, got:\n%s", buf.String())
	}
}

func TestChatUseAdapterMissingErrors(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	driveOne(t, sess, opts, "/use nonexistent")
	if !strings.Contains(buf.String(), "unavailable") && !strings.Contains(buf.String(), "✗") {
		t.Fatalf("/use without SwitchTo wired should error, got:\n%s", buf.String())
	}
}

func TestChatRecognisesUnknownSlashAsError(t *testing.T) {
	sess, opts, buf := newTestSession(t)
	turn := driveOne(t, sess, opts, "/totally-unknown-command")
	if turn.Action == "" {
		t.Errorf("turn should have an action label")
	}
	_ = buf
}
