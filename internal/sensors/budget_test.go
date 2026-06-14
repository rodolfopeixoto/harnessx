package sensors

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeBudget(t *testing.T, root string, budgets map[string]any) {
	t.Helper()
	dir := filepath.Join(root, ".harness", "project")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	body := map[string]any{
		"generated_at": "2026-06-14T00:00:00Z",
		"budgets":      budgets,
	}
	b, _ := json.MarshalIndent(body, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "performance-budget.json"), b, 0o644))
}

func writeSnap(t *testing.T, root string, snap map[string]any) {
	t.Helper()
	dir := filepath.Join(root, ".harness", "artifacts", "perf")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	b, _ := json.MarshalIndent(snap, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "snap.json"), b, 0o644))
}

func TestPerformanceBudget_NoBudget_Skips(t *testing.T) {
	root := t.TempDir()
	res := PerformanceBudgetSensor{}.Run(RunCtx{Ctx: context.Background(), Root: root})
	require.Equal(t, StatusSkipped, res.Status)
}

func TestPerformanceBudget_NoSnapshot_Skips(t *testing.T) {
	root := t.TempDir()
	writeBudget(t, root, map[string]any{"deps_total": 100})
	res := PerformanceBudgetSensor{}.Run(RunCtx{Ctx: context.Background(), Root: root})
	require.Equal(t, StatusSkipped, res.Status)
}

func TestPerformanceBudget_WithinBudget_Passes(t *testing.T) {
	root := t.TempDir()
	writeBudget(t, root, map[string]any{"deps_total": 100, "harness_dir_bytes": 100000})
	writeSnap(t, root, map[string]any{
		"deps": map[string]any{"total": 10},
		"disk": map[string]any{"harness_dir_bytes": 12345},
	})
	res := PerformanceBudgetSensor{}.Run(RunCtx{Ctx: context.Background(), Root: root})
	require.Equal(t, StatusPassed, res.Status)
}

func TestPerformanceBudget_Exceeds_Fails(t *testing.T) {
	root := t.TempDir()
	writeBudget(t, root, map[string]any{"deps_total": 5})
	writeSnap(t, root, map[string]any{
		"deps": map[string]any{"total": 10},
	})
	res := PerformanceBudgetSensor{}.Run(RunCtx{
		Ctx: context.Background(), Root: root,
		OutputDir: filepath.Join(root, ".harness", "artifacts", "sensors"),
	})
	require.Equal(t, StatusFailed, res.Status)
	require.Contains(t, res.Detail, "deps_total")
}
