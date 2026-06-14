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

type Caches struct{}

var cacheTargets = []string{
	"node_modules/.cache",
	".npm",
	".pnpm-store",
	".cargo/target",
	".cache/pip",
	"go/pkg/mod/cache",
	"vendor/cache",
}

func (Caches) Name() string { return constants.KindCleanupCache }

func (Caches) Detect(_ context.Context, root string) ([]cleanup.Finding, error) {
	var out []cleanup.Finding
	for _, suffix := range cacheTargets {
		path := filepath.Join(root, suffix)
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		size := dirSize(path)
		risk := cleanup.RiskLow
		if size > constants.CleanupLargeFileThresholdB {
			risk = cleanup.RiskMedium
		}
		out = append(out, cleanup.Finding{
			Kind:        constants.KindCleanupCache,
			Path:        path,
			Risk:        risk,
			Reason:      "package cache",
			SizeBytes:   size,
			LastTouched: lastTouchedIn(path),
		})
	}
	return out, nil
}

func dirSize(root string) int64 {
	var total int64
	_ = filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

func lastTouchedIn(root string) time.Time {
	var newest time.Time
	_ = filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
		return nil
	})
	return newest
}
