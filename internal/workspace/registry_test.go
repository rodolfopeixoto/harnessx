// SPDX-License-Identifier: MIT

package workspace

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func openTemp(t *testing.T) *Registry {
	t.Helper()
	r, err := Open(filepath.Join(t.TempDir(), "registry.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close() })
	return r
}

func TestOpen_AppliesMigrations(t *testing.T) {
	r := openTemp(t)
	for _, table := range []string{"projects", "active_project", "registry_meta"} {
		_, err := r.DB().Exec("select count(*) from " + table)
		require.NoErrorf(t, err, "table %s missing", table)
	}
}

func TestAdd_NewProject(t *testing.T) {
	r := openTemp(t)
	dir := t.TempDir()
	p, err := r.Add(context.Background(), dir, "", "")
	require.NoError(t, err)
	require.NotEmpty(t, p.ID)
	require.Equal(t, Slugify(filepath.Base(dir)), p.Slug)
	require.Equal(t, dir, p.RootPath)
	require.Contains(t, p.DBPath, ".harness")
}

func TestAdd_Idempotent(t *testing.T) {
	r := openTemp(t)
	dir := t.TempDir()
	first, err := r.Add(context.Background(), dir, "First", "")
	require.NoError(t, err)
	second, err := r.Add(context.Background(), dir, "Second", "")
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)
	require.Equal(t, "Second", second.DisplayName)
}

func TestAdd_SlugCollisionFails(t *testing.T) {
	r := openTemp(t)
	a := t.TempDir()
	b := t.TempDir()
	_, err := r.Add(context.Background(), a, "", "shared-slug")
	require.NoError(t, err)
	_, err = r.Add(context.Background(), b, "", "shared-slug")
	require.Error(t, err)
}

func TestList_ExcludesArchivedByDefault(t *testing.T) {
	r := openTemp(t)
	d1, d2 := t.TempDir(), t.TempDir()
	p1, _ := r.Add(context.Background(), d1, "", "")
	_, _ = r.Add(context.Background(), d2, "", "")
	_, err := r.Archive(context.Background(), p1.ID)
	require.NoError(t, err)

	active, err := r.List(context.Background(), false)
	require.NoError(t, err)
	require.Len(t, active, 1)
	require.Nil(t, active[0].ArchivedAt)

	all, err := r.List(context.Background(), true)
	require.NoError(t, err)
	require.Len(t, all, 2)
}

func TestResolve_ByIDSlugAndPath(t *testing.T) {
	r := openTemp(t)
	dir := t.TempDir()
	p, _ := r.Add(context.Background(), dir, "", "myproj")

	byID, err := r.Resolve(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, p.ID, byID.ID)

	bySlug, err := r.Resolve(context.Background(), "myproj")
	require.NoError(t, err)
	require.Equal(t, p.ID, bySlug.ID)

	byPath, err := r.Resolve(context.Background(), dir)
	require.NoError(t, err)
	require.Equal(t, p.ID, byPath.ID)

	_, err = r.Resolve(context.Background(), "no-such")
	require.ErrorIs(t, err, ErrNotFound)

	_, err = r.Resolve(context.Background(), "")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestActive_SetGetClear(t *testing.T) {
	r := openTemp(t)
	dir := t.TempDir()
	p, _ := r.Add(context.Background(), dir, "", "")

	_, err := r.Active(context.Background())
	require.ErrorIs(t, err, ErrNotFound)

	_, err = r.SetActive(context.Background(), p.ID)
	require.NoError(t, err)
	active, err := r.Active(context.Background())
	require.NoError(t, err)
	require.Equal(t, p.ID, active.ID)

	_, err = r.SetActive(context.Background(), "")
	require.NoError(t, err)
	_, err = r.Active(context.Background())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestArchive_RoundTrip(t *testing.T) {
	r := openTemp(t)
	p, _ := r.Add(context.Background(), t.TempDir(), "", "")
	arch, err := r.Archive(context.Background(), p.ID)
	require.NoError(t, err)
	require.NotNil(t, arch.ArchivedAt)
	un, err := r.Unarchive(context.Background(), p.ID)
	require.NoError(t, err)
	require.Nil(t, un.ArchivedAt)
}

func TestForget_RemovesRow(t *testing.T) {
	r := openTemp(t)
	dir := t.TempDir()
	p, _ := r.Add(context.Background(), dir, "", "")
	require.NoError(t, r.Forget(context.Background(), p.ID))
	_, err := r.Resolve(context.Background(), p.ID)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestTouch_UpdatesLastSeen(t *testing.T) {
	r := openTemp(t)
	p, _ := r.Add(context.Background(), t.TempDir(), "", "")
	require.NoError(t, r.Touch(context.Background(), p.ID))
	resolved, err := r.Resolve(context.Background(), p.ID)
	require.NoError(t, err)
	require.NotNil(t, resolved.LastSeenAt)
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Aurora Commerce": "aurora-commerce",
		"~/dev/foo":       "dev-foo",
		"  -- ":           "project",
		"already-ok":      "already-ok",
		"":                "project",
		"UPPER_CASE.1":    "upper-case-1",
	}
	for in, want := range cases {
		require.Equal(t, want, Slugify(in), in)
	}
}

func TestConcurrent_AddSameRoot(t *testing.T) {
	r := openTemp(t)
	dir := t.TempDir()
	var wg sync.WaitGroup
	var errs []error
	var mu sync.Mutex
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := r.Add(context.Background(), dir, "", ""); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	require.Empty(t, errs)
	all, err := r.List(context.Background(), true)
	require.NoError(t, err)
	require.Len(t, all, 1, "concurrent adds must not duplicate")
}
