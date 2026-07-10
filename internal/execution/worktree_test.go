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

// gitInitEmpty runs `git init` in dir but creates no commits (unborn HEAD).
func gitInitEmpty(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
}

func TestPrepare_EmptyGitRepo_FallsBackToCopy(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	gitInitEmpty(t, root)
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mgr := NewManager(root)
	wt, err := mgr.Prepare(context.Background(), "run-empty")
	if err != nil {
		t.Fatalf("prepare on empty repo should not error, got: %v", err)
	}
	if wt.Kind != "copy" {
		t.Fatalf("expected copy fallback for empty repo, got %q", wt.Kind)
	}
	if _, err := os.Stat(filepath.Join(wt.Path, "main.go")); err != nil {
		t.Fatalf("copied file missing: %v", err)
	}
}

func TestPrepare_NonGit_UsesCopy(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mgr := NewManager(root)
	wt, err := mgr.Prepare(context.Background(), "run-nongit")
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if wt.Kind != "copy" {
		t.Fatalf("expected copy, got %q", wt.Kind)
	}
}

func TestPrepare_GitRepoWithCommit_UsesWorktree(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	gitInit(t, root)
	mgr := NewManager(root)
	ctx := context.Background()
	wt, err := mgr.Prepare(ctx, "run-wt-ok")
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if wt.Kind != "git_worktree" {
		t.Fatalf("expected git_worktree, got %q", wt.Kind)
	}
	t.Cleanup(func() { _ = mgr.Cleanup(ctx, wt) })
}

func TestPrepare_RejectsUnsafeRunID(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	gitInit(t, root)
	mgr := NewManager(root)
	ctx := context.Background()
	bad := []string{
		"../evil",
		"run/../evil",
		"has space",
		"has/slash",
		"back\\slash",
		"semi;colon",
		".hidden",
	}
	for _, id := range bad {
		if _, err := mgr.Prepare(ctx, id); err == nil {
			t.Errorf("expected error for runID %q, got nil", id)
		}
	}
}

func TestPrepare_ReclaimsStaleWorktreeDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	gitInit(t, root)
	mgr := NewManager(root)
	ctx := context.Background()

	// Prepare once, then simulate a crash by wiping git's knowledge of the
	// worktree while leaving the directory behind.
	wt, err := mgr.Prepare(ctx, "run-stale")
	if err != nil {
		t.Fatalf("prepare #1: %v", err)
	}
	// Kill git's registration (simulating a corrupt/orphan state).
	adminDir := filepath.Join(root, ".git", "worktrees", "run-stale")
	_ = os.RemoveAll(adminDir)
	// Leave wt.Path on disk. Also leave the branch behind.

	// Second Prepare with the same runID must succeed.
	wt2, err := mgr.Prepare(ctx, "run-stale")
	if err != nil {
		t.Fatalf("prepare #2 after stale state should succeed, got: %v", err)
	}
	if wt2.Kind != "git_worktree" {
		t.Fatalf("expected git_worktree after reclaim, got %q", wt2.Kind)
	}
	if _, err := os.Stat(wt2.Path); err != nil {
		t.Fatalf("worktree dir missing after reclaim: %v", err)
	}
	t.Cleanup(func() {
		_ = mgr.Cleanup(ctx, wt2)
		_ = mgr.Cleanup(ctx, wt)
	})
}

func TestPrepare_ReclaimsStaleBranch(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	gitInit(t, root)
	// Pre-create the branch name that Prepare would want.
	branch := "harness/run/run-branchdup"
	head, _ := gitHead(context.Background(), root)
	create := exec.Command("git", "branch", branch, head)
	create.Dir = root
	if out, err := create.CombinedOutput(); err != nil {
		t.Fatalf("pre-create branch: %v: %s", err, out)
	}

	mgr := NewManager(root)
	ctx := context.Background()
	wt, err := mgr.Prepare(ctx, "run-branchdup")
	if err != nil {
		t.Fatalf("prepare should reclaim stale branch, got: %v", err)
	}
	if wt.Kind != "git_worktree" {
		t.Fatalf("expected git_worktree, got %q", wt.Kind)
	}
	t.Cleanup(func() { _ = mgr.Cleanup(ctx, wt) })
}

func TestPrepare_ReclaimsStaleCopyDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Pre-populate the target so we exercise the stale-copy reclaim path.
	staleDir := filepath.Join(root, ".harness", "worktrees", "run-copydup")
	if err := os.MkdirAll(staleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staleDir, "leftover.txt"), []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mgr := NewManager(root)
	wt, err := mgr.Prepare(context.Background(), "run-copydup")
	if err != nil {
		t.Fatalf("prepare with stale copy dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wt.Path, "leftover.txt")); !os.IsNotExist(err) {
		t.Fatalf("stale leftover.txt should have been removed, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(wt.Path, "a.txt")); err != nil {
		t.Fatalf("expected fresh a.txt after reclaim: %v", err)
	}
}

func TestCleanup_Idempotent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "x.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mgr := NewManager(root)
	ctx := context.Background()
	wt, err := mgr.Prepare(ctx, "run-idem")
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if err := mgr.Cleanup(ctx, wt); err != nil {
		t.Fatalf("cleanup #1: %v", err)
	}
	if err := mgr.Cleanup(ctx, wt); err != nil {
		t.Fatalf("cleanup #2 (idempotent) should not error, got: %v", err)
	}
}
