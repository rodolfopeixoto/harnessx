// SPDX-License-Identifier: MIT

package health

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func TestCompute_DeterministicGolden(t *testing.T) {
	in := Inputs{
		TestsPassPct:       100,
		SensorsPassPct:     100,
		SecurityFindings:   0,
		PerfBudgetExceeded: false,
		OutdatedDeps:       0,
		DocsCoverage:       80,
		DesignParityPct:    90,
		RoadmapClearPct:    70,
		MemoryFreshDays:    5,
		InvalidConfigs:     0,
	}
	score := in.Compute()
	require.Equal(t, 94, score.Total)
	require.Len(t, score.Subsystems, 10)
}

func TestCompute_WorstCase(t *testing.T) {
	in := Inputs{
		TestsPassPct:       0,
		SensorsPassPct:     0,
		SecurityFindings:   100,
		PerfBudgetExceeded: true,
		OutdatedDeps:       100,
		DocsCoverage:       0,
		DesignParityPct:    0,
		RoadmapClearPct:    0,
		MemoryFreshDays:    365,
		InvalidConfigs:     100,
	}
	score := in.Compute()
	require.LessOrEqual(t, score.Total, 5)
}

func TestPercent_NegativeUsesDefault(t *testing.T) {
	require.Equal(t, constants.HealthDefaultScore, percentOrDefault(-1))
	require.Equal(t, constants.HealthMaxScore, percentOrDefault(1000))
}

func TestClampInverseBoundaries(t *testing.T) {
	require.Equal(t, constants.HealthMaxScore, clampInverse(0, 5))
	require.Equal(t, 0, clampInverse(10, 5))
}

func TestFreshnessBuckets(t *testing.T) {
	require.Equal(t, constants.HealthMaxScore, freshnessScore(3))
	require.Equal(t, 70, freshnessScore(20))
	require.Equal(t, 40, freshnessScore(60))
	require.Equal(t, 10, freshnessScore(120))
	require.Equal(t, constants.HealthDefaultScore, freshnessScore(-1))
}
