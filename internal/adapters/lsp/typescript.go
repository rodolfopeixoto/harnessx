// SPDX-License-Identifier: MIT

package lsp

// NewTypeScript wraps typescript-language-server (Microsoft's official Node
// LSP wrapper around tsserver). Single binary, stdio mode.
func NewTypeScript(root string) *Stdio {
	return NewStdio("typescript-language-server", []string{"--stdio"}, "typescript", "typescript", root)
}
