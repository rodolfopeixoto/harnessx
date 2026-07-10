// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

const (
	kindGitWorktree = "git_worktree"
	kindCopy        = "copy"
	branchPrefix    = "harness/run/"
)

// validRunID enforces safe path components: alphanumeric plus `_` and `-`.
// Blocks path traversal (`..`, `/`, `\`) and shell metacharacters.
var validRunID = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Prepare creates an isolation workspace for runID. Returns the workspace
// path and the kind chosen. The caller invokes Cleanup when done.
func (m *Manager) Prepare(ctx context.Context, runID string) (Worktree, error) {
	if runID == "" {
		return Worktree{}, errors.New("execution: runID required")
	}
	if !validRunID.MatchString(runID) {
		return Worktree{}, fmt.Errorf("execution: invalid runID %q: must match [a-zA-Z0-9_-]+", runID)
	}
	wtDir := filepath.Join(m.ProjectRoot, ".harness", "worktrees", runID)
	if err := os.MkdirAll(filepath.Dir(wtDir), 0o755); err != nil {
		return Worktree{}, fmt.Errorf("execution: mkdir worktrees parent: %w", err)
	}
	if isGitRepo(m.ProjectRoot) && hasHead(ctx, m.ProjectRoot) {
		base, err := gitHead(ctx, m.ProjectRoot)
		if err != nil {
			return Worktree{}, fmt.Errorf("execution: read HEAD: %w", err)
		}
		branch := branchPrefix + runID
		// Clear any residue from a previous crashed run so `git worktree add`
		// doesn't fail with "already exists" or "already checked out".
		m.reclaimStaleWorktree(ctx, wtDir, branch)
		cmd := exec.CommandContext(ctx, "git", "worktree", "add", "-b", branch, wtDir, base)
		cmd.Dir = m.ProjectRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			return Worktree{}, fmt.Errorf("execution: git worktree add: %w: %s", err, strings.TrimSpace(string(out)))
		}
		writeWorktreeGitignore(wtDir)
		return Worktree{RunID: runID, Kind: kindGitWorktree, Path: wtDir, Branch: branch, BaseHead: base}, nil
	}
	if isGitRepo(m.ProjectRoot) {
		// Repo exists but HEAD is unborn (fresh `git init` with no commits).
		// Fall back to copy so the run works without mutating the user's git state.
		slog.Debug("execution: git worktree fallback to copy", "reason", "empty repo (no HEAD)", "root", m.ProjectRoot)
	}
	// Same reclaim story for the copy branch: a leftover directory from a
	// crashed prior run would cause files to mingle.
	if _, err := os.Stat(wtDir); err == nil {
		slog.Warn("execution: removing stale copy worktree", "path", wtDir)
		if rmErr := os.RemoveAll(wtDir); rmErr != nil {
			return Worktree{}, fmt.Errorf("execution: remove stale copy dir: %w", rmErr)
		}
	}
	if err := copyTree(m.ProjectRoot, wtDir); err != nil {
		return Worktree{}, fmt.Errorf("execution: copy tree: %w", err)
	}
	return Worktree{RunID: runID, Kind: kindCopy, Path: wtDir}, nil
}

// reclaimStaleWorktree tries to recover from a previous crashed run leaving
// behind (a) a directory git no longer tracks, (b) a registered worktree at
// the same path, or (c) a lingering branch. Best-effort: failures are logged
// and the outer `git worktree add` will surface the real error.
func (m *Manager) reclaimStaleWorktree(ctx context.Context, wtDir, branch string) {
	pathExists := false
	if _, err := os.Stat(wtDir); err == nil {
		pathExists = true
	}

	// Ask git if it still tracks this worktree; if so, remove it cleanly.
	if pathExists {
		slog.Warn("execution: reclaiming stale worktree", "path", wtDir, "branch", branch)
		rm := exec.CommandContext(ctx, "git", "worktree", "remove", "--force", wtDir)
		rm.Dir = m.ProjectRoot
		if out, err := rm.CombinedOutput(); err != nil {
			slog.Debug("execution: worktree remove failed (may be orphan)",
				"err", err, "out", strings.TrimSpace(string(out)))
		}
		// Whether git knew about it or not, drop the directory.
		if err := os.RemoveAll(wtDir); err != nil {
			slog.Warn("execution: could not remove stale worktree dir", "path", wtDir, "err", err)
		}
	}

	// Prune stale registrations regardless — even without a directory,
	// git may still have the entry in .git/worktrees.
	prune := exec.CommandContext(ctx, "git", "worktree", "prune")
	prune.Dir = m.ProjectRoot
	if out, err := prune.CombinedOutput(); err != nil {
		slog.Debug("execution: worktree prune failed",
			"err", err, "out", strings.TrimSpace(string(out)))
	}

	// Delete the branch if it exists from a prior aborted run.
	if branchExists(ctx, m.ProjectRoot, branch) {
		slog.Warn("execution: deleting stale branch", "branch", branch)
		br := exec.CommandContext(ctx, "git", "branch", "-D", branch)
		br.Dir = m.ProjectRoot
		if out, err := br.CombinedOutput(); err != nil {
			slog.Debug("execution: branch delete failed",
				"branch", branch, "err", err, "out", strings.TrimSpace(string(out)))
		}
	}
}

