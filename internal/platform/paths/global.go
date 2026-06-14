// SPDX-License-Identifier: MIT

package paths

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

// GlobalHarnessDir resolves the user-scoped HarnessX home where the
// cross-project registry lives. Precedence:
//  1. HARNESS_HOME (escape hatch for tests and packagers).
//  2. XDG_DATA_HOME/harness (Linux/macOS spec).
//  3. ~/.local/share/harness on Linux, ~/Library/Application Support/harness on macOS,
//     %LOCALAPPDATA%/harness on Windows.
//  4. ~/.harness as final fallback.
func GlobalHarnessDir() string {
	if v := os.Getenv(constants.EnvHarnessHome); v != "" {
		return v
	}
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return filepath.Join(v, constants.GlobalHarnessDirName)
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), constants.GlobalHarnessDirName)
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", constants.GlobalHarnessDirName)
	case "windows":
		if v := os.Getenv("LOCALAPPDATA"); v != "" {
			return filepath.Join(v, constants.GlobalHarnessDirName)
		}
		return filepath.Join(home, "AppData", "Local", constants.GlobalHarnessDirName)
	default:
		return filepath.Join(home, ".local", "share", constants.GlobalHarnessDirName)
	}
}

// GlobalRegistryPath returns the absolute path to the registry SQLite file.
func GlobalRegistryPath() string {
	return filepath.Join(GlobalHarnessDir(), constants.GlobalRegistryFilename)
}

// GlobalRegistryLockPath returns the absolute path to the advisory-lock file
// held during registry write transactions.
func GlobalRegistryLockPath() string {
	return filepath.Join(GlobalHarnessDir(), constants.GlobalRegistryLockFilename)
}
