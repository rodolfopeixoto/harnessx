// SPDX-License-Identifier: MIT

package router

import (
	"context"
	"testing"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type strengthFake struct {
	id        string
	strengths []string
}

func (f strengthFake) ID() string   { return f.id }
func (f strengthFake) Name() string { return f.id }
func (f strengthFake) Capabilities() agents.Capabilities {
	return agents.Capabilities{Strengths: f.strengths}
}
func (f strengthFake) Healthcheck(context.Context) agents.HealthcheckResult {
	return agents.HealthcheckResult{OK: true}
}
func (f strengthFake) Run(context.Context, agents.AgentRequest) agents.AgentResult {
	return agents.AgentResult{}
}
func (f strengthFake) ParseUsage(agents.AgentOutput) agents.Usage { return agents.Usage{} }
func (f strengthFake) ClassifyFailure(agents.AgentOutput, int, error) agents.FailureType {
	return agents.FailureNone
}

func newRegistryWith(t *testing.T, adapters ...strengthFake) *agents.Registry {
	t.Helper()
	r := agents.NewRegistry()
	for _, a := range adapters {
		if err := r.Register(a); err != nil {
			t.Fatal(err)
		}
	}
	return r
}

func TestMatchReturnsRankedByScore(t *testing.T) {
	reg := newRegistryWith(t,
		strengthFake{id: "claude", strengths: []string{"code", "reasoning"}},
		strengthFake{id: "gemini", strengths: []string{"vision", "image"}},
		strengthFake{id: "codex", strengths: []string{"code", "tests"}},
	)
	got := Match([]string{"code"}, reg)
	if len(got) < 2 {
		t.Fatalf("want ≥2 matches, got %d", len(got))
	}
	if got[0].Score != 1.0 {
		t.Errorf("top score: want 1.0, got %v", got[0].Score)
	}
}

func TestMatchEmptyOnNoOverlap(t *testing.T) {
	reg := newRegistryWith(t, strengthFake{id: "x", strengths: []string{"docs"}})
	got := Match([]string{"image"}, reg)
	if len(got) != 0 {
		t.Errorf("want 0 matches, got %v", got)
	}
}

func TestMatchEmptyTags(t *testing.T) {
	reg := newRegistryWith(t, strengthFake{id: "x", strengths: []string{"code"}})
	got := Match(nil, reg)
	if got != nil {
		t.Errorf("nil tags should return nil, got %v", got)
	}
}

func TestMatchNilRegistry(t *testing.T) {
	got := Match([]string{"code"}, nil)
	if got != nil {
		t.Errorf("nil registry should return nil, got %v", got)
	}
}

func TestPickReturnsTopOrFalse(t *testing.T) {
	reg := newRegistryWith(t, strengthFake{id: "claude", strengths: []string{"code"}})
	c, ok := Pick([]string{"code"}, reg)
	if !ok || c.AdapterID != "claude" {
		t.Errorf("Pick failed: ok=%v id=%q", ok, c.AdapterID)
	}
	_, ok2 := Pick([]string{"audio"}, reg)
	if ok2 {
		t.Error("Pick should fail when no overlap")
	}
}

func TestStableTieBreakByID(t *testing.T) {
	reg := newRegistryWith(t,
		strengthFake{id: "zeta", strengths: []string{"code"}},
		strengthFake{id: "alpha", strengths: []string{"code"}},
	)
	c, _ := Pick([]string{"code"}, reg)
	if c.AdapterID != "alpha" {
		t.Errorf("alphabetical tiebreak: want alpha, got %q", c.AdapterID)
	}
}
