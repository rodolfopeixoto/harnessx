// SPDX-License-Identifier: MIT

package router

import (
	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/intent"
)

// PickModel selects the adapter model alias most appropriate for the
// classified complexity. Adapters advertise three aliases on
// Capabilities.Models: `cheap`, `default`, `deep`. A missing alias falls
// back to `default`, then to empty (adapter picks its own model).
func PickModel(adapter agents.AgentAdapter, complexity intent.Complexity) string {
	if adapter == nil {
		return ""
	}
	models := adapter.Capabilities().Models
	if models == nil {
		return ""
	}
	want := "default"
	switch complexity {
	case intent.ComplexityTrivial:
		want = "cheap"
	case intent.ComplexityComplex:
		want = "deep"
	}
	if m, ok := models[want]; ok && m != "" {
		return m
	}
	if m, ok := models["default"]; ok {
		return m
	}
	return ""
}
