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

// ApplyWorktreeDiff brings the worktree's diff into the project root. For
// git worktrees: apply diff.patch with git apply --3way. For copy mode:
// rsync-like copy of changed files. Both refuse to run if the project
// root has uncommitted changes that overlap the patch.
func ApplyWorktreeDiff(ctx context.Context, projectRoot string, wt Worktree, runDir string) error {
	if wt.Kind == "git_worktree" {
		patch := filepath.Join(runDir, "diff.patch")
		info, err := os.Stat(patch)
		if err != nil {
			return fmt.Errorf("apply: stat patch: %w", err)
		}
		if info.Size() == 0 {
			return errors.New("apply: empty patch")
		}
		cmd := exec.CommandContext(ctx, "git", "apply", "--3way", "--index", patch)
		cmd.Dir = projectRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("apply: git apply: %w: %s", err, strings.TrimSpace(string(out)))
		}
		return nil
	}
	return applyCopy(wt, projectRoot, runDir)
}

func applyCopy(wt Worktree, projectRoot, runDir string) error {
	return filepath.Walk(wt.Path, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(wt.Path, p)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		target := filepath.Join(projectRoot, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
