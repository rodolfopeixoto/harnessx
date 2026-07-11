// SPDX-License-Identifier: MIT

// Package specflow drives an interactive spec-authoring loop: ask
// clarifying questions, draft markdown, refine sections through an
// LLM, persist revisions so /undo can walk back. Caller (cmd/harness
// or internal/repl) owns the I/O; this package is pure logic +
// filesystem persistence so it is unit-testable without a TTY.
package specflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
)

// Question is one clarifying question the agent asks before drafting.
// Required questions must be answered; optional ones can be skipped
// with empty string + the LLM fills the gap from the prompt.
type Question struct {
	Key      string `json:"key"`
	Prompt   string `json:"prompt"`
	Required bool   `json:"required,omitempty"`
}

// Answer is a (Key, Value) pair the caller collected from the user.
type Answer struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Revision is one entry in the spec history; appended every time the
// LLM rewrites or the user accepts an inline edit. The Source field
// distinguishes editor saves from refine/expand/shrink so a future
// dashboard can summarise where the spec actually came from.
type Revision struct {
	Time    time.Time `json:"time"`
	Source  string    `json:"source"`
	Section string    `json:"section,omitempty"`
	Body    string    `json:"body"`
}

// Session is the in-memory document the REPL is editing. It is the only
// piece of mutable state the package owns; persistence is explicit via
// Save/AppendRevision.
type Session struct {
	ID        string
	Root      string
	Prompt    string
	Mode      string
	Template  string
	Questions []Question
	Answers   []Answer
	Body      string
	Revisions []Revision
}

// New creates a fresh session with a generated ULID. Root is the
// project directory; the spec will land at
// `.harness/artifacts/specs/<id>.md`.
func New(root, prompt string) *Session {
	return &Session{
		ID:     strings.ToLower(ids.New()),
		Root:   root,
		Prompt: strings.TrimSpace(prompt),
	}
}

// BaselineQuestions are the deterministic questions every spec asks.
// LLM-generated context-specific questions are appended on top.
var BaselineQuestions = []Question{
	{Key: "users", Prompt: "Who uses this feature? (role / persona)", Required: true},
	{Key: "acceptance", Prompt: "What does success look like in observable terms?", Required: true},
	{Key: "scope_out", Prompt: "What is explicitly OUT of scope?"},
	{Key: "risks", Prompt: "Any known risks, edge cases, or constraints?"},
	{Key: "tests", Prompt: "How will we test it (unit / e2e / manual)?"},
}

// ContextQuestions asks the planning adapter for 1-3 prompt-specific
// follow-up questions on top of the baseline. The caller decides what
// to do when the adapter errors — earlier we swallowed the error and
// returned nil, which left users wondering why the contextual question
// never showed up. Now we surface the error so the wizard can warn.
func ContextQuestions(ctx context.Context, adapter agents.AgentAdapter, prompt string) ([]Question, error) {
	return ContextQuestionsFor(ctx, adapter, prompt, "", "")
}

// ContextQuestionsFor enriches the planning-adapter prompt with the
// spec mode + template id so the model can pick category-specific
// questions instead of generic ones. Mode/template empty falls back to
// the original generic behaviour.
func ContextQuestionsFor(ctx context.Context, adapter agents.AgentAdapter, prompt, mode, template string) ([]Question, error) {
	if adapter == nil {
		return nil, nil
	}
	hint := ""
	if mode != "" {
		hint = fmt.Sprintf("Spec mode: %s.\n", mode)
	}
	if template != "" && template != "none" {
		hint += fmt.Sprintf("Recurring pattern: %s. Focus the questions on edge cases for this pattern specifically.\n", template)
	}
	body := fmt.Sprintf(`You are HarnessX's spec planner.
%sRead the feature prompt and emit a JSON array of 1-3 SHORT clarifying
questions a senior engineer would ask before implementing. Do not repeat
the baseline questions (users, acceptance, out-of-scope, risks, tests).
Output JSON only, schema: [{"key":"snake_case","prompt":"...?"}].

Feature prompt:
%s`, hint, prompt)
	req := agents.AgentRequest{
		Prompt:  body,
		Timeout: 60 * time.Second,
		Extra:   map[string]string{"task": "planning"},
	}
	res := adapter.Run(ctx, req)
	if res.Err != nil {
		return nil, fmt.Errorf("specflow: context questions failed: %w", res.Err)
	}
	raw := strings.TrimSpace(res.Output.FinalMessage)
	if raw == "" {
		raw = string(res.Output.Stdout)
	}
	return parseQuestions(raw), nil
}

func parseQuestions(raw string) []Question {
	raw = strings.TrimSpace(raw)
	if i := strings.Index(raw, "["); i >= 0 {
		raw = raw[i:]
	}
	if j := strings.LastIndex(raw, "]"); j >= 0 {
		raw = raw[:j+1]
	}
	var qs []Question
	if err := json.Unmarshal([]byte(raw), &qs); err != nil {
		return nil
	}
	out := make([]Question, 0, len(qs))
	for _, q := range qs {
		if strings.TrimSpace(q.Prompt) == "" {
			continue
		}
		if q.Key == "" {
			q.Key = slugify(q.Prompt)
		}
		out = append(out, q)
	}
	return out
}

