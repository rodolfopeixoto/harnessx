// SPDX-License-Identifier: MIT

package workspace

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

// scan.go isolates the row-scanning + single-column lookup helpers used
// by registry queries. Kept sibling to registry.go so the public
// Registry API stays unchanged.

func (r *Registry) byID(ctx context.Context, id string) (Project, error) {
	return r.scanOne(ctx, `select id, slug, display_name, root_path, db_path, added_at, last_seen_at, archived_at, schema_version
		from projects where id = ?`, id)
}

func (r *Registry) bySlug(ctx context.Context, slug string) (Project, error) {
	return r.scanOne(ctx, `select id, slug, display_name, root_path, db_path, added_at, last_seen_at, archived_at, schema_version
		from projects where slug = ?`, slug)
}

func (r *Registry) byRoot(ctx context.Context, root string) (Project, error) {
	return r.scanOne(ctx, `select id, slug, display_name, root_path, db_path, added_at, last_seen_at, archived_at, schema_version
		from projects where root_path = ?`, root)
}

func (r *Registry) scanOne(ctx context.Context, q string, args ...any) (Project, error) {
	row := r.db.QueryRowContext(ctx, q, args...)
	p, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	return p, err
}

type rowScanner interface {
	Scan(...any) error
}

func scanProject(s rowScanner) (Project, error) {
	var (
		p              Project
		addedAt        string
		lastSeen, arch sql.NullString
	)
	if err := s.Scan(&p.ID, &p.Slug, &p.DisplayName, &p.RootPath, &p.DBPath,
		&addedAt, &lastSeen, &arch, &p.SchemaVer); err != nil {
		return Project{}, err
	}
	p.AddedAt, _ = time.Parse(timeFmt, addedAt)
	if lastSeen.Valid {
		p.LastSeenAt = parseTimePtr(lastSeen.String)
	}
	if arch.Valid {
		p.ArchivedAt = parseTimePtr(arch.String)
	}
	return p, nil
}

func parseTimePtr(s string) *time.Time {
	t, err := time.Parse(timeFmt, s)
	if err != nil {
		return nil
	}
	return &t
}

func defaultProjectDBPath(root string) string {
	return filepath.Join(root, constants.HarnessDir, constants.DBSubdir, constants.DBFilename)
}
