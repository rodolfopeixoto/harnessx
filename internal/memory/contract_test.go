// SPDX-License-Identifier: MIT

package memory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestKindValuesLocked pins the paper §3.2.1–§3.2.5 kind taxonomy
// strings. These land in the memories.kind column and are read back
// by the recall pipeline — renaming any of them silently orphans
// every existing row.
func TestKindValuesLocked(t *testing.T) {
	t.Parallel()
	require.Equal(t, "working", KindWorking)
	require.Equal(t, "semantic", KindSemantic)
	require.Equal(t, "experiential", KindExperiential)
	require.Equal(t, "long_term", KindLongTerm)
	require.Equal(t, "multi_agent", KindMultiAgent)
}

// TestKnownKindsSetLocked ensures the taxonomy stays exactly 5 kinds
// in documented order. Downstream sensors ("unknown kind") emit this
// list verbatim in their error output.
func TestKnownKindsSetLocked(t *testing.T) {
	t.Parallel()
	want := []string{"working", "semantic", "experiential", "long_term", "multi_agent"}
	require.Equal(t, want, KnownKinds())
}

// TestConfidenceFloorLocked pins the 0.4 gate from CLAUDE.md
// ("Mutate project memory without an evidence_run_id + confidence
// ≥ 0.4"). Lowering the floor silently weakens the promotion gate.
func TestConfidenceFloorLocked(t *testing.T) {
	t.Parallel()
	require.InDelta(t, 0.4, confidenceFloor, 1e-9,
		"CLAUDE.md pins the memory confidence floor at 0.4")
}

// TestSensitiveReMatchesDocumentedShapes exercises the sensitive
// content gate on every shape called out in the regex comment. If
// any of these stop matching, the gate is silently weaker.
func TestSensitiveReMatchesDocumentedShapes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		payload string
	}{
		{"aws_access_key", "AKIAIOSFODNN7EXAMPLE"},
		{"aws_secret", "aws_secret_access_key=abcdefghijklmnopqrstuvwxyz0123456789ABCD"},
		{"slack_bot_token", "xox\x62-1234567890-1234567890-abcdefghijklmno"},
		{"private_key", "-----BEGIN RSA PRIVATE KEY-----"},
		{"github_token", "ghp_1234567890123456789012345678901234567890"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.True(t, sensitiveRe.MatchString(tc.payload),
				"sensitiveRe must catch %s shape", tc.name)
		})
	}
}

// TestSensitiveReAvoidsBenignPhrases guards against the regex being
// widened to a level where docs would trip it.
func TestSensitiveReAvoidsBenignPhrases(t *testing.T) {
	t.Parallel()
	benign := []string{
		"how to rotate your AWS credentials",
		"the reviewer approved the PR",
		"we support Slack, Discord and Teams",
	}
	for _, b := range benign {
		require.False(t, sensitiveRe.MatchString(b), "should not flag: %q", b)
	}
}

// TestValidKindGate exercises the private validKind through
// KnownKinds — happy path + rejection of an unknown taxonomy value.
func TestValidKindGate(t *testing.T) {
	t.Parallel()
	for _, k := range KnownKinds() {
		require.True(t, validKind(k), "known kind %q must pass gate", k)
	}
	require.False(t, validKind("procedural"), "unknown kind must be rejected")
	require.False(t, validKind(""), "empty kind must be rejected")
}
