// SPDX-License-Identifier: MIT

package detectors

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNames(t *testing.T) {
	names := []string{
		Worktrees{}.Name(),
		Containers{}.Name(),
		Caches{}.Name(),
		AbandonedHarness{}.Name(),
		LargeFiles{}.Name(),
		VMLeftovers{}.Name(),
		ClaudeLeftovers{}.Name(),
		HarnessWorktrees{}.Name(),
	}
	for _, n := range names {
		if n == "" {
			t.Errorf("empty name returned")
		}
	}
}

func TestHarnessWorktreesDetect(t *testing.T) {
	root := t.TempDir()
	out, err := HarnessWorktrees{}.Detect(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("expected zero findings on empty root, got %d", len(out))
	}

	dir := filepath.Join(root, ".harness", "worktrees", "child")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	out, err = HarnessWorktrees{}.Detect(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(out))
	}
	if out[0].Path != dir {
		t.Fatalf("path=%q want %q", out[0].Path, dir)
	}
}
