package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/domain"
)

func TestNew_Defaults(t *testing.T) {
	p := New("spec-1", domain.ModeFeature)
	require.NotEmpty(t, p.ID)
	require.Equal(t, "spec-1", p.SpecID)
	require.Equal(t, domain.ModeFeature, p.Mode)
	require.Equal(t, "pending", p.ConfirmationStatus)
}

func TestWrite_ContainsAllSections(t *testing.T) {
	root := t.TempDir()
	p := New("spec-2", domain.ModeBugfix)
	p.Summary = "fix the bug"
	p.DetectedStack = []string{"go", "rails"}
	p.SensorsToRun = []string{"forbidden_files"}
	p.Risks = []string{"breaking change"}
	p.AgentChain = []string{"codex", "claude"}
	out, err := p.Write(root)
	require.NoError(t, err)
	b, err := os.ReadFile(out)
	require.NoError(t, err)
	body := string(b)
	for _, want := range []string{
		"# Plan", "## 1. Summary", "## 3. Detected stack",
		"## 6. Files likely to change", "## 8. Sensors to run",
		"## 10. Risks", "## 14. Agent chain",
		"fix the bug", "go, rails", "forbidden_files",
		"breaking change", "codex → claude",
	} {
		require.Contains(t, body, want)
	}
}

func TestLatestPlanPath(t *testing.T) {
	root := t.TempDir()
	first := New("s", domain.ModeFeature)
	p1, err := first.Write(root)
	require.NoError(t, err)
	got, err := LatestPlanPath(root)
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(got, filepath.Base(p1)))
}

func TestLatestPlanPath_EmptyDir(t *testing.T) {
	root := t.TempDir()
	_, err := LatestPlanPath(root)
	require.Error(t, err)
}
