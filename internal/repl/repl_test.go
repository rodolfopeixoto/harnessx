package repl

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/agents"
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

func TestStartSpinnerPlainPrintsStaticLine(t *testing.T) {
	var buf bytes.Buffer
	stop := startSpinner(&buf, true)
	stop()
	if !strings.Contains(buf.String(), "agent: working…") {
		t.Errorf("plain spinner must print static fallback line, got %q", buf.String())
	}
	if strings.Contains(buf.String(), "⠋") {
		t.Errorf("plain spinner must NOT print braille glyphs")
	}
}

func TestStartSpinnerNonTTYDoesNotEmitGlyphs(t *testing.T) {
	var buf bytes.Buffer
	stop := startSpinner(&buf, false)
	stop()
	if strings.Contains(buf.String(), "⠋") || strings.Contains(buf.String(), "\r") {
		t.Errorf("non-TTY writer must not receive CR-clobbered glyphs, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "agent: working…") {
		t.Errorf("non-TTY writer must get static fallback, got %q", buf.String())
	}
}

func TestWithConversationContextEmptyReturnsPrompt(t *testing.T) {
	if got := withConversationContext(&Session{}, "", "", "hi"); got != "hi" {
		t.Errorf("empty session must pass prompt through; got %q", got)
	}
}

func TestWithConversationContextThreadsRecentTurns(t *testing.T) {
	sess := &Session{Goal: intentplan.GoalDev, Turns: []Turn{
		{Input: "add /products", Action: "chat"},
		{Input: "/exec foo", Action: "executed"},
		{Input: "fix 500 in cart", Action: "chat"},
	}}
	out := withConversationContext(sess, "", "", "checkout next?")
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
	out := withConversationContext(sess, "", "", "now")
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

func TestRunUseUnavailableWithoutSwitcher(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/use kimi\n/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "/use unavailable") {
		t.Errorf("/use with no SwitchTo should print unavailable: %s", buf.String())
	}
}

func TestRunBudgetRejectsWhenExhausted(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	prior := &Session{ID: "X", Goal: intentplan.GoalDev, BudgetUSD: 0.01, Turns: []Turn{
		{Action: "chat", CostUSD: 0.05},
	}}
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:     strings.NewReader("/cost\n/exit\n"),
		Out:    &buf,
		Resume: prior,
	})
	if !strings.Contains(buf.String(), "0.0500") {
		t.Errorf("expected cost output to reflect resumed turn cost: %s", buf.String())
	}
}

func TestRunBudgetSetAndClear(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/budget 0.5\n/budget off\n/exit\n"),
		Out: &buf,
	})
	for _, want := range []string{"budget set to $0.5000", "budget cleared"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q: %s", want, buf.String())
		}
	}
}

func TestRunBudgetRejectsBadInput(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/budget abc\n/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "needs a non-negative USD") {
		t.Errorf("missing budget validation: %s", buf.String())
	}
}

func TestSetBudgetDirect(t *testing.T) {
	sess := &Session{}
	var buf bytes.Buffer
	setBudget(sess, &buf, "1.25")
	if sess.BudgetUSD != 1.25 {
		t.Errorf("budget not set: %f", sess.BudgetUSD)
	}
}

func TestCheckBudgetBlocksWhenSpent(t *testing.T) {
	sess := &Session{BudgetUSD: 0.10, Turns: []Turn{{CostUSD: 0.20}}}
	var buf bytes.Buffer
	if checkBudget(sess, &buf) {
		t.Error("budget should block when spent exceeds cap")
	}
	if !strings.Contains(buf.String(), "budget exhausted") {
		t.Errorf("missing exhausted message: %s", buf.String())
	}
}

func TestCheckBudgetAllowsWhenWithinCap(t *testing.T) {
	sess := &Session{BudgetUSD: 1.0, Turns: []Turn{{CostUSD: 0.20}}}
	var buf bytes.Buffer
	if !checkBudget(sess, &buf) {
		t.Error("budget should allow when under cap")
	}
}

