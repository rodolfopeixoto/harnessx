package installcmd

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/ropeixoto/harnessx/internal/install"
)

type ToolPack struct {
	Stack string
	Tools []string
}

var defaultPacks = map[string][]string{
	"go":      {"gopls", "golangci-lint", "govulncheck"},
	"python":  {"basedpyright"},
	"node":    {"tsserver"},
	"react":   {"tsserver"},
	"ruby":    {"ruby-lsp", "solargraph"},
	"rust":    {"rust-analyzer"},
	"java":    {},
	"kotlin":  {"ktlint", "detekt"},
	"swift":   {"swiftlint", "swift-format"},
	"elixir":  {"credo"},
	"php":     {"phpstan", "psalm", "php-cs-fixer"},
	"laravel": {"phpstan", "psalm", "php-cs-fixer"},
	"symfony": {"phpstan", "psalm", "php-cs-fixer"},
	"dotnet":  {"dotnet"},
	"dart":    {"dart"},
	"flutter": {"dart"},
}

func ToolsForStack(stack string) []string {
	return defaultPacks[stack]
}

type InstallOptions struct {
	Stack  string
	DryRun bool
	Probe  func(binary string) bool
}

func Install(ctx context.Context, out io.Writer, opts InstallOptions) error {
	tools := ToolsForStack(opts.Stack)
	if len(tools) == 0 {
		return fmt.Errorf("install tools: no tool pack defined for stack %q", opts.Stack)
	}
	probe := opts.Probe
	if probe == nil {
		probe = func(b string) bool {
			_, err := exec.LookPath(b)
			return err == nil
		}
	}
	fmt.Fprintf(out, "stack %q: %d tools — %v\n", opts.Stack, len(tools), tools)
	reg := install.NewRegistry()
	for _, name := range tools {
		m, err := install.LoadBundled(name)
		if err != nil {
			fmt.Fprintf(out, "  [skip] %s: %v\n", name, err)
			continue
		}
		if probe(m.Probe.Binary) {
			fmt.Fprintf(out, "  [ok]   %s already installed\n", name)
			continue
		}
		plan, err := reg.Pick(m)
		if err != nil {
			fmt.Fprintf(out, "  [skip] %s: %v\n", name, err)
			continue
		}
		fmt.Fprintf(out, "  [run]  %s\n", name)
		if err := install.Execute(ctx, plan, opts.DryRun, out, out); err != nil {
			fmt.Fprintf(out, "  [fail] %s: %v\n", name, err)
		}
	}
	return nil
}
