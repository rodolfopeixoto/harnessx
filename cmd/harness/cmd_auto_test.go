// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewAutoCmdExposesFlags(t *testing.T) {
	c := newAutoCmd()
	for _, f := range []string{"max-attempts", "budget-usd", "agent", "watch", "dry-run", "resume"} {
		if c.Flags().Lookup(f) == nil {
			t.Errorf("flag %q missing", f)
		}
	}
}

func TestAutoStateRoundTrip(t *testing.T) {
	root := t.TempDir()
	st := &autoState{
		RunID:       "abc123",
		Prompt:      "add /readyz",
		AgentID:     "kimi",
		MaxAttempts: 5,
		BudgetUSD:   2.0,
		Phases: map[string]any{
			"plan": map[string]any{"ok": true, "elapsed_ms": float64(42)},
		},
		CostUSD: 0.0123,
	}
	if err := saveAutoState(root, st); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(autoStatePath(root, "abc123"))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got["run_id"] != "abc123" || got["prompt"] != "add /readyz" {
		t.Errorf("round-trip wrong: %+v", got)
	}

	loaded, err := loadAutoState(root, "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.AgentID != "kimi" || loaded.MaxAttempts != 5 {
		t.Errorf("loaded state wrong: %+v", loaded)
	}
}

func TestAutoCmdDryRunListsPhasesWithoutExecuting(t *testing.T) {
	root := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	c := newAutoCmd()
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetErr(&out)
	c.SetArgs([]string{"--dry-run", "add /healthz"})
	if err := c.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("dry-run must not error: %v", err)
	}
	for _, w := range []string{"plan", "spec", "tests", "impl", "ci"} {
		if !strings.Contains(out.String(), w) {
			t.Errorf("dry-run missing phase %q\n%s", w, out.String())
		}
	}
	matches, _ := filepath.Glob(filepath.Join(root, ".harness", "runs", "_agent", "*", "state.json"))
	if len(matches) != 1 {
		t.Errorf("want 1 state file, got %d", len(matches))
	}
}

func TestAutoCmdEmptyPromptErrors(t *testing.T) {
	c := newAutoCmd()
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetErr(&out)
	c.SetArgs([]string{"--dry-run"})
	if err := c.ExecuteContext(context.Background()); err == nil {
		t.Error("empty prompt must error")
	}
}

func TestAutoResumeReadsExistingState(t *testing.T) {
	root := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	st := &autoState{
		RunID: "deadbeef", Prompt: "old prompt", AgentID: "kimi",
		Phases: map[string]any{
			"plan": map[string]any{"ok": true},
			"spec": map[string]any{"ok": true},
		},
	}
	if err := saveAutoState(root, st); err != nil {
		t.Fatal(err)
	}
	c := newAutoCmd()
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetErr(&out)
	c.SetArgs([]string{"--resume", "deadbeef", "--dry-run"})
	if err := c.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "resuming run deadbeef") {
		t.Errorf("missing resume banner: %s", out.String())
	}
}
