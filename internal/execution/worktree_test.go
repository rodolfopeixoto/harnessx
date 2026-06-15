// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func gitInit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "-A"},
		{"commit", "-q", "-m", "seed"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

func TestPrepareAndCleanup_Worktree(t *testing.T) {
	root := t.TempDir()
	gitInit(t, root)
	mgr := NewManager(root)
	ctx := context.Background()
	wt, err := mgr.Prepare(ctx, "run-001")
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if wt.Kind != "git_worktree" {
		t.Fatalf("expected git_worktree, got %q", wt.Kind)
	}
	if _, err := os.Stat(wt.Path); err != nil {
		t.Fatalf("worktree path missing: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wt.Path, "hello.txt"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runDir := filepath.Join(root, ".harness", "runs", "run-001")
	changed, err := CaptureDiff(ctx, wt, runDir)
	if err != nil {
		t.Fatalf("capture diff: %v", err)
	}
	if len(changed) != 1 || changed[0] != "hello.txt" {
		t.Fatalf("unexpected changed files: %v", changed)
	}
	if _, err := os.Stat(filepath.Join(runDir, "diff.patch")); err != nil {
		t.Fatalf("diff.patch missing: %v", err)
	}
	if err := mgr.Cleanup(ctx, wt); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if _, err := os.Stat(wt.Path); !os.IsNotExist(err) {
		t.Fatalf("worktree path still exists after cleanup")
	}
}

func TestPrepare_NonGitFallsBackToCopy(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mgr := NewManager(root)
	wt, err := mgr.Prepare(context.Background(), "run-002")
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if wt.Kind != "copy" {
		t.Fatalf("expected copy, got %q", wt.Kind)
	}
	if _, err := os.Stat(filepath.Join(wt.Path, "main.go")); err != nil {
		t.Fatalf("copied file missing: %v", err)
	}
}
