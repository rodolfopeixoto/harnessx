// SPDX-License-Identifier: MIT

// Package recall scans past .harness/runs/<id>/report.md files and
// scores them against a query using a simple bag-of-words overlap.
// Cheap, deterministic, no external index. Replaceable later by
// SQLite FTS if scale demands it.
package recall

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

// Hit is one scored report.
type Hit struct {
	RunID   string
	Path    string
	Score   float64
	Snippet string
}

var wordRE = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9_-]+`)

// Recall returns the top N reports ordered by descending score for the
// given query. Score is bag-of-words overlap normalised by query
// length; zero-overlap reports are excluded.
func Recall(startDir, query string, limit int) ([]Hit, error) {
	root, err := paths.FindProjectRoot(startDir)
	if err != nil {
		return nil, err
	}
	runsDir := filepath.Join(paths.HarnessDir(root), "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		return nil, nil
	}
	qTerms := tokenise(query)
	if len(qTerms) == 0 {
		return nil, nil
	}
	var hits []Hit
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(runsDir, e.Name(), "report.md")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		score, snippet := scoreReport(string(raw), qTerms)
		if score == 0 {
			continue
		}
		hits = append(hits, Hit{RunID: e.Name(), Path: path, Score: score, Snippet: snippet})
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
	}
	return hits, nil
}

func tokenise(s string) []string {
	var out []string
	for _, w := range wordRE.FindAllString(strings.ToLower(s), -1) {
		if len(w) < 3 {
			continue
		}
		if stopWords[w] {
			continue
		}
		out = append(out, w)
	}
	return out
}

var stopWords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "from": true,
	"this": true, "that": true, "into": true, "out": true, "not": true,
	"are": true, "was": true, "but": true, "all": true, "you": true,
}

func scoreReport(body string, qTerms []string) (float64, string) {
	bodyTerms := tokenise(body)
	if len(bodyTerms) == 0 {
		return 0, ""
	}
	set := map[string]bool{}
	for _, t := range bodyTerms {
		set[t] = true
	}
	hits := 0
	for _, q := range qTerms {
		if set[q] {
			hits++
		}
	}
	score := float64(hits) / float64(len(qTerms))
	snippet := firstNonEmptyLine(body)
	return score, snippet
}

func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(line) > 120 {
			line = line[:117] + "..."
		}
		return line
	}
	return ""
}
