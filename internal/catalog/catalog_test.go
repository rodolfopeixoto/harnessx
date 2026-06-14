// SPDX-License-Identifier: MIT

package catalog

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/domain"
)

type fakeKind struct {
	kind  domain.CapabilityKind
	caps  []domain.Capability
	ops   []domain.FileOp
	plErr error
	dErr  error
}

func (f fakeKind) Kind() domain.CapabilityKind { return f.kind }
func (f fakeKind) Discover(_ context.Context, _ string) ([]domain.Capability, error) {
	return f.caps, f.dErr
}
func (f fakeKind) Plan(_ context.Context, _, _ string) ([]domain.FileOp, error) {
	return f.ops, f.plErr
}

func TestCatalog_RegisterAndDiscover(t *testing.T) {
	c := New()
	c.Register(fakeKind{kind: domain.KindMCP, caps: []domain.Capability{{Kind: domain.KindMCP, Name: "z"}, {Kind: domain.KindMCP, Name: "a"}}})
	c.Register(fakeKind{kind: domain.KindHook, caps: []domain.Capability{{Kind: domain.KindHook, Name: "m"}}})
	got, err := c.Discover(context.Background(), "/x")
	require.NoError(t, err)
	require.Len(t, got, 3)
	// sorted by (kind, name)
	require.Equal(t, "m", got[0].Name)
	require.Equal(t, "a", got[1].Name)
	require.Equal(t, "z", got[2].Name)
	require.ElementsMatch(t, []domain.CapabilityKind{domain.KindMCP, domain.KindHook}, c.Kinds())
}

func TestCatalog_DiscoverPropagatesError(t *testing.T) {
	c := New()
	c.Register(fakeKind{kind: domain.KindMCP, dErr: errors.New("boom")})
	_, err := c.Discover(context.Background(), "/x")
	require.Error(t, err)
}

func TestCatalog_DiscoverKindUnknown(t *testing.T) {
	c := New()
	_, err := c.DiscoverKind(context.Background(), "/x", domain.KindAgent)
	require.Error(t, err)
}

func TestCatalog_ShowFound(t *testing.T) {
	c := New()
	c.Register(fakeKind{kind: domain.KindMCP, caps: []domain.Capability{{Kind: domain.KindMCP, Name: "fs"}}})
	got, err := c.Show(context.Background(), "/x", domain.KindMCP, "fs")
	require.NoError(t, err)
	require.Equal(t, "fs", got.Name)
}

func TestCatalog_ShowMissing(t *testing.T) {
	c := New()
	c.Register(fakeKind{kind: domain.KindMCP, caps: []domain.Capability{{Kind: domain.KindMCP, Name: "fs"}}})
	_, err := c.Show(context.Background(), "/x", domain.KindMCP, "nope")
	require.ErrorIs(t, err, ErrUnknownCapability)
}

func TestCatalog_Plan(t *testing.T) {
	c := New()
	want := []domain.FileOp{{Op: domain.FileCreate, Path: "/x/foo", Body: []byte("hi")}}
	c.Register(fakeKind{kind: domain.KindMCP, ops: want})
	got, err := c.Plan(context.Background(), "/x", domain.KindMCP, "any")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestApply_CreateOverwriteAppendDeleteMkdir(t *testing.T) {
	root := t.TempDir()
	ops := []domain.FileOp{
		{Op: domain.FileMkdir, Path: filepath.Join(root, "dir")},
		{Op: domain.FileCreate, Path: filepath.Join(root, "dir", "a.txt"), Body: []byte("hello")},
		{Op: domain.FileAppend, Path: filepath.Join(root, "dir", "a.txt"), Body: []byte(" world")},
		{Op: domain.FileOverwrite, Path: filepath.Join(root, "dir", "b.txt"), Body: []byte("bye")},
	}
	res, err := Apply(context.Background(), root, ops)
	require.NoError(t, err)
	require.Len(t, res.Written, 4)
	body, _ := os.ReadFile(filepath.Join(root, "dir", "a.txt"))
	require.Equal(t, "hello world", string(body))

	del := []domain.FileOp{{Op: domain.FileDelete, Path: filepath.Join(root, "dir", "b.txt")}}
	res, err = Apply(context.Background(), root, del)
	require.NoError(t, err)
	require.Len(t, res.Deleted, 1)
}

func TestApply_RejectsPathOutsideRoot(t *testing.T) {
	root := t.TempDir()
	ops := []domain.FileOp{{Op: domain.FileCreate, Path: "/etc/evil", Body: []byte("x")}}
	_, err := Apply(context.Background(), root, ops)
	require.Error(t, err)
}

func TestApply_CreateRefusesOverwrite(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "x.txt")
	require.NoError(t, os.WriteFile(dest, []byte("old"), 0o644))
	_, err := Apply(context.Background(), root, []domain.FileOp{{Op: domain.FileCreate, Path: dest, Body: []byte("new")}})
	require.Error(t, err)
}

func TestApply_UnsupportedOp(t *testing.T) {
	root := t.TempDir()
	_, err := Apply(context.Background(), root, []domain.FileOp{{Op: "weird", Path: filepath.Join(root, "x")}})
	require.Error(t, err)
}

func TestApply_NoOpsReturnsEmpty(t *testing.T) {
	res, err := Apply(context.Background(), t.TempDir(), nil)
	require.NoError(t, err)
	require.Empty(t, res.Written)
}

func TestHashOps_Stable(t *testing.T) {
	a := HashOps([]domain.FileOp{{Op: domain.FileCreate, Path: "/x", Body: []byte("a")}})
	b := HashOps([]domain.FileOp{{Op: domain.FileCreate, Path: "/x", Body: []byte("a")}})
	require.Equal(t, a, b)
	c := HashOps([]domain.FileOp{{Op: domain.FileCreate, Path: "/x", Body: []byte("b")}})
	require.NotEqual(t, a, c)
}

func TestDiscoverByGlobs_ParsesManifests(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "templates", "capabilities", "mcp", "demo")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	body := []byte("kind: mcp\nname: demo\nversion: 0.1.0\ndescription: hi\nbody: |\n  ok\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "manifest.yaml"), body, 0o644))
	caps, err := DiscoverByGlobs(root, domain.KindMCP, []string{"templates/capabilities/mcp/*/manifest.yaml"})
	require.NoError(t, err)
	require.Len(t, caps, 1)
	require.Equal(t, "demo", caps[0].Name)
	require.Equal(t, "bundled", caps[0].Source)
}

func TestDiscoverByGlobs_SkipsMalformed(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "templates", "capabilities", "mcp", "bad"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "templates", "capabilities", "mcp", "bad", "manifest.yaml"), []byte("not: [valid"), 0o644))
	caps, err := DiscoverByGlobs(root, domain.KindMCP, []string{"templates/capabilities/mcp/*/manifest.yaml"})
	require.NoError(t, err)
	require.Empty(t, caps)
}

func TestSafeJoin_RelativeBecomesRooted(t *testing.T) {
	root := t.TempDir()
	got, err := safeJoin(root, "sub/file")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(root, "sub/file"), got)
}
