// SPDX-License-Identifier: MIT

package repl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/intentplan"
)

type LLMPlannerOptions struct {
	Adapter         agents.AgentAdapter
	Model           string
	RequestTimeout  time.Duration
	WorkingDir      string
	PromptTemplate  string
	OnParseFallback func(prompt string, raw string, err error)
}

func NewLLMPlanner(opts LLMPlannerOptions) (Planner, error) {
	if opts.Adapter == nil {
		return nil, errors.New("repl: LLM planner needs an adapter")
	}
	if opts.RequestTimeout == 0 {
		opts.RequestTimeout = 90 * time.Second
	}
	if opts.PromptTemplate == "" {
		opts.PromptTemplate = DefaultLLMPrompt
	}
	return func(ctx context.Context, goal intentplan.Goal, prompt string) (intentplan.Plan, error) {
		req := agents.AgentRequest{
			Prompt:     renderPlannerPrompt(opts.PromptTemplate, goal, prompt),
			Model:      opts.Model,
			WorkingDir: opts.WorkingDir,
			Timeout:    opts.RequestTimeout,
			Extra:      map[string]string{"task": "planning", "goal": string(goal)},
		}
		res := opts.Adapter.Run(ctx, req)
		if res.Err != nil {
			return intentplan.Plan{}, fmt.Errorf("planner adapter: %w", res.Err)
		}
		raw := strings.TrimSpace(res.Output.FinalMessage)
		if raw == "" {
			raw = strings.TrimSpace(string(res.Output.Stdout))
		}
		plan, err := decodePlan(goal, raw)
		if err != nil {
			if opts.OnParseFallback != nil {
				opts.OnParseFallback(prompt, raw, err)
			}
			plan = DefaultPlan(goal, prompt)
			plan.Intent = prompt
			return plan, nil
		}
		if plan.Intent == "" {
			plan.Intent = prompt
		}
		if plan.Goal == "" {
			plan.Goal = goal
		}
		if err := plan.Validate(); err != nil {
			return intentplan.Plan{}, fmt.Errorf("planner returned invalid plan: %w", err)
		}
		return plan, nil
	}, nil
}

const DefaultLLMPrompt = `You are the HarnessX planner.

Goal: {{goal}}
User prompt: {{prompt}}

Return a single JSON object matching:
{
  "goal": "{{goal}}",
  "intent": "<one-line restatement>",
  "steps": [
    {"kind": "harness", "title": "<short>", "cmd": ["<allowed-cmd>", "<arg>", ...]}
  ],
  "exit_when": {"all_pass": ["ci"]}
}

Constraints:
- "goal" must equal "{{goal}}".
- Each "harness" step's first "cmd" entry must be in the goal palette.
- Prefer the cheapest deterministic step that still verifies the change.
- Do not add commentary outside the JSON object.
`

func renderPlannerPrompt(tmpl string, goal intentplan.Goal, prompt string) string {
	out := strings.ReplaceAll(tmpl, "{{goal}}", string(goal))
	out = strings.ReplaceAll(out, "{{prompt}}", prompt)
	return out
}

func decodePlan(goal intentplan.Goal, body string) (intentplan.Plan, error) {
	body = strings.TrimSpace(body)
	body = extractJSONObject(body)
	if body == "" {
		return intentplan.Plan{}, errors.New("planner: no JSON object found in response")
	}
	var plan intentplan.Plan
	if err := json.Unmarshal([]byte(body), &plan); err != nil {
		return intentplan.Plan{}, err
	}
	if plan.Goal != "" && plan.Goal != goal {
		return intentplan.Plan{}, fmt.Errorf("planner: returned goal %q but session is %q", plan.Goal, goal)
	}
	return plan, nil
}

func extractJSONObject(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end <= start {
		return ""
	}
	return s[start : end+1]
}