func TestRunSaveLabelsSessionAndPersists(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	dir := t.TempDir()
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: dir, HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/save ecommerce-cart\n/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "labelled \"ecommerce-cart\"") {
		t.Errorf("missing label confirmation: %s", buf.String())
	}
	sessions, _ := ListSessions(dir)
	if len(sessions) == 0 || sessions[0].Label != "ecommerce-cart" {
		t.Errorf("label not surfaced in ListSessions: %+v", sessions)
	}
}

func TestSetSessionLabelRejectsBadChars(t *testing.T) {
	sess := &Session{}
	var buf bytes.Buffer
	setSessionLabel(sess, &buf, "with/slash")
	if sess.Label != "" {
		t.Error("label with slash should be refused")
	}
	if !strings.Contains(buf.String(), "unsupported character") {
		t.Errorf("missing rejection message: %s", buf.String())
	}
}

func TestSetSessionLabelEmptyRejected(t *testing.T) {
	sess := &Session{}
	var buf bytes.Buffer
	setSessionLabel(sess, &buf, "")
	if !strings.Contains(buf.String(), "needs a name") {
		t.Errorf("missing empty-name rejection: %s", buf.String())
	}
}

func TestRunRecapWithoutAdapterListsInputs(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	prior := &Session{ID: "X", Goal: intentplan.GoalDev, Turns: []Turn{
		{Input: "add /products", Action: "chat"},
		{Input: "fix cart 500", Action: "chat"},
	}}
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In: strings.NewReader("/recap\n/exit\n"), Out: &buf, Resume: prior,
	})
	for _, want := range []string{"no adapter wired", "add /products", "fix cart 500"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q: %s", want, buf.String())
		}
	}
}

func TestRunRecapEmptyReportsNoTurns(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In: strings.NewReader("/recap\n/exit\n"), Out: &buf,
	})
	if !strings.Contains(buf.String(), "no turns to recap") {
		t.Errorf("missing empty-recap message: %s", buf.String())
	}
}

func TestRunReplayBlocksMutatingInputs(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	prior := &Session{ID: "X", Goal: intentplan.GoalDev, ReadOnly: true, Turns: []Turn{
		{Input: "old", Action: "chat"},
	}}
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:     strings.NewReader("hello\n/ship something\n/history\n/exit\n"),
		Out:    &buf,
		Resume: prior,
	})
	if !strings.Contains(buf.String(), "session is read-only") {
		t.Errorf("missing read-only block: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "old") {
		t.Errorf("/history should still work in replay: %s", buf.String())
	}
}

func TestIsMutatingInputClassification(t *testing.T) {
	tests := map[string]bool{
		"plain text":    true,
		"/help":         false,
		"/history":      false,
		"/agents":       false,
		"/diff":         false,
		"/cost":         false,
		"/ship foo":     true,
		"/exec foo":     true,
		"/use kimi":     true,
		"/budget 0.50":  true,
		"/auto-gate on": true,
		"/recap":        true,
		"!ls":           true,
	}
	for in, want := range tests {
		if got := isMutatingInput(in); got != want {
			t.Errorf("isMutatingInput(%q) = %v; want %v", in, got, want)
		}
	}
}

func TestResolveSessionIDFallsBackToInput(t *testing.T) {
	if got := ResolveSessionID(t.TempDir(), "01KX-not-real"); got != "01KX-not-real" {
		t.Errorf("unknown id should be returned as-is; got %q", got)
	}
}

func TestResolveSessionIDByLabel(t *testing.T) {
	dir := t.TempDir()
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: dir, HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/save my-feature\n/exit\n"),
		Out: &buf,
	})
	sessions, _ := ListSessions(dir)
	if len(sessions) == 0 {
		t.Fatal("expected session persisted")
	}
	if got := ResolveSessionID(dir, "my-feature"); got != sessions[0].ID {
		t.Errorf("label lookup failed: got %q want %q", got, sessions[0].ID)
	}
	if got := ResolveSessionID(dir, sessions[0].ID); got != sessions[0].ID {
		t.Errorf("ulid lookup should pass through: got %q", got)
	}
}

