// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree isolates an agent run in a dedicated git worktree under
// .harness/worktrees/<run-id>. P31 prefers worktrees so the agent can
// modify files without touching the project root, sensors run against
// the isolated tree, and a clean discard is `git worktree remove`.
//
// Callers fall back to controlled temp copies only when the project
// isn't a git repo (handled by Manager.Prepare returning kind=copy).
type Worktree struct {
	RunID    string
	Kind     string // "git_worktree" | "copy"
	Path     string
	Branch   string
	BaseHead string
}

type Manager struct {
	ProjectRoot string
}

func NewManager(projectRoot string) *Manager {
	return &Manager{ProjectRoot: projectRoot}
}

// Prepare creates an isolation workspace for runID. Returns the workspace
// path and the kind chosen. The caller invokes Cleanup when done.
func (m *Manager) Prepare(ctx context.Context, runID string) (Worktree, error) {
	if runID == "" {
		return Worktree{}, errors.New("execution: runID required")
	}
	wtDir := filepath.Join(m.ProjectRoot, ".harness", "worktrees", runID)
	if err := os.MkdirAll(filepath.Dir(wtDir), 0o755); err != nil {
		return Worktree{}, fmt.Errorf("execution: mkdir worktrees parent: %w", err)
	}
	if isGitRepo(m.ProjectRoot) {
		base, err := gitHead(ctx, m.ProjectRoot)
		if err != nil {
			return Worktree{}, fmt.Errorf("execution: read HEAD: %w", err)
		}
		branch := fmt.Sprintf("harness/run/%s", runID)
		cmd := exec.CommandContext(ctx, "git", "worktree", "add", "-b", branch, wtDir, base)
		cmd.Dir = m.ProjectRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			return Worktree{}, fmt.Errorf("execution: git worktree add: %w: %s", err, strings.TrimSpace(string(out)))
		}
		return Worktree{RunID: runID, Kind: "git_worktree", Path: wtDir, Branch: branch, BaseHead: base}, nil
	}
	if err := copyTree(m.ProjectRoot, wtDir); err != nil {
		return Worktree{}, fmt.Errorf("execution: copy tree: %w", err)
	}
	return Worktree{RunID: runID, Kind: "copy", Path: wtDir}, nil
}

// Cleanup removes the worktree (git worktree remove + branch delete when
// applicable, rm -rf for plain copies). Idempotent.
func (m *Manager) Cleanup(ctx context.Context, wt Worktree) error {
	if wt.Path == "" {
		return nil
	}
	if _, err := os.Stat(wt.Path); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if wt.Kind == "git_worktree" {
		rm := exec.CommandContext(ctx, "git", "worktree", "remove", "--force", wt.Path)
		rm.Dir = m.ProjectRoot
		_ = rm.Run()
		if wt.Branch != "" {
			br := exec.CommandContext(ctx, "git", "branch", "-D", wt.Branch)
			br.Dir = m.ProjectRoot
			_ = br.Run()
		}
	}
	return os.RemoveAll(wt.Path)
}

func isGitRepo(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil
}

func gitHead(ctx context.Context, root string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		if strings.HasPrefix(rel, ".harness"+string(os.PathSeparator)) || strings.HasPrefix(rel, ".git"+string(os.PathSeparator)) {
			return filepath.SkipDir
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
