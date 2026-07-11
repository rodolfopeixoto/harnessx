// SPDX-License-Identifier: MIT

package specflow

import (
	"strings"

	"github.com/ropeixoto/harnessx/internal/agents"
)

// helpers.go carries the pure text-manipulation helpers used by the
// planning + persistence flows. Kept sibling to specflow.go so tests
// keep the same package surface.

func renderBaseline(s *Session) string {
	var b strings.Builder
	b.WriteString("## Summary\n")
	b.WriteString(s.Prompt + "\n\n")
	b.WriteString("## Users\n- " + answerOr(s, "users", "TBD") + "\n\n")
	b.WriteString("## Acceptance Criteria\n- " + answerOr(s, "acceptance", "TBD") + "\n\n")
	b.WriteString("## Out of Scope\n- " + answerOr(s, "scope_out", "TBD") + "\n\n")
	b.WriteString("## Risks\n- " + answerOr(s, "risks", "TBD") + "\n\n")
	b.WriteString("## Test Plan\n- " + answerOr(s, "tests", "TBD") + "\n\n")
	b.WriteString("## Implementation Notes\n- TBD\n")
	return b.String()
}

func answerOr(s *Session, key, fallback string) string {
	for _, a := range s.Answers {
		if a.Key == key && strings.TrimSpace(a.Value) != "" {
			return a.Value
		}
	}
	return fallback
}

func formatAnswers(answers []Answer) string {
	if len(answers) == 0 {
		return "(none)"
	}
	var b strings.Builder
	for _, a := range answers {
		if strings.TrimSpace(a.Value) == "" {
			continue
		}
		b.WriteString("- " + a.Key + ": " + a.Value + "\n")
	}
	if b.Len() == 0 {
		return "(none)"
	}
	return b.String()
}

func pickBody(o agents.AgentOutput) string {
	if strings.TrimSpace(o.FinalMessage) != "" {
		return strings.TrimSpace(o.FinalMessage)
	}
	return strings.TrimSpace(string(o.Stdout))
}

func slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ', r == '-', r == '_':
			if b.Len() > 0 && b.String()[b.Len()-1] != '_' {
				b.WriteRune('_')
			}
		}
		if b.Len() >= 24 {
			break
		}
	}
	return strings.Trim(b.String(), "_")
}

func unifiedLines(a, b string) string {
	al := strings.Split(a, "\n")
	bl := strings.Split(b, "\n")
	set := func(lines []string) map[string]int {
		m := map[string]int{}
		for _, l := range lines {
			m[l]++
		}
		return m
	}
	aSet, bSet := set(al), set(bl)
	var out strings.Builder
	for _, l := range al {
		if bSet[l] == 0 {
			out.WriteString("- " + l + "\n")
		}
	}
	for _, l := range bl {
		if aSet[l] == 0 {
			out.WriteString("+ " + l + "\n")
		}
	}
	return out.String()
}
