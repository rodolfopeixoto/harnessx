// SPDX-License-Identifier: MIT

package router

import "github.com/ropeixoto/harnessx/internal/agents"

// Defaults returns the canonical task → chain map (spec §17) projected
// over whichever adapters are currently registered. Missing adapters are
// silently dropped; chains stay useful as long as at least one entry
// remains. Three callers (`workflow.executeAgents`, `routescmd.Run`,
// `explaincmd.Run`) share this so they can't drift apart.
func Defaults(reg *agents.Registry) map[string]RouteConfig {
	present := func(ids ...string) []string {
		var out []string
		for _, id := range ids {
			if _, ok := reg.Get(id); ok {
				out = append(out, id)
			}
		}
		return out
	}
	pick := func(primary string, fallback ...string) RouteConfig {
		all := append([]string{primary}, fallback...)
		filtered := present(all...)
		if len(filtered) == 0 {
			return RouteConfig{Primary: primary, Fallback: fallback}
		}
		return RouteConfig{Primary: filtered[0], Fallback: filtered[1:]}
	}
	return map[string]RouteConfig{
		"prompt_refinement":     pick("antigravity", "kimi", "claude"),
		"planning":              pick("claude", "kimi", "antigravity"),
		"codebase_exploration":  pick("kimi", "claude", "antigravity"),
		"implementation":        pick("codex", "claude", "antigravity", "kimi"),
		"design_to_product":     pick("claude", "codex", "kimi", "antigravity"),
		"resource_optimization": pick("claude", "codex", "kimi", "antigravity"),
		"dependency_audit":      pick("kimi", "claude", "antigravity"),
		"security_review":       pick("claude", "kimi", "codex"),
		"cheap_review":          pick("antigravity", "kimi", "codex"),
	}
}
