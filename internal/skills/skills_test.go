package skills

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
)

func TestPromote_FirstVersionAccepted(t *testing.T) {
	root := t.TempDir()
	repo, err := sqlite.Open(filepath.Join(root, "db.sqlite"))
	require.NoError(t, err)
	defer repo.Close()
	s, err := Promote(context.Background(), repo.DB(), root, "tdd-cycle",
		"step 1\nstep 2\nstep 3\n", nil)
	require.NoError(t, err)
	require.Equal(t, 1, s.Version)
	require.True(t, s.Accepted)
}

func TestPromote_NoImprovementRejected(t *testing.T) {
	root := t.TempDir()
	repo, err := sqlite.Open(filepath.Join(root, "db.sqlite"))
	require.NoError(t, err)
	defer repo.Close()
	_, err = Promote(context.Background(), repo.DB(), root, "x", "a\nb\nc\n", nil)
	require.NoError(t, err)
	_, err = Promote(context.Background(), repo.DB(), root, "x", "a\nb\nc\n", nil)
	require.True(t, errors.Is(err, ErrNoImprovement))
}

func TestPromote_ImprovementAccepted(t *testing.T) {
	root := t.TempDir()
	repo, err := sqlite.Open(filepath.Join(root, "db.sqlite"))
	require.NoError(t, err)
	defer repo.Close()
	_, err = Promote(context.Background(), repo.DB(), root, "x", "one\ntwo\n", nil)
	require.NoError(t, err)
	s, err := Promote(context.Background(), repo.DB(), root, "x", "one\ntwo\nthree\nfour\n", nil)
	require.NoError(t, err)
	require.Equal(t, 2, s.Version)
}

func TestList(t *testing.T) {
	root := t.TempDir()
	repo, err := sqlite.Open(filepath.Join(root, "db.sqlite"))
	require.NoError(t, err)
	defer repo.Close()
	_, _ = Promote(context.Background(), repo.DB(), root, "a", "x\n", nil)
	_, _ = Promote(context.Background(), repo.DB(), root, "b", "y\nz\n", nil)
	list, err := List(context.Background(), repo.DB())
	require.NoError(t, err)
	require.Len(t, list, 2)
}
