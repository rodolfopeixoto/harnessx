// SPDX-License-Identifier: MIT

package router

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/fake"
)

func TestDefaults_PicksFirstPresentAdapter(t *testing.T) {
	reg := agents.NewRegistry()
	require.NoError(t, reg.Register(fake.New("codex")))
	require.NoError(t, reg.Register(fake.New("claude")))
	got := Defaults(reg)
	require.Equal(t, "codex", got["implementation"].Primary)
	require.Contains(t, got["implementation"].Fallback, "claude")
}

func TestDefaults_KeepsConfigEvenIfNoneRegistered(t *testing.T) {
	reg := agents.NewRegistry()
	got := Defaults(reg)
	require.NotEmpty(t, got["implementation"].Primary)
}

func TestDefaults_EveryTaskPresent(t *testing.T) {
	reg := agents.NewRegistry()
	require.NoError(t, reg.Register(fake.New("fake")))
	got := Defaults(reg)
	for _, task := range []string{
		"prompt_refinement", "planning", "codebase_exploration",
		"implementation", "design_to_product", "resource_optimization",
		"dependency_audit", "security_review", "cheap_review",
	} {
		require.Containsf(t, got, task, "missing task %s", task)
	}
}
