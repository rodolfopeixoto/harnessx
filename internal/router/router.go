// SPDX-License-Identifier: MIT

// Package router implements deterministic agent selection with an
// explainable decision trail and a fallback executor.
package router

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
)

// RouteConfig describes the policy for a single task type (e.g.
// "implementation", "planning", "security_review"). Order in Fallback
// matters — earlier IDs are tried first.
type RouteConfig struct {
	Primary   string   `yaml:"primary" json:"primary"`
	Fallback  []string `yaml:"fallback" json:"fallback"`
	BudgetUSD float64  `yaml:"budget_usd" json:"budget_usd"`
	Model     string   `yaml:"model,omitempty" json:"model,omitempty"`
}

// Router maps task types to deterministic chains over a Registry of agents.
type Router struct {
	routes   map[string]RouteConfig
	registry *agents.Registry
}

func New(registry *agents.Registry, routes map[string]RouteConfig) *Router {
	return &Router{routes: routes, registry: registry}
}

// Decision is the explainable output of Select. Chain[0] is the primary
// pick; subsequent entries are fallbacks in order.
type Decision struct {
	Task      string
	Chain     []agents.AgentAdapter
	Model     string
	BudgetUSD float64
	Reasons   []string
}

// Select resolves a task type to a chain of adapters. Missing adapters in
// the chain are silently dropped from the result but recorded in Reasons.
func (r *Router) Select(task string) (Decision, error) {
	cfg, ok := r.routes[task]
	if !ok {
		return Decision{}, fmt.Errorf("router: no route configured for task %q", task)
	}
	d := Decision{Task: task, BudgetUSD: cfg.BudgetUSD, Model: cfg.Model}
	d.Reasons = append(d.Reasons, fmt.Sprintf("task=%q primary=%q fallback=%v budget=$%.2f", task, cfg.Primary, cfg.Fallback, cfg.BudgetUSD))

	ids := append([]string{cfg.Primary}, cfg.Fallback...)
	for _, id := range ids {
		if id == "" {
			continue
		}
		a, ok := r.registry.Get(id)
		if !ok {
			d.Reasons = append(d.Reasons, fmt.Sprintf("skipped %q: not registered", id))
			continue
		}
		d.Chain = append(d.Chain, a)
		d.Reasons = append(d.Reasons, fmt.Sprintf("included %q (%s)", a.ID(), a.Name()))
	}
	if len(d.Chain) == 0 {
		return d, errors.New("router: no registered adapters in chain")
	}
	return d, nil
}

// Routes returns a copy of the configured route map for introspection.
func (r *Router) Routes() map[string]RouteConfig {
	out := make(map[string]RouteConfig, len(r.routes))
	for k, v := range r.routes {
		out[k] = v
	}
	return out
}

// FallbackEvent records one agent's failure during Execute. The order of
// events matches the order in which the chain was attempted.
type FallbackEvent struct {
	From    string
	Failure agents.FailureType
	Detail  string
	At      time.Time
}

// ExecuteResult is what Execute returns to the caller. Result is the
// final attempt's outcome; Tried lists the agents tried before success
// (or the full chain if every attempt failed).
type ExecuteResult struct {
	Decision  Decision
	Selected  agents.AgentAdapter
	Result    agents.AgentResult
	Fallbacks []FallbackEvent
	Succeeded bool
}

// Execute runs the chain returned by Select(task), advancing to the next
// agent on any recoverable failure. Auth + permanent + nil-recoverable
// failures stop the chain immediately.
func (r *Router) Execute(ctx context.Context, task string, req agents.AgentRequest, clock func() time.Time) (ExecuteResult, error) {
	if clock == nil {
		clock = time.Now
	}
	d, err := r.Select(task)
	if err != nil {
		return ExecuteResult{Decision: d}, err
	}
	out := ExecuteResult{Decision: d}
	for _, a := range d.Chain {
		if req.Model == "" {
			req.Model = pickModel(a, d.Model)
		}
		res := a.Run(ctx, req)
		out.Selected = a
		out.Result = res
		if res.Failure == agents.FailureNone && res.ExitCode == 0 && res.Err == nil {
			out.Succeeded = true
			return out, nil
		}
		out.Fallbacks = append(out.Fallbacks, FallbackEvent{
			From: a.ID(), Failure: res.Failure,
			Detail: trim(res.Output.Stderr, 240),
			At:     clock(),
		})
		if !res.Failure.IsRecoverable() {
			return out, nil
		}
		// reset model so the next adapter picks its own default
		req.Model = ""
	}
	return out, nil
}

func pickModel(a agents.AgentAdapter, override string) string {
	if override != "" {
		return override
	}
	caps := a.Capabilities()
	if caps.Models != nil {
		if m, ok := caps.Models["default"]; ok {
			return m
		}
	}
	return ""
}

func trim(b []byte, max int) string {
	s := string(b)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
