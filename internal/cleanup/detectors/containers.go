// SPDX-License-Identifier: MIT

package detectors

import (
	"context"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/cleanup"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/runtime/containers"
)

type Containers struct {
	Lister containers.Lister
}

func (c Containers) Name() string { return constants.KindCleanupContainer }

func (c Containers) Detect(ctx context.Context, _ string) ([]cleanup.Finding, error) {
	lister := c.Lister
	if lister == nil {
		lister = containers.RealLister{}
	}
	items, err := lister.List(ctx)
	if err != nil {
		return nil, err
	}
	var out []cleanup.Finding
	for _, item := range items {
		out = append(out, cleanup.Finding{
			Kind:        constants.KindCleanupContainer,
			Path:        item.ID,
			Risk:        riskFor(item),
			Reason:      "container " + item.Name + " (" + item.Status + ")",
			SizeBytes:   item.SizeBytes,
			LastTouched: item.CreatedAt,
		})
	}
	return out, nil
}

func riskFor(item containers.Item) cleanup.Risk {
	if strings.HasPrefix(item.Status, "Exited") {
		if time.Since(item.CreatedAt).Hours() > constants.CleanupStaleThresholdHours {
			return cleanup.RiskHigh
		}
		return cleanup.RiskMedium
	}
	return cleanup.RiskLow
}
