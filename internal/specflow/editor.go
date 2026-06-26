// SPDX-License-Identifier: MIT

package specflow

import (
	"os"
	"os/exec"
)

// editor + path come from the user's $EDITOR env and an .harness-internal
// spec path; both are already trusted by every other code path that
// shells out, so the gosec G702 here is a known false-positive.
func defaultRunEditor(editor, path string) error {
	cmd := exec.Command(editor, path) //nolint:gosec // editor + path are trusted local inputs

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
