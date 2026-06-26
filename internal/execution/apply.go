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

// ErrApplyConflict signals the patch would not merge cleanly. Callers
// MUST treat this as a terminal non-applied state — applying anyway
// would leak `<<<<<<< ours` markers into the user's working tree.
var ErrApplyConflict = errors.New("apply: patch conflicts with project root")

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

		checkCmd := exec.CommandContext(ctx, "git", "apply", "--check", "--3way", patch)
		checkCmd.Dir = projectRoot
		if checkOut, checkErr := checkCmd.CombinedOutput(); checkErr != nil {
			dumpRejectedHunks(patch, runDir, checkOut)
			return fmt.Errorf("%w: %s", ErrApplyConflict, strings.TrimSpace(string(checkOut)))
		}

		cmd := exec.CommandContext(ctx, "git", "apply", "--3way", "--index", patch)
		cmd.Dir = projectRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			dumpRejectedHunks(patch, runDir, out)
			return fmt.Errorf("%w: %s", ErrApplyConflict, strings.TrimSpace(string(out)))
		}
		return nil
	}
	return applyCopy(wt, projectRoot, runDir)
}

func dumpRejectedHunks(patch, runDir string, checkOutput []byte) {
	rejDir := filepath.Join(runDir, "rejects")
	if err := os.MkdirAll(rejDir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(rejDir, "git-apply-check.stderr"), checkOutput, 0o644)
	if data, err := os.ReadFile(patch); err == nil {
		_ = os.WriteFile(filepath.Join(rejDir, "diff.patch"), data, 0o644)
	}
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
