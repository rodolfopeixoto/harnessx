// SPDX-License-Identifier: MIT

package index

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// BuildCommands reads stack-specific manifests and returns a Commands map.
// Anti-pattern guard: only emit a hint when there is concrete evidence
// (a script in package.json, a target in Makefile, a known convention).
func BuildCommands(root string, stacks []Stack) Commands {
	c := Commands{}
	hasStack := func(name string) bool {
		for _, s := range stacks {
			if s.Name == name {
				return true
			}
		}
		return false
	}

	// Node: parse package.json scripts.
	if hasStack("react") || hasStack("nextjs") || hasStack("vite") || hasStack("node") {
		if scripts := readNodeScripts(filepath.Join(root, "package.json")); scripts != nil {
			for _, name := range sortedKeys(scripts) {
				cmd := "npm run " + name
				hint := CommandHint{
					Stack: "node", Command: cmd,
					Source: "package.json#scripts." + name, Confidence: ConfidenceHigh,
				}
				switch {
				case name == "build":
					c.Build = append(c.Build, hint)
				case name == "test", strings.HasPrefix(name, "test:"):
					c.Test = append(c.Test, hint)
				case name == "lint", strings.HasPrefix(name, "lint:"):
					c.Lint = append(c.Lint, hint)
				case name == "typecheck", name == "tsc":
					c.Typecheck = append(c.Typecheck, hint)
				case name == "format", name == "prettier":
					c.Format = append(c.Format, hint)
				case name == "dev", name == "start":
					c.Run = append(c.Run, hint)
				}
			}
		}
	}

	if hasStack("go") {
		add := func(category *[]CommandHint, cmd string) {
			*category = append(*category, CommandHint{
				Stack: "go", Command: cmd, Source: "convention", Confidence: ConfidenceHigh,
			})
		}
		add(&c.Build, "go build ./...")
		add(&c.Test, "go test ./...")
		add(&c.Lint, "go vet ./...")
		add(&c.Format, "gofmt -l -w .")
	}

	if hasStack("rails") {
		add := func(category *[]CommandHint, cmd, src string, conf Confidence) {
			*category = append(*category, CommandHint{
				Stack: "rails", Command: cmd, Source: src, Confidence: conf,
			})
		}
		add(&c.Test, "bundle exec rspec", "convention", ConfidenceMedium)
		add(&c.Lint, "bundle exec rubocop", "convention", ConfidenceMedium)
		add(&c.Run, "bin/rails server", "convention", ConfidenceHigh)
	}

	if hasStack("python") {
		add := func(category *[]CommandHint, cmd, src string, conf Confidence) {
			*category = append(*category, CommandHint{
				Stack: "python", Command: cmd, Source: src, Confidence: conf,
			})
		}
		add(&c.Test, "pytest", "convention", ConfidenceMedium)
		add(&c.Lint, "ruff check .", "convention", ConfidenceMedium)
		add(&c.Format, "ruff format .", "convention", ConfidenceMedium)
	}

	if hasStack("rust") {
		add := func(category *[]CommandHint, cmd string) {
			*category = append(*category, CommandHint{
				Stack: "rust", Command: cmd, Source: "convention", Confidence: ConfidenceHigh,
			})
		}
		add(&c.Build, "cargo build")
		add(&c.Test, "cargo test --all")
		add(&c.Lint, "cargo clippy --all-targets")
		add(&c.Format, "cargo fmt --check")
	}

	// Makefile targets — augment, do not replace stack defaults.
	if targets := readMakefileTargets(filepath.Join(root, "Makefile")); len(targets) > 0 {
		for _, t := range targets {
			cmd := "make " + t
			hint := CommandHint{
				Stack: "make", Command: cmd, Source: "Makefile:" + t, Confidence: ConfidenceHigh,
			}
			switch t {
			case "build":
				c.Build = append(c.Build, hint)
			case "test":
				c.Test = append(c.Test, hint)
			case "lint":
				c.Lint = append(c.Lint, hint)
			case "check":
				c.Test = append(c.Test, hint)
			case "fmt", "format":
				c.Format = append(c.Format, hint)
			}
		}
	}

	return c
}

func readNodeScripts(path string) map[string]string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(b, &pkg); err != nil {
		return nil
	}
	return pkg.Scripts
}

func readMakefileTargets(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		// Match "target:" or "target: deps" at column 0, ignore special targets.
		if len(line) == 0 || line[0] == '\t' || line[0] == '#' || line[0] == '.' {
			continue
		}
		i := strings.IndexByte(line, ':')
		if i <= 0 {
			continue
		}
		target := strings.TrimSpace(line[:i])
		if target == "" || strings.ContainsAny(target, " /\\$") {
			continue
		}
		if seen[target] {
			continue
		}
		seen[target] = true
		out = append(out, target)
	}
	sort.Strings(out)
	return out
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
