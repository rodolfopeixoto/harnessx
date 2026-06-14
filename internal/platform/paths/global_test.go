// SPDX-License-Identifier: MIT

package paths

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func TestGlobalHarnessDir_HonoursHarnessHome(t *testing.T) {
	t.Setenv(constants.EnvHarnessHome, "/tmp/override-home")
	t.Setenv("XDG_DATA_HOME", "/should/be/ignored")
	require.Equal(t, "/tmp/override-home", GlobalHarnessDir())
}

func TestGlobalHarnessDir_FallsBackToXDG(t *testing.T) {
	t.Setenv(constants.EnvHarnessHome, "")
	t.Setenv("XDG_DATA_HOME", "/tmp/xdg")
	require.Equal(t, filepath.Join("/tmp/xdg", constants.GlobalHarnessDirName), GlobalHarnessDir())
}

func TestGlobalHarnessDir_PlatformFallback(t *testing.T) {
	t.Setenv(constants.EnvHarnessHome, "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "/tmp/home")
	got := GlobalHarnessDir()
	require.True(t, strings.HasPrefix(got, "/tmp/home/"), "unexpected dir %s", got)
	switch runtime.GOOS {
	case "darwin":
		require.Contains(t, got, "Library/Application Support")
	case "windows":
		// Skip on non-windows test environments.
	default:
		require.Contains(t, got, ".local/share")
	}
}

func TestGlobalRegistryPath_JoinsFilename(t *testing.T) {
	t.Setenv(constants.EnvHarnessHome, "/tmp/h")
	require.Equal(t, "/tmp/h/registry.sqlite", GlobalRegistryPath())
	require.Equal(t, "/tmp/h/registry.lock", GlobalRegistryLockPath())
}
