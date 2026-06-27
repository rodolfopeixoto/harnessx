// SPDX-License-Identifier: MIT

// Package spec renders the spec-driven-development template (spec §8).
// A spec is always a markdown file under .harness/artifacts/specs/. Tests
// generate one and assert the section headers exist; downstream phases
// (sensor `spec_gate`, agent contexts) read the same file.
package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type Spec struct {
	ID                 string
	Title              string
	Mode               domain.Mode
	GeneratedAt        time.Time
	Prompt             string
	UserProblem        string
	ExpectedOutcome    string
	Scope              []string
	OutOfScope         []string
	BusinessRules      []string
	UXExpectations     []string
	APIExpectations    []string
	DataModel          []string
	Security           []string
	Performance        []string
	Observability      []string
	TestPlan           []string
	E2EPlan            []string
	AcceptanceCriteria []string
	RollbackPlan       []string
	DefinitionOfDone   []string
	Assumptions        []string
	OpenQuestions      []string
}

// NewFromPrompt seeds a Spec with sensible defaults derived from the prompt
// and mode. Empty fields are kept empty — the user (or a downstream LLM
// step in Phase 7+) fills them in. This keeps the spec honest.
func NewFromPrompt(prompt string, mode domain.Mode) Spec {
	now := time.Now().UTC()
	title := summary(prompt, 60)
	s := Spec{
		ID:          ids.New(),
		Title:       title,
		Mode:        mode,
		GeneratedAt: now,
		Prompt:      prompt,
		AcceptanceCriteria: []string{
			"All sensors pass via `harness ci`.",
			"Changes have at least one passing test that covers the new behaviour.",
			"No new entries in `forbidden_files` / `secrets_scan`.",
		},
		DefinitionOfDone: []string{
			"Spec exists and matches the implemented behaviour.",
			"Plan is acknowledged.",
			"Code review is recorded as approved.",
			"Report under `.harness/artifacts/reports/` references this spec ID.",
		},
		RollbackPlan: []string{
			"Revert the implementation commit(s).",
			"Re-run `harness ci` to confirm green baseline.",
		},
	}
	switch mode {
	case domain.ModeBugfix:
		s.UserProblem = "Reproduce, root-cause, and fix the reported regression."
		s.AcceptanceCriteria = append(s.AcceptanceCriteria,
			"A regression test exists that fails on `main` and passes on the fix branch.")
	case domain.ModeFeature:
		s.UserProblem = "Deliver the requested behaviour as described in the prompt."
	case domain.ModeOptimization:
		s.UserProblem = "Reduce resource use without degrading behaviour, security, observability, or DX."
		s.AcceptanceCriteria = append(s.AcceptanceCriteria,
			"Baseline `perf-snapshot` exists before any change.",
			"`perf-compare` shows non-regressive deltas.")
	}
	return s
}

const tmplBody = "# Spec: {{.Title}}\n\n" +
	"> ⚠️ **Stub template.** This spec was generated deterministically from the\n" +
	"> prompt — no LLM was invoked. Sections marked `_TODO:_` need a human\n" +
	"> (or an agent invoked with `--agent <id>`) to fill in.\n\n" +
	`- **ID:** {{.ID}}
- **Mode:** {{.Mode}}
- **Generated:** {{.GeneratedAt.Format "2006-01-02T15:04:05Z07:00"}}

## Prompt
{{.Prompt}}

## Feature Name
{{.Title}}

## User Problem
{{or .UserProblem "_TODO: state the user's underlying problem._"}}

## Expected Outcome
{{or .ExpectedOutcome "_TODO: describe the observable outcome for the user._"}}

## Scope
{{listOrTODO .Scope "_TODO: list what is in scope._"}}

## Out of Scope
{{listOrTODO .OutOfScope "_TODO: list what is intentionally out of scope._"}}

## Business Rules
{{listOrTODO .BusinessRules "_TODO: enumerate domain rules the change must respect._"}}

## UX Expectations
{{listOrTODO .UXExpectations "_TODO: describe the user experience (loading/empty/error states)._"}}

## API Expectations
{{listOrTODO .APIExpectations "_TODO: describe API endpoints + contracts touched._"}}

## Data Model Expectations
{{listOrTODO .DataModel "_TODO: list table / column changes; flag destructive migrations._"}}

## Security Considerations
{{listOrTODO .Security "_TODO: authn/authz, secrets, audit log, input validation._"}}

## Performance Considerations
{{listOrTODO .Performance "_TODO: budgets, hotspots, expected p95._"}}

## Observability Expectations
{{listOrTODO .Observability "_TODO: logs, metrics, traces, alerting thresholds._"}}

## Test Plan
{{listOrTODO .TestPlan "_TODO: unit + integration tests to add/update._"}}

## E2E Plan
{{listOrTODO .E2EPlan "_TODO: critical user flows to cover end-to-end._"}}

## Acceptance Criteria
{{list .AcceptanceCriteria}}

## Rollback Plan
{{list .RollbackPlan}}

## Definition of Done
{{list .DefinitionOfDone}}

## Assumptions
{{listOrTODO .Assumptions "_TODO: explicit assumptions taken as safe defaults._"}}

## Open Questions
{{listOrTODO .OpenQuestions "_None — proceed via safe defaults._"}}
`

func list(items []string) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	for _, it := range items {
		b.WriteString("- ")
		b.WriteString(it)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func listOrTODO(items []string, todo string) string {
	if len(items) == 0 {
		return todo
	}
	return list(items)
}

var tmpl = template.Must(template.New("spec").Funcs(template.FuncMap{
	"list":       list,
	"listOrTODO": listOrTODO,
}).Parse(tmplBody))

// Write renders the spec to .harness/artifacts/specs/<id>.md and returns
// the absolute path.
func (s Spec) Write(root string) (string, error) {
	dir := filepath.Join(paths.HarnessDir(root), "artifacts", "specs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(dir, s.ID+".md")
	f, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := tmpl.Execute(f, s); err != nil {
		return "", err
	}
	return p, nil
}

// LatestSpecPath returns the most recently modified spec path under
// .harness/artifacts/specs/. Used by `harness report --last` and the
// future `spec_gate` sensor.
func LatestSpecPath(root string) (string, error) {
	dir := filepath.Join(paths.HarnessDir(root), "artifacts", "specs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var newest string
	var newestMod time.Time
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(newestMod) {
			newestMod = info.ModTime()
			newest = filepath.Join(dir, e.Name())
		}
	}
	if newest == "" {
		return "", fmt.Errorf("no specs under %s", dir)
	}
	return newest, nil
}

// summary trims a prompt to a short title, ending on a word boundary
// when possible.
func summary(s string, max int) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	cut := s[:max]
	if i := strings.LastIndexByte(cut, ' '); i > max/2 {
		cut = cut[:i]
	}
	return cut + "…"
}
