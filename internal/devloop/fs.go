// SPDX-License-Identifier: MIT

package devloop

import "os"

var (
	statFn    = os.Stat
	mkdirAll  = func(p string) error { return os.MkdirAll(p, 0o755) }
	writeFile = func(p string, data []byte) error { return os.WriteFile(p, data, 0o644) }
)
