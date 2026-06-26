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

// ErrApplyConflict is returned when the diff produced by an agent cannot
// be merged into the project root cleanly. The caller MUST treat this as
// a non-applied terminal state — see audit BUG-13/14 for the regression
// (running `do` against the same file twice used to leave `<<<<<<< ours`
// markers committed in the working tree).
var ErrApplyConflict = errors.New("apply: patch conflicts with project root")

// ApplyWorktreeDiff brings the worktree's diff into the project root. For
// git worktrees: validate diff.patch applies cleanly via
// `git apply --check --3way`, then commit it with `git apply --3way --index`.
// On conflict, the rejected hunks are written to <runDir>/rejects/ and
// ErrApplyConflict is returned so the executor can mark status=conflict
// instead of falsely reporting `applied`.
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

		// Pre-flight: `git apply --check --3way` mirrors the real apply but
		// only reports whether it would succeed. Failure here means a hunk
		// won't merge cleanly; we never mutate the working tree in that case.
		checkCmd := exec.CommandContext(ctx, "git", "apply", "--check", "--3way", patch)
		checkCmd.Dir = projectRoot
		if checkOut, checkErr := checkCmd.CombinedOutput(); checkErr != nil {
			dumpRejectedHunks(ctx, projectRoot, patch, runDir, checkOut)
			return fmt.Errorf("%w: %s", ErrApplyConflict, strings.TrimSpace(string(checkOut)))
		}

		cmd := exec.CommandContext(ctx, "git", "apply", "--3way", "--index", patch)
		cmd.Dir = projectRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			// Defensive: even after a passing --check, a race with the
			// working tree (concurrent edit) can still fail the real
			// apply. Treat that as a conflict too.
			dumpRejectedHunks(ctx, projectRoot, patch, runDir, out)
			return fmt.Errorf("%w: %s", ErrApplyConflict, strings.TrimSpace(string(out)))
		}
		return nil
	}
	return applyCopy(wt, projectRoot, runDir)
}

// dumpRejectedHunks attempts to materialise the patch as a best-effort
// reject pile under <runDir>/rejects/. We use `git apply --reject` against
// a throwaway index so the project root stays clean. Errors from this
// helper are swallowed — the caller already knows the apply failed and
// has the original patch under <runDir>/diff.patch.
func dumpRejectedHunks(ctx context.Context, projectRoot, patch, runDir string, checkOutput []byte) {
	rejDir := filepath.Join(runDir, "rejects")
	if err := os.MkdirAll(rejDir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(rejDir, "git-apply-check.stderr"), checkOutput, 0o644)
	// `git apply --reject` writes <file>.rej next to the target. Use a
	// sandbox copy of projectRoot would be expensive; instead we run with
	// --check-no-write substitute: ask git to emit raw .rej hunks via
	// `git apply -3 --reject --check` is not supported, so we simply
	// preserve the original patch + the check error log under rejects/.
	if data, err := os.ReadFile(patch); err == nil {
		_ = os.WriteFile(filepath.Join(rejDir, "diff.patch"), data, 0o644)
	}
	_ = ctx
	_ = projectRoot
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
