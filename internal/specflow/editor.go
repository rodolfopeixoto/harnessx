// SPDX-License-Identifier: MIT

package specflow

import (
	"os"
	"os/exec"
)

func defaultRunEditor(editor, path string) error {
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
