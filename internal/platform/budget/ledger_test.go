// SPDX-License-Identifier: MIT

package budget

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChargeWithRecordsLedgerEntry(t *testing.T) {
	g := New(1.0)
	err := g.ChargeWith(Entry{Label: "claude-sonnet-4-6", USD: 0.0144, Tag: "reported", Note: "in=2140 out=512"})
	require.NoError(t, err)
	entries := g.Entries()
	require.Len(t, entries, 1)
	require.Equal(t, "claude-sonnet-4-6", entries[0].Label)
	require.Equal(t, "reported", entries[0].Tag)
	require.InDelta(t, 0.0144, entries[0].USD, 1e-9)
}

func TestChargeFallsBackToUncategorised(t *testing.T) {
	g := New(1.0)
	require.NoError(t, g.Charge(0.05))
	entries := g.Entries()
	require.Len(t, entries, 1)
	require.Equal(t, "uncategorised", entries[0].Label)
}

func TestEntriesReturnsCopyNotAlias(t *testing.T) {
	g := New(1.0)
	_ = g.ChargeWith(Entry{Label: "a", USD: 0.01})
	got := g.Entries()
	got[0].Label = "mutated"
	again := g.Entries()
	require.Equal(t, "a", again[0].Label, "Entries must return a copy")
}

func TestCapAccessor(t *testing.T) {
	g := New(2.5)
	require.InDelta(t, 2.5, g.Cap(), 1e-9)
}

func TestMultipleChargesAccumulate(t *testing.T) {
	g := New(1.0)
	for i := 0; i < 5; i++ {
		_ = g.ChargeWith(Entry{Label: "step", USD: 0.10})
	}
	require.InDelta(t, 0.50, g.Spent(), 1e-9)
	require.Len(t, g.Entries(), 5)
}
