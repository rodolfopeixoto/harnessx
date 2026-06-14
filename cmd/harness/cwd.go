// SPDX-License-Identifier: MIT

package main

import "os"

// cwd returns the current working directory or an error suitable for a
// Cobra RunE return. Centralising the call avoids repeating the same
// three-line dance in every subcommand.
func cwd() (string, error) { return os.Getwd() }
