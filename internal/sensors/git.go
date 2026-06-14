// SPDX-License-Identifier: MIT

package sensors

import (
	"context"
	"os/exec"
	"time"
)

// runGitCapture executes git with the given args inside dir, returning
// trimmed stdout. Bounded by 10s; errors swallowed (we only need best-effort
// data for ChangedFilesSensor).
func runGitCapture(dir string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}
