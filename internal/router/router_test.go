package router

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/fake"
)

func newReg(t *testing.T, ads ...agents.AgentAdapter) *agents.Registry {
	t.Helper()
	r := agents.NewRegistry()
	for _, a := range ads {
		require.NoError(t, r.Register(a))
	}
	return r
}

func TestSelect_PrimaryAndFallback(t *testing.T) {
	a := fake.New("a")
	b := fake.New("b")
	c := fake.New("c")
	reg := newReg(t, a, b, c)

	r := New(reg, map[string]RouteConfig{
		"implementation": {Primary: "a", Fallback: []string{"b", "c"}, BudgetUSD: 1.0},
	})
	d, err := r.Select("implementation")
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, idsOf(d.Chain))
}

func TestSelect_MissingAgentsDropped(t *testing.T) {
	a := fake.New("a")
	reg := newReg(t, a)
	r := New(reg, map[string]RouteConfig{
		"implementation": {Primary: "a", Fallback: []string{"missing"}},
	})
	d, err := r.Select("implementation")
	require.NoError(t, err)
	require.Equal(t, []string{"a"}, idsOf(d.Chain))
}

func TestSelect_NoRoute(t *testing.T) {
	r := New(newReg(t), map[string]RouteConfig{})
	_, err := r.Select("nothing")
	require.Error(t, err)
}

func TestExecute_FallsBackOnRecoverableFailure(t *testing.T) {
	a := fake.New("a")
	a.ForceFailure = agents.FailureRateLimit
	a.ExitCode = 1
	a.RunErr = errors.New("rate limited")

	b := fake.New("b")
	b.FinalMessage = "did it"

	reg := newReg(t, a, b)
	r := New(reg, map[string]RouteConfig{
		"implementation": {Primary: "a", Fallback: []string{"b"}, BudgetUSD: 1.0},
	})

	out, err := r.Execute(context.Background(), "implementation", agents.AgentRequest{Prompt: "p"}, nil)
	require.NoError(t, err)
	require.True(t, out.Succeeded)
	require.Equal(t, "b", out.Selected.ID())
	require.Len(t, out.Fallbacks, 1)
	require.Equal(t, "a", out.Fallbacks[0].From)
	require.Equal(t, agents.FailureRateLimit, out.Fallbacks[0].Failure)
}

func TestExecute_StopsOnAuthFailure(t *testing.T) {
	a := fake.New("a")
	a.ForceFailure = agents.FailureAuth
	a.ExitCode = 1
	a.RunErr = errors.New("unauthorized")

	b := fake.New("b") // should not be reached

	reg := newReg(t, a, b)
	r := New(reg, map[string]RouteConfig{
		"implementation": {Primary: "a", Fallback: []string{"b"}, BudgetUSD: 1.0},
	})

	out, err := r.Execute(context.Background(), "implementation", agents.AgentRequest{Prompt: "p"}, nil)
	require.NoError(t, err)
	require.False(t, out.Succeeded)
	require.Equal(t, "a", out.Selected.ID())
	require.Len(t, out.Fallbacks, 1)
}

func idsOf(as []agents.AgentAdapter) []string {
	out := make([]string, len(as))
	for i, a := range as {
		out[i] = a.ID()
	}
	return out
}
