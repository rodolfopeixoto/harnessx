package router

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Missing(t *testing.T) {
	got, err := LoadConfig(filepath.Join(t.TempDir(), "absent.yaml"))
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestLoadConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "routes.yaml")
	body := `routes:
  implementation:
    primary: codex
    fallback: [claude, gemini]
    budget_usd: 1.5
  security_review:
    primary: claude
    fallback: [kimi]
    budget_usd: 0.5
`
	require.NoError(t, os.WriteFile(p, []byte(body), 0o644))
	got, err := LoadConfig(p)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "codex", got["implementation"].Primary)
	require.Equal(t, []string{"claude", "gemini"}, got["implementation"].Fallback)
	require.InDelta(t, 1.5, got["implementation"].BudgetUSD, 0.0001)
}

func TestLoadConfig_BadYAML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "routes.yaml")
	require.NoError(t, os.WriteFile(p, []byte(":::not yaml"), 0o644))
	_, err := LoadConfig(p)
	require.Error(t, err)
}
