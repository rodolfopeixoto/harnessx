// SPDX-License-Identifier: MIT

// Package intent classifies a natural-language prompt into one of the
// HarnessX modes (spec §7). Rules first; an optional LLM fallback can be
// wired in later when no rule matches with sufficient confidence.
package intent

import (
	"strings"

	"github.com/ropeixoto/harnessx/internal/domain"
)

type Classification struct {
	Mode       domain.Mode
	Confidence float64
	Reasons    []string
}

type rule struct {
	mode   domain.Mode
	weight float64
	any    []string
	starts []string
}

// rules are evaluated in declaration order; the highest-scoring mode wins.
// Each match contributes weight; the highest scoring mode is selected.
// Reasons captures every rule that fired so callers can explain the pick.
var rules = []rule{
	{
		mode: domain.ModeQuestion, weight: 1.2,
		starts: []string{"what ", "how ", "where ", "why ", "when ", "who ", "explain ", "describe ", "is ", "are ", "does ", "do ", "can ", "could "},
	},
	{
		mode: domain.ModeBugfix, weight: 1.0,
		any: []string{"fix ", "bug ", "broken", "regression", "crash", "panic", "error", "failing test"},
	},
	{
		mode: domain.ModeDesignToProduct, weight: 1.5,
		any: []string{"claude design", "design zip", "design.zip", "prototype", "design-to-product", "figma", "react parity"},
	},
	{
		mode: domain.ModeOptimization, weight: 1.0,
		any: []string{"optimize", "optimise", "reduce memory", "shrink", "bundle size", "image size", "performance budget", "perf budget", "remove unused"},
	},
	{
		mode: domain.ModeAudit, weight: 0.9,
		any: []string{"audit", "scan ", "lint ", "typecheck", "dependency audit", "security scan", "log audit"},
	},
	{
		mode: domain.ModeReview, weight: 0.9,
		any: []string{"review", "code review", "pr review", "diff review"},
	},
	{
		mode: domain.ModeSetup, weight: 1.0,
		any: []string{"scaffold", "new project", "init project", "bootstrap project", "greenfield"},
	},
	{
		mode: domain.ModeFeature, weight: 1.0,
		any: []string{"add ", "create ", "build ", "implement ", "introduce ", "support ", "expose ", "new feature"},
	},
}

// Classify returns the best-fit mode for prompt. When no rule fires, the
// classification falls back to Feature with low confidence — Feature mode
// requires a spec gate so the user gets explicitly asked downstream.
func Classify(prompt string) Classification {
	lower := strings.ToLower(strings.TrimSpace(prompt)) + " "
	scores := map[domain.Mode]float64{}
	reasons := map[domain.Mode][]string{}
	for _, r := range rules {
		hit := false
		for _, s := range r.starts {
			if strings.HasPrefix(lower, s) {
				scores[r.mode] += r.weight
				reasons[r.mode] = append(reasons[r.mode], "starts-with:"+strings.TrimSpace(s))
				hit = true
				break
			}
		}
		if !hit {
			for _, s := range r.any {
				if strings.Contains(lower, s) {
					scores[r.mode] += r.weight
					reasons[r.mode] = append(reasons[r.mode], "contains:"+strings.TrimSpace(s))
					hit = true
				}
			}
		}
	}

	best := domain.ModeFeature
	bestScore := 0.0
	for m, s := range scores {
		if s > bestScore {
			best = m
			bestScore = s
		}
	}
	if bestScore == 0 {
		return Classification{
			Mode: domain.ModeFeature, Confidence: 0.3,
			Reasons: []string{"no rule matched — defaulting to feature mode (will gate on spec)"},
		}
	}
	conf := bestScore / (bestScore + 1.0) // squash to 0..1
	return Classification{
		Mode: best, Confidence: conf, Reasons: reasons[best],
	}
}
