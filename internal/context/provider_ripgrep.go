// SPDX-License-Identifier: MIT

package context

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strings"
)

// RipgrepProvider runs `rg --json -m1 <keyword>` for each keyword extracted
// from the task description, capping at MaxHits files per keyword to keep
// the pack minimal.
type RipgrepProvider struct {
	MaxKeywords int
	MaxHits     int
}

func (r RipgrepProvider) Name() string { return "ripgrep" }

func (r RipgrepProvider) Apply(ctx context.Context, root string, pack *Pack) error {
	if !hasBinary("rg") {
		pack.Stats.ProvidersSkipped++
		return nil
	}
	maxK := r.MaxKeywords
	if maxK <= 0 {
		maxK = 6
	}
	maxH := r.MaxHits
	if maxH <= 0 {
		maxH = 8
	}

	keywords := extractKeywords(pack.Task, maxK)
	for _, kw := range keywords {
		hits := rgSearch(ctx, root, kw, maxH)
		for _, p := range hits {
			pack.RelevantFiles = appendFile(pack.RelevantFiles, FileEntry{
				Path: p, Reason: "ripgrep:" + kw,
			})
		}
	}
	pack.Stats.ProvidersRan++
	return nil
}

func rgSearch(ctx context.Context, root, kw string, limit int) []string {
	cmd := exec.CommandContext(ctx, "rg",
		"--json", "--ignore-case", "--max-count", "1",
		"--max-filesize", "1M", "--type-not", "lock",
		"--", kw, ".")
	cmd.Dir = root
	var out bytes.Buffer
	cmd.Stdout = &out
	// rg returns 1 when no matches; ignore.
	_ = cmd.Run()

	var paths []string
	scanner := bufio.NewScanner(&out)
	scanner.Buffer(make([]byte, 1024*1024), 4*1024*1024)
	for scanner.Scan() && len(paths) < limit {
		var ev struct {
			Type string `json:"type"`
			Data struct {
				Path struct {
					Text string `json:"text"`
				} `json:"path"`
			} `json:"data"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			continue
		}
		if ev.Type != "match" {
			continue
		}
		if ev.Data.Path.Text != "" {
			paths = append(paths, ev.Data.Path.Text)
		}
	}
	return paths
}

// extractKeywords picks bare-word identifiers from the task prompt. Common
// stop words and 1-char tokens are dropped. Deterministic alphabetical order
// keeps the resulting pack hash stable across re-runs of the same prompt.
func extractKeywords(task string, max int) []string {
	stop := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "of": true,
		"to": true, "in": true, "on": true, "for": true, "with": true,
		"is": true, "are": true, "was": true, "be": true, "by": true, "as": true,
		"this": true, "that": true, "from": true, "into": true, "at": true,
		"create": true, "make": true, "fix": true, "add": true, "update": true,
		"please": true, "using": true,
	}
	seen := map[string]bool{}
	var words []string
	for _, raw := range strings.FieldsFunc(task, isSep) {
		w := strings.ToLower(strings.Trim(raw, "_-."))
		if len(w) < 3 || stop[w] || seen[w] {
			continue
		}
		seen[w] = true
		words = append(words, w)
	}
	// Stable, length-capped output.
	if len(words) > max {
		words = words[:max]
	}
	return words
}

func isSep(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z',
		r >= 'A' && r <= 'Z',
		r >= '0' && r <= '9',
		r == '_', r == '-', r == '.':
		return false
	}
	return true
}
