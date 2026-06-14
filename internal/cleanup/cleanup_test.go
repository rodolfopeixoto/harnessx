// SPDX-License-Identifier: MIT

package cleanup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type stubDetector struct {
	name string
	caps []Finding
	err  error
}

func (s stubDetector) Name() string { return s.name }
func (s stubDetector) Detect(_ context.Context, _ string) ([]Finding, error) {
	return s.caps, s.err
}

func TestScanner_SortsByRiskAndKind(t *testing.T) {
	scanner := New(
		stubDetector{name: "a", caps: []Finding{
			{Kind: "z", Risk: RiskLow, Path: "/z"},
			{Kind: "a", Risk: RiskHigh, Path: "/a-high"},
		}},
		stubDetector{name: "b", caps: []Finding{
			{Kind: "a", Risk: RiskMedium, Path: "/a-med"},
		}},
	)
	out, err := scanner.Scan(context.Background(), "/")
	require.NoError(t, err)
	require.Len(t, out, 3)
	require.Equal(t, RiskHigh, out[0].Risk)
	require.Equal(t, RiskMedium, out[1].Risk)
	require.Equal(t, RiskLow, out[2].Risk)
}

func TestScanner_DetectorError(t *testing.T) {
	scanner := New(stubDetector{name: "broken", err: errors.New("boom")})
	_, err := scanner.Scan(context.Background(), "/")
	var derr *DetectorError
	require.ErrorAs(t, err, &derr)
	require.Equal(t, "broken", derr.Name)
}

func TestPolicy_MatchAllowlist(t *testing.T) {
	policy := Policy{
		Rules: []PolicyRule{
			{Kind: "cache", Allowlist: []string{"/tmp/*"}, MaxRisk: constants.CleanupRiskMedium},
		},
	}
	_, matched := policy.Match(Finding{Kind: "cache", Path: "/tmp/x", Risk: RiskLow})
	require.True(t, matched)
	_, matched = policy.Match(Finding{Kind: "cache", Path: "/other/x", Risk: RiskLow})
	require.False(t, matched)
}

func TestPolicy_MaxRiskCaps(t *testing.T) {
	policy := Policy{
		Rules: []PolicyRule{
			{Kind: "cache", Allowlist: []string{"/tmp/*"}, MaxRisk: constants.CleanupRiskLow},
		},
	}
	_, matched := policy.Match(Finding{Kind: "cache", Path: "/tmp/x", Risk: RiskHigh})
	require.False(t, matched)
}

func TestExecutor_PolicyMatchDeletesAndAudits(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "junk.txt")
	require.NoError(t, os.WriteFile(target, []byte("hello"), 0o644))
	policy := Policy{
		Globals: PolicyGlobals{RequireAcknowledgement: false},
		Rules: []PolicyRule{
			{Kind: "test", Allowlist: []string{filepath.Join(dir, "*")}, MaxRisk: constants.CleanupRiskHigh},
		},
	}
	var events []AuditEvent
	exec := NewExecutor(policy, sinkFn(func(e AuditEvent) error {
		events = append(events, e)
		return nil
	}))
	outcome, err := exec.Apply(context.Background(), Finding{Kind: "test", Path: target, Risk: RiskMedium})
	require.NoError(t, err)
	_, err = os.Stat(target)
	require.True(t, errors.Is(err, os.ErrNotExist))
	require.Equal(t, int64(5), outcome.SizeBytes)
	require.Len(t, events, 1)
	require.NotEmpty(t, events[0].ContentHash)
}

func TestExecutor_RequiresAcknowledgementWhenPolicyDemands(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "junk")
	require.NoError(t, os.WriteFile(target, []byte("x"), 0o644))
	policy := Policy{
		Globals: PolicyGlobals{RequireAcknowledgement: true},
		Rules: []PolicyRule{
			{Kind: "test", Allowlist: []string{filepath.Join(dir, "*")}, MaxRisk: constants.CleanupRiskHigh},
		},
	}
	exec := NewExecutor(policy, nil)
	exec.Acknowledgement = ""
	_, err := exec.Apply(context.Background(), Finding{Kind: "test", Path: target, Risk: RiskLow})
	require.ErrorIs(t, err, ErrAcknowledgementMissing)
}

func TestExecutor_InteractiveDenialBlocks(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "j")
	require.NoError(t, os.WriteFile(target, []byte("x"), 0o644))
	exec := NewExecutor(Policy{}, nil)
	exec.Interactive = func(Finding) (bool, error) { return false, nil }
	_, err := exec.Apply(context.Background(), Finding{Kind: "x", Path: target, Risk: RiskLow})
	require.ErrorIs(t, err, ErrUserDenied)
}

func TestExecutor_NoPolicyNoApproverFails(t *testing.T) {
	exec := NewExecutor(Policy{}, nil)
	_, err := exec.Apply(context.Background(), Finding{Kind: "x", Path: "/nope", Risk: RiskLow})
	require.ErrorIs(t, err, ErrPolicyMissing)
}

func TestLoadPolicyFile_ParsesAndDefaultsVersion(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "p.yaml")
	require.NoError(t, os.WriteFile(target, []byte("rules:\n  - kind: cache\n    allowlist:\n      - '/tmp/*'\n"), 0o644))
	policy, err := LoadPolicyFile(target)
	require.NoError(t, err)
	require.Equal(t, 1, policy.Version)
	require.Len(t, policy.Rules, 1)
	require.True(t, strings.HasSuffix(policy.Source, "p.yaml"))
}

type sinkFn func(AuditEvent) error

func (s sinkFn) Write(e AuditEvent) error { return s(e) }

func TestAgeHours_UsesInjectedClock(t *testing.T) {
	prev := nowFn
	defer func() { nowFn = prev }()
	nowFn = func() time.Time { return time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC) }
	hrs := ageHours(Finding{LastTouched: time.Date(2026, 6, 14, 10, 0, 0, 0, time.UTC)})
	require.Equal(t, 2, hrs)
}
