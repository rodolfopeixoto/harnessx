package memory

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
)

type sqlAdapter struct{ db *sql.DB }

func (a sqlAdapter) ExecContext(ctx context.Context, q string, args ...any) (any, error) {
	return a.db.ExecContext(ctx, q, args...)
}

func TestPromote_Success(t *testing.T) {
	repo, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	defer repo.Close()

	adapter := sqlAdapter{db: repo.DB()}
	m, err := Promote(context.Background(), repo, Candidate{
		Scope: "project", Kind: "convention",
		Content:       "tests use rspec, not minitest",
		EvidenceRunID: "run-1", Confidence: 0.8,
	}, adapter)
	require.NoError(t, err)
	require.NotEmpty(t, m.ID)
}

func TestPromote_MissingEvidence(t *testing.T) {
	repo, _ := sqlite.Open(":memory:")
	defer repo.Close()
	_, err := Promote(context.Background(), repo, Candidate{
		Scope: "p", Kind: "k", Content: "x", Confidence: 0.9,
	}, sqlAdapter{db: repo.DB()})
	require.True(t, errors.Is(err, ErrMissingEvidence))
}

func TestPromote_LowConfidence(t *testing.T) {
	repo, _ := sqlite.Open(":memory:")
	defer repo.Close()
	_, err := Promote(context.Background(), repo, Candidate{
		Scope: "p", Kind: "k", Content: "x", EvidenceRunID: "run-1", Confidence: 0.2,
	}, sqlAdapter{db: repo.DB()})
	require.True(t, errors.Is(err, ErrLowConfidence))
}

func TestPromote_RejectsSensitive(t *testing.T) {
	repo, _ := sqlite.Open(":memory:")
	defer repo.Close()
	_, err := Promote(context.Background(), repo, Candidate{
		Scope: "p", Kind: "k",
		Content:       "found AKIAIOSFODNN7EXAMPLE in the codebase",
		EvidenceRunID: "run-1", Confidence: 0.9,
	}, sqlAdapter{db: repo.DB()})
	require.True(t, errors.Is(err, ErrSensitive))
}