func TestRunBranchEmptyRejected(t *testing.T) {
	var buf bytes.Buffer
	runBranch(context.Background(), &Session{}, Options{Out: &buf, Root: t.TempDir()}, "")
	if !strings.Contains(buf.String(), "needs a name") {
		t.Errorf("missing /branch validation: %s", buf.String())
	}
}

func TestRunBranchLabelsSessionWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	// init bare git repo so checkout -B can succeed
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "x@y.z"},
		{"config", "user.name", "t"},
		{"commit", "--allow-empty", "-q", "-m", "init"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			t.Skipf("git not available: %v", err)
		}
	}
	sess := &Session{}
	var buf bytes.Buffer
	runBranch(context.Background(), sess, Options{Out: &buf, Root: dir}, "feature/cart")
	if sess.Label != "feature-cart" {
		t.Errorf("expected label feature-cart, got %q", sess.Label)
	}
}

func TestSignalAwareCtxCancelsOnInterrupt(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buf bytes.Buffer
	ctx, stop := signalAwareCtx(parent, &buf)
	defer stop()
	// Closing the parent should drain the goroutine cleanly.
	cancel()
	<-ctx.Done()
}

func TestRunSavePromptCapturesLastPlainText(t *testing.T) {
	dir := t.TempDir()
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: dir, HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("add /products with pytest\n/save-prompt add-endpoint\n/prompts\n/exit\n"),
		Out: &buf,
	})
	for _, want := range []string{
		"prompt \"add-endpoint\" saved",
		"add-endpoint",
	} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q: %s", want, buf.String())
		}
	}
}

func TestRunSavePromptRejectsBadName(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("plain text\n/save-prompt BAD NAME\n/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "not a valid name") {
		t.Errorf("missing validation: %s", buf.String())
	}
}

func TestRunSavePromptNoPriorTurn(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/save-prompt only-slashes\n/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "no plain-text turn") {
		t.Errorf("missing 'no plain-text' guard: %s", buf.String())
	}
}

func TestRunPromptMissingErrors(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/prompt ghost\n/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "/prompt ghost") {
		t.Errorf("missing /prompt error: %s", buf.String())
	}
}

func TestRunPromptsEmpty(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/prompts\n/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "no saved prompts") {
		t.Errorf("missing empty-prompts message: %s", buf.String())
	}
}

func TestRunNoAdapterRefusesPlainText(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:        strings.NewReader("just hello there\n/exit\n"),
		Out:       &buf,
		NoAdapter: true,
	})
	if !strings.Contains(buf.String(), "no adapter wired") {
		t.Errorf("missing refusal: %s", buf.String())
	}
	if strings.Contains(buf.String(), "✓ plan green") {
		t.Errorf("plan should not have executed: %s", buf.String())
	}
}

func TestLevenshteinBasic(t *testing.T) {
	cases := map[string]struct {
		a, b string
		want int
	}{
		"identical":  {"abc", "abc", 0},
		"single sub": {"abc", "abx", 1},
		"insert":     {"abc", "abcd", 1},
		"delete":     {"abcd", "abc", 1},
		"swap":       {"abc", "acb", 2},
		"empty":      {"", "abc", 3},
	}
	for name, c := range cases {
		if got := levenshtein(c.a, c.b); got != c.want {
			t.Errorf("%s: levenshtein(%q,%q)=%d want %d", name, c.a, c.b, got, c.want)
		}
	}
}

func TestSuggestSessionByCloseLabel(t *testing.T) {
	dir := t.TempDir()
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: dir, HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/save ecommerce-products\n/exit\n"),
		Out: &buf,
	})
	got := SuggestSession(dir, "ecommerce-prodcts")
	if got != "ecommerce-products" {
		t.Errorf("expected suggestion 'ecommerce-products', got %q", got)
	}
}

func TestSuggestSessionTooFarReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: dir, HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/save foo\n/exit\n"),
		Out: &buf,
	})
	if got := SuggestSession(dir, "totally-unrelated-name"); got != "" {
		t.Errorf("expected empty for distant arg, got %q", got)
	}
}

