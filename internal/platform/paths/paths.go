// SPDX-License-Identifier: MIT

package paths

import (
	"errors"
	"os"
	"path/filepath"
)

// rootMarkers are filenames or directories whose presence identifies a
// project root. Order does not matter — the first ancestor containing any
// of them wins.
var rootMarkers = []string{
	".git",
	".harness",
	"go.mod",
	"package.json",
	"Gemfile",
	"Cargo.toml",
	"pyproject.toml",
	"requirements.txt",
}

// FindProjectRoot walks up from start looking for a marker. If none is
// found, start itself is returned. start must be an absolute path.
func FindProjectRoot(start string) (string, error) {
	if !filepath.IsAbs(start) {
		return "", errors.New("paths: start must be absolute")
	}
	cur := filepath.Clean(start)
	for {
		for _, m := range rootMarkers {
			if _, err := os.Stat(filepath.Join(cur, m)); err == nil {
				return cur, nil
			}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return start, nil
		}
		cur = parent
	}
}

// HarnessDir returns <root>/.harness.
func HarnessDir(root string) string {
	return filepath.Join(root, ".harness")
}
