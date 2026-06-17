// SPDX-License-Identifier: MIT

package main

import (
	stdstrings "strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	for in, want := range map[string]string{
		"Add /healthz endpoint":    "add-healthz-endpoint",
		"Fix: bug #42 in router!!": "fix-bug-42-in-router",
		"   ":                      "change",
		stdstrings.Repeat("x", 60): stdstrings.Repeat("x", 50),
		"!!!---!!!":                "change",
	} {
		got := slugify(in)
		if got != want {
			t.Errorf("slugify(%q): want %q got %q", in, want, got)
		}
	}
}

func TestConventionalSubjectPrefixes(t *testing.T) {
	for _, c := range []struct{ prefix, prompt, want string }{
		{"feature", "add /healthz", "feat: add /healthz"},
		{"fix", "router 500s", "fix: router 500s"},
		{"chore", "bump dep", "chore: bump dep"},
		{"refactor", "split file", "refactor: split file"},
		{"docs", "tutorial", "docs: tutorial"},
	} {
		got := conventionalSubject(c.prefix, c.prompt)
		if got != c.want {
			t.Errorf("conventionalSubject(%q,%q): want %q got %q", c.prefix, c.prompt, c.want, got)
		}
	}
}

func TestIsRateLimitDetectsCommonSignals(t *testing.T) {
	hits := []string{
		"HTTP 429 too many requests",
		"rate limit exceeded",
		"rate-limit",
		"Rate Limit reached",
		"quota exceeded for project",
		"TOO MANY REQUESTS",
	}
	for _, s := range hits {
		if !isRateLimit(s) {
			t.Errorf("missed rate-limit signal in %q", s)
		}
	}
	misses := []string{"ok", "200 success", "auth denied"}
	for _, s := range misses {
		if isRateLimit(s) {
			t.Errorf("false positive on %q", s)
		}
	}
}

func TestShipCmdHasExpectedFlags(t *testing.T) {
	c := newShipCmd()
	for _, f := range []string{"max-attempts", "rate-limit-wait", "rate-limit-retries", "base", "branch-prefix", "autonomy", "dry-run", "skip-commit"} {
		if c.Flags().Lookup(f) == nil {
			t.Errorf("ship flag missing: %s", f)
		}
	}
}
