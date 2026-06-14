// SPDX-License-Identifier: MIT

package lsp

// NewRustAnalyzer wraps rust-analyzer; defaults to stdio when no args are
// provided.
func NewRustAnalyzer(root string) *Stdio {
	return NewStdio("rust-analyzer", nil, "rust", "rust", root)
}
