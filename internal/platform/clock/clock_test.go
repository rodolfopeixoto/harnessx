package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReal_Now_UTC(t *testing.T) {
	got := Real{}.Now()
	require.Equal(t, time.UTC, got.Location())
	require.WithinDuration(t, time.Now(), got, time.Second)
}

func TestFake_Advance(t *testing.T) {
	start := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	f := &Fake{T: start}
	require.Equal(t, start, f.Now())
	f.Advance(2 * time.Hour)
	require.Equal(t, start.Add(2*time.Hour), f.Now())
}