func TestRunTimelineRendersTurnsAndCost(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	prior := &Session{ID: "X", Goal: intentplan.GoalDev, Turns: []Turn{
		{Input: "add /products", Action: "chat", CostUSD: 0.02},
		{Input: "fix cart", Action: "chat", CostUSD: 0.03},
	}}
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In: strings.NewReader("/timeline\n/exit\n"), Out: &buf, Resume: prior,
	})
	for _, want := range []string{
		"add /products", "fix cart", "$0.0200", "$0.0300", "total:", "~$0.0500",
	} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q: %s", want, buf.String())
		}
	}
}

func TestRunTimelineEmpty(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In: strings.NewReader("/timeline\n/exit\n"), Out: &buf,
	})
	if !strings.Contains(buf.String(), "no turns yet") {
		t.Errorf("missing empty timeline message: %s", buf.String())
	}
}

func TestAggregateCostGroupsByAdapterAndTask(t *testing.T) {
	turns := []Turn{
		{Action: "chat", AdapterID: "claude", TaskTag: "implementation", InTokens: 100, OutTokens: 50, CostUSD: 0.10},
		{Action: "chat", AdapterID: "claude", TaskTag: "implementation", InTokens: 200, OutTokens: 100, CostUSD: 0.20},
		{Action: "chat", AdapterID: "gemini", TaskTag: "cheap_review", InTokens: 50, OutTokens: 20, CostUSD: 0.01},
		{Action: "save", Input: "/save x"},
	}
	totals := aggregateCost(turns)
	if totals.ChatTurns != 3 {
		t.Errorf("want 3 chat turns, got %d", totals.ChatTurns)
	}
	if len(totals.PerAgent) != 2 {
		t.Fatalf("want 2 adapter rows, got %d", len(totals.PerAgent))
	}
	if totals.PerAgent[0].AdapterID != "claude" {
		t.Errorf("first row should be claude (highest cost): %+v", totals.PerAgent)
	}
	if totals.PerAgent[0].Turns != 2 || totals.PerAgent[0].CostUSD < 0.299 {
		t.Errorf("claude row aggregated wrong: %+v", totals.PerAgent[0])
	}
	if totals.Total.CostUSD < 0.30 || totals.Total.CostUSD > 0.32 {
		t.Errorf("total cost wrong: %f", totals.Total.CostUSD)
	}
}

func TestRenderCostReportShowsAdapterColumns(t *testing.T) {
	var buf bytes.Buffer
	totals := costTotals{
		ChatTurns: 2,
		Total:     costRow{Turns: 2, InTokens: 300, OutTokens: 150, CostUSD: 0.30},
		PerAgent: []costRow{
			{AdapterID: "claude", Task: "implementation", Turns: 1, InTokens: 200, OutTokens: 100, CostUSD: 0.25},
			{AdapterID: "gemini", Task: "cheap_review", Turns: 1, InTokens: 100, OutTokens: 50, CostUSD: 0.05},
		},
	}
	renderCostReport(&buf, "01ABC", totals)
	for _, want := range []string{
		"session 01ABC",
		"ADAPTER",
		"claude", "implementation",
		"gemini", "cheap_review",
		"TOTAL",
		"0.3000",
	} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q\n%s", want, buf.String())
		}
	}
}

func TestRenderCostReportEmpty(t *testing.T) {
	var buf bytes.Buffer
	renderCostReport(&buf, "X", costTotals{ChatTurns: 0})
	if !strings.Contains(buf.String(), "no recorded usage") {
		t.Errorf("missing empty marker: %s", buf.String())
	}
}

func TestRunSlashAloneShowsMenu(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/\n/exit\n"),
		Out: &buf,
	})
	for _, want := range []string{
		"slash commands",
		"chat",
		"plain text",
		"/drive",
		"/cost",
		"session",
		"exit",
	} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q in slash menu: %s", want, buf.String())
		}
	}
}

func TestPrintSlashMenuQuestionMark(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/?\n/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "slash commands") {
		t.Errorf("/? should also open the menu: %s", buf.String())
	}
}

func TestPadRightFillsToWidth(t *testing.T) {
	if padRight("ab", 5) != "ab   " {
		t.Errorf("padRight wrong")
	}
	if padRight("hello", 3) != "hello" {
		t.Errorf("padRight should not truncate")
	}
}

