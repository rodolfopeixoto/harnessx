// SPDX-License-Identifier: MIT

package spec

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ropeixoto/harnessx/internal/agents"
)

// Filler runs a single adapter call to populate Spec fields the
// deterministic `NewFromPrompt` left as `_TODO:` placeholders. Audit
// BUG-26: previously the spec emitted by `harness feature` was a TODO
// shell, so spec-driven development could not start without a manual
// rewrite. The adapter is asked for a strict JSON object so the parsing
// is unambiguous and budget-bounded.
type Filler struct {
	Adapter agents.AgentAdapter
	// BudgetUSD caps the cost. 0 means no cap from this helper (caller
	// should usually pass 0.05 — same default as `harness ask`).
	BudgetUSD float64
}

// fillResponse is the JSON contract we coerce the adapter into. Every
// field maps to the matching Spec field; unknown keys are ignored.
type fillResponse struct {
	UserProblem     string   `json:"user_problem"`
	ExpectedOutcome string   `json:"expected_outcome"`
	Scope           []string `json:"scope"`
	OutOfScope      []string `json:"out_of_scope"`
	BusinessRules   []string `json:"business_rules"`
	UXExpectations  []string `json:"ux_expectations"`
	APIExpectations []string `json:"api_expectations"`
	DataModel       []string `json:"data_model"`
	Security        []string `json:"security"`
	Performance     []string `json:"performance"`
	Observability   []string `json:"observability"`
	TestPlan        []string `json:"test_plan"`
	E2EPlan         []string `json:"e2e_plan"`
	Assumptions     []string `json:"assumptions"`
	OpenQuestions   []string `json:"open_questions"`
}

const fillSystemPrompt = `You are filling a spec-driven-development template.
Return a SINGLE JSON object with these string-or-array-of-string keys:
user_problem, expected_outcome, scope, out_of_scope, business_rules,
ux_expectations, api_expectations, data_model, security, performance,
observability, test_plan, e2e_plan, assumptions, open_questions.

Rules:
  1. NEVER include explanations, markdown, or text outside the JSON.
  2. Omit keys you cannot answer from the prompt — do NOT invent.
  3. Keep each list element under 120 characters.
  4. Be concrete: cite endpoints, table names, file paths when present.`

// Fill enriches s with adapter-generated content. Returns the new Spec
// and the adapter cost (0 on no-op). Errors from the adapter are
// returned untouched so the caller can decide whether to surface or log.
func (f Filler) Fill(ctx context.Context, s Spec) (Spec, float64, error) {
	if f.Adapter == nil {
		return s, 0, nil
	}
	prompt := fillSystemPrompt + "\n\nOriginal prompt:\n" + s.Prompt
	res := f.Adapter.Run(ctx, agents.AgentRequest{Prompt: prompt})
	cost := res.Usage.EstimatedCostUSD
	if res.Err != nil {
		return s, cost, res.Err
	}
	if f.BudgetUSD > 0 && cost > f.BudgetUSD {
		return s, cost, fmt.Errorf("spec fill: cost $%.4f exceeded --budget-usd $%.4f", cost, f.BudgetUSD)
	}
	payload := extractJSONObject(res.Output.FinalMessage)
	if payload == "" {
		return s, cost, fmt.Errorf("spec fill: adapter did not return a JSON object")
	}
	var fr fillResponse
	if err := json.Unmarshal([]byte(payload), &fr); err != nil {
		return s, cost, fmt.Errorf("spec fill: decode JSON: %w", err)
	}
	return mergeFillResponse(s, fr), cost, nil
}

func mergeFillResponse(s Spec, fr fillResponse) Spec {
	if v := strings.TrimSpace(fr.UserProblem); v != "" {
		s.UserProblem = v
	}
	if v := strings.TrimSpace(fr.ExpectedOutcome); v != "" {
		s.ExpectedOutcome = v
	}
	s.Scope = appendIfEmpty(s.Scope, fr.Scope)
	s.OutOfScope = appendIfEmpty(s.OutOfScope, fr.OutOfScope)
	s.BusinessRules = appendIfEmpty(s.BusinessRules, fr.BusinessRules)
	s.UXExpectations = appendIfEmpty(s.UXExpectations, fr.UXExpectations)
	s.APIExpectations = appendIfEmpty(s.APIExpectations, fr.APIExpectations)
	s.DataModel = appendIfEmpty(s.DataModel, fr.DataModel)
	s.Security = appendIfEmpty(s.Security, fr.Security)
	s.Performance = appendIfEmpty(s.Performance, fr.Performance)
	s.Observability = appendIfEmpty(s.Observability, fr.Observability)
	s.TestPlan = appendIfEmpty(s.TestPlan, fr.TestPlan)
	s.E2EPlan = appendIfEmpty(s.E2EPlan, fr.E2EPlan)
	s.Assumptions = appendIfEmpty(s.Assumptions, fr.Assumptions)
	s.OpenQuestions = appendIfEmpty(s.OpenQuestions, fr.OpenQuestions)
	return s
}

func appendIfEmpty(dst, src []string) []string {
	if len(dst) > 0 {
		return dst
	}
	out := make([]string, 0, len(src))
	for _, v := range src {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// extractJSONObject pulls the first balanced {...} block from raw,
// tolerating adapters that wrap JSON in markdown fences (```json ... ```).
func extractJSONObject(raw string) string {
	raw = strings.TrimSpace(raw)
	start := strings.IndexByte(raw, '{')
	end := strings.LastIndexByte(raw, '}')
	if start < 0 || end <= start {
		return ""
	}
	return raw[start : end+1]
}
