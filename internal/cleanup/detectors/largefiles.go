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

type LargeFiles struct {
	Threshold int64
}

func (l LargeFiles) Name() string { return constants.KindCleanupLargeFile }

func (l LargeFiles) Detect(_ context.Context, root string) ([]cleanup.Finding, error) {
	threshold := l.Threshold
	if threshold <= 0 {
		threshold = constants.CleanupLargeFileThresholdB
	}
	var out []cleanup.Finding
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info == nil || info.IsDir() {
			return nil
		}
		if info.Size() < threshold {
			return nil
		}
		risk := cleanup.RiskMedium
		if time.Since(info.ModTime()).Hours() > constants.CleanupStaleThresholdHours {
			risk = cleanup.RiskHigh
		}
		out = append(out, cleanup.Finding{
			Kind:        constants.KindCleanupLargeFile,
			Path:        path,
			Risk:        risk,
			Reason:      "large file over threshold",
			SizeBytes:   info.Size(),
			LastTouched: info.ModTime(),
		})
		return nil
	})
	return out, err
}
