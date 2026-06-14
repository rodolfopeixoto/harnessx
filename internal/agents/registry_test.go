// SPDX-License-Identifier: MIT

package agents

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubAdapter struct {
	id   string
	name string
}

func (s stubAdapter) ID() string                 { return s.id }
func (s stubAdapter) Name() string               { return s.name }
func (s stubAdapter) Capabilities() Capabilities { return Capabilities{} }
func (s stubAdapter) Healthcheck(_ context.Context) HealthcheckResult {
	return HealthcheckResult{OK: true}
}
func (s stubAdapter) Run(_ context.Context, _ AgentRequest) AgentResult { return AgentResult{} }
func (s stubAdapter) ParseUsage(_ AgentOutput) Usage                    { return Usage{} }
func (s stubAdapter) ClassifyFailure(_ AgentOutput, _ int, _ error) FailureType {
	return FailureNone
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(stubAdapter{id: "a", name: "Alpha"}))
	got, ok := r.Get("a")
	require.True(t, ok)
	require.Equal(t, "Alpha", got.Name())
}

func TestRegistry_RejectsNilAndDuplicate(t *testing.T) {
	r := NewRegistry()
	require.Error(t, r.Register(nil))
	require.NoError(t, r.Register(stubAdapter{id: "x"}))
	require.Error(t, r.Register(stubAdapter{id: "x"}))
}

func TestRegistry_IDsAndAllSorted(t *testing.T) {
	r := NewRegistry()
	for _, id := range []string{"c", "a", "b"} {
		require.NoError(t, r.Register(stubAdapter{id: id}))
	}
	require.Equal(t, []string{"a", "b", "c"}, r.IDs())
	all := r.All()
	require.Len(t, all, 3)
	require.Equal(t, "a", all[0].ID())
}

func TestFailureType_IsRecoverable(t *testing.T) {
	cases := map[FailureType]bool{
		FailureNone:         false,
		FailureRateLimit:    true,
		FailureContextLimit: true,
		FailureTransient:    true,
		FailureTimeout:      true,
		FailureAuth:         false,
		FailurePermanent:    false,
	}
	for f, want := range cases {
		require.Equalf(t, want, f.IsRecoverable(), "failure %s", f)
	}
}
