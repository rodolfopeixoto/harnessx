package repl

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/intentplan"
)

func writeFakeBin(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "fake.sh")
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDefaultPlanDevHasCIStep(t *testing.T) {
	p := DefaultPlan(intentplan.GoalDev, "add /healthz")
	if p.Goal != intentplan.GoalDev {
		t.Errorf("goal: %s", p.Goal)
	}
	if len(p.Steps) == 0 {
		t.Fatal("steps empty")
	}
	found := false
	for _, s := range p.Steps {
		if s.Cmd[0] == "ci" {
			found = true
		}
	}
	if !found {
		t.Errorf("dev default plan must include ci: %+v", p.Steps)
	}
}

func TestDefaultPlanForEveryGoal(t *testing.T) {
	for _, g := range intentplan.KnownGoals() {
		p := DefaultPlan(g, "x")
		if p.Goal != g {
			t.Errorf("goal mismatch for %s", g)
		}
	}
}

func TestRunRejectsUnknownGoal(t *testing.T) {
	err := Run(context.Background(), Options{Goal: "alien"})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestRunHandlesExitImmediately(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var out bytes.Buffer
	dir := t.TempDir()
	err := Run(context.Background(), Options{
		Root:       dir,
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("/exit\n"),
		Out:        &out,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "bye") {
		t.Errorf("missing bye: %s", out.String())
	}
}

func TestRunSwitchesGoal(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var out bytes.Buffer
	err := Run(context.Background(), Options{
		Root:       t.TempDir(),
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("/goal ops\n/exit\n"),
		Out:        &out,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "goal → ops") {
		t.Errorf("missing goal switch: %s", out.String())
	}
}

func TestRunPlanCommandPrintsJSON(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var out bytes.Buffer
	_ = Run(context.Background(), Options{
		Root:       t.TempDir(),
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("/plan add /healthz\n/exit\n"),
		Out:        &out,
	})
	if !strings.Contains(out.String(), `"intent": "add /healthz"`) {
		t.Errorf("missing plan json: %s", out.String())
	}
}

func TestRunExecutesPromptAsPlan(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var out bytes.Buffer
	dir := t.TempDir()
	err := Run(context.Background(), Options{
		Root:       dir,
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("add /healthz\n/exit\n"),
		Out:        &out,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "✓ plan green") {
		t.Errorf("missing success: %s", out.String())
	}
	entries, _ := os.ReadDir(filepath.Join(dir, ".harness", "sessions"))
	if len(entries) == 0 {
		t.Errorf("no session file persisted")
	}
}

func TestRunReportsRedPlan(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 1\n")
	var out bytes.Buffer
	_ = Run(context.Background(), Options{
		Root:       t.TempDir(),
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("buggy\n/exit\n"),
		Out:        &out,
	})
	if !strings.Contains(out.String(), "✗ plan red") {
		t.Errorf("missing red: %s", out.String())
	}
}

func TestRunHelpListsCommands(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var out bytes.Buffer
	_ = Run(context.Background(), Options{
		Root:       t.TempDir(),
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("/help\n/exit\n"),
		Out:        &out,
	})
	for _, want := range []string{"/goal", "/plan", "/exit"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("help missing %q", want)
		}
	}
}

func TestRunRejectsUnknownGoalInline(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var out bytes.Buffer
	_ = Run(context.Background(), Options{
		Root:       t.TempDir(),
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("/goal alien\n/exit\n"),
		Out:        &out,
	})
	if !strings.Contains(out.String(), "unknown goal") {
		t.Errorf("missing rejection: %s", out.String())
	}
}

func TestLoadSessionMissingReturnsError(t *testing.T) {
	if _, err := LoadSession(t.TempDir(), "ghost"); err == nil {
		t.Fatal("want error for missing session id")
	}
}

func TestLoadSessionRoundTripsTurns(t *testing.T) {
	dir := t.TempDir()
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root:       dir,
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("foo\nbar\n/exit\n"),
		Out:        &buf,
	})
	sessions, err := ListSessions(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) == 0 {
		t.Fatal("expected at least one persisted session")
	}
	got, err := LoadSession(dir, sessions[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Turns) < 2 {
		t.Errorf("expected at least 2 turns, got %d", len(got.Turns))
	}
}

func TestListSessionsEmptyProjectReturnsNil(t *testing.T) {
	got, err := ListSessions(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("want 0 sessions, got %d", len(got))
	}
}

func TestRunResumeReplaysTurnsBeforeNewInput(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	prior := &Session{ID: "RESUMED", Goal: intentplan.GoalDev, Turns: []Turn{
		{Input: "added /products", Action: "chat"},
	}}
	var buf bytes.Buffer
	if err := Run(context.Background(), Options{
		Root:       t.TempDir(),
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("/history\n/exit\n"),
		Out:        &buf,
		Resume:     prior,
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "added /products") {
		t.Errorf("resumed turn not visible in /history: %s", buf.String())
	}
}

func TestStartSpinnerPlainNoop(t *testing.T) {
	var buf bytes.Buffer
	stop := startSpinner(&buf, true)
	stop()
	if buf.Len() != 0 {
		t.Errorf("plain spinner wrote %d bytes; want 0", buf.Len())
	}
}

func TestWithConversationContextEmptyReturnsPrompt(t *testing.T) {
	if got := withConversationContext(&Session{}, "hi"); got != "hi" {
		t.Errorf("empty session must pass prompt through; got %q", got)
	}
}

func TestWithConversationContextThreadsRecentTurns(t *testing.T) {
	sess := &Session{Goal: intentplan.GoalDev, Turns: []Turn{
		{Input: "add /products", Action: "chat"},
		{Input: "/exec foo", Action: "executed"},
		{Input: "fix 500 in cart", Action: "chat"},
	}}
	out := withConversationContext(sess, "checkout next?")
	for _, want := range []string{
		"# Conversation so far",
		"add /products",
		"fix 500 in cart",
		"# New user message",
		"checkout next?",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "/exec foo") {
		t.Errorf("slash-command turn should be filtered: %s", out)
	}
}

func TestWithConversationContextCapsAtFiveTurns(t *testing.T) {
	sess := &Session{}
	for i := 0; i < 12; i++ {
		sess.Turns = append(sess.Turns, Turn{Input: "msg-" + string(rune('a'+i)), Action: "chat"})
	}
	out := withConversationContext(sess, "now")
	if strings.Contains(out, "msg-a") {
		t.Errorf("oldest turn should be trimmed: %s", out)
	}
	if !strings.Contains(out, "msg-l") {
		t.Errorf("newest turn missing: %s", out)
	}
}

func TestRunAgentsListsAdapters(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root:         t.TempDir(),
		HarnessBin:   bin,
		Goal:         intentplan.GoalDev,
		In:           strings.NewReader("/agents\n/exit\n"),
		Out:          &buf,
		AdapterID:    "claude",
		AdaptersList: []string{"claude", "codex", "gemini"},
	})
	for _, want := range []string{"active: claude", "→ claude", "codex", "gemini"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q in /agents output: %s", want, buf.String())
		}
	}
}

func TestRunCostEmptyReports(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/cost\n/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "no agent turns yet") {
		t.Errorf("missing empty-cost message: %s", buf.String())
	}
}

func TestRunClearSetsContextMark(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	prior := &Session{ID: "X", Goal: intentplan.GoalDev, Turns: []Turn{
		{Input: "old chatter", Action: "chat"},
	}}
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:     strings.NewReader("/clear\n/exit\n"),
		Out:    &buf,
		Resume: prior,
	})
	if !strings.Contains(buf.String(), "working memory cleared") {
		t.Errorf("missing /clear confirmation: %s", buf.String())
	}
}

func TestRunAutoGateToggle(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/auto-gate on\n/auto-gate off\n/exit\n"),
		Out: &buf,
	})
	for _, want := range []string{"auto-gate ON", "auto-gate OFF"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q: %s", want, buf.String())
		}
	}
}

func TestRunIgnoresBlankLines(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var out bytes.Buffer
	if err := Run(context.Background(), Options{
		Root:       t.TempDir(),
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("\n\n/exit\n"),
		Out:        &out,
	}); err != nil {
		t.Fatal(err)
	}
}
