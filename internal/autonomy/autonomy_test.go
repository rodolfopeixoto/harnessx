// SPDX-License-Identifier: MIT

package autonomy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGate_KnownLevels(t *testing.T) {
	cases := []struct {
		level Level
		op    Operation
		want  Decision
	}{
		{Manual, OpRead, DecisionAllow},
		{Manual, OpExecuteLowRisk, DecisionDeny},
		{PlanAndAsk, OpExecuteLowRisk, DecisionApproval},
		{SafeExecute, OpExecuteLowRisk, DecisionAllow},
		{SafeExecute, OpExecuteHighRisk, DecisionApproval},
		{FullProjectLoop, OpClean, DecisionApproval},
		{ScheduledMaintenance, OpSchedule, DecisionAllow},
		{ScheduledMaintenance, OpExecuteHighRisk, DecisionDeny},
	}
	for _, c := range cases {
		d, err := Gate(c.level, c.op)
		require.NoError(t, err)
		require.Equal(t, c.want, d, "%s/%s", c.level, c.op)
	}
}

func TestGate_UnknownLevelErrors(t *testing.T) {
	_, err := Gate(Level("nope"), OpRead)
	require.ErrorIs(t, err, ErrUnknownLevel)
}

func TestGate_UnknownOperationDenies(t *testing.T) {
	d, err := Gate(Manual, Operation("frobnicate"))
	require.NoError(t, err)
	require.Equal(t, DecisionDeny, d)
}

func TestAllLevels(t *testing.T) {
	require.Len(t, AllLevels(), 5)
}

func TestDefaultSetting(t *testing.T) {
	require.Equal(t, PlanAndAsk, DefaultSetting().Level)
}
