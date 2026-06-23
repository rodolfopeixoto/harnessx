// SPDX-License-Identifier: MIT

// Package specflow drives an interactive spec-authoring loop: ask
// clarifying questions, draft markdown, refine sections through an
// LLM, persist revisions so /undo can walk back. Caller (cmd/harness
// or internal/repl) owns the I/O; this package is pure logic +
// filesystem persistence so it is unit-testable without a TTY.
package specflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
// follow-up questions on top of the baseline. Errors degrade to the
// baseline alone so the wizard never blocks on a flaky network call.
func ContextQuestions(ctx context.Context, adapter agents.AgentAdapter, prompt string) []Question {
	if adapter == nil {
		return nil
	}
	body := fmt.Sprintf(`You are HarnessX's spec planner.
Read the feature prompt and emit a JSON array of 1-3 SHORT clarifying
questions a senior engineer would ask before implementing. Do not repeat
the baseline questions (users, acceptance, out-of-scope, risks, tests).
Output JSON only, schema: [{"key":"snake_case","prompt":"...?"}].

Feature prompt:
%s`, prompt)
	req := agents.AgentRequest{
		Prompt:  body,
		Timeout: 60 * time.Second,
		Extra:   map[string]string{"task": "planning"},
	}
	res := adapter.Run(ctx, req)
	if res.Err != nil {
		return nil
	}
	raw := strings.TrimSpace(res.Output.FinalMessage)
	if raw == "" {
		raw = string(res.Output.Stdout)
	}
	return parseQuestions(raw)
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
	body := fmt.Sprintf(`You are HarnessX's spec writer.
Given a feature prompt and clarifying answers, produce a single
markdown spec with these sections (use these exact H2 headings):

## Summary
## Users
## Acceptance Criteria
## Out of Scope
## Risks
## Test Plan
## Implementation Notes

Be specific. Use bullet lists. No preamble, no trailing chatter.

Prompt:
%s

Clarifying answers:
%s`, s.Prompt, formatAnswers(s.Answers))
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

// Save writes the spec markdown to `.harness/artifacts/specs/<id>.md`
// and the revision history alongside as `<id>.history.jsonl`. Returns
// the spec path.
func (s *Session) Save() (string, error) {
	if strings.TrimSpace(s.Body) == "" {
		return "", fmt.Errorf("specflow save: empty draft")
	}
	dir := filepath.Join(s.Root, ".harness", "artifacts", "specs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	specPath := filepath.Join(dir, s.ID+".md")
	header := s.metadataHeader()
	if err := os.WriteFile(specPath, []byte(header+s.Body+"\n"), 0o644); err != nil {
		return "", err
	}
	histPath := filepath.Join(dir, s.ID+".history.jsonl")
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, r := range s.Revisions {
		_ = enc.Encode(r)
	}
	if err := os.WriteFile(histPath, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	return specPath, nil
}

func (s *Session) metadataHeader() string {
	var b strings.Builder
	b.WriteString("<!--\n")
	b.WriteString("harness-spec-id: " + s.ID + "\n")
	b.WriteString("prompt: " + strings.ReplaceAll(s.Prompt, "\n", " ") + "\n")
	if len(s.Answers) > 0 {
		b.WriteString("answers:\n")
		for _, a := range s.Answers {
			b.WriteString("  " + a.Key + ": " + strings.ReplaceAll(a.Value, "\n", " ") + "\n")
		}
	}
	b.WriteString("-->\n\n")
	return b.String()
}

// EditViaEditor writes the current draft to a temp file, shells out
// to $EDITOR (fallback $VISUAL, then `vi`), and re-reads the file.
// The returned body is the user-edited content; caller invokes
// ApplyEdit to record it on the session.
func EditViaEditor(initial string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}
	f, err := os.CreateTemp("", "harness-spec-*.md")
	if err != nil {
		return "", err
	}
	tmp := f.Name()
	if _, err := f.WriteString(initial); err != nil {
		_ = f.Close()
		return "", err
	}
	_ = f.Close()
	defer os.Remove(tmp)
	if err := runEditor(editor, tmp); err != nil {
		return "", err
	}
	body, err := os.ReadFile(tmp)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// runEditor is split into its own var so tests can stub it without
// spawning a real editor process.
var runEditor = defaultRunEditor

// helpers --------------------------------------------------------

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
