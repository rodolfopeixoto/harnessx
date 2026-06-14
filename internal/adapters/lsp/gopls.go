// SPDX-License-Identifier: MIT

package lsp

// NewGopls returns a Stdio LSP client configured for gopls. The shared
// machinery lives in stdio_client.go; this file is intentionally thin so
// each language server's quirks (binary, args, languageId, init opts) can
// evolve in isolation.
func NewGopls(root string) *Stdio {
	return NewStdio("gopls", []string{"-mode=stdio"}, "go", "go", root)
}
