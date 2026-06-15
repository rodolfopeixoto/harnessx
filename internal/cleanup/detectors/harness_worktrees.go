// SPDX-License-Identifier: MIT

package detectors

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/cleanup"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type HarnessWorktrees struct{}

func (HarnessWorktrees) Name() string { return constants.KindCleanupWorktree }

func (HarnessWorktrees) Detect(_ context.Context, root string) ([]cleanup.Finding, error) {
	dir := filepath.Join(root, ".harness", "worktrees")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []cleanup.Finding
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		child := filepath.Join(dir, e.Name())
		stat, err := os.Stat(child)
		if err != nil {
			continue
		}
		risk := cleanup.RiskMedium
		if time.Since(stat.ModTime()).Hours() > constants.CleanupStaleThresholdHours {
			risk = cleanup.RiskHigh
		}
		out = append(out, cleanup.Finding{
			Kind:        constants.KindCleanupWorktree,
			Path:        child,
			Risk:        risk,
			Reason:      "orphan harness worktree under " + child,
			SizeBytes:   dirSize(child),
			LastTouched: stat.ModTime(),
		})
	}
	return out, nil
}
