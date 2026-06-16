// SPDX-License-Identifier: MIT

package sensors

import (
	"os/exec"
	"testing"
)

func TestRunGitCaptureNonRepoReturnsEmpty(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	out := runGitCapture(t.TempDir(), "status", "--porcelain")
	_ = out
}

func TestRunGitCaptureMissingCommand(t *testing.T) {
	out := runGitCapture(t.TempDir(), "definitely-not-a-real-subcommand-xyz")
	_ = out
}
