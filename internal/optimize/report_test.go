// SPDX-License-Identifier: MIT

package optimize

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirSizeEmpty(t *testing.T) {
	root := t.TempDir()
	if got := dirSize(root); got != 0 {
		t.Errorf("empty dir: want 0, got %d", got)
	}
}

func TestDirSizeCountsFiles(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "a.txt"), []byte("12345"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "b.txt"), []byte("67890"), 0o644)
	if got := dirSize(root); got != 10 {
		t.Errorf("want 10, got %d", got)
	}
}

func TestLoadSnapshotMissingFile(t *testing.T) {
	if _, err := LoadSnapshot("/tmp/__nope_snap__.json"); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestRemovalCandidateUnknownEcosystem(t *testing.T) {
	got, _ := removalCandidate("unknown-eco", "anything")
	if got {
		t.Error("unknown ecosystem should not flag removal")
	}
}
