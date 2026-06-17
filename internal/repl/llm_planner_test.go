package repl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/intentplan"
)

type stubAdapter struct {
	final string
	err   error
}

func (s stubAdapter) ID() string                        { return "stub" }
func (s stubAdapter) Name() string                      { return "stub" }
func (s stubAdapter) Capabilities() agents.Capabilities { return agents.Capabilities{} }
func (s stubAdapter) Healthcheck(ctx context.Context) agents.HealthcheckResult {
	return agents.HealthcheckResult{}
}
func (s stubAdapter) ParseUsage(o agents.AgentOutput) agents.Usage { return agents.Usage{} }
func (s stubAdapter) ClassifyFailure(o agents.AgentOutput, c int, e error) agents.FailureType {
	return agents.FailureNone
}
func (s stubAdapter) Run(ctx context.Context, req agents.AgentRequest) agents.AgentResult {
	return agents.AgentResult{
		Output: agents.AgentOutput{FinalMessage: s.final},
		Err:    s.err,
	}
}

func TestNewLLMPlannerRequiresAdapter(t *testing.T) {
	if _, err := NewLLMPlanner(LLMPlannerOptions{}); err == nil {
		t.Fatal("want error")
	}
}

func TestLLMPlannerDecodesValidJSON(t *testing.T) {
	a := stubAdapter{final: `{"goal":"dev","intent":"add /healthz","steps":[{"kind":"harness","cmd":["ci"]}]}`}
	p, err := NewLLMPlanner(LLMPlannerOptions{Adapter: a, RequestTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := p(context.Background(), intentplan.GoalDev, "add /healthz")
	if err != nil {
		t.Fatal(err)
	}
	if plan.Intent != "add /healthz" {
		t.Errorf("intent: %q", plan.Intent)
	}
}

func TestLLMPlannerStripsTextAroundJSON(t *testing.T) {
	a := stubAdapter{final: "Sure! here it is:\n```json\n{\"goal\":\"dev\",\"intent\":\"x\",\"steps\":[{\"kind\":\"harness\",\"cmd\":[\"ci\"]}]}\n```\nDone."}
	p, _ := NewLLMPlanner(LLMPlannerOptions{Adapter: a})
	plan, err := p(context.Background(), intentplan.GoalDev, "x")
	if err != nil {
		t.Fatal(err)
	}
	if plan.Intent != "x" {
		t.Errorf("intent: %q", plan.Intent)
	}
}

func TestLLMPlannerFallsBackOnBadJSON(t *testing.T) {
	fallback := 0
	a := stubAdapter{final: "no json here"}
	p, _ := NewLLMPlanner(LLMPlannerOptions{
		Adapter:         a,
		OnParseFallback: func(prompt, raw string, err error) { fallback++ },
	})
	plan, err := p(context.Background(), intentplan.GoalDev, "add /healthz")
	if err != nil {
		t.Fatal(err)
	}
	if fallback != 1 {
		t.Errorf("fallback hook should fire once, got %d", fallback)
	}
	if len(plan.Steps) == 0 {
		t.Errorf("fallback plan should have steps")
	}
}

func TestLLMPlannerPropagatesAdapterError(t *testing.T) {
	a := stubAdapter{err: errors.New("network")}
	p, _ := NewLLMPlanner(LLMPlannerOptions{Adapter: a})
	if _, err := p(context.Background(), intentplan.GoalDev, "x"); err == nil {
		t.Fatal("want error")
	}
}

func TestDecodePlanRejectsGoalMismatch(t *testing.T) {
	_, err := decodePlan(intentplan.GoalDev, `{"goal":"ops","intent":"x","steps":[{"kind":"harness","cmd":["doctor"]}]}`)
	if err == nil {
		t.Fatal("want error")
	}
}

func TestExtractJSONObjectFindsObject(t *testing.T) {
	if got := extractJSONObject("noise {x:1} tail"); got != "{x:1}" {
		t.Errorf("got %q", got)
	}
	if got := extractJSONObject("no braces"); got != "" {
		t.Errorf("got %q", got)
	}
}
