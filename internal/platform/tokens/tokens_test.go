package tokens

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeuristic4_KnownInputs(t *testing.T) {
	h := Heuristic4{}
	require.Equal(t, 0, h.Estimate(""))
	require.Equal(t, 1, h.Estimate("ab"))
	require.Equal(t, 1, h.Estimate("abcd"))
	require.Equal(t, 2, h.Estimate("abcdefgh"))
	require.Equal(t, 250, h.Estimate(strings.Repeat("x", 1000)))
}

func TestEstimateBytes(t *testing.T) {
	require.Equal(t, 0, EstimateBytes(nil))
	require.Equal(t, 1, EstimateBytes([]byte("ab")))
	require.Equal(t, 25, EstimateBytes(make([]byte, 100)))
}
