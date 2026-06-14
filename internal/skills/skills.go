// SPDX-License-Identifier: MIT

// Package skills implements spec §27 — versioned playbooks gated on
// benchmark improvement. A skill is a markdown body stored on disk; a
// hash + score row in `skill_versions` tracks every version. Promotion
// only succeeds when the new version's benchmark beats the prior best.
package skills

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"crypto/sha256"

	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

var (
	ErrNoImprovement = errors.New("skills: new version did not improve over previous best")
)

// Skill is the in-memory representation of a versioned playbook.
type Skill struct {
	Name        string
	Version     int
	Content     string
	ContentHash string
	Score       float64
	Accepted    bool
	CreatedAt   time.Time
}

// dir returns .harness/skills under root.
func dir(root string) string { return filepath.Join(paths.HarnessDir(root), "skills") }

// Promote validates the new version against the current best and writes
// both the markdown file + a `skill_versions` row when accepted.
//
// scoreFn computes a benchmark score for the new content; higher is better.
// Pass a deterministic stub in tests.
func Promote(ctx context.Context, db *sql.DB, root, name, content string, scoreFn func(string) float64) (Skill, error) {
	if name == "" || content == "" {
		return Skill{}, errors.New("skills: missing name or content")
	}
	if scoreFn == nil {
		scoreFn = staticScore
	}
	score := scoreFn(content)
	best, err := BestScore(ctx, db, name)
	if err != nil {
		return Skill{}, err
	}
	if score <= best {
		return Skill{}, fmt.Errorf("%w (new=%.3f best=%.3f)", ErrNoImprovement, score, best)
	}

	version, err := nextVersion(ctx, db, name)
	if err != nil {
		return Skill{}, err
	}

	hash := contentHash(content)
	if err := os.MkdirAll(dir(root), 0o755); err != nil {
		return Skill{}, err
	}
	path := filepath.Join(dir(root), fmt.Sprintf("%s.v%d.md", name, version))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return Skill{}, err
	}
	now := time.Now().UTC()
	if _, err := db.ExecContext(ctx, `
		insert into skill_versions (id, skill_name, version, content_hash, score, accepted, created_at)
		values (?, ?, ?, ?, ?, ?, ?)`,
		ids.New(), name, version, hash, score, 1, now.Format(time.RFC3339Nano),
	); err != nil {
		return Skill{}, err
	}
	return Skill{
		Name: name, Version: version, Content: content, ContentHash: hash,
		Score: score, Accepted: true, CreatedAt: now,
	}, nil
}

// BestScore returns the highest accepted score for a skill, or 0 when
// the skill has no versions yet.
func BestScore(ctx context.Context, db *sql.DB, name string) (float64, error) {
	var v sql.NullFloat64
	err := db.QueryRowContext(ctx,
		`select max(score) from skill_versions where skill_name = ? and accepted = 1`,
		name).Scan(&v)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if !v.Valid {
		return 0, nil
	}
	return v.Float64, nil
}

func nextVersion(ctx context.Context, db *sql.DB, name string) (int, error) {
	var v sql.NullInt64
	err := db.QueryRowContext(ctx,
		`select max(version) from skill_versions where skill_name = ?`, name).Scan(&v)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if !v.Valid {
		return 1, nil
	}
	return int(v.Int64) + 1, nil
}

// List returns all known skill versions, newest first.
func List(ctx context.Context, db *sql.DB) ([]Skill, error) {
	rows, err := db.QueryContext(ctx, `
		select skill_name, version, content_hash, coalesce(score, 0), accepted, created_at
		from skill_versions order by created_at desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Skill
	for rows.Next() {
		var s Skill
		var acc int
		var created string
		if err := rows.Scan(&s.Name, &s.Version, &s.ContentHash, &s.Score, &acc, &created); err != nil {
			return nil, err
		}
		s.Accepted = acc == 1
		s.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		out = append(out, s)
	}
	return out, rows.Err()
}

func contentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// staticScore is a deliberately weak default. Real workflows pass a
// benchmark harness via scoreFn. Counts non-comment, non-blank lines as
// a proxy for "specificity" — better than nothing, terrible at scale.
func staticScore(content string) float64 {
	n := 0
	for _, line := range splitLines(content) {
		trim := trimSpace(line)
		if trim == "" || trim[0] == '#' {
			continue
		}
		n++
	}
	return float64(n)
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
