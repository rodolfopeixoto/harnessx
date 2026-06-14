// SPDX-License-Identifier: MIT

package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func TestResolve_FlagWins(t *testing.T) {
	t.Setenv(constants.EnvProjectOverride, "ignored")
	r := openTemp(t)
	dir := t.TempDir()
	p, _ := r.Add(context.Background(), dir, "", "byflag")
	res, err := Resolve(context.Background(), r, ResolveOptions{Flag: "byflag"})
	require.NoError(t, err)
	require.Equal(t, SourceFlag, res.Source)
	require.Equal(t, p.ID, res.Project.ID)
}

func TestResolve_EnvWhenNoFlag(t *testing.T) {
	r := openTemp(t)
	dir := t.TempDir()
	p, _ := r.Add(context.Background(), dir, "", "byenv")
	t.Setenv(constants.EnvProjectOverride, "byenv")
	res, err := Resolve(context.Background(), r, ResolveOptions{})
	require.NoError(t, err)
	require.Equal(t, SourceEnv, res.Source)
	require.Equal(t, p.ID, res.Project.ID)
}

func TestResolve_CWDWalkUp(t *testing.T) {
	r := openTemp(t)
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0o755))
	sub := filepath.Join(dir, "nested", "deep")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	t.Setenv(constants.EnvProjectOverride, "")
	res, err := Resolve(context.Background(), r, ResolveOptions{CWD: sub})
	require.NoError(t, err)
	require.Equal(t, SourceCWD, res.Source)
	require.Equal(t, dir, res.Project.RootPath)
}

func TestResolve_FallsBackToActive(t *testing.T) {
	r := openTemp(t)
	dir := t.TempDir()
	p, _ := r.Add(context.Background(), dir, "", "")
	_, err := r.SetActive(context.Background(), p.ID)
	require.NoError(t, err)
	scratch := t.TempDir()
	t.Setenv(constants.EnvProjectOverride, "")
	res, err := Resolve(context.Background(), r, ResolveOptions{CWD: scratch})
	require.NoError(t, err)
	require.Equal(t, SourceActive, res.Source)
	require.Equal(t, p.ID, res.Project.ID)
}

func TestResolve_NoProjectReturnsError(t *testing.T) {
	r := openTemp(t)
	t.Setenv(constants.EnvProjectOverride, "")
	_, err := Resolve(context.Background(), r, ResolveOptions{CWD: t.TempDir()})
	require.ErrorIs(t, err, ErrNoProject)
}

func TestResolve_NilRegistryStillResolvesCWD(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".harness"), 0o755))
	t.Setenv(constants.EnvProjectOverride, "")
	res, err := Resolve(context.Background(), nil, ResolveOptions{CWD: dir})
	require.NoError(t, err)
	require.Equal(t, SourceCWD, res.Source)
	require.Equal(t, dir, res.Project.RootPath)
}

func TestResolve_AutoTouchUpdatesLastSeen(t *testing.T) {
	r := openTemp(t)
	dir := t.TempDir()
	p, _ := r.Add(context.Background(), dir, "", "tap")
	_, err := Resolve(context.Background(), r, ResolveOptions{Flag: "tap", AutoTouch: true})
	require.NoError(t, err)
	fetched, _ := r.Resolve(context.Background(), p.ID)
	require.NotNil(t, fetched.LastSeenAt)
}

func TestResolve_FlagUnknownErrors(t *testing.T) {
	r := openTemp(t)
	_, err := Resolve(context.Background(), r, ResolveOptions{Flag: "no-such"})
	require.Error(t, err)
}
