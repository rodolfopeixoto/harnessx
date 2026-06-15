// SPDX-License-Identifier: MIT

package hookscan

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Hook struct {
	Name       string `json:"name"`
	Event      string `json:"event"`
	ConfigPath string `json:"config_path"`
	Scope      string `json:"scope"`
	Source     string `json:"source"`
	Blocking   bool   `json:"blocking"`
	Status     string `json:"status"`
	Risk       string `json:"risk"`
}

const (
	StatusEnabled  = "enabled"
	StatusDisabled = "disabled"
	ScopeGlobal    = "global"
	ScopeProject   = "project"
	SourceHarness  = "harness"
	SourceGit      = "git"
	SourceClaude   = "claude"
	SourceCodex    = "codex"
	SourceGemini   = "gemini"
	SourceKimi     = "kimi"
	RiskLow        = "low"
	RiskMedium     = "medium"
	RiskHigh       = "high"
)

var defaultGlobs = []string{
	".harness/hooks/**",
	"scripts/git-hooks/*",
	".claude/hooks/**",
	".codex/hooks/**",
	".gemini/hooks/**",
	".kimi/hooks/**",
}

func Scan(root string) ([]Hook, error) {
	if root == "" {
		return nil, errors.New("hookscan: empty root")
	}
	seen := map[string]struct{}{}
	var out []Hook
	for _, g := range defaultGlobs {
		matches, _ := filepath.Glob(filepath.Join(root, g))
		for _, m := range matches {
			info, err := os.Stat(m)
			if err != nil || info.IsDir() {
				continue
			}
			if _, dup := seen[m]; dup {
				continue
			}
			seen[m] = struct{}{}
			out = append(out, classify(root, m, info))
		}
	}
	if err := walkForHookNamed(root, seen, &out); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return out, err
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func walkForHookNamed(root string, seen map[string]struct{}, out *[]Hook) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" || base == "dist" || base == "tmp" {
				return filepath.SkipDir
			}
			return nil
		}
		base := filepath.Base(path)
		if !strings.Contains(strings.ToLower(base), "hook") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(base))
		if ext != ".json" && ext != ".yml" && ext != ".yaml" && ext != ".sh" {
			return nil
		}
		if _, dup := seen[path]; dup {
			return nil
		}
		seen[path] = struct{}{}
		info, _ := d.Info()
		*out = append(*out, classify(root, path, info))
		return nil
	})
}

func classify(root, path string, info os.FileInfo) Hook {
	rel, _ := filepath.Rel(root, path)
	base := filepath.Base(path)
	source := sourceOf(rel)
	event := eventOf(base)
	scope := ScopeProject
	if source == SourceHarness && strings.HasPrefix(rel, ".harness/hooks/global/") {
		scope = ScopeGlobal
	}
	blocking := strings.HasPrefix(event, "pre-") || event == "commit-msg"
	risk := RiskLow
	if event == "pre-push" || event == "pre-commit" || event == "commit-msg" {
		risk = RiskMedium
	}
	status := StatusEnabled
	if info != nil && info.Mode()&0o111 == 0 && strings.HasSuffix(base, ".sh") {
		status = StatusDisabled
	}
	return Hook{
		Name:       strings.TrimSuffix(base, filepath.Ext(base)),
		Event:      event,
		ConfigPath: path,
		Scope:      scope,
		Source:     source,
		Blocking:   blocking,
		Status:     status,
		Risk:       risk,
	}
}

func eventOf(base string) string {
	trimmed := strings.TrimSuffix(base, filepath.Ext(base))
	switch trimmed {
	case "pre-commit", "pre-push", "commit-msg", "post-commit", "post-merge", "post-checkout":
		return trimmed
	}
	if strings.HasPrefix(trimmed, "pre-") || strings.HasPrefix(trimmed, "post-") {
		return trimmed
	}
	return "custom"
}

func sourceOf(rel string) string {
	switch {
	case strings.HasPrefix(rel, "scripts/git-hooks/"):
		return SourceGit
	case strings.HasPrefix(rel, ".harness/"):
		return SourceHarness
	case strings.HasPrefix(rel, ".claude/"):
		return SourceClaude
	case strings.HasPrefix(rel, ".codex/"):
		return SourceCodex
	case strings.HasPrefix(rel, ".gemini/"):
		return SourceGemini
	case strings.HasPrefix(rel, ".kimi/"):
		return SourceKimi
	}
	return SourceHarness
}
