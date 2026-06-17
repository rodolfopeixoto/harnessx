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
