// SPDX-License-Identifier: MIT

package autonomy

import (
	"testing"
)

func TestAppendAndListRoundTrip(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{"src/app.py", "src/app.py", "src/db.py"} {
		if err := AppendApproval(root, Event{Path: p, Decision: "approve"}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := ListApprovals(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3, got %d", len(got))
	}
	for _, e := range got {
		if e.At.IsZero() {
			t.Error("At should be set")
		}
	}
}

func TestAppendRejectsEmptyFields(t *testing.T) {
	root := t.TempDir()
	if err := AppendApproval(root, Event{}); err == nil {
		t.Error("expected error for empty event")
	}
	if err := AppendApproval(root, Event{Path: "x"}); err == nil {
		t.Error("expected error for missing decision")
	}
}

func TestListAbsentLogReturnsNil(t *testing.T) {
	got, err := ListApprovals(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("absent log should return nil, got %v", got)
	}
}
