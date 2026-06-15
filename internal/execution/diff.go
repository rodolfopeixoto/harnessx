// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CaptureDiff records the post-agent state of the worktree as three
// artifacts under runDir: diff.patch (unified), diff-stat.txt, and
// changed-files.json. Returns the list of changed paths (worktree
// relative) or empty slice when no diff exists.
func CaptureDiff(ctx context.Context, wt Worktree, runDir string) ([]string, error) {
	if wt.Path == "" {
		return nil, errors.New("execution: worktree path required")
	}
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return nil, fmt.Errorf("execution: mkdir runDir: %w", err)
	}
	if wt.Kind != "git_worktree" {
		return captureCopyDiff(wt, runDir)
	}
	staging := exec.CommandContext(ctx, "git", "add", "-A")
	staging.Dir = wt.Path
	if out, err := staging.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("execution: git add: %w: %s", err, strings.TrimSpace(string(out)))
	}
	names := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-only")
	names.Dir = wt.Path
	nameOut, err := names.Output()
	if err != nil {
		return nil, fmt.Errorf("execution: git diff names: %w", err)
	}
	changed := splitLines(string(nameOut))
	patch := exec.CommandContext(ctx, "git", "diff", "--cached")
	patch.Dir = wt.Path
	patchOut, err := patch.Output()
	if err != nil {
		return nil, fmt.Errorf("execution: git diff patch: %w", err)
	}
	stat := exec.CommandContext(ctx, "git", "diff", "--cached", "--stat")
	stat.Dir = wt.Path
	statOut, _ := stat.Output()
	if err := os.WriteFile(filepath.Join(runDir, "diff.patch"), patchOut, 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(runDir, "diff-stat.txt"), statOut, 0o644); err != nil {
		return nil, err
	}
	cfBytes, _ := json.MarshalIndent(changed, "", "  ")
	if err := os.WriteFile(filepath.Join(runDir, "changed-files.json"), cfBytes, 0o644); err != nil {
		return nil, err
	}
	return changed, nil
}

func captureCopyDiff(wt Worktree, runDir string) ([]string, error) {
	var changed []string
	err := filepath.Walk(wt.Path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(wt.Path, p)
		if err != nil {
			return err
		}
		changed = append(changed, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	cfBytes, _ := json.MarshalIndent(changed, "", "  ")
	return changed, os.WriteFile(filepath.Join(runDir, "changed-files.json"), cfBytes, 0o644)
}

func splitLines(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
