package repl

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/intentplan"
)

func TestDefaultPlanAdsResearchOps(t *testing.T) {
	for _, g := range []intentplan.Goal{intentplan.GoalAds, intentplan.GoalResearch, intentplan.GoalOps} {
		p := DefaultPlan(g, "x")
		if len(p.Steps) == 0 {
			t.Errorf("goal %s should have steps", g)
		}
	}
}

func TestRunPersistsTurnsAcrossPlanAndExecute(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	dir := t.TempDir()
	var out bytes.Buffer
	if err := Run(context.Background(), Options{
		Root:       dir,
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		In:         strings.NewReader("first prompt\n/plan second\n/exit\n"),
		Out:        &out,
	}); err != nil {
		t.Fatal(err)
	}
	allEntries, _ := os.ReadDir(filepath.Join(dir, ".harness", "sessions"))
	var entries []os.DirEntry
	for _, e := range allEntries {
		if strings.HasSuffix(e.Name(), ".jsonl") {
			entries = append(entries, e)
		}
	}
	if len(entries) != 1 {
		t.Fatalf("want 1 jsonl session file, got %d (all: %d)", len(entries), len(allEntries))
	}
	body, _ := os.ReadFile(filepath.Join(dir, ".harness", "sessions", entries[0].Name()))
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(lines) != 2 {
		t.Errorf("want 2 jsonl turns, got %d: %s", len(lines), body)
	}
}

func TestRunHandlesPlannerError(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var out bytes.Buffer
	planner := func(ctx context.Context, g intentplan.Goal, prompt string) (intentplan.Plan, error) {
		return intentplan.Plan{}, errors.New("simulated")
	}
	if err := Run(context.Background(), Options{
		Root:       t.TempDir(),
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		Planner:    planner,
		In:         strings.NewReader("buggy\n/exit\n"),
		Out:        &out,
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "planner:") {
		t.Errorf("missing planner error: %s", out.String())
	}
}

func TestRunHandlesPlanCmdError(t *testing.T) {
	bin := writeFakeBin(t, "#!/bin/sh\nexit 0\n")
	var out bytes.Buffer
	planner := func(ctx context.Context, g intentplan.Goal, prompt string) (intentplan.Plan, error) {
		return intentplan.Plan{}, errors.New("plan boom")
	}
	_ = Run(context.Background(), Options{
		Root:       t.TempDir(),
		HarnessBin: bin,
		Goal:       intentplan.GoalDev,
		Planner:    planner,
		In:         strings.NewReader("/plan add x\n/exit\n"),
		Out:        &out,
	})
	if !strings.Contains(out.String(), "plan error") {
		t.Errorf("missing plan error: %s", out.String())
	}
}

func TestShouldExitVariants(t *testing.T) {
	for _, s := range []string{"/exit", "/quit", "exit", "quit"} {
		if !shouldExit(s) {
			t.Errorf("%s should exit", s)
		}
	}
	if shouldExit("hello") {
		t.Error("hello should not exit")
	}
}

func TestPersistCreatesDir(t *testing.T) {
	dir := t.TempDir()
	s := Session{ID: "abc", Turns: []Turn{{Action: "x"}}}
	if err := persist(dir, s); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".harness", "sessions", "abc.jsonl")); err != nil {
		t.Fatalf("file missing: %v", err)
	}
}
