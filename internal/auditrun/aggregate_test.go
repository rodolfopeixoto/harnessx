// SPDX-License-Identifier: MIT

package auditrun

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func TestAggregate_CountsAndPassRate(t *testing.T) {
	features := []Feature{
		{ID: "a", Priority: constants.AuditSeverityP0},
		{ID: "b", Priority: constants.AuditSeverityP1},
	}
	results := []Result{
		{FeatureID: "a", Status: constants.AuditStatusPassed},
		{FeatureID: "a", Status: constants.AuditStatusFailed},
		{FeatureID: "b", Status: constants.AuditStatusPassed, Visual: &VisualDiff{VisualStatus: constants.AuditVisualMinorDiff}},
	}
	s := Aggregate(results, features)
	require.Equal(t, 2, s.Counts[constants.AuditStatusPassed])
	require.Equal(t, 1, s.Counts[constants.AuditStatusFailed])
	require.Equal(t, 1, s.Visual[constants.AuditVisualMinorDiff])
	require.Equal(t, 2, s.Severity[constants.AuditSeverityP0]+s.Severity[constants.AuditSeverityP1])
	require.InDelta(t, 2.0/3.0, s.PassRate, 0.001)
	require.Equal(t, 3, s.TotalResults)
	require.Equal(t, 2, s.TotalFeatures)
}

func TestBuildBacklog_SortsBySeverity(t *testing.T) {
	features := []Feature{
		{ID: "a", Name: "Alpha", Route: "/a", Role: RoleOperator, ExpectedHTTPStatus: 200, Priority: constants.AuditSeverityP1},
		{ID: "b", Name: "Beta", Route: "/b", Role: RoleAdmin, ExpectedHTTPStatus: 200, Priority: constants.AuditSeverityP2},
	}
	results := []Result{
		{FeatureID: "a", Viewport: "desktop", Status: constants.AuditStatusFailed, Reason: "boom"},
		{FeatureID: "b", Viewport: "mobile", Status: constants.AuditStatusPermissionError, Reason: "401"},
	}
	backlog := BuildBacklog(results, features)
	require.Len(t, backlog, 2)
	require.Equal(t, constants.AuditSeverityP0, backlog[0].Severity)
	require.Contains(t, backlog[0].Reproduce, "AUDIT_FEATURE=b")
}

func TestBuildBacklog_OmitsPassed(t *testing.T) {
	features := []Feature{{ID: "a"}}
	results := []Result{{FeatureID: "a", Status: constants.AuditStatusPassed}}
	require.Empty(t, BuildBacklog(results, features))
}

func TestWriteJSONFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "f.json")
	require.NoError(t, WriteJSONFile(path, map[string]int{"a": 1}))
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	var got map[string]int
	require.NoError(t, json.Unmarshal(body, &got))
	require.Equal(t, 1, got["a"])
}

func TestSuggestionFor_KnownStatuses(t *testing.T) {
	require.Contains(t, suggestionFor(Result{Status: constants.AuditStatusSelectorMissing}), "data-testid")
	require.Contains(t, suggestionFor(Result{Status: constants.AuditStatusAPIError}), "API handler")
	require.Contains(t, suggestionFor(Result{Status: constants.AuditStatusPermissionError}), "role gate")
	require.Contains(t, suggestionFor(Result{Status: constants.AuditStatusLayoutCollapsed}), "tokens")
	require.Contains(t, suggestionFor(Result{Status: constants.AuditStatusDataMissing}), "seed fixture")
}

func TestClassifySeverity_HighestForBreakingFailures(t *testing.T) {
	cases := map[string]string{
		constants.AuditStatusPermissionError: constants.AuditSeverityP0,
		constants.AuditStatusWrongScreen:     constants.AuditSeverityP0,
		constants.AuditStatusAPIError:        constants.AuditSeverityP1,
		constants.AuditStatusPartial:         constants.AuditSeverityP2,
	}
	for status, want := range cases {
		got := classifySeverity(Feature{Priority: constants.AuditSeverityP3}, Result{Status: status})
		require.Equal(t, want, got, status)
	}
}
