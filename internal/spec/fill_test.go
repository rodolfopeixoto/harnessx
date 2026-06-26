// SPDX-License-Identifier: MIT

package spec

import (
	"context"
	"testing"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/domain"
)

type stubAdapter struct {
	resp string
	cost float64
	err  error
}

func (s stubAdapter) ID() string                        { return "stub" }
func (s stubAdapter) Name() string                      { return "stub" }
func (s stubAdapter) Capabilities() agents.Capabilities { return agents.Capabilities{} }
func (s stubAdapter) Healthcheck(_ context.Context) agents.HealthcheckResult {
	return agents.HealthcheckResult{OK: true}
}
func (s stubAdapter) ParseUsage(_ agents.AgentOutput) agents.Usage { return agents.Usage{} }
func (s stubAdapter) ClassifyFailure(_ agents.AgentOutput, _ int, _ error) agents.FailureType {
	return agents.FailureNone
}
func (s stubAdapter) Run(_ context.Context, _ agents.AgentRequest) agents.AgentResult {
	return agents.AgentResult{
		Output: agents.AgentOutput{FinalMessage: s.resp},
		Usage:  agents.Usage{EstimatedCostUSD: s.cost},
		Err:    s.err,
	}
}

func TestFillReplacesEmptyFields_BUG26(t *testing.T) {
	s := NewFromPrompt("add /healthz", domain.ModeFeature)
	// Sanity: NewFromPrompt does not pre-fill Scope etc.
	if len(s.Scope) != 0 {
		t.Fatalf("expected empty Scope, got %v", s.Scope)
	}
	adapter := stubAdapter{
		resp: `{"user_problem":"observability gap","scope":["add /healthz returning 200","docs"],"test_plan":["unit test","e2e via curl"]}`,
		cost: 0.01,
	}
	filled, cost, err := Filler{Adapter: adapter, BudgetUSD: 0.05}.Fill(context.Background(), s)
	if err != nil {
		t.Fatalf("Fill error: %v", err)
	}
	if cost != 0.01 {
		t.Fatalf("cost = %v, want 0.01", cost)
	}
	if filled.UserProblem != "observability gap" {
		t.Fatalf("UserProblem = %q", filled.UserProblem)
	}
	if len(filled.Scope) != 2 || filled.Scope[0] != "add /healthz returning 200" {
		t.Fatalf("Scope = %v", filled.Scope)
	}
	if len(filled.TestPlan) != 2 {
		t.Fatalf("TestPlan = %v", filled.TestPlan)
	}
}

func TestFillRespectsBudget_BUG26(t *testing.T) {
	s := NewFromPrompt("add x", domain.ModeFeature)
	adapter := stubAdapter{resp: `{"user_problem":"y"}`, cost: 0.20}
	_, _, err := Filler{Adapter: adapter, BudgetUSD: 0.05}.Fill(context.Background(), s)
	if err == nil {
		t.Fatal("expected budget error")
	}
}

func TestFillTolerantsMarkdownFence(t *testing.T) {
	s := NewFromPrompt("add x", domain.ModeFeature)
	adapter := stubAdapter{resp: "```json\n{\"user_problem\":\"clear\"}\n```"}
	filled, _, err := Filler{Adapter: adapter}.Fill(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if filled.UserProblem != "clear" {
		t.Fatalf("UserProblem = %q", filled.UserProblem)
	}
}

func TestFillNoAdapterIsNoOp(t *testing.T) {
	s := NewFromPrompt("noop", domain.ModeFeature)
	filled, cost, err := Filler{}.Fill(context.Background(), s)
	if err != nil || cost != 0 {
		t.Fatalf("Fill noop err=%v cost=%v", err, cost)
	}
	if filled.Prompt != s.Prompt {
		t.Fatalf("noop should be identity")
	}
}
