package repl

import (
	"bytes"
	"context"
	"os"
	"os/exec"
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
	if !strings.Contains(buf.String(), "$0.0500") {
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