// Cleanup removes the worktree (git worktree remove + branch delete when
// applicable, rm -rf for plain copies). Idempotent. Surfaces the final
// filesystem error rather than swallowing it, but git-level failures are
// logged and best-effort.
func (m *Manager) Cleanup(ctx context.Context, wt Worktree) error {
	if wt.Path == "" {
		return nil
	}
	if _, err := os.Stat(wt.Path); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if wt.Kind == kindGitWorktree {
		rm := exec.CommandContext(ctx, "git", "worktree", "remove", "--force", wt.Path)
		rm.Dir = m.ProjectRoot
		if out, err := rm.CombinedOutput(); err != nil {
			slog.Warn("execution: git worktree remove failed",
				"path", wt.Path, "err", err, "out", strings.TrimSpace(string(out)))
		}
		if wt.Branch != "" {
			br := exec.CommandContext(ctx, "git", "branch", "-D", wt.Branch)
			br.Dir = m.ProjectRoot
			if out, err := br.CombinedOutput(); err != nil {
				slog.Warn("execution: git branch delete failed",
					"branch", wt.Branch, "err", err, "out", strings.TrimSpace(string(out)))
			}
		}
	}
	if err := os.RemoveAll(wt.Path); err != nil {
		slog.Error("execution: RemoveAll worktree path failed", "path", wt.Path, "err", err)
		return fmt.Errorf("execution: cleanup remove %s: %w", wt.Path, err)
	}
	return nil
}

func isGitRepo(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil
}

// hasHead reports whether the repository at root has a resolvable HEAD.
// Returns false for a fresh `git init` where no commit exists yet (unborn HEAD).
func hasHead(ctx context.Context, root string) bool {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", "--quiet", "HEAD")
	cmd.Dir = root
	return cmd.Run() == nil
}

// branchExists reports whether a local branch exists.
func branchExists(ctx context.Context, root, branch string) bool {
	cmd := exec.CommandContext(ctx, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = root
	return cmd.Run() == nil
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

var defaultWorktreeSkipDirs = []string{
	".harness", ".git", ".venv", "venv", "__pycache__", "node_modules",
	"target", "dist", "build", "_build", "deps", ".gradle", ".idea",
	".pytest_cache", ".mypy_cache", ".ruff_cache", "bin", "obj", "coverage",
}

func writeWorktreeGitignore(dir string) {
	body := "# harness: auto-excluded from agent worktree diffs\n"
	for _, d := range defaultWorktreeSkipDirs {
		if d == ".harness" || d == ".git" {
			continue
		}
		body += d + "/\n"
	}
	body += "*.tsbuildinfo\n*.log\n"
	dotGit := filepath.Join(dir, ".git")
	info, err := os.Stat(dotGit)
	if err != nil {
		return
	}
	infoDir := ""
	if info.IsDir() {
		infoDir = filepath.Join(dotGit, "info")
	} else {
		data, err := os.ReadFile(dotGit)
		if err != nil {
			return
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "gitdir:") {
				infoDir = filepath.Join(strings.TrimSpace(strings.TrimPrefix(line, "gitdir:")), "info")
				break
			}
		}
	}
	if infoDir == "" {
		return
	}
	_ = os.MkdirAll(infoDir, 0o755)
	excludePath := filepath.Join(infoDir, "exclude")
	existing, _ := os.ReadFile(excludePath)
	full := string(existing) + "\n" + body
	_ = os.WriteFile(excludePath, []byte(full), 0o644)
}

func isSkippedWorktreePart(part string) bool {
	for _, d := range defaultWorktreeSkipDirs {
		if part == d {
			return true
		}
	}
	return false
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
		first := strings.SplitN(rel, string(os.PathSeparator), 2)[0]
		if isSkippedWorktreePart(first) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
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
