// SPDX-License-Identifier: MIT

// Package workspace owns the cross-project registry: which projects are
// known, where their per-project SQLite lives, and which one is active.
// Per-project data (sessions, runs, sensors, memory) stays in each
// project's own .harness/db/harness.sqlite.
package workspace

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // registers the "sqlite" driver via init()

	"github.com/ropeixoto/harnessx/internal/platform/constants"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const timeFmt = time.RFC3339Nano

// Project is the registry's view of a known project root.
type Project struct {
	ID          string
	Slug        string
	DisplayName string
	RootPath    string
	DBPath      string
	AddedAt     time.Time
	LastSeenAt  *time.Time
	ArchivedAt  *time.Time
	SchemaVer   int
}

// Registry is the typed handle for the workspace registry SQLite DB.
type Registry struct {
	db   *sql.DB
	path string
}

// Open opens (or creates) the registry at the given absolute path, applies
// migrations and returns a Registry. The empty path resolves to
// paths.GlobalRegistryPath().
func Open(path string) (*Registry, error) {
	if path == "" {
		path = paths.GlobalRegistryPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("workspace: mkdir registry dir: %w", err)
	}
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", url.PathEscape(path))
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("workspace: open: %w", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("workspace: ping: %w", err)
	}
	// Serialise writers — UPSERT + busy_timeout still leave a window where two
	// IMMEDIATE writers from the same process race to acquire the WAL lock.
	// Capping the write pool at 1 connection removes that window without
	// affecting read throughput.
	db.SetMaxOpenConns(1)
	r := &Registry{db: db, path: path}
	if err := r.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return r, nil
}

// Close releases the registry handle.
func (r *Registry) Close() error { return r.db.Close() }

// Path returns the on-disk path of the registry database.
func (r *Registry) Path() string { return r.path }

// DB exposes the underlying *sql.DB for advanced callers and tests.
func (r *Registry) DB() *sql.DB { return r.db }

func (r *Registry) migrate() error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("workspace: read migrations: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return fmt.Errorf("workspace: read %s: %w", e.Name(), err)
		}
		if _, err := r.db.Exec(string(b)); err != nil {
			return fmt.Errorf("workspace: apply %s: %w", e.Name(), err)
		}
	}
	return nil
}

// ErrNotFound is returned when a requested project does not exist.
var ErrNotFound = errors.New("workspace: project not found")

// Add registers a project root. The slug is derived from the basename of
// rootPath unless explicitly provided. Add is idempotent: re-adding the same
// rootPath updates last_seen_at + display_name instead of failing.
// Concurrent Add calls against the same root collapse into a single row
// thanks to the IMMEDIATE transaction + UNIQUE(root_path) upsert.
func (r *Registry) Add(ctx context.Context, rootPath, displayName, slug string) (Project, error) {
	abs, err := filepath.Abs(rootPath)
	if err != nil {
		return Project{}, fmt.Errorf("workspace: abs path: %w", err)
	}
	if slug == "" {
		slug = Slugify(filepath.Base(abs))
	}
	if displayName == "" {
		displayName = filepath.Base(abs)
	}
	dbPath := defaultProjectDBPath(abs)
	now := time.Now().UTC().Format(timeFmt)

	var result Project
	if err := r.inTx(ctx, func(tx *sql.Tx) error {
		existing, terr := txByRoot(ctx, tx, abs)
		if terr == nil {
			newSlug := existing.Slug
			if slug != "" && slug != existing.Slug {
				if other, serr := txBySlug(ctx, tx, slug); serr == nil && other.ID != existing.ID {
					return fmt.Errorf("workspace: slug %q already used by %s", slug, other.RootPath)
				} else if serr != nil && !errors.Is(serr, ErrNotFound) {
					return serr
				}
				newSlug = slug
			}
			if _, terr := tx.ExecContext(ctx,
				`update projects set last_seen_at = ?, display_name = ?, slug = ? where id = ?`,
				now, displayName, newSlug, existing.ID); terr != nil {
				return fmt.Errorf("workspace: refresh: %w", terr)
			}
			existing.LastSeenAt = parseTimePtr(now)
			existing.DisplayName = displayName
			existing.Slug = newSlug
			result = existing
			return nil
		}
		if !errors.Is(terr, ErrNotFound) {
			return terr
		}
		if other, serr := txBySlug(ctx, tx, slug); serr == nil {
			return fmt.Errorf("workspace: slug %q already used by %s", slug, other.RootPath)
		} else if !errors.Is(serr, ErrNotFound) {
			return serr
		}
		p := Project{
			ID:          ids.New(),
			Slug:        slug,
			DisplayName: displayName,
			RootPath:    abs,
			DBPath:      dbPath,
			AddedAt:     time.Now().UTC(),
			LastSeenAt:  parseTimePtr(now),
			SchemaVer:   1,
		}
		if _, terr := tx.ExecContext(ctx, `
			insert into projects (id, slug, display_name, root_path, db_path, added_at, last_seen_at, schema_version)
			values (?, ?, ?, ?, ?, ?, ?, ?)
			on conflict(root_path) do update set
				display_name = excluded.display_name,
				last_seen_at = excluded.last_seen_at`,
			p.ID, p.Slug, p.DisplayName, p.RootPath, p.DBPath,
			p.AddedAt.Format(timeFmt), now, p.SchemaVer); terr != nil {
			return fmt.Errorf("workspace: insert: %w", terr)
		}
		// Re-read so concurrent winners surface the canonical row.
		canonical, terr := txByRoot(ctx, tx, abs)
		if terr != nil {
			return terr
		}
		result = canonical
		return nil
	}); err != nil {
		return Project{}, err
	}
	return result, nil
}

