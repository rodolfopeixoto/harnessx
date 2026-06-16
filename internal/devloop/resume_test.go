// SPDX-License-Identifier: MIT

package devloop

import (
	"path/filepath"
	"testing"
	"time"
)

func TestWriteAndLoadStateRoundTrip(t *testing.T) {
	root := t.TempDir()
	s := State{
		RunID:              "01ABC",
		OriginalPrompt:     "fix x",
		BaselineLintOK:     true,
		BaselineTestOK:     false,
		Attempts:           []Attempt{{Number: 1, LintOK: true, TestOK: false, Elapsed: 200 * time.Millisecond}},
		BudgetUSDRemaining: 0.45,
		MaxAttempts:        5,
		AgentID:            "claude",
		Autonomy:           "safe_execute",
		Apply:              true,
		LintCmd:            "ruff",
		TestCmd:            "pytest",
	}
	if err := WriteState(root, s); err != nil {
		t.Fatal(err)
	}
	got, err := LoadState(root, "01ABC")
	if err != nil {
		t.Fatal(err)
	}
	if got.RunID != s.RunID || got.OriginalPrompt != s.OriginalPrompt {
		t.Errorf("round-trip mismatch: got %+v", got)
	}
	if len(got.Attempts) != 1 || got.Attempts[0].Number != 1 {
		t.Errorf("attempts mismatch: got %+v", got.Attempts)
	}
	if got.SchemaVersion != StateSchemaVersion {
		t.Errorf("schema_version: got %d", got.SchemaVersion)
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set on write")
	}
}

func TestWriteStateRejectsEmptyRunID(t *testing.T) {
	err := WriteState(t.TempDir(), State{})
	if err == nil {
		t.Fatal("empty RunID should error")
	}
}

func TestLoadStateMissingFile(t *testing.T) {
	_, err := LoadState(t.TempDir(), "missing")
	if err == nil {
		t.Error("expected error for missing state")
	}
}

func TestListResumableSortsNewestFirst(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"01OLD", "01NEW", "01MID"} {
		s := State{RunID: id, OriginalPrompt: id}
		if err := WriteState(root, s); err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	got, err := ListResumable(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3 resumable, got %d", len(got))
	}
	if got[0].RunID != "01MID" {
		t.Errorf("newest first should be 01MID, got %q", got[0].RunID)
	}
}

func TestListResumableAbsentDirReturnsNil(t *testing.T) {
	got, err := ListResumable(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("absent dir should return nil, got %v", got)
	}
}

func TestStartAttemptCountsAttempts(t *testing.T) {
	cases := []struct {
		attempts int
		want     int
	}{
		{0, 1},
		{1, 2},
		{5, 6},
	}
	for _, c := range cases {
		s := State{Attempts: make([]Attempt, c.attempts)}
		if got := StartAttempt(s); got != c.want {
			t.Errorf("attempts=%d: want %d, got %d", c.attempts, c.want, got)
		}
	}
}

func TestStateDirIsUnderRunsLoop(t *testing.T) {
	root := t.TempDir()
	d := StateDir(root, "01XYZ")
	want := filepath.Join(root, ".harness", "runs", "_loop", "01XYZ")
	if d != want {
		t.Errorf("StateDir: want %q, got %q", want, d)
	}
}
