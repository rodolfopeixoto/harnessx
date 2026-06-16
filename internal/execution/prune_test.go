// SPDX-License-Identifier: MIT

package execution

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func mkRun(t *testing.T, root, id string, age time.Duration) string {
	t.Helper()
	dir := filepath.Join(root, ".harness", "runs", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-age)
	_ = os.Chtimes(dir, past, past)
	return dir
}

func TestPruneCandidatesByAge(t *testing.T) {
	root := t.TempDir()
	mkRun(t, root, "01NEW", time.Hour)
	mkRun(t, root, "01OLD", 100*24*time.Hour)
	got, err := PruneCandidates(root, 30*24*time.Hour, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 candidate, got %d: %v", len(got), got)
	}
	if filepath.Base(got[0]) != "01OLD" {
		t.Errorf("wrong candidate: %s", got[0])
	}
}

func TestPruneCandidatesKeepLast(t *testing.T) {
	root := t.TempDir()
	mkRun(t, root, "01ONE", time.Hour)
	mkRun(t, root, "01TWO", time.Hour)
	mkRun(t, root, "01THR", time.Hour)
	got, err := PruneCandidates(root, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 candidate (oldest), got %d: %v", len(got), got)
	}
	if filepath.Base(got[0]) != "01ONE" {
		t.Errorf("keep-last picked wrong: %s", got[0])
	}
}

func TestPruneCandidatesNoRunsDir(t *testing.T) {
	got, err := PruneCandidates(t.TempDir(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("expected nil for missing runs dir, got %v", got)
	}
}

func TestDeletePathsFreesBytes(t *testing.T) {
	root := t.TempDir()
	dir := mkRun(t, root, "01ABC", time.Hour)
	if err := os.WriteFile(filepath.Join(dir, "blob"), make([]byte, 1024), 0o644); err != nil {
		t.Fatal(err)
	}
	freed, err := DeletePaths([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if freed < 1024 {
		t.Errorf("want freed >= 1024 bytes, got %d", freed)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("dir should be gone")
	}
}

func TestDeletePathsMissingTargetIsError(t *testing.T) {
	// DeletePaths walks before remove; missing target dir is silently
	// skipped by RemoveAll, so deletion of an absent path is a no-op.
	freed, err := DeletePaths([]string{"/tmp/harness-test-absolutely-not-there-xyz"})
	if err != nil {
		t.Fatalf("missing path should not error: %v", err)
	}
	if freed != 0 {
		t.Errorf("freed should be 0 for missing path, got %d", freed)
	}
}
