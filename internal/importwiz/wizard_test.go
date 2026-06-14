// SPDX-License-Identifier: MIT

package importwiz

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/workspace"
)

func newRegistry(t *testing.T) *workspace.Registry {
	t.Helper()
	r, err := workspace.Open(filepath.Join(t.TempDir(), "reg.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close() })
	return r
}

func TestPlan_ProducesAllSteps(t *testing.T) {
	steps := Plan(Options{Path: "/tmp/x"})
	require.Len(t, steps, 5)
	ids := []string{StepDetect, StepStack, StepRegister, StepIndex, StepDone}
	for i, want := range ids {
		require.Equal(t, want, steps[i].ID)
		require.Equal(t, StatusPending, steps[i].Status)
	}
}

func TestRun_RegistersAndFingerprints(t *testing.T) {
	reg := newRegistry(t)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644))
	res, err := Run(context.Background(), reg, Options{Path: dir, DisplayName: "Demo"})
	require.NoError(t, err)
	require.NotEmpty(t, res.Project.ID)
	require.Contains(t, res.Stack, "node")
	for _, s := range res.Steps {
		require.Equal(t, StatusOK, s.Status, s.ID)
	}
}

func TestRun_MissingFolderFails(t *testing.T) {
	reg := newRegistry(t)
	_, err := Run(context.Background(), reg, Options{Path: "/tmp/this-folder-does-not-exist-xyz"})
	require.Error(t, err)
}

func TestRun_EmptyPathRejected(t *testing.T) {
	reg := newRegistry(t)
	_, err := Run(context.Background(), reg, Options{Path: ""})
	require.Error(t, err)
}

func TestDetectStack_MultipleMarkers(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch"), 0o644))
	stack := DetectStack(dir)
	require.Contains(t, stack, "go")
	require.Contains(t, stack, "container")
}

func TestDetectStack_DefaultsToFallback(t *testing.T) {
	dir := t.TempDir()
	stack := DetectStack(dir)
	require.NotEmpty(t, stack)
}
