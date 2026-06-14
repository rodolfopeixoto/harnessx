// SPDX-License-Identifier: MIT

package context

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/adapters/lsp"
)

// AutoLSP returns the default provider chain plus an LSPProvider when at
// least one supported language server is on PATH and the project shows
// signs of that language. Today we wire gopls when `go.mod` exists.
func AutoLSP(root string) []Provider {
	chain := DefaultProviders()
	clients := autoClients(root)
	if len(clients) > 0 {
		chain = append(chain, LSPProvider{Clients: clients})
	}
	return chain
}

// autoClients spawns every LSP server that (a) has its binary on PATH and
// (b) shows project-level evidence (manifest file). Each spawn is bounded
// by Start's internal 15 s handshake timeout; failures are silent so a
// flaky server never blocks a context build.
func autoClients(root string) []lsp.Client {
	type candidate struct {
		needPATH []string
		needFile []string
		factory  func(string) lsp.Client
	}
	candidates := []candidate{
		{
			needPATH: []string{"gopls"},
			needFile: []string{"go.mod"},
			factory:  func(r string) lsp.Client { return lsp.NewGopls(r) },
		},
		{
			needPATH: []string{"ruby-lsp"},
			needFile: []string{"Gemfile", "Gemfile.lock"},
			factory:  func(r string) lsp.Client { return lsp.NewRubyLSP(r) },
		},
		{
			needPATH: []string{"solargraph"},
			needFile: []string{"Gemfile"},
			factory:  func(r string) lsp.Client { return lsp.NewSolargraph(r) },
		},
		{
			needPATH: []string{"pyright-langserver"},
			needFile: []string{"pyproject.toml", "requirements.txt", "setup.py"},
			factory:  func(r string) lsp.Client { return lsp.NewPyright(r) },
		},
		{
			needPATH: []string{"basedpyright-langserver"},
			needFile: []string{"pyproject.toml", "requirements.txt", "setup.py"},
			factory:  func(r string) lsp.Client { return lsp.NewBasedPyright(r) },
		},
		{
			needPATH: []string{"rust-analyzer"},
			needFile: []string{"Cargo.toml"},
			factory:  func(r string) lsp.Client { return lsp.NewRustAnalyzer(r) },
		},
		{
			needPATH: []string{"typescript-language-server"},
			needFile: []string{"tsconfig.json", "package.json"},
			factory:  func(r string) lsp.Client { return lsp.NewTypeScript(r) },
		},
	}
	// Pick at most one server per language (avoid starting both ruby-lsp
	// and solargraph against the same project).
	taken := map[string]bool{}
	var out []lsp.Client
	for _, c := range candidates {
		if !anyPath(c.needPATH) || !anyFile(root, c.needFile) {
			continue
		}
		client := c.factory(root)
		s, _ := client.(*lsp.Stdio)
		if s != nil {
			if taken[s.Language()] {
				continue
			}
			taken[s.Language()] = true
		}
		if err := startable(client); err != nil {
			continue
		}
		out = append(out, client)
	}
	return out
}

func anyPath(names []string) bool {
	for _, n := range names {
		if _, err := exec.LookPath(n); err == nil {
			return true
		}
	}
	return false
}

func anyFile(root string, names []string) bool {
	for _, n := range names {
		if _, err := os.Stat(filepath.Join(root, n)); err == nil {
			return true
		}
	}
	return false
}

// startable invokes Start on a typed Stdio client and returns the error.
// Kept as a tiny indirection so a future Client interface that exposes
// Start can replace the type assertion without churn elsewhere.
func startable(c lsp.Client) error {
	if s, ok := c.(*lsp.Stdio); ok {
		return s.Start(context.Background())
	}
	return nil
}

// LSPProvider asks every registered language server about symbols and
// diagnostics for the relevant files. Phase 5 ships the abstraction +
// cache layout; real per-language clients (gopls, ruby-lsp, pyright,
// rust-analyzer, typescript-language-server) plug in later by satisfying
// the lsp.Client interface.
type LSPProvider struct {
	Clients []lsp.Client
}

func (LSPProvider) Name() string { return "lsp" }

func (p LSPProvider) Apply(ctx context.Context, root string, pack *Pack) error {
	if len(p.Clients) == 0 {
		pack.Stats.ProvidersSkipped++
		return nil
	}
	for _, c := range p.Clients {
		for _, f := range pack.RelevantFiles {
			pack.Stats.LSPQueries++
			syms, hit, err := c.DocumentSymbols(ctx, root, f.Path)
			if err != nil {
				continue
			}
			if hit {
				pack.Stats.LSPCacheHits++
			}
			for _, s := range syms {
				pack.LSPSymbols = append(pack.LSPSymbols, Symbol{
					Name: s.Name, Path: s.Path, Line: s.Line,
				})
			}
			diags, hit, err := c.Diagnostics(ctx, root, f.Path)
			pack.Stats.LSPQueries++
			if err != nil {
				continue
			}
			if hit {
				pack.Stats.LSPCacheHits++
			}
			for _, d := range diags {
				pack.LSPDiagnostics = append(pack.LSPDiagnostics, Diagnostic{
					Path: d.Path, Line: d.Line, Severity: d.Severity, Message: d.Message,
				})
			}
		}
	}
	pack.Stats.ProvidersRan++
	return nil
}
