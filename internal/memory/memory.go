// SPDX-License-Identifier: MIT

// Package memory implements evidence-gated project memory promotion
// (spec §11). A learning can be promoted only when it is backed by a run id,
// has evidence, is non-sensitive, improves future execution, and has a
// confidence score.
package memory

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
)

var (
	ErrMissingEvidence = errors.New("memory: missing run id / evidence")
	ErrSensitive       = errors.New("memory: content contains forbidden tokens")
	ErrLowConfidence   = errors.New("memory: confidence below floor")
	ErrUnknownKind     = errors.New("memory: unknown kind (paper §3.2)")
)

// Kinds taxonomy mirrors paper "Code as Agent Harness" §3.2.1–§3.2.5.
// Unknown kinds are rejected at Promote time to keep the taxonomy
// observable for downstream context-engineering decisions.
const (
	KindWorking      = "working"
	KindSemantic     = "semantic"
	KindExperiential = "experiential"
	KindLongTerm     = "long_term"
	KindMultiAgent   = "multi_agent"
)

func KnownKinds() []string {
	return []string{KindWorking, KindSemantic, KindExperiential, KindLongTerm, KindMultiAgent}
}

func validKind(k string) bool {
	for _, v := range KnownKinds() {
		if v == k {
			return true
		}
	}
	return false
}

const confidenceFloor = 0.4

// sensitiveRe blocks promotion of strings that look like secrets. It's
// the same vocabulary as the secrets sensor — kept in sync deliberately.
var sensitiveRe = regexp.MustCompile(`(?i)(AKIA[0-9A-Z]{16}|aws_secret_access_key|xox[abprs]-[A-Za-z0-9-]{20,}|-----BEGIN [A-Z ]*PRIVATE KEY-----|gh[opusr]_[A-Za-z0-9]{36,})`)

type Candidate struct {
	Scope         string
	Kind          string
	Content       string
	EvidenceRunID string
	Confidence    float64
}

// Promote runs the gate. On success it writes a memory row and returns it.
func Promote(ctx context.Context, repo *sqlite.Repo, c Candidate, db SQLExec) (domain.Memory, error) {
	if c.EvidenceRunID == "" {
		return domain.Memory{}, ErrMissingEvidence
	}
	if strings.HasPrefix(c.EvidenceRunID, "--") {
		return domain.Memory{}, fmt.Errorf("%w: run id looks like a flag (%q) — likely empty shell variable", ErrMissingEvidence, c.EvidenceRunID)
	}
	if c.Confidence < confidenceFloor {
		return domain.Memory{}, ErrLowConfidence
	}
	if sensitiveRe.MatchString(c.Content) {
		return domain.Memory{}, ErrSensitive
	}
	if strings.TrimSpace(c.Content) == "" {
		return domain.Memory{}, errors.New("memory: empty content")
	}
	if c.Kind == "" {
		c.Kind = KindSemantic
	}
	if !validKind(c.Kind) {
		return domain.Memory{}, fmt.Errorf("%w: got %q, want one of %v", ErrUnknownKind, c.Kind, KnownKinds())
	}

	now := time.Now().UTC()
	m := domain.Memory{
		ID: ids.New(), Scope: c.Scope, Kind: c.Kind, Content: c.Content,
		EvidenceRunID: c.EvidenceRunID, Confidence: c.Confidence,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := writeMemory(ctx, db, m); err != nil {
		return domain.Memory{}, err
	}
	return m, nil
}

// SQLExec is the minimal contract from sqlite.Repo we need; allows tests
// to inject a stub without depending on a real DB.
type SQLExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (any, error)
}

// writeMemory uses the underlying *sql.DB exposed by sqlite.Repo via the
// SQLExec adapter so this package doesn't import database/sql directly.
func writeMemory(ctx context.Context, db SQLExec, m domain.Memory) error {
	_, err := db.ExecContext(ctx, `
		insert into memories
		  (id, scope, kind, content, evidence_run_id, confidence, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.Scope, m.Kind, m.Content, m.EvidenceRunID, m.Confidence,
		m.CreatedAt.Format(time.RFC3339Nano), m.UpdatedAt.Format(time.RFC3339Nano))
	return err
}
