// SPDX-License-Identifier: MIT

package http

import "os/exec"

func binaryOnPath(bin string) bool {
	if bin == "" {
		return false
	}
	_, err := exec.LookPath(bin)
	return err == nil
}
