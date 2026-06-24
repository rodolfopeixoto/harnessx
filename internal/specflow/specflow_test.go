// SPDX-License-Identifier: MIT

package specflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/fake"
)

func freezeTime(t *testing.T) {
	t.Helper()
	orig := now
	now = func() time.Time { return time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = orig })
}

func newFakeReturning(body string) *fake.Adapter {
	a := fake.New("fake")
	a.FinalMessage = body
	return a
}

func TestNewGeneratesID(t *testing.T) {
	s := New(t.TempDir(), "do the thing")
	if s.ID == "" || len(s.ID) < 16 {
		t.Errorf("ID looks wrong: %q", s.ID)
	}
	if s.Prompt != "do the thing" {
		t.Errorf("prompt not stored: %q", s.Prompt)
	}
}

func TestDraftOfflineRendersBaseline(t *testing.T) {
	freezeTime(t)
	s := New(t.TempDir(), "add /healthz")
	s.Answers = []Answer{{Key: "users", Value: "ops"}}
	body, err := s.Draft(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range []string{"## Summary", "add /healthz", "## Users", "ops"} {
		if !strings.Contains(body, w) {
			t.Errorf("missing %q in baseline: %s", w, body)
		}
	}
	if len(s.Revisions) != 1 || s.Revisions[0].Source != "baseline" {
		t.Errorf("revision history wrong: %+v", s.Revisions)
	}
}

func TestDraftWithAdapterRecordsRevision(t *testing.T) {
	freezeTime(t)
	a := newFakeReturning("## Summary\nhello\n")
	s := New(t.TempDir(), "x")
	body, err := s.Draft(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "hello") {
		t.Errorf("draft missing fake body: %s", body)
	}
	if len(s.Revisions) != 1 || s.Revisions[0].Source != "draft" {
		t.Errorf("revision wrong: %+v", s.Revisions)
	}
}

func TestDraftPropagatesAdapterError(t *testing.T) {
	a := fake.New("err")
	a.RunErr = errors.New("boom")
	s := New(t.TempDir(), "x")
	if _, err := s.Draft(context.Background(), a); err == nil {
		t.Error("expected error")
	}
}

func TestDraftEmptyResponseErrors(t *testing.T) {
	a := newFakeReturning("")
	s := New(t.TempDir(), "x")
	if _, err := s.Draft(context.Background(), a); err == nil {
		t.Error("expected empty-body error")
	}
}

func TestRefineRequiresDraft(t *testing.T) {
	s := New(t.TempDir(), "x")
	a := newFakeReturning("ok")
	if _, err := s.Refine(context.Background(), a, "", "do it"); err == nil {
		t.Error("expected no-draft error")
	}
}

func TestRefineReplacesBody(t *testing.T) {
	freezeTime(t)
	a := newFakeReturning("## Summary\nv1\n")
	s := New(t.TempDir(), "x")
	if _, err := s.Draft(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	a.FinalMessage = "## Summary\nv2\n"
	body, err := s.Refine(context.Background(), a, "Summary", "make it v2")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "v2") {
		t.Errorf("refine body wrong: %s", body)
	}
	if len(s.Revisions) != 2 {
		t.Errorf("want 2 revisions, got %d", len(s.Revisions))
	}
	if s.Revisions[1].Source != "refine" || s.Revisions[1].Section != "Summary" {
		t.Errorf("refine revision tagged wrong: %+v", s.Revisions[1])
	}
}

func TestRefineNoAdapter(t *testing.T) {
	s := New(t.TempDir(), "x")
	s.Body = "## Summary\nv1\n"
	if _, err := s.Refine(context.Background(), nil, "", "x"); err == nil {
		t.Error("expected no-adapter error")
	}
}

func TestUndoWalksBack(t *testing.T) {
	freezeTime(t)
	a := newFakeReturning("## Summary\nv1\n")
	s := New(t.TempDir(), "x")
	if _, err := s.Draft(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	a.FinalMessage = "## Summary\nv2\n"
	if _, err := s.Refine(context.Background(), a, "", "v2 please"); err != nil {
		t.Fatal(err)
	}
	body, err := s.Undo()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "v1") {
		t.Errorf("undo body wrong: %s", body)
	}
	if _, err := s.Undo(); !errors.Is(err, ErrNoUndo) {
		t.Errorf("want ErrNoUndo, got %v", err)
	}
}

func TestApplyEditRecordsAndSkipsEmpty(t *testing.T) {
	freezeTime(t)
	s := New(t.TempDir(), "x")
	s.Body = "v1"
	s.Revisions = []Revision{{Time: now(), Source: "draft", Body: "v1"}}
	s.ApplyEdit("  ")
	if len(s.Revisions) != 1 {
		t.Error("empty edit must not append")
	}
	s.ApplyEdit("v1")
	if len(s.Revisions) != 1 {
		t.Error("identical edit must not append")
	}
	s.ApplyEdit("v2")
	if len(s.Revisions) != 2 || s.Revisions[1].Source != "editor" {
		t.Errorf("editor revision wrong: %+v", s.Revisions)
	}
}

func TestSaveWritesSpecAndHistory(t *testing.T) {
	freezeTime(t)
	root := t.TempDir()
	s := New(root, "x")
	s.Body = "## Summary\nhi\n"
	s.Revisions = []Revision{
		{Time: now(), Source: "draft", Body: s.Body},
		{Time: now(), Source: "editor", Body: s.Body},
	}
	p, err := s.Save()
	if err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(p)
	if !strings.Contains(string(body), "harness-spec-id:") {
		t.Errorf("metadata header missing: %s", body)
	}
	hist, err := os.ReadFile(filepath.Join(filepath.Dir(p), s.ID+".history.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSpace(hist), []byte("\n"))
	if len(lines) != 2 {
		t.Errorf("want 2 history lines, got %d (%s)", len(lines), hist)
	}
	var r Revision
	if err := json.Unmarshal(lines[0], &r); err != nil {
		t.Fatalf("history malformed: %v", err)
	}
}

func TestSaveRefusesEmptyDraft(t *testing.T) {
	s := New(t.TempDir(), "x")
	if _, err := s.Save(); err == nil {
		t.Error("expected empty draft error")
	}
}

func TestSectionListExtractsH2Headings(t *testing.T) {
	s := New(t.TempDir(), "x")
	s.Body = "## Summary\nabc\n## Users\ndef\n### sub\n## Risks\nghi\n"
	got := s.SectionList()
	if len(got) != 3 || got[0] != "Summary" || got[2] != "Risks" {
		t.Errorf("section list wrong: %v", got)
	}
}

func TestDiffShowsDelta(t *testing.T) {
	freezeTime(t)
	a := newFakeReturning("## Summary\nv1\n")
	s := New(t.TempDir(), "x")
	if _, err := s.Draft(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	a.FinalMessage = "## Summary\nv2\nextra\n"
	if _, err := s.Refine(context.Background(), a, "", "go"); err != nil {
		t.Fatal(err)
	}
	d := s.Diff()
	if !strings.Contains(d, "- v1") {
		t.Errorf("diff missing removed line: %s", d)
	}
	if !strings.Contains(d, "+ v2") || !strings.Contains(d, "+ extra") {
		t.Errorf("diff missing added lines: %s", d)
	}
}

func TestDiffEmptyForSingleRevision(t *testing.T) {
	s := New(t.TempDir(), "x")
	s.Revisions = []Revision{{Body: "x"}}
	if got := s.Diff(); got != "" {
		t.Errorf("want empty diff, got %q", got)
	}
}

func TestParseQuestionsToleratesPreamble(t *testing.T) {
	raw := "Here are the questions:\n[\n  {\"key\":\"data\",\"prompt\":\"Where does the data live?\"},\n  {\"key\":\"\",\"prompt\":\"How big?\"}\n]\nThanks."
	qs := parseQuestions(raw)
	if len(qs) != 2 {
		t.Fatalf("want 2 questions, got %d", len(qs))
	}
	if qs[1].Key == "" {
		t.Error("second question must get auto-slugified key")
	}
}

func TestParseQuestionsRejectsGarbage(t *testing.T) {
	if got := parseQuestions("not json"); got != nil {
		t.Errorf("garbage should return nil, got %v", got)
	}
}

func TestContextQuestionsParsesAdapter(t *testing.T) {
	a := newFakeReturning(`[{"key":"db","prompt":"Which DB?"}]`)
	got, err := ContextQuestions(context.Background(), a, "x")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Key != "db" {
		t.Errorf("want 1 question 'db', got %v", got)
	}
}

func TestContextQuestionsNilAdapter(t *testing.T) {
	got, err := ContextQuestions(context.Background(), nil, "x")
	if err != nil {
		t.Error("nil adapter must not error")
	}
	if got != nil {
		t.Error("nil adapter must return nil")
	}
}

func TestContextQuestionsForIncludesModeAndTemplate(t *testing.T) {
	a := newFakeReturning(`[{"key":"refresh","prompt":"Refresh token TTL?"}]`)
	got, err := ContextQuestionsFor(context.Background(), a, "JWT for /tasks", "feature", "auth")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Key != "refresh" {
		t.Errorf("want refresh question, got %v", got)
	}
}

func TestExpandShrinkDelegateToRefine(t *testing.T) {
	freezeTime(t)
	a := newFakeReturning("## Summary\nv1\n")
	s := New(t.TempDir(), "x")
	if _, err := s.Draft(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	a.FinalMessage = "## Summary\nlonger\n"
	if _, err := s.Expand(context.Background(), a, "Summary"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Shrink(context.Background(), a, "Summary"); err != nil {
		t.Fatal(err)
	}
	if len(s.Revisions) != 3 {
		t.Errorf("want 3 revisions, got %d", len(s.Revisions))
	}
}

func TestEditViaEditorRoundTrips(t *testing.T) {
	orig := runEditor
	t.Cleanup(func() { runEditor = orig })
	runEditor = func(editor, path string) error {
		return os.WriteFile(path, []byte("## Summary\nedited\n"), 0o644)
	}
	body, err := EditViaEditor("## Summary\nold\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "edited") {
		t.Errorf("edit lost: %s", body)
	}
}

func TestEditViaEditorPropagatesError(t *testing.T) {
	orig := runEditor
	t.Cleanup(func() { runEditor = orig })
	runEditor = func(editor, path string) error { return errors.New("editor crashed") }
	if _, err := EditViaEditor("x"); err == nil {
		t.Error("expected error")
	}
}

func TestSlugifyTrimsAndBoundsLength(t *testing.T) {
	if got := slugify("Hello World!!"); got != "hello_world" {
		t.Errorf("want hello_world, got %q", got)
	}
	long := slugify("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if len(long) > 24 {
		t.Errorf("slug should cap at 24, got %d (%q)", len(long), long)
	}
}

func TestAnswerOrFallsBack(t *testing.T) {
	s := New(t.TempDir(), "x")
	s.Answers = []Answer{{Key: "users", Value: "  "}, {Key: "tests", Value: "vitest"}}
	if got := answerOr(s, "users", "TBD"); got != "TBD" {
		t.Errorf("blank should fall back, got %q", got)
	}
	if got := answerOr(s, "tests", "TBD"); got != "vitest" {
		t.Errorf("want vitest, got %q", got)
	}
}

func TestPickBodyPrefersFinalMessage(t *testing.T) {
	o := agents.AgentOutput{FinalMessage: "  final  ", Stdout: []byte("stdout")}
	if got := pickBody(o); got != "final" {
		t.Errorf("want final, got %q", got)
	}
	o2 := agents.AgentOutput{Stdout: []byte("  stdout  ")}
	if got := pickBody(o2); got != "stdout" {
		t.Errorf("want stdout fallback, got %q", got)
	}
}

func TestFormatAnswersDropsBlanksAndReturnsNoneSentinel(t *testing.T) {
	if got := formatAnswers(nil); got != "(none)" {
		t.Errorf("nil should produce (none), got %q", got)
	}
	if got := formatAnswers([]Answer{{Key: "x", Value: "  "}}); got != "(none)" {
		t.Errorf("all blank should produce (none), got %q", got)
	}
	got := formatAnswers([]Answer{{Key: "users", Value: "ops"}})
	if !strings.Contains(got, "- users: ops") {
		t.Errorf("format wrong: %q", got)
	}
}

func TestMetadataHeaderWritesAnswers(t *testing.T) {
	s := New(t.TempDir(), "multi\nline\nprompt")
	s.Answers = []Answer{{Key: "users", Value: "ops"}}
	h := s.metadataHeader()
	for _, w := range []string{"harness-spec-id:", "prompt: multi line prompt", "answers:", "users: ops"} {
		if !strings.Contains(h, w) {
			t.Errorf("header missing %q\n%s", w, h)
		}
	}
}

func TestRefineEmptyResponseErrors(t *testing.T) {
	a := newFakeReturning("## Summary\nv1\n")
	s := New(t.TempDir(), "x")
	if _, err := s.Draft(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	a.FinalMessage = ""
	if _, err := s.Refine(context.Background(), a, "", "x"); err == nil {
		t.Error("expected empty-response error")
	}
}

func TestContextQuestionsErrorReturnsError(t *testing.T) {
	a := fake.New("bad")
	a.RunErr = errors.New("nope")
	got, err := ContextQuestions(context.Background(), a, "x")
	if err == nil {
		t.Error("error path must surface error")
	}
	if got != nil {
		t.Errorf("error path must return nil questions, got %v", got)
	}
}

func TestRenderPrintsSummary(t *testing.T) {
	var buf bytes.Buffer
	s := New(t.TempDir(), "x")
	s.Body = "## Summary\nhi"
	s.Revisions = []Revision{{Source: "draft", Body: s.Body}}
	s.Render(&buf)
	if !strings.Contains(buf.String(), "revisions=1") {
		t.Errorf("render missing revisions: %s", buf.String())
	}
}
