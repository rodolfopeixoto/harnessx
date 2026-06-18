// SPDX-License-Identifier: MIT

// Package projectcfg manages .harness/config/project.yaml — the per-project
// command catalogue (test/lint/run/bench/profile/...) used by the
// `harness test`, `harness lint`, `harness dev`, ... wrappers so users
// never need to remember stack-specific invocations.
package projectcfg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	relPath = ".harness/config/project.yaml"
)

type Project struct {
	Stack    string            `yaml:"stack"`
	Commands map[string]string `yaml:"commands"`
}

func path(root string) string { return filepath.Join(root, relPath) }

func Load(root string) (Project, error) {
	raw, err := os.ReadFile(path(root))
	if err != nil {
		return Project{}, err
	}
	var p Project
	if err := yaml.Unmarshal(raw, &p); err != nil {
		return Project{}, fmt.Errorf("projectcfg: parse: %w", err)
	}
	return p, nil
}

func Save(root string, p Project) error {
	if err := os.MkdirAll(filepath.Dir(path(root)), 0o755); err != nil {
		return err
	}
	out, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path(root), out, 0o644)
}

func Resolve(root, command string) (string, error) {
	if p, err := Load(root); err == nil {
		if cmd, ok := p.Commands[command]; ok && strings.TrimSpace(cmd) != "" {
			return cmd, nil
		}
	}
	stack, err := Detect(root)
	if err != nil {
		return "", err
	}
	if cmd, ok := defaults(stack)[command]; ok {
		return cmd, nil
	}
	return "", fmt.Errorf("projectcfg: no command %q for stack %q", command, stack)
}

func Detect(root string) (string, error) {
	if _, err := os.Stat(filepath.Join(root, "config", "application.rb")); err == nil {
		return "rails", nil
	}
	for _, c := range []struct {
		probe, stack string
	}{
		{"go.mod", "go"},
		{"Cargo.toml", "rust"},
		{"pyproject.toml", "python"},
		{"requirements.txt", "python"},
		{"Gemfile", "ruby"},
		{"package.json", "node"},
	} {
		if _, err := os.Stat(filepath.Join(root, c.probe)); err == nil {
			return c.stack, nil
		}
	}
	return "", errors.New("projectcfg: cannot detect stack (no recognised manifest)")
}

func defaults(stack string) map[string]string {
	switch stack {
	case "python":
		return map[string]string{
			"test":        ".venv/bin/pytest -q",
			"lint":        ".venv/bin/ruff check .",
			"run":         ".venv/bin/uvicorn app:app --reload",
			"dev":         ".venv/bin/uvicorn app:app --reload",
			"bench":       ".venv/bin/pytest --benchmark-only",
			"profile-mem": "python -X tracemalloc=25 -m app",
			"profile-cpu": "python -m cProfile -o profile.out -m app",
		}
	case "python-ecommerce":
		return map[string]string{
			"test":  ".venv/bin/pytest -q",
			"lint":  ".venv/bin/ruff check .",
			"run":   ".venv/bin/uvicorn app.main:app --reload",
			"dev":   ".venv/bin/uvicorn app.main:app --reload",
			"bench": ".venv/bin/pytest --benchmark-only",
		}
	case "go":
		return map[string]string{
			"test":        "go test -race ./...",
			"lint":        "go vet ./...",
			"run":         "go run ./...",
			"dev":         "go run ./...",
			"bench":       "go test -bench=. -benchmem ./...",
			"profile-mem": "go test -memprofile=mem.pprof -bench=. ./...",
			"profile-cpu": "go test -cpuprofile=cpu.pprof -bench=. ./...",
		}
	case "rust":
		return map[string]string{
			"test":        "cargo test",
			"lint":        "cargo clippy",
			"run":         "cargo run",
			"dev":         "cargo run",
			"bench":       "cargo bench",
			"profile-mem": "cargo build --release && valgrind --tool=massif target/release/$NAME",
			"profile-cpu": "cargo build --release && perf record target/release/$NAME",
		}
	case "ruby":
		return map[string]string{
			"test":  "bundle exec rspec",
			"lint":  "bundle exec rubocop",
			"run":   "bundle exec ruby app.rb",
			"dev":   "bundle exec ruby app.rb",
			"bench": "bundle exec ruby benchmark.rb",
		}
	case "rails":
		return map[string]string{
			"test":  "bundle exec rspec",
			"lint":  "bundle exec rubocop",
			"run":   "bundle exec rails server",
			"dev":   "bundle exec rails server",
			"bench": "bundle exec rake bench",
		}
	case "node":
		return map[string]string{
			"test": "npm test",
			"lint": "npm run lint",
			"run":  "npm run dev",
			"dev":  "npm run dev",
		}
	}
	return nil
}

func KnownCommands() []string {
	keys := map[string]struct{}{}
	for _, s := range []string{"python", "go", "rust", "ruby", "rails", "node"} {
		for k := range defaults(s) {
			keys[k] = struct{}{}
		}
	}
	out := make([]string, 0, len(keys))
	for k := range keys {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func FromMeta(stack string, commands map[string]string) Project {
	clean := map[string]string{}
	for k, v := range commands {
		if v = strings.TrimSpace(v); v != "" {
			clean[k] = v
		}
	}
	return Project{Stack: stack, Commands: clean}
}
