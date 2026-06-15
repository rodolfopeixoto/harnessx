// SPDX-License-Identifier: MIT

// Package promptenh deterministically augments a raw user prompt with
// project conventions, skill snippets, and a compact context summary.
// No LLM, no network — pure local enrichment so token spend stays
// predictable and the enhancement is reproducible from the same inputs.
package promptenh

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	hxcontext "github.com/ropeixoto/harnessx/internal/context"
	"github.com/ropeixoto/harnessx/internal/domain"
)

// SkillSource is the minimum subset of internal/skills needed to
// prefix the prompt without importing the whole skills package
// (which would create an import cycle with workflow).
type SkillSource interface {
	List() ([]SkillSnippet, error)
}

type SkillSnippet struct {
	ID      string
	Mode    string
	Body    string
	Version string
	Score   float64
}

type Enhancement struct {
	Original       string   `json:"original"`
	Enhanced       string   `json:"enhanced"`
	Mode           string   `json:"mode"`
	SkillPrefixes  []string `json:"skill_prefixes,omitempty"`
	ContextSummary string   `json:"context_summary,omitempty"`
	TokensAdded    int      `json:"tokens_added"`
}

// Enhance returns the augmented prompt and a structured artifact. The
// caller persists artifact via Write. The order of additions is fixed:
//
//  1. Skill prefixes for the chosen mode (alphabetical by skill ID).
//  2. Context summary (up to 5 bullet lines from pack.RelevantFiles).
//  3. The original prompt.
func Enhance(prompt string, mode domain.Mode, pack *hxcontext.Pack, skills SkillSource) Enhancement {
	out := Enhancement{Original: prompt, Mode: string(mode)}
	var parts []string

	if skills != nil {
		if list, err := skills.List(); err == nil {
			matched := pickSkillsForMode(list, mode)
			for _, s := range matched {
				out.SkillPrefixes = append(out.SkillPrefixes, s.ID)
				parts = append(parts, "## Skill: "+s.ID+"\n"+strings.TrimSpace(s.Body))
			}
		}
	}

	if pack != nil && len(pack.RelevantFiles) > 0 {
		summary := contextSummary(pack)
		out.ContextSummary = summary
		parts = append(parts, "## Project context\n"+summary)
	}

	parts = append(parts, "## Task\n"+strings.TrimSpace(prompt))
	out.Enhanced = strings.Join(parts, "\n\n")
	out.TokensAdded = estimateTokens(out.Enhanced) - estimateTokens(prompt)
	return out
}

func pickSkillsForMode(list []SkillSnippet, mode domain.Mode) []SkillSnippet {
	var matches []SkillSnippet
	want := string(mode)
	for _, s := range list {
		if s.Mode == "" || s.Mode == "*" || s.Mode == want {
			matches = append(matches, s)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return matches[i].ID < matches[j].ID
	})
	if len(matches) > 5 {
		matches = matches[:5]
	}
	return matches
}

func contextSummary(pack *hxcontext.Pack) string {
	var b strings.Builder
	limit := 5
	if len(pack.RelevantFiles) < limit {
		limit = len(pack.RelevantFiles)
	}
	for i := 0; i < limit; i++ {
		f := pack.RelevantFiles[i]
		fmt.Fprintf(&b, "- %s (%s)\n", f.Path, f.Reason)
	}
	if len(pack.RelevantFiles) > limit {
		fmt.Fprintf(&b, "- ... %d more files\n", len(pack.RelevantFiles)-limit)
	}
	return strings.TrimRight(b.String(), "\n")
}

func estimateTokens(s string) int {
	if s == "" {
		return 0
	}
	return len(s) / 4
}

// Write serialises the enhancement as enhancement.json under runDir.
func Write(runDir string, e Enhancement) (string, error) {
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(runDir, "enhancement.json")
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}
