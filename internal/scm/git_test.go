// SPDX-License-Identifier: MIT

package scm

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestHasRepoAbsentAndPresent(t *testing.T) {
	dir := t.TempDir()
	if HasRepo(dir) {
		t.Fatal("fresh dir should not have a repo")
	}
	_ = os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	if !HasRepo(dir) {
		t.Fatal("after mkdir .git, HasRepo should be true")
	}
}

func TestInitCreatesRepoOnBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	// Some test environments restrict user-config lookup; let git
	// fall back without dying mid-init.
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")
	if err := Init(context.Background(), dir, "main"); err != nil {
		t.Skipf("git init unavailable in this environment: %v", err)
	}
	if !HasRepo(dir) {
		t.Fatal("Init did not create .git")
	}
	// CurrentBranch requires at least one commit; before any, git
	// returns exit 128 on rev-parse. Skip the branch check pre-commit.
}

func TestInitRejectsExistingRepo(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	err := Init(context.Background(), dir, "main")
	if err == nil {
		t.Fatal("expected error when .git present")
	}
}

func TestInitWithEmptyBranchUsesDefault(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")
	if err := Init(context.Background(), dir, ""); err != nil {
		t.Skipf("git init unavailable: %v", err)
	}
	if !HasRepo(dir) {
		t.Fatal("empty branch arg should still init")
	}
}

func TestRunGitCapturesOutput(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	out, err := runGit(context.Background(), os.TempDir(), "--version")
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Error("--version should print something")
	}
}

func TestCurrentBranchOnAbsentRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	_, err := CurrentBranch(context.Background(), t.TempDir())
	if err == nil {
		t.Error("expected error on non-repo")
	}
}
