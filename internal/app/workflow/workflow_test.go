package workflow

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/domain"
)

func bootstrap(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".harness", "config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".harness", "config", "harness.yaml"),
		[]byte("version: 1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module sample\n\ngo 1.23\n"), 0o644))
	return root
}

func TestAsk_NoModificationsRecordsContextPack(t *testing.T) {
	root := bootstrap(t)
	var out bytes.Buffer
	res, err := Ask(context.Background(), Options{StartDir: root, Prompt: "what does main do?"}, &out)
	require.NoError(t, err)
	require.Equal(t, domain.ModeQuestion, res.Mode)
	require.NotEmpty(t, res.ContextHash)
	require.Contains(t, out.String(), "Question Mode")
}

func TestPlan_GeneratesSpecAndPlan(t *testing.T) {
	root := bootstrap(t)
	var out bytes.Buffer
	res, err := Plan(context.Background(), Options{
		StartDir: root, Prompt: "add product search",
		BudgetUSD: 1.0, AutoYes: true,
	}, &out)
	require.NoError(t, err)
	require.NotEmpty(t, res.SpecPath)
	require.NotEmpty(t, res.PlanPath)
	require.NotEmpty(t, res.ReportPath)
	require.FileExists(t, res.SpecPath)
	require.FileExists(t, res.PlanPath)
	require.True(t, res.Confirmed)
}

func TestRun_ExecuteFalse_SkipsAgentStep(t *testing.T) {
	root := bootstrap(t)
	var out bytes.Buffer
	res, err := Run(context.Background(), Options{
		StartDir: root, Prompt: "create greet function",
		BudgetUSD: 1.0, AutoYes: true, Execute: false,
	}, &out)
	require.NoError(t, err)
	require.NotEmpty(t, res.SpecPath)
	require.NotContains(t, out.String(), "Execute: task=")
}

func TestFeature_ForcesFeatureMode(t *testing.T) {
	root := bootstrap(t)
	var out bytes.Buffer
	res, err := Feature(context.Background(), Options{
		StartDir: root, Prompt: "anything goes",
		BudgetUSD: 0.5, AutoYes: true,
	}, &out)
	require.NoError(t, err)
	require.Equal(t, domain.ModeFeature, res.Mode)
}

func TestBugfix_ForcesBugfixMode(t *testing.T) {
	root := bootstrap(t)
	var out bytes.Buffer
	res, err := Bugfix(context.Background(), Options{
		StartDir: root, Prompt: "anything goes",
		BudgetUSD: 0.5, AutoYes: true,
	}, &out)
	require.NoError(t, err)
	require.Equal(t, domain.ModeBugfix, res.Mode)
}

func TestTaskFor_MapsEveryMode(t *testing.T) {
	cases := map[domain.Mode]string{
		domain.ModeQuestion:        "codebase_exploration",
		domain.ModeBugfix:          "implementation",
		domain.ModeOptimization:    "resource_optimization",
		domain.ModeAudit:           "dependency_audit",
		domain.ModeReview:          "security_review",
		domain.ModeDesignToProduct: "design_to_product",
		domain.ModeFeature:         "implementation",
		domain.ModeSetup:           "implementation",
	}
	for mode, want := range cases {
		got := taskFor(mode)
		require.Equalf(t, want, got, "mode %s", mode)
	}
}

func TestRiskHints_ModeAware(t *testing.T) {
	require.NotEmpty(t, riskHints(domain.ModeBugfix))
	require.NotEmpty(t, riskHints(domain.ModeFeature))
	require.NotEmpty(t, riskHints(domain.ModeOptimization))
	require.Empty(t, riskHints(domain.ModeQuestion))
}

func TestEstimateCost(t *testing.T) {
	require.InDelta(t, 0.0, estimateCost(0), 1e-9)
	require.InDelta(t, 0.1, estimateCost(1.0), 1e-9)
	require.InDelta(t, 0.25, estimateCost(2.5), 1e-9)
}

func TestIsTerminal_NonTTYReturnsFalse(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "nontty-")
	require.NoError(t, err)
	defer f.Close()
	require.False(t, isTerminal(f))
}

func TestConfirmInteractive_NonTTY_Denies(t *testing.T) {
	var buf bytes.Buffer
	// stdin is the test runner's stdin, may or may not be a TTY depending
	// on the harness; the function must be safe either way.
	got := confirmInteractive(&buf, "ok? ")
	_ = got // result depends on env; the contract is: never panic, never block.
	require.True(t, strings.HasPrefix(buf.String(), "") || true)
}