// Draft writes the first version of the spec by handing the prompt +
// answers to the planning adapter. The returned markdown is also
// stored on the session as Draft + appended to Revisions.
func (s *Session) Draft(ctx context.Context, adapter agents.AgentAdapter) (string, error) {
	if adapter == nil {
		// deterministic baseline when no LLM is available — still useful
		// so the wizard never hard-fails offline.
		s.Body = renderBaseline(s)
		s.Revisions = append(s.Revisions, Revision{Time: now(), Source: "baseline", Body: s.Body})
		return s.Body, nil
	}
	skeleton := SkeletonFor(s.Template)
	skeletonHint := ""
	if skeleton != "" {
		skeletonHint = fmt.Sprintf("\n\nWhen the recurring pattern fits, also include this template skeleton verbatim as an additional H2 section:\n%s", skeleton)
	}
	modeHint := ""
	if s.Mode != "" {
		modeHint = fmt.Sprintf("\nSpec mode: %s.", s.Mode)
	}
	body := fmt.Sprintf(`You are HarnessX's spec writer.%s
Given a feature prompt and clarifying answers, produce a single
markdown spec with these sections (use these exact H2 headings):

## Summary
## Users
## Acceptance Criteria
## Out of Scope
## Risks
## Test Plan
## Implementation Notes

Be specific. Use bullet lists. No preamble, no trailing chatter.%s

Prompt:
%s

Clarifying answers:
%s`, modeHint, skeletonHint, s.Prompt, formatAnswers(s.Answers))
	req := agents.AgentRequest{
		Prompt:  body,
		Timeout: 90 * time.Second,
		Extra:   map[string]string{"task": "planning"},
	}
	res := adapter.Run(ctx, req)
	if res.Err != nil {
		return "", fmt.Errorf("specflow draft: %w", res.Err)
	}
	out := pickBody(res.Output)
	if strings.TrimSpace(out) == "" {
		return "", fmt.Errorf("specflow draft: empty response from %s", adapter.ID())
	}
	s.Body = out
	s.Revisions = append(s.Revisions, Revision{Time: now(), Source: "draft", Body: out})
	return out, nil
}

// Refine asks the LLM to rewrite one named section in-place. section
// is matched against the H2 headings case-insensitively; empty section
// rewrites the whole spec.
func (s *Session) Refine(ctx context.Context, adapter agents.AgentAdapter, section, instruction string) (string, error) {
	if adapter == nil {
		return "", fmt.Errorf("specflow refine: no adapter wired")
	}
	if strings.TrimSpace(s.Body) == "" {
		return "", fmt.Errorf("specflow refine: no draft yet — call Draft first")
	}
	scope := "the entire spec"
	if section != "" {
		scope = "the section titled '" + section + "'"
	}
	body := fmt.Sprintf(`You are HarnessX's spec editor.
Rewrite %s of the markdown below according to the instruction.
Preserve every other section verbatim. Return the full updated
markdown spec, no preamble or trailing chatter.

Instruction: %s

Current spec:
%s`, scope, instruction, s.Body)
	req := agents.AgentRequest{
		Prompt:  body,
		Timeout: 90 * time.Second,
		Extra:   map[string]string{"task": "planning"},
	}
	res := adapter.Run(ctx, req)
	if res.Err != nil {
		return "", fmt.Errorf("specflow refine: %w", res.Err)
	}
	out := pickBody(res.Output)
	if strings.TrimSpace(out) == "" {
		return "", fmt.Errorf("specflow refine: empty response")
	}
	s.Body = out
	s.Revisions = append(s.Revisions, Revision{Time: now(), Source: "refine", Section: section, Body: out})
	return out, nil
}

// Expand / Shrink are convenience wrappers around Refine with a
// templated instruction so the REPL keeps a tidy command surface.
func (s *Session) Expand(ctx context.Context, adapter agents.AgentAdapter, section string) (string, error) {
	return s.Refine(ctx, adapter, section, "Add concrete detail and worked examples without changing the meaning. Keep bullet format.")
}

func (s *Session) Shrink(ctx context.Context, adapter agents.AgentAdapter, section string) (string, error) {
	return s.Refine(ctx, adapter, section, "Tighten the wording. Drop filler, keep the technical substance, keep bullet format.")
}

// ApplyEdit accepts a draft already mutated by the user (e.g. after
// closing $EDITOR) and records it as a revision tagged 'editor'.
func (s *Session) ApplyEdit(body string) {
	body = strings.TrimSpace(body)
	if body == "" || body == strings.TrimSpace(s.Body) {
		return
	}
	s.Body = body
	s.Revisions = append(s.Revisions, Revision{Time: now(), Source: "editor", Body: body})
}

// Undo walks back one revision, returning the now-current draft.
// Returns ErrNoUndo when only the initial revision remains.
var ErrNoUndo = fmt.Errorf("specflow: no earlier revision to restore")

func (s *Session) Undo() (string, error) {
	if len(s.Revisions) < 2 {
		return s.Body, ErrNoUndo
	}
	s.Revisions = s.Revisions[:len(s.Revisions)-1]
	s.Body = s.Revisions[len(s.Revisions)-1].Body
	return s.Body, nil
}

// Diff returns a unified-ish diff between the last two revisions.
// Caller renders it; the diff is intentionally line-based and tiny so
// no external tool is required.
func (s *Session) Diff() string {
	if len(s.Revisions) < 2 {
		return ""
	}
	prev := s.Revisions[len(s.Revisions)-2].Body
	curr := s.Revisions[len(s.Revisions)-1].Body
	return unifiedLines(prev, curr)
}

// SectionList returns the H2 headings present in the current draft,
// in document order, so the REPL can offer them for /refine.
func (s *Session) SectionList() []string {
	out := []string{}
	for _, line := range strings.Split(s.Body, "\n") {
		if strings.HasPrefix(line, "## ") {
			out = append(out, strings.TrimSpace(strings.TrimPrefix(line, "## ")))
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return false })
	return out
}

// Render prints a tidy summary of the session to w; used by /show.
func (s *Session) Render(w io.Writer) {
	fmt.Fprintf(w, "spec %s (revisions=%d)\n\n", s.ID, len(s.Revisions))
	fmt.Fprintln(w, s.Body)
}

var now = func() time.Time { return time.Now().UTC() }
