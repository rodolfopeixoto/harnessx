// SPDX-License-Identifier: MIT

package lsp

// NewRubyLSP returns a Stdio LSP client for the official Ruby LSP server
// (`ruby-lsp` binary, ships with the `ruby-lsp` gem). Falls back to
// solargraph by spawning the alternate binary when needed — caller decides
// which to try via AutoLSP.
func NewRubyLSP(root string) *Stdio {
	return NewStdio("ruby-lsp", nil, "ruby", "ruby", root)
}

// NewSolargraph returns a Stdio client for the solargraph language server.
// Wraps solargraph in stdio mode for projects that haven't adopted ruby-lsp.
func NewSolargraph(root string) *Stdio {
	return NewStdio("solargraph", []string{"stdio"}, "ruby", "ruby", root)
}
