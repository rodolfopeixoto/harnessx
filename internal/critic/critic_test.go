// SPDX-License-Identifier: MIT

package critic

import (
	"context"
	"testing"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type stubAdapter struct {
	id        string
	strengths []string
	output    string
	err       error
}

func (s stubAdapter) ID() string   { return s.id }
func (s stubAdapter) Name() string { return s.id }
func (s stubAdapter) Capabilities() agents.Capabilities {
	return agents.Capabilities{Strengths: s.strengths}
}
func (s stubAdapter) Healthcheck(context.Context) agents.HealthcheckResult {
	return agents.HealthcheckResult{OK: true}
}
func (s stubAdapter) Run(_ context.Context, _ agents.AgentRequest) agents.AgentResult {
	return agents.AgentResult{Output: agents.AgentOutput{Stdout: []byte(s.output)}, Err: s.err}
}
func (s stubAdapter) ParseUsage(agents.AgentOutput) agents.Usage { return agents.Usage{} }
func (s stubAdapter) ClassifyFailure(agents.AgentOutput, int, error) agents.FailureType {
	return agents.FailureNone
}

func newReg(t *testing.T, adapters ...stubAdapter) *agents.Registry {
	t.Helper()
	r := agents.NewRegistry()
	for _, a := range adapters {
		if err := r.Register(a); err != nil {
			t.Fatal(err)
		}
	}
	return r
}

func TestPickCriticPrefersReviewCritic(t *testing.T) {
	reg := newReg(t,
		stubAdapter{id: "a", strengths: []string{"code"}},
		stubAdapter{id: "b", strengths: []string{"review", "critic"}},
		stubAdapter{id: "c", strengths: []string{"review"}},
	)
	id, ok := PickCritic(reg)
	if !ok || id != "b" {
		t.Errorf("want b, got %q ok=%v", id, ok)
	}
}

func TestPickCriticFallsBackToReview(t *testing.T) {
	reg := newReg(t,
		stubAdapter{id: "a", strengths: []string{"code"}},
		stubAdapter{id: "c", strengths: []string{"review"}},
	)
	id, ok := PickCritic(reg)
	if !ok || id != "c" {
		t.Errorf("want c, got %q ok=%v", id, ok)
	}
}

func TestPickCriticAnyFallback(t *testing.T) {
	reg := newReg(t, stubAdapter{id: "only", strengths: []string{"image"}})
	id, ok := PickCritic(reg)
	if !ok || id != "only" {
		t.Errorf("want only, got %q ok=%v", id, ok)
	}
}

func TestPickCriticEmptyRegistry(t *testing.T) {
	if _, ok := PickCritic(agents.NewRegistry()); ok {
		t.Error("empty registry should yield ok=false")
	}
}

func TestCritiqueReturnsVerdict(t *testing.T) {
	output := "score: 7/10\nconcerns:\n- handler missing nil check\nsuggestions:\n- add logging\n- bump test coverage\n"
	reg := newReg(t, stubAdapter{id: "rev", strengths: []string{"review"}, output: output})
	v, err := Critique(context.Background(), Request{Diff: "+x", OriginalPrompt: "add handler"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if v.AdapterID != "rev" {
		t.Errorf("adapter id: got %q", v.AdapterID)
	}
	if v.Score != 7 {
		t.Errorf("score: want 7, got %v", v.Score)
	}
	if len(v.Concerns) != 1 || v.Concerns[0] != "handler missing nil check" {
		t.Errorf("concerns: %+v", v.Concerns)
	}
	if len(v.Suggestions) != 2 {
		t.Errorf("suggestions: %+v", v.Suggestions)
	}
}

func TestCritiqueNilRegistry(t *testing.T) {
	if _, err := Critique(context.Background(), Request{}, nil); err == nil {
		t.Error("expected error for nil registry")
	}
}

func TestCritiqueUnregisteredAdapter(t *testing.T) {
	reg := newReg(t, stubAdapter{id: "x", strengths: []string{"review"}})
	if _, err := Critique(context.Background(), Request{AdapterID: "missing"}, reg); err == nil {
		t.Error("expected error for missing adapter id")
	}
}

func TestParseScoreEdgeCases(t *testing.T) {
	cases := map[string]float64{
		"score: 9/10":  9,
		"Score: 10/10": 10,
		"score:":       0,
		"score: abc":   0,
		"score: 7":     7,
	}
	for in, want := range cases {
		if got := parseScore(in); got != want {
			t.Errorf("parseScore(%q)=%v, want %v", in, got, want)
		}
	}
}

func TestParseVerdictHandlesMissingSections(t *testing.T) {
	v := parseVerdict("x", "score: 5/10")
	if v.Score != 5 || len(v.Concerns) != 0 || len(v.Suggestions) != 0 {
		t.Errorf("unexpected: %+v", v)
	}
}
