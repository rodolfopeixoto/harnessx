package optimize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePercent(t *testing.T) {
	require.InDelta(t, 12.5, parsePercent("12.5%"), 0.001)
	require.InDelta(t, 0, parsePercent(""), 0.001)
}

func TestSplitMemUsage(t *testing.T) {
	u, l := splitMemUsage("12.34MiB / 1.95GiB")
	require.InDelta(t, 12.34, u, 0.001)
	require.InDelta(t, 1996.8, l, 0.5)
}

func TestParseBytes_Units(t *testing.T) {
	require.InDelta(t, 1, parseBytes("1MiB"), 0.001)
	require.InDelta(t, 1024, parseBytes("1GiB"), 0.5)
	require.InDelta(t, 0.5, parseBytes("512KiB"), 0.001)
}

func TestCaptureRuntime_ReportsHost(t *testing.T) {
	m := captureRuntime()
	require.Greater(t, m.ProcessNumGoroutines, 0)
	require.Greater(t, m.ProcessRSSMB, 0.0)
}
