// SPDX-License-Identifier: MIT

package lsp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

func repoHash(root string) string {
	sum := sha256.Sum256([]byte(root))
	return hex.EncodeToString(sum[:8])
}

func queryHash(method, path string) string {
	sum := sha256.Sum256([]byte(method + "|" + path))
	return hex.EncodeToString(sum[:16])
}

type cachedResult struct {
	Symbols     []Symbol     `json:"symbols,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

func readCache(path string) (cachedResult, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return cachedResult{}, false
	}
	var c cachedResult
	if err := json.Unmarshal(b, &c); err != nil {
		return cachedResult{}, false
	}
	return c, true
}

func writeCache(path string, c cachedResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
