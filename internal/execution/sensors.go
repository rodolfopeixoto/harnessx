// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/index"
	"github.com/ropeixoto/harnessx/internal/sensors"
)

// RunSensors executes the supplied sensors against the worktree path,
// writes per-sensor artifacts under runDir/sensors/, and maps each
// sensors.Result into a SensorOutcome.
//
// Sensors are configured by the workflow plan; this bridge stays
// independent of which catalog of sensors is plugged in so adding new
// ones (vet, build, audit) doesn't require touching the executor.
func RunSensors(ctx context.Context, list []sensors.Sensor, profile index.Profile, root, runDir string) []SensorOutcome {
	if len(list) == 0 {
		return nil
	}
	rc := sensors.RunCtx{
		Ctx:       ctx,
		Root:      root,
		OutputDir: filepath.Join(runDir, "sensors"),
	}
	runner := sensors.Runner{}
	results := runner.Run(ctx, applicable(list, profile), rc)
	out := make([]SensorOutcome, 0, len(results))
	for _, r := range results {
		out = append(out, SensorOutcome{
			ID:         r.ID,
			Status:     string(r.Status),
			Output:     r.OutputPath,
			DurationMs: r.Duration.Milliseconds(),
		})
	}
	return out
}

func applicable(in []sensors.Sensor, p index.Profile) []sensors.Sensor {
	out := make([]sensors.Sensor, 0, len(in))
	for _, s := range in {
		if s.AppliesTo(p) {
			out = append(out, s)
		}
	}
	return out
}
