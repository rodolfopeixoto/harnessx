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

type AbandonedHarness struct{}

func (AbandonedHarness) Name() string { return constants.KindCleanupAbandonedHX }

func (AbandonedHarness) Detect(_ context.Context, root string) ([]cleanup.Finding, error) {
	var out []cleanup.Finding
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		if filepath.Base(path) != constants.HarnessDir {
			return nil
		}
		dbPath := filepath.Join(path, constants.DBSubdir, constants.DBFilename)
		_, dbErr := os.Stat(dbPath)
		if dbErr != nil {
			out = append(out, finding(path, "missing harness db"))
			return filepath.SkipDir
		}
		stat, _ := os.Stat(dbPath)
		if stat != nil && time.Since(stat.ModTime()).Hours() > constants.CleanupStaleThresholdHours {
			out = append(out, finding(path, "stale harness db"))
		}
		return filepath.SkipDir
	})
	return out, err
}

func finding(path, reason string) cleanup.Finding {
	stat, _ := os.Stat(path)
	var (
		mod  time.Time
		size int64
	)
	if stat != nil {
		mod = stat.ModTime()
		size = stat.Size()
	}
	return cleanup.Finding{
		Kind:        constants.KindCleanupAbandonedHX,
		Path:        path,
		Risk:        cleanup.RiskHigh,
		Reason:      reason,
		SizeBytes:   size,
		LastTouched: mod,
	}
}
