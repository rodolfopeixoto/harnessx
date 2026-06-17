package configwiz

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/router"
)

func TestLoadAbsentFileReturnsEmpty(t *testing.T) {
	snap, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if snap.Routes != nil {
		t.Errorf("want nil routes, got %v", snap.Routes)
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	in := Snapshot{Routes: map[string]router.RouteConfig{
		"planning": {Primary: "claude", Fallback: []string{"gemini"}, BudgetUSD: 0.5, Model: "sonnet"},
	}}
	if err := Save(dir, in); err != nil {
		t.Fatal(err)
	}
	out, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if out.Routes["planning"].Primary != "claude" {
		t.Errorf("primary mismatch")
	}
	if out.Routes["planning"].BudgetUSD != 0.5 {
		t.Errorf("budget mismatch")
	}
}

func TestSetTaskAppendsAudit(t *testing.T) {
	dir := t.TempDir()
	if err := SetTaskPrimary(dir, "planning", "claude", []string{"gemini"}, 0.5, "sonnet"); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, ".harness", "logs", "config-mutations.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var m Mutation
	if err := json.Unmarshal(bytes.TrimSpace(body), &m); err != nil {
		t.Fatalf("invalid audit json: %v", err)
	}
	if m.Action != "set" || m.Task != "planning" {
		t.Errorf("audit fields: %+v", m)
	}
}

func TestSetTaskOverwritesAndRecordsBefore(t *testing.T) {
	dir := t.TempDir()
	_ = SetTaskPrimary(dir, "planning", "claude", nil, 0, "")
	_ = SetTaskPrimary(dir, "planning", "gemini", nil, 0, "")
	body, _ := os.ReadFile(filepath.Join(dir, ".harness", "logs", "config-mutations.jsonl"))
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 audit lines, got %d", len(lines))
	}
	var second Mutation
	_ = json.Unmarshal([]byte(lines[1]), &second)
	if second.Before.Primary != "claude" {
		t.Errorf("before mismatch: %+v", second.Before)
	}
}

func TestDeleteTaskErrorsWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	if err := DeleteTask(dir, "missing"); err == nil {
		t.Fatal("want error")
	}
}

func TestDeleteTaskRemovesEntry(t *testing.T) {
	dir := t.TempDir()
	_ = SetTaskPrimary(dir, "planning", "claude", nil, 0, "")
	if err := DeleteTask(dir, "planning"); err != nil {
		t.Fatal(err)
	}
	snap, _ := Load(dir)
	if _, ok := snap.Routes["planning"]; ok {
		t.Error("planning still present")
	}
}

func TestDiffShowsAddRemoveChange(t *testing.T) {
	before := Snapshot{Routes: map[string]router.RouteConfig{
		"a": {Primary: "x"},
		"b": {Primary: "old"},
	}}
	after := Snapshot{Routes: map[string]router.RouteConfig{
		"b": {Primary: "new"},
		"c": {Primary: "y"},
	}}
	got := Diff(before, after)
	joined := strings.Join(got, "\n")
	for _, want := range []string{"- a:", "+ c:", "~ b:"} {
		if !strings.Contains(joined, want) {
			t.Errorf("diff missing %q: %s", want, joined)
		}
	}
}

func TestRunWizardWithCannedInput(t *testing.T) {
	dir := t.TempDir()
	in := strings.NewReader("claude\ngemini,kimi\n0.5\n")
	var out bytes.Buffer
	err := RunWizard(WizardOptions{
		Root:         dir,
		AvailableIDs: []string{"claude", "gemini", "kimi"},
		Tasks:        []string{"planning"},
		In:           in,
		Out:          &out,
	})
	if err != nil {
		t.Fatal(err)
	}
	snap, _ := Load(dir)
	r := snap.Routes["planning"]
	if r.Primary != "claude" || len(r.Fallback) != 2 || r.BudgetUSD != 0.5 {
		t.Fatalf("snapshot mismatch: %+v", r)
	}
}

func TestSplitCSVTrimsEntries(t *testing.T) {
	got := SplitCSVForCLI(" a , b ,, c ")
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Errorf("split: %v", got)
	}
}

func TestRunWizardErrorsWithoutAdapters(t *testing.T) {
	err := RunWizard(WizardOptions{Root: t.TempDir(), AvailableIDs: nil})
	if err == nil {
		t.Fatal("expected error")
	}
}
