// SPDX-License-Identifier: MIT

package router

import (
	"context"
	"testing"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/intent"
)

type fakeAdapter struct{ models map[string]string }

func (f fakeAdapter) ID() string   { return "fake" }
func (f fakeAdapter) Name() string { return "fake" }
func (f fakeAdapter) Capabilities() agents.Capabilities {
	return agents.Capabilities{Models: f.models}
}
func (f fakeAdapter) Healthcheck(ctx context.Context) agents.HealthcheckResult {
	return agents.HealthcheckResult{OK: true}
}
func (f fakeAdapter) Run(ctx context.Context, req agents.AgentRequest) agents.AgentResult {
	return agents.AgentResult{}
}
func (f fakeAdapter) ParseUsage(o agents.AgentOutput) agents.Usage { return agents.Usage{} }
func (f fakeAdapter) ClassifyFailure(o agents.AgentOutput, code int, err error) agents.FailureType {
	return agents.FailureNone
}

func TestPickModel_TrivialPicksCheap(t *testing.T) {
	a := fakeAdapter{models: map[string]string{"cheap": "haiku", "default": "sonnet", "deep": "opus"}}
	if got := PickModel(a, intent.ComplexityTrivial); got != "haiku" {
		t.Fatalf("expected haiku, got %q", got)
	}
}

func TestPickModel_ComplexPicksDeep(t *testing.T) {
	a := fakeAdapter{models: map[string]string{"cheap": "haiku", "default": "sonnet", "deep": "opus"}}
	if got := PickModel(a, intent.ComplexityComplex); got != "opus" {
		t.Fatalf("expected opus, got %q", got)
	}
}

func TestPickModel_StandardPicksDefault(t *testing.T) {
	a := fakeAdapter{models: map[string]string{"cheap": "haiku", "default": "sonnet", "deep": "opus"}}
	if got := PickModel(a, intent.ComplexityStandard); got != "sonnet" {
		t.Fatalf("expected sonnet, got %q", got)
	}
}

func TestPickModel_FallsBackToDefaultWhenAliasMissing(t *testing.T) {
	a := fakeAdapter{models: map[string]string{"default": "sonnet"}}
	if got := PickModel(a, intent.ComplexityTrivial); got != "sonnet" {
		t.Fatalf("expected default fallback, got %q", got)
	}
}

func TestPickModel_NilAdapter(t *testing.T) {
	if got := PickModel(nil, intent.ComplexityTrivial); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