func TestAdapterBillingModeKnownIDs(t *testing.T) {
	cases := map[string]string{
		"claude":             "oneshot · API-billed",
		"codex":              "oneshot · API-billed",
		"gemini":             "oneshot · API-billed",
		"anthropic-api":      "oneshot · API-billed",
		"openai-api":         "oneshot · API-billed",
		"kimi":               "interactive · plan/local",
		"claude-interactive": "interactive · plan/local",
		"ollama":             "interactive · plan/local",
		"fake":               "fake · free",
		"unknown":            "unknown billing",
	}
	for id, want := range cases {
		if got := adapterBillingMode(id); got != want {
			t.Errorf("adapterBillingMode(%q)=%q want %q", id, got, want)
		}
	}
}

func TestSuggestSlashPicksClosestMatch(t *testing.T) {
	var buf bytes.Buffer
	suggestSlash(&buf, "/drv")
	if !strings.Contains(buf.String(), "/drive") {
		t.Errorf("want /drive suggestion, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "did you mean") {
		t.Errorf("missing 'did you mean' phrasing: %q", buf.String())
	}
}

func TestSuggestSlashEmptyDoesNothing(t *testing.T) {
	var buf bytes.Buffer
	suggestSlash(&buf, "")
	if buf.Len() != 0 {
		t.Errorf("empty attempt should print nothing, got %q", buf.String())
	}
}

func TestSuggestSlashFarAttemptFallsBack(t *testing.T) {
	var buf bytes.Buffer
	suggestSlash(&buf, "/totally-unrelated-name-here")
	if !strings.Contains(buf.String(), "/ for the menu") {
		t.Errorf("missing fallback message: %q", buf.String())
	}
}

func TestFirstToken(t *testing.T) {
	cases := map[string]string{
		"/drive foo bar": "/drive",
		"/cost":          "/cost",
		"  /agents  ":    "/agents",
		"":               "",
	}
	for in, want := range cases {
		if got := firstToken(in); got != want {
			t.Errorf("firstToken(%q)=%q want %q", in, got, want)
		}
	}
}

func TestSummariseSessionPrintsRecap(t *testing.T) {
	sess := &Session{
		ID: "01ABC", Goal: intentplan.GoalDev, Label: "ecommerce",
		Turns: []Turn{
			{Action: "chat", AdapterID: "claude", InTokens: 100, OutTokens: 50, CostUSD: 0.10},
			{Action: "save", Input: "/save x"},
		},
	}
	var buf bytes.Buffer
	summariseSession(&buf, sess)
	for _, want := range []string{"session recap", "01ABC", "ecommerce", "dev", "turns  2", "tokens", "0.1000"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q\n%s", want, buf.String())
		}
	}
}

func TestSummariseSessionEmpty(t *testing.T) {
	var buf bytes.Buffer
	summariseSession(&buf, &Session{})
	if buf.Len() != 0 {
		t.Errorf("empty session should print nothing, got %q", buf.String())
	}
}

func TestRunUnknownSlashSuggests(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/drv\n/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "/drive") {
		t.Errorf("expected /drive suggestion: %s", buf.String())
	}
}

func TestRunExitPrintsRecap(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/help\n/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "session recap") {
		t.Errorf("missing recap on exit: %s", buf.String())
	}
}

func TestRunPipeModeSuppressesGreetAndRecap(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:    strings.NewReader("/help\n/exit\n"),
		Out:   &buf,
		Pipe:  true,
		Plain: true,
	})
	for _, banned := range []string{
		"harness chat — session",
		"session recap",
		"bye",
	} {
		if strings.Contains(buf.String(), banned) {
			t.Errorf("pipe mode should not print %q\n%s", banned, buf.String())
		}
	}
	if !strings.Contains(buf.String(), "/exec") {
		t.Errorf("pipe mode should still print /help output: %s", buf.String())
	}
}

