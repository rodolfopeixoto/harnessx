// SPDX-License-Identifier: MIT

package index

import (
	"os"
	"path/filepath"
	"sort"
)

// DetectStacks inspects root (non-recursive for primary markers, plus a
// shallow look at apps/* for monorepos) and returns the stack list.
func DetectStacks(root string) []Stack {
	exists := func(p ...string) bool {
		_, err := os.Stat(filepath.Join(append([]string{root}, p...)...))
		return err == nil
	}
	var stacks []Stack
	add := func(name string, conf Confidence, ev ...string) {
		stacks = append(stacks, Stack{Name: name, Evidence: ev, Confidence: conf})
	}

	// Go
	if exists("go.mod") {
		add("go", ConfidenceHigh, "go.mod")
	}

	// Node ecosystem; refine into react/nextjs by content.
	if exists("package.json") {
		raw, _ := os.ReadFile(filepath.Join(root, "package.json"))
		s := string(raw)
		switch {
		case contains(s, `"next"`):
			add("nextjs", ConfidenceHigh, "package.json:next")
		case contains(s, `"react"`):
			conf := ConfidenceHigh
			if !contains(s, `"react-dom"`) {
				conf = ConfidenceMedium
			}
			add("react", conf, "package.json:react")
		default:
			add("node", ConfidenceMedium, "package.json")
		}
		if exists("vite.config.ts") || exists("vite.config.js") {
			add("vite", ConfidenceHigh, "vite.config.*")
		}
	}

	// Ruby on Rails
	if exists("Gemfile") {
		raw, _ := os.ReadFile(filepath.Join(root, "Gemfile"))
		if contains(string(raw), "rails") || exists("config", "application.rb") {
			add("rails", ConfidenceHigh, "Gemfile:rails")
		} else {
			add("ruby", ConfidenceMedium, "Gemfile")
		}
	}

	// Python
	if exists("pyproject.toml") {
		add("python", ConfidenceHigh, "pyproject.toml")
	} else if exists("requirements.txt") {
		add("python", ConfidenceHigh, "requirements.txt")
	}

	// Rust
	if exists("Cargo.toml") {
		add("rust", ConfidenceHigh, "Cargo.toml")
	}

	// Docker
	if exists("Dockerfile") || exists("docker-compose.yml") || exists("docker-compose.yaml") || exists("compose.yaml") {
		ev := []string{}
		for _, p := range []string{"Dockerfile", "docker-compose.yml", "docker-compose.yaml", "compose.yaml"} {
			if exists(p) {
				ev = append(ev, p)
			}
		}
		add("docker", ConfidenceHigh, ev...)
	}

	sort.Slice(stacks, func(i, j int) bool { return stacks[i].Name < stacks[j].Name })
	return stacks
}

// DetectLanguages returns the de-duplicated set of programming languages
// implied by the detected stacks. Order is alphabetical for stable output.
func DetectLanguages(stacks []Stack) []string {
	set := map[string]struct{}{}
	for _, s := range stacks {
		switch s.Name {
		case "go":
			set["go"] = struct{}{}
		case "nextjs", "react", "vite", "node":
			set["typescript"] = struct{}{}
			set["javascript"] = struct{}{}
		case "rails", "ruby":
			set["ruby"] = struct{}{}
		case "python":
			set["python"] = struct{}{}
		case "rust":
			set["rust"] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// DetectMarkers returns the root-relative marker filenames that explain
// the detection.
func DetectMarkers(root string) []string {
	known := []string{
		".git", ".harness",
		"go.mod", "package.json", "pnpm-lock.yaml", "yarn.lock", "package-lock.json",
		"Gemfile", "Gemfile.lock", "config/application.rb",
		"pyproject.toml", "requirements.txt",
		"Cargo.toml", "Cargo.lock",
		"Dockerfile", "docker-compose.yml", "docker-compose.yaml", "compose.yaml",
		"Makefile", "AGENTS.md", "CLAUDE.md", "README.md",
	}
	var out []string
	for _, m := range known {
		if _, err := os.Stat(filepath.Join(root, m)); err == nil {
			out = append(out, m)
		}
	}
	sort.Strings(out)
	return out
}

func contains(haystack, needle string) bool {
	// tiny indexOf to avoid pulling strings.Contains symbol into hot path tests
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
