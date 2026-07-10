// SPDX-License-Identifier: MIT

package orchestrate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRoleValuesLocked pins the wire-level string values of the paper
// §4.1.1 role taxonomy. Flows on disk (.harness/flows/*.yaml) address
// roles by these exact strings; renaming any of them silently breaks
// every user-shipped flow, so this test fails loudly on drift.
func TestRoleValuesLocked(t *testing.T) {
	t.Parallel()
	require.Equal(t, Role("manager"), RoleManager)
	require.Equal(t, Role("planner"), RolePlanner)
	require.Equal(t, Role("coder"), RoleCoder)
	require.Equal(t, Role("reviewer"), RoleReviewer)
	require.Equal(t, Role("tester"), RoleTester)
}

// TestKnownRolesSetLocked ensures the canonical role set stays a
// 5-tuple in the documented order (manager, planner, coder, reviewer,
// tester). Reordering breaks error messages ("want [manager planner
// coder reviewer tester]") documented in the CLI help.
func TestKnownRolesSetLocked(t *testing.T) {
	t.Parallel()
	want := []Role{"manager", "planner", "coder", "reviewer", "tester"}
	require.Equal(t, want, KnownRoles())
}

// TestTopologyValuesLocked pins the two supported topologies. Same
// contract concern: flows on disk reference these strings.
func TestTopologyValuesLocked(t *testing.T) {
	t.Parallel()
	require.Equal(t, Topology("chain"), TopologyChain)
	require.Equal(t, Topology("cyclic"), TopologyCyclic)
}

// TestValidateRejectsUnknownRole exercises the public Validate seam
// with every combination the docs promise: known role + known
// topology passes; unknown role fails; unknown topology fails.
func TestValidateRejectsUnknownRole_Contract(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		flow    Flow
		wantErr bool
	}{
		{
			name: "valid_chain_coder",
			flow: Flow{Name: "f", Topology: TopologyChain,
				Steps: []Step{{Role: RoleCoder, Command: []string{"true"}}}},
			wantErr: false,
		},
		{
			name: "unknown_role",
			flow: Flow{Name: "f", Topology: TopologyChain,
				Steps: []Step{{Role: Role("architect"), Command: []string{"true"}}}},
			wantErr: true,
		},
		{
			name: "unknown_topology",
			flow: Flow{Name: "f", Topology: Topology("mesh"),
				Steps: []Step{{Role: RoleCoder, Command: []string{"true"}}}},
			wantErr: true,
		},
		{
			name:    "empty_steps",
			flow:    Flow{Name: "f", Topology: TopologyChain, Steps: nil},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.flow.Validate()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
