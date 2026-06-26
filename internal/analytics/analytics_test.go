// SPDX-License-Identifier: MIT

package analytics

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeChatSession(t *testing.T, root, id, label string, turns []map[string]any) {
	t.Helper()
	dir := filepath.Join(root, ".harness", "sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	for _, turn := range turns {
		if err := enc.Encode(turn); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, id+".jsonl"), b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	meta := map[string]any{"id": id, "goal": "dev", "label": label}
	body, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(dir, id+".meta.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWalkAggregatesAcrossStacks(t *testing.T) {
	work := t.TempDir()
	py := filepath.Join(work, "api-python")
	go_ := filepath.Join(work, "api-go")
	if err := os.MkdirAll(py, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(go_, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(py, "pyproject.toml"), []byte(""), 0o644)
	_ = os.WriteFile(filepath.Join(go_, "go.mod"), []byte("module x\n"), 0o644)

	now := time.Now().UTC().Format(time.RFC3339Nano)
	writeChatSession(t, py, "01PY", "auth", []map[string]any{
		{"time": now, "input": "add auth", "action": "chat", "adapter_id": "claude", "task_tag": "implementation", "in_tokens": 100, "out_tokens": 50, "cost_usd": 0.10},
	})
	writeChatSession(t, go_, "01GO", "lists", []map[string]any{
		{"time": now, "input": "add lists", "action": "chat", "adapter_id": "kimi", "task_tag": "cheap_review", "in_tokens": 20, "out_tokens": 10, "cost_usd": 0.01},
	})

	rep, err := Walk([]string{work}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if rep.TotalTurns < 2 {
		t.Errorf("want >=2 turns, got %d", rep.TotalTurns)
	}
	if len(rep.Stacks) < 2 {
		t.Errorf("want >=2 stack rows, got %d (%+v)", len(rep.Stacks), rep.Stacks)
	}
	stacksByName := map[string]Row{}
	for _, s := range rep.Stacks {
		stacksByName[s.Stack] = s
	}
	if stacksByName["python"].CostUSD < 0.099 {
		t.Errorf("python stack cost wrong: %+v", stacksByName["python"])
	}
	if stacksByName["go"].CostUSD < 0.009 {
		t.Errorf("go stack cost wrong: %+v", stacksByName["go"])
	}
}

func TestRenderTextHasAllSections(t *testing.T) {
	var buf bytes.Buffer
	Render(&buf, Report{
		Roots:      []string{"/x"},
		TotalTurns: 3, TotalUSD: 0.12,
		Stacks: []Row{{Stack: "python", Sessions: 1, Turns: 2, ChatTurns: 1, InTokens: 100, OutTokens: 50, CostUSD: 0.10}},
		Adapters: []AdapterRow{
			{AdapterID: "claude", Task: "implementation", Turns: 1, CostUSD: 0.10},
		},
		Days: []DayRow{{Day: "2026-06-21", Turns: 1, CostUSD: 0.10}},
	})
	for _, w := range []string{"harness analytics", "by stack", "python", "by adapter", "claude", "by day", "2026-06-21"} {
		if !strings.Contains(buf.String(), w) {
			t.Errorf("missing %q\n%s", w, buf.String())
		}
	}
}

func TestRenderJSONRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	rep := Report{TotalTurns: 1, TotalUSD: 0.5}
	if err := RenderJSON(&buf, rep); err != nil {
		t.Fatal(err)
	}
	var got Report
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.TotalUSD != 0.5 || got.TotalTurns != 1 {
		t.Errorf("round-trip wrong: %+v", got)
	}
}

func TestDetectStackProbes(t *testing.T) {
	dir := t.TempDir()
	if got := detectStack(dir); got != "unknown" {
		t.Errorf("want unknown, got %q", got)
	}
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte(""), 0o644)
	if got := detectStack(dir); got != "go" {
		t.Errorf("want go, got %q", got)
	}
}

func TestDetectStackRailsBeatsRuby(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "Gemfile"), []byte(""), 0o644)
	_ = os.MkdirAll(filepath.Join(dir, "config"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "config", "application.rb"), []byte(""), 0o644)
	if got := detectStack(dir); got != "rails" {
		t.Errorf("want rails, got %q", got)
	}
}

func TestWalkSinceFiltersOldTurns(t *testing.T) {
	work := t.TempDir()
	_ = os.WriteFile(filepath.Join(work, "go.mod"), []byte(""), 0o644)
	oldTs := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339Nano)
	writeChatSession(t, work, "01OLD", "old", []map[string]any{
		{"time": oldTs, "input": "old", "action": "chat", "adapter_id": "claude", "cost_usd": 0.20},
	})
	rep, err := Walk([]string{work}, time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if rep.TotalTurns != 0 {
		t.Errorf("expected old turns filtered, got %d", rep.TotalTurns)
	}
}

func TestWalkAggregatesRunMeta_BUG19(t *testing.T) {
	work := t.TempDir()
	_ = os.WriteFile(filepath.Join(work, "go.mod"), []byte(""), 0o644)
	runDir := filepath.Join(work, ".harness", "runs", "run_01TEST")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	meta := []byte(`{"run_id":"run_01TEST","agent_id":"claude","task_tag":"implementation",` +
		`"started_at":"` + now + `","estimated_cost_usd":0.42,"input_tokens":1000,"output_tokens":250}`)
	if err := os.WriteFile(filepath.Join(runDir, "meta.json"), meta, 0o644); err != nil {
		t.Fatal(err)
	}

	rep, err := Walk([]string{work}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if rep.TotalUSD < 0.41 || rep.TotalUSD > 0.43 {
		t.Fatalf("TotalUSD = %v, want ~0.42", rep.TotalUSD)
	}
	if rep.TotalTurns != 1 {
		t.Fatalf("TotalTurns = %d, want 1", rep.TotalTurns)
	}
	if len(rep.Adapters) == 0 || rep.Adapters[0].AdapterID != "claude" {
		t.Fatalf("expected claude adapter row, got %+v", rep.Adapters)
	}
}