// inTx runs fn inside a single deferred transaction. With WAL + UPSERT,
// SQLite serialises writers on first write inside the tx, which is enough
// to make Add idempotent under concurrency.
func (r *Registry) inTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("workspace: begin: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func txByRoot(ctx context.Context, tx *sql.Tx, root string) (Project, error) {
	row := tx.QueryRowContext(ctx, `select id, slug, display_name, root_path, db_path, added_at, last_seen_at, archived_at, schema_version
		from projects where root_path = ?`, root)
	p, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	return p, err
}

func txBySlug(ctx context.Context, tx *sql.Tx, slug string) (Project, error) {
	row := tx.QueryRowContext(ctx, `select id, slug, display_name, root_path, db_path, added_at, last_seen_at, archived_at, schema_version
		from projects where slug = ?`, slug)
	p, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	return p, err
}

// List returns registered projects. When includeArchived is false, archived
// rows are filtered out. Sort order: active first (most-recent last_seen_at),
// then archived (most-recent archived_at).
func (r *Registry) List(ctx context.Context, includeArchived bool) ([]Project, error) {
	q := `select id, slug, display_name, root_path, db_path, added_at, last_seen_at, archived_at, schema_version
	      from projects`
	if !includeArchived {
		q += " where archived_at is null"
	}
	q += " order by archived_at is null desc, coalesce(last_seen_at, added_at) desc"
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Resolve returns a project by id, slug or absolute path. Returns ErrNotFound
// when nothing matches.
func (r *Registry) Resolve(ctx context.Context, ref string) (Project, error) {
	if ref == "" {
		return Project{}, ErrNotFound
	}
	if p, err := r.byID(ctx, ref); err == nil {
		return p, nil
	}
	if p, err := r.bySlug(ctx, ref); err == nil {
		return p, nil
	}
	if abs, absErr := filepath.Abs(ref); absErr == nil {
		if p, err := r.byRoot(ctx, abs); err == nil {
			return p, nil
		}
	}
	return Project{}, ErrNotFound
}

// Archive marks the project as archived.
func (r *Registry) Archive(ctx context.Context, ref string) (Project, error) {
	p, err := r.Resolve(ctx, ref)
	if err != nil {
		return Project{}, err
	}
	now := time.Now().UTC().Format(timeFmt)
	if _, err := r.db.ExecContext(ctx,
		`update projects set archived_at = ? where id = ?`, now, p.ID); err != nil {
		return Project{}, fmt.Errorf("workspace: archive: %w", err)
	}
	p.ArchivedAt = parseTimePtr(now)
	return p, nil
}

// Unarchive clears the archived_at timestamp.
func (r *Registry) Unarchive(ctx context.Context, ref string) (Project, error) {
	p, err := r.Resolve(ctx, ref)
	if err != nil {
		return Project{}, err
	}
	if _, err := r.db.ExecContext(ctx,
		`update projects set archived_at = null where id = ?`, p.ID); err != nil {
		return Project{}, fmt.Errorf("workspace: unarchive: %w", err)
	}
	p.ArchivedAt = nil
	return p, nil
}

// SetActive records the active project. Pass an empty ref to clear.
func (r *Registry) SetActive(ctx context.Context, ref string) (Project, error) {
	if ref == "" {
		if _, err := r.db.ExecContext(ctx, `delete from active_project where singleton = 1`); err != nil {
			return Project{}, fmt.Errorf("workspace: clear active: %w", err)
		}
		return Project{}, nil
	}
	p, err := r.Resolve(ctx, ref)
	if err != nil {
		return Project{}, err
	}
	if _, err := r.db.ExecContext(ctx, `
		insert into active_project (singleton, project_id) values (1, ?)
		on conflict(singleton) do update set project_id = excluded.project_id`, p.ID); err != nil {
		return Project{}, fmt.Errorf("workspace: set active: %w", err)
	}
	return p, nil
}

// Active returns the currently active project. ErrNotFound when none is set.
func (r *Registry) Active(ctx context.Context) (Project, error) {
	var id sql.NullString
	if err := r.db.QueryRowContext(ctx,
		`select project_id from active_project where singleton = 1`).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		return Project{}, err
	}
	if !id.Valid {
		return Project{}, ErrNotFound
	}
	return r.byID(ctx, id.String)
}

// Touch updates last_seen_at on the resolved project.
func (r *Registry) Touch(ctx context.Context, ref string) error {
	p, err := r.Resolve(ctx, ref)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`update projects set last_seen_at = ? where id = ?`,
		time.Now().UTC().Format(timeFmt), p.ID)
	return err
}

// Forget removes the project row entirely. Used by Scan reconciliation when
// the on-disk root has disappeared. Never touches per-project files.
func (r *Registry) Forget(ctx context.Context, ref string) error {
	p, err := r.Resolve(ctx, ref)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `delete from projects where id = ?`, p.ID)
	return err
}

// --- helpers --------------------------------------------------------------

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

func Slugify(in string) string {
	var b strings.Builder
	prevSeparator := false
	for _, r := range strings.ToLower(in) {
		if isSlugRune(r) {
			b.WriteRune(r)
			prevSeparator = false
			continue
		}
		if !prevSeparator {
			b.WriteString(constants.SlugSeparator)
			prevSeparator = true
		}
	}
	if out := strings.Trim(b.String(), constants.SlugSeparator); out != "" {
		return out
	}
	return constants.SlugFallbackName
}

func isSlugRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}
