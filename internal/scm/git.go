// SPDX-License-Identifier: MIT

// Package scm wraps the minimum git operations harness needs (presence
// check, init, branch query). Keeps git off the hot path so the rest of
// the codebase never shells out directly.
package scm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func HasRepo(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil
}

func Init(ctx context.Context, root, branch string) error {
	if HasRepo(root) {
		return errors.New("scm: .git already present")
	}
	if branch == "" {
		branch = "main"
	}
	if out, err := runGit(ctx, root, "init", "-b", branch); err == nil {
		_ = out
		return nil
	}
	if out, err := runGit(ctx, root, "init"); err != nil {
		return fmt.Errorf("git init: %w: %s", err, strings.TrimSpace(out))
	}
	if out, err := runGit(ctx, root, "checkout", "-b", branch); err != nil {
		return fmt.Errorf("git checkout -b %s: %w: %s", branch, err, strings.TrimSpace(out))
	}
	return nil
}

func CurrentBranch(ctx context.Context, root string) (string, error) {
	out, err := runGit(ctx, root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func runGit(ctx context.Context, root string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	return string(out), err
}
