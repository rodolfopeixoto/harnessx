// SPDX-License-Identifier: MIT

// Package doctor probes the local toolchain and known agent CLIs and
// returns a structured report. Presentation lives in internal/ui.
package doctor

import (
	"context"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/execprobe"
)

type ProbeSpec struct {
	Binary       string
	Label        string
	VersionArgs  []string
	VersionRegex string
	Required     bool
	Category     string
	InstallID    string
}

// DefaultProbes mirror docs/cli-reference.md.
func DefaultProbes() []ProbeSpec {
	return []ProbeSpec{
		// Toolchain.
		{Binary: "git", Label: "Git", VersionArgs: []string{"--version"}, Required: true, Category: "tool"},
		{Binary: "docker", Label: "Docker", VersionArgs: []string{"--version"}, Required: false, Category: "tool"},
		{Binary: "sqlite3", Label: "SQLite", VersionArgs: []string{"-version"}, Required: false, Category: "tool"},
		{Binary: "rg", Label: "ripgrep", VersionArgs: []string{"--version"}, Required: false, Category: "tool", InstallID: "ripgrep"},
		{Binary: "node", Label: "Node", VersionArgs: []string{"--version"}, Required: false, Category: "tool"},
		{Binary: "npm", Label: "npm", VersionArgs: []string{"--version"}, Required: false, Category: "tool"},
		{Binary: "go", Label: "Go", VersionArgs: []string{"version"}, VersionRegex: `go version go(\d+\.\d+(?:\.\d+)?)`, Required: false, Category: "tool", InstallID: "go"},
		{Binary: "ruby", Label: "Ruby", VersionArgs: []string{"--version"}, Required: false, Category: "tool"},
		{Binary: "python3", Label: "Python", VersionArgs: []string{"--version"}, Required: false, Category: "tool"},
		{Binary: "rustc", Label: "Rust", VersionArgs: []string{"--version"}, Required: false, Category: "tool"},

		// LSP servers (auto-wired by internal/context.AutoLSP when manifests match).
		{Binary: "gopls", Label: "gopls", VersionArgs: []string{"version"}, Required: false, Category: "lsp", InstallID: "gopls"},
		{Binary: "ruby-lsp", Label: "ruby-lsp", VersionArgs: []string{"--version"}, Required: false, Category: "lsp", InstallID: "ruby-lsp"},
		{Binary: "solargraph", Label: "solargraph", VersionArgs: []string{"--version"}, Required: false, Category: "lsp", InstallID: "solargraph"},
		{Binary: "pyright-langserver", Label: "pyright", VersionArgs: []string{"--version"}, Required: false, Category: "lsp", InstallID: "pyright"},
		{Binary: "basedpyright-langserver", Label: "basedpyright", VersionArgs: []string{"--version"}, Required: false, Category: "lsp", InstallID: "basedpyright"},
		{Binary: "rust-analyzer", Label: "rust-analyzer", VersionArgs: []string{"--version"}, Required: false, Category: "lsp", InstallID: "rust-analyzer"},
		{Binary: "typescript-language-server", Label: "tsserver", VersionArgs: []string{"--version"}, Required: false, Category: "lsp", InstallID: "tsserver"},

		// Supply-chain + quality tooling.
		{Binary: "golangci-lint", Label: "golangci-lint", VersionArgs: []string{"version"}, Required: false, Category: "quality", InstallID: "golangci-lint"},
		{Binary: "govulncheck", Label: "govulncheck", VersionArgs: []string{"-version"}, Required: false, Category: "quality", InstallID: "govulncheck"},
		{Binary: "go-licenses", Label: "go-licenses", VersionArgs: []string{"--help"}, Required: false, Category: "quality"},
		{Binary: "gitleaks", Label: "gitleaks", VersionArgs: []string{"version"}, Required: false, Category: "quality", InstallID: "gitleaks"},
		{Binary: "syft", Label: "syft", VersionArgs: []string{"version"}, Required: false, Category: "quality", InstallID: "syft"},

		// Coding-agent CLIs.
		{Binary: "claude", Label: "Claude Code", VersionArgs: []string{"--version"}, Required: false, Category: "agent", InstallID: "claude"},
		{Binary: "codex", Label: "Codex", VersionArgs: []string{"--version"}, Required: false, Category: "agent", InstallID: "codex"},
		{Binary: "gemini", Label: "Gemini", VersionArgs: []string{"--version"}, Required: false, Category: "agent", InstallID: "gemini"},
		{Binary: "kimi", Label: "Kimi", VersionArgs: []string{"--version"}, Required: false, Category: "agent", InstallID: "kimi"},
	}
}

type Entry struct {
	Spec   ProbeSpec
	Result execprobe.Result
}

type ProjectInfo struct {
	Root         string
	HarnessDir   string
	HarnessReady bool // .harness/db/harness.sqlite exists
}

type Report struct {
	OS      string
	Arch    string
	Tools   []Entry
	LSPs    []Entry
	Quality []Entry
	Agents  []Entry
	Project ProjectInfo
}

// AllRequiredPresent reports whether every required tool was found.
func (r Report) AllRequiredPresent() bool {
	for _, e := range r.Tools {
		if e.Spec.Required && !e.Result.Present {
			return false
		}
	}
	return true
}

// Run probes every spec concurrently with a per-probe timeout.
func Run(ctx context.Context, probe *execprobe.Probe, specs []ProbeSpec, project ProjectInfo, perTimeout time.Duration) Report {
	if perTimeout <= 0 {
		perTimeout = 2 * time.Second
	}
	entries := make([]Entry, len(specs))
	var wg sync.WaitGroup
	for i, s := range specs {
		wg.Add(1)
		go func(i int, s ProbeSpec) {
			defer wg.Done()
			entries[i] = Entry{Spec: s, Result: probe.RunSpec(ctx, execprobe.Spec{
				Binary: s.Binary, Args: s.VersionArgs, Timeout: perTimeout, VersionRegex: s.VersionRegex,
			})}
		}(i, s)
	}
	wg.Wait()

	r := Report{OS: runtime.GOOS, Arch: runtime.GOARCH, Project: project}
	for _, e := range entries {
		switch e.Spec.Category {
		case "agent":
			r.Agents = append(r.Agents, e)
		case "lsp":
			r.LSPs = append(r.LSPs, e)
		case "quality":
			r.Quality = append(r.Quality, e)
		default:
			r.Tools = append(r.Tools, e)
		}
	}
	return r
}

// DetectProject inspects root for .harness/db/harness.sqlite.
func DetectProject(root string) ProjectInfo {
	hd := root + string(os.PathSeparator) + ".harness"
	dbReady := false
	if _, err := os.Stat(hd + string(os.PathSeparator) + "db" + string(os.PathSeparator) + "harness.sqlite"); err == nil {
		dbReady = true
	}
	return ProjectInfo{Root: root, HarnessDir: hd, HarnessReady: dbReady}
}