func TestIntentRedirectTurnRecordedAsZeroCost(t *testing.T) {
	// BUG 6: an input that fires the shell-or-slash guard must be
	// persisted as action=intent_redirect with cost_usd=0 so /cost
	// and harness analytics never count it as a paid agent turn.
	fakeAdapter := &noopAdapter{}
	sess := &Session{Goal: intentplan.GoalDev}
	opts := &Options{
		Adapter:   fakeAdapter,
		AdapterID: "fake",
		Out:       &bytes.Buffer{},
	}
	turn := handleInput(context.Background(), sess, opts, "exec something")
	if turn.Action != "intent_redirect" {
		t.Errorf("guard turn must be action=intent_redirect, got %q", turn.Action)
	}
	if turn.CostUSD != 0 {
		t.Errorf("intent_redirect cost must be 0, got %f", turn.CostUSD)
	}
	if fakeAdapter.runs != 0 {
		t.Errorf("adapter must NOT be billed for guard hint, got %d Run calls", fakeAdapter.runs)
	}
}

type noopAdapter struct {
	runs int
}

func (n *noopAdapter) ID() string                        { return "fake" }
func (n *noopAdapter) Name() string                      { return "fake" }
func (n *noopAdapter) Capabilities() agents.Capabilities { return agents.Capabilities{} }
func (n *noopAdapter) Healthcheck(_ context.Context) agents.HealthcheckResult {
	return agents.HealthcheckResult{OK: true}
}
func (n *noopAdapter) Run(_ context.Context, _ agents.AgentRequest) agents.AgentResult {
	n.runs++
	return agents.AgentResult{}
}
func (n *noopAdapter) ParseUsage(_ agents.AgentOutput) agents.Usage { return agents.Usage{} }
func (n *noopAdapter) ClassifyFailure(_ agents.AgentOutput, _ int, _ error) agents.FailureType {
	return agents.FailureNone
}

func TestDetectAdapterSwitchExtractsID(t *testing.T) {
	cases := map[string]string{
		"use kimi":                 "kimi",
		"USE Kimi":                 "kimi",
		"harness use kimi":         "kimi",
		"switch to codex":          "codex",
		"switch codex":             "codex",
		"change adapter to gemini": "gemini",
		"change adapter gemini":    "gemini",
		"model claude-3-opus":      "claude-3-opus",
		"  use kimi  ":             "kimi",
		"hello agent":              "",
		"use the new feature":      "",
		"please change something":  "",
		"":                         "",
	}
	for in, want := range cases {
		if got := detectAdapterSwitch(in); got != want {
			t.Errorf("detectAdapterSwitch(%q)=%q want %q", in, got, want)
		}
	}
}

func TestLooksLikeShellOrSlash(t *testing.T) {
	cases := map[string]bool{
		"harness use 4":                  true,
		"use kimi":                       true,
		"exec add /readyz":               true,
		"hello agent":                    false,
		"please add a /healthz endpoint": false,
		"":                               false,
		"!ls":                            false,
	}
	for in, wantHint := range cases {
		got := looksLikeShellOrSlash(in)
		if (got != "") != wantHint {
			t.Errorf("looksLikeShellOrSlash(%q)=%q want hint=%v", in, got, wantHint)
		}
	}
}

func TestRunOutputJSONEmitsTurnEnvelope(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:         strings.NewReader("/help\n/exit\n"),
		Out:        &buf,
		Pipe:       true,
		Plain:      true,
		OutputJSON: true,
	})
	found := false
	for _, line := range strings.Split(buf.String(), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var env map[string]any
		if err := json.Unmarshal([]byte(line), &env); err != nil {
			continue
		}
		if env["input"] == "/help" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("OutputJSON did not emit /help envelope\n%s", buf.String())
	}
}

func TestRunNonPipePrintsGreetAndRecap(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var buf bytes.Buffer
	_ = Run(context.Background(), Options{
		Root: t.TempDir(), HarnessBin: bin, Goal: intentplan.GoalDev,
		In:  strings.NewReader("/exit\n"),
		Out: &buf,
	})
	if !strings.Contains(buf.String(), "harness chat — session") {
		t.Errorf("non-pipe greet missing: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "bye") {
		t.Errorf("non-pipe bye missing: %s", buf.String())
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
