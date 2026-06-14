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

type VMLeftovers struct{}

var vmFingerprints = map[string]string{
	".vagrant":       "vagrant working dir",
	".parallels":     "parallels working dir",
	".vmware":        "vmware working dir",
	"VirtualBox VMs": "virtualbox vm dir",
}

func (VMLeftovers) Name() string { return constants.KindCleanupVMLeftover }

func (VMLeftovers) Detect(_ context.Context, root string) ([]cleanup.Finding, error) {
	return scanFingerprints(root, vmFingerprints, constants.KindCleanupVMLeftover, cleanup.RiskMedium), nil
}

type ClaudeLeftovers struct{}

var claudeFingerprints = map[string]string{
	".claude":          "claude config",
	".claude-cache":    "claude cache",
	".claude-history":  "claude history",
	".claude-leftover": "orphan claude directory",
}

func (ClaudeLeftovers) Name() string { return constants.KindCleanupClaudeLeftover }

func (ClaudeLeftovers) Detect(_ context.Context, root string) ([]cleanup.Finding, error) {
	return scanFingerprints(root, claudeFingerprints, constants.KindCleanupClaudeLeftover, cleanup.RiskLow), nil
}

func scanFingerprints(root string, prints map[string]string, kind string, baseRisk cleanup.Risk) []cleanup.Finding {
	var out []cleanup.Finding
	for suffix, reason := range prints {
		path := filepath.Join(root, suffix)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		risk := baseRisk
		if time.Since(info.ModTime()).Hours() > constants.CleanupStaleThresholdHours {
			risk = cleanup.RiskHigh
		}
		out = append(out, cleanup.Finding{
			Kind:        kind,
			Path:        path,
			Risk:        risk,
			Reason:      reason,
			SizeBytes:   sizeOf(info),
			LastTouched: info.ModTime(),
		})
	}
	return out
}

func sizeOf(info os.FileInfo) int64 {
	if info.IsDir() {
		return dirSize(info.Name())
	}
	return info.Size()
}
