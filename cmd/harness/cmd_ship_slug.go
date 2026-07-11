// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"regexp"
	"strings"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// slugify normalises a free-form prompt into a git-branch-safe token.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 50 {
		s = s[:50]
		s = strings.Trim(s, "-")
	}
	if s == "" {
		s = "change"
	}
	return s
}

func conventionalSubject(prefix, prompt string) string {
	conv := "feat"
	switch prefix {
	case "fix", "hotfix":
		conv = "fix"
	case "chore":
		conv = "chore"
	case "refactor":
		conv = "refactor"
	case "docs":
		conv = "docs"
	}
	short := strings.TrimSpace(prompt)
	if len(short) > 50-len(conv)-2 {
		short = short[:50-len(conv)-2]
	}
	return fmt.Sprintf("%s: %s", conv, short)
}
