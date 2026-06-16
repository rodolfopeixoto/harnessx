// SPDX-License-Identifier: MIT

package sharedstate

import (
	"path/filepath"
	"testing"
)

func TestDetectNoConflictsOnDisjointSets(t *testing.T) {
	s := Snapshot{
		RunID: "r1",
		Tasks: []Task{
			{Index: 1, WriteSet: []string{"a.go"}, Version: 1},
			{Index: 2, ReadSet: []string{"b.go"}},
		},
	}
	if got := Detect(s); len(got) != 0 {
		t.Errorf("disjoint sets should not conflict, got %+v", got)
	}
}

func TestDetectFlagsStaleAssumption(t *testing.T) {
	s := Snapshot{
		RunID: "r1",
		Tasks: []Task{
			{Index: 1, WriteSet: []string{"app.py"}, Version: 2},
			{Index: 2, ReadSet: []string{"app.py"}, Assumptions: map[string]int{"app.py": 1}},
		},
	}
	got := Detect(s)
	if len(got) != 1 {
		t.Fatalf("want 1 conflict, got %d: %+v", len(got), got)
	}
	if got[0].OverlappingOn != "app.py" {
		t.Errorf("overlapping_on: got %q", got[0].OverlappingOn)
	}
	if got[0].LaterIdx != 2 || got[0].EarlierIdx != 1 {
		t.Errorf("indexes: got later=%d earlier=%d", got[0].LaterIdx, got[0].EarlierIdx)
	}
}

func TestDetectNoConflictWhenAssumptionCurrent(t *testing.T) {
	s := Snapshot{
		RunID: "r1",
		Tasks: []Task{
			{Index: 1, WriteSet: []string{"app.py"}, Version: 2},
			{Index: 2, ReadSet: []string{"app.py"}, Assumptions: map[string]int{"app.py": 2}},
		},
	}
	if got := Detect(s); len(got) != 0 {
		t.Errorf("current assumption should not conflict, got %+v", got)
	}
}

func TestDetectMissingAssumptionFlags(t *testing.T) {
	s := Snapshot{
		RunID: "r1",
		Tasks: []Task{
			{Index: 1, WriteSet: []string{"x"}, Version: 1},
			{Index: 2, ReadSet: []string{"x"}},
		},
	}
	got := Detect(s)
	if len(got) != 1 {
		t.Fatalf("missing assumption should flag conflict, got %d", len(got))
	}
}

func TestWriteAndReadRoundTrip(t *testing.T) {
	root := t.TempDir()
	s := Snapshot{
		RunID: "01ABC",
		Tasks: []Task{
			{Index: 1, Kind: "scaffold", WriteSet: []string{"app.py"}, Version: 1},
		},
	}
	if err := Write(root, s); err != nil {
		t.Fatal(err)
	}
	got, err := Read(root, "01ABC")
	if err != nil {
		t.Fatal(err)
	}
	if got.RunID != s.RunID || len(got.Tasks) != 1 {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("schema_version: got %d", got.SchemaVersion)
	}
}

func TestWriteRejectsEmptyRunID(t *testing.T) {
	if err := Write(t.TempDir(), Snapshot{}); err == nil {
		t.Fatal("empty RunID should error")
	}
}

func TestReadMissingFile(t *testing.T) {
	if _, err := Read(t.TempDir(), "missing"); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestPathConvention(t *testing.T) {
	root := t.TempDir()
	got := Path(root, "01XYZ")
	want := filepath.Join(root, ".harness", "runs", "01XYZ", "shared.json")
	if got != want {
		t.Errorf("Path: want %q, got %q", want, got)
	}
}

func TestContainsHelper(t *testing.T) {
	if !contains([]string{"a", "b"}, "a") {
		t.Error("expected hit")
	}
	if contains([]string{"a", "b"}, "c") {
		t.Error("unexpected hit")
	}
	if contains(nil, "x") {
		t.Error("nil slice should not contain")
	}
}
