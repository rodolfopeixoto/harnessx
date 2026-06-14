package budget

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGuard_ChargeWithinBudget(t *testing.T) {
	g := New(1.0)
	require.NoError(t, g.Charge(0.5))
	require.InDelta(t, 0.5, g.Remaining(), 1e-9)
	require.InDelta(t, 0.5, g.Spent(), 1e-9)
}

func TestGuard_ExceedsBudget(t *testing.T) {
	g := New(0.10)
	require.NoError(t, g.Charge(0.05))
	err := g.Charge(0.10)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrBudgetExceeded))
}

func TestGuard_DefaultCap(t *testing.T) {
	g := New(0)
	require.InDelta(t, 1.0, g.Remaining(), 1e-9)
}
