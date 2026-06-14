package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/domain"
)

func TestSpec_Write_ContainsRequiredSections(t *testing.T) {
	root := t.TempDir()
	s := NewFromPrompt("add product search with filters", domain.ModeFeature)
	p, err := s.Write(root)
	require.NoError(t, err)
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	body := string(b)
	for _, header := range []string{
		"## Feature Name", "## User Problem", "## Expected Outcome",
		"## Scope", "## Out of Scope", "## Business Rules",
		"## UX Expectations", "## API Expectations", "## Data Model Expectations",
		"## Security Considerations", "## Performance Considerations",
		"## Observability Expectations", "## Test Plan", "## E2E Plan",
		"## Acceptance Criteria", "## Rollback Plan", "## Definition of Done",
	} {
		require.Contains(t, body, header)
	}
	require.Contains(t, body, "add product search with filters")
}

func TestLatestSpecPath(t *testing.T) {
	root := t.TempDir()
	s := NewFromPrompt("first", domain.ModeFeature)
	p, err := s.Write(root)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(p, filepath.Join(root, ".harness", "artifacts", "specs")))

	got, err := LatestSpecPath(root)
	require.NoError(t, err)
	require.Equal(t, p, got)
}
