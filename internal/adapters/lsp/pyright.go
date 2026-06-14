// SPDX-License-Identifier: MIT

package lsp

// NewPyright wraps the pyright language server in stdio mode. pyright ships
// `pyright-langserver` as the LSP-mode entrypoint; the bare `pyright`
// binary runs as a one-shot CLI.
func NewPyright(root string) *Stdio {
	return NewStdio("pyright-langserver", []string{"--stdio"}, "python", "python", root)
}

// NewBasedPyright is the community fork. Same protocol; different binary.
func NewBasedPyright(root string) *Stdio {
	return NewStdio("basedpyright-langserver", []string{"--stdio"}, "python", "python", root)
}
