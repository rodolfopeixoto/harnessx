// SPDX-License-Identifier: MIT

package detectors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/cleanup"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type Worktrees struct{}

func (Worktrees) Name() string { return constants.KindCleanupWorktree }

func (Worktrees) Detect(_ context.Context, root string) ([]cleanup.Finding, error) {
	var out []cleanup.Finding
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base != "worktrees" {
			return nil
		}
		parent := filepath.Base(filepath.Dir(path))
		if parent != ".git" {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}
		for _, e := range entries {
			child := filepath.Join(path, e.Name())
			stat, err := os.Stat(child)
			if err != nil {
				continue
			}
			risk := cleanup.RiskMedium
			if time.Since(stat.ModTime()).Hours() > constants.CleanupStaleThresholdHours {
				risk = cleanup.RiskHigh
			}
			if strings.HasPrefix(e.Name(), "active-") {
				risk = cleanup.RiskLow
			}
			out = append(out, cleanup.Finding{
				Kind:        constants.KindCleanupWorktree,
				Path:        child,
				Risk:        risk,
				Reason:      "git worktree under " + child,
				SizeBytes:   stat.Size(),
				LastTouched: stat.ModTime(),
			})
		}
		return filepath.SkipDir
	})
	return out, err
}
