// SPDX-License-Identifier: MIT

package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	// modernc.org/sqlite registers the "sqlite" driver via init().
	_ "modernc.org/sqlite"

	"github.com/ropeixoto/harnessx/internal/domain"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const timeFmt = time.RFC3339Nano

// Repo is a thin typed wrapper around *sql.DB. All times are stored as
// RFC3339 UTC strings; conversions live in this file.
type Repo struct {
	db *sql.DB
}

// Open opens (and creates if missing) a SQLite database at path, applies
// pending migrations, and returns the Repo. path may be ":memory:".
func Open(path string) (*Repo, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("sqlite: mkdir: %w", err)
		}
	}

	dsn := path
	if path != ":memory:" {
		// foreign_keys is off by default in SQLite; turn it on so cascade works.
		dsn = fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", url.PathEscape(path))
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open: %w", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: ping: %w", err)
	}
	r := &Repo{db: db}
	if err := r.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return r, nil
}

func (r *Repo) Close() error { return r.db.Close() }

// DB exposes the underlying *sql.DB for tests and ad-hoc queries.
func (r *Repo) DB() *sql.DB { return r.db }

func (r *Repo) migrate() error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("sqlite: read migrations: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return fmt.Errorf("sqlite: read %s: %w", e.Name(), err)
		}
		if _, err := r.db.Exec(string(b)); err != nil {
			return fmt.Errorf("sqlite: apply %s: %w", e.Name(), err)
		}
	}
	return nil
}

// --- Sessions ---------------------------------------------------------------

func (r *Repo) CreateSession(ctx context.Context, s domain.Session) error {
	_, err := r.db.ExecContext(ctx, `
		insert into sessions (id, project_path, mode, status, started_at)
		values (?, ?, ?, ?, ?)`,
		s.ID, s.ProjectPath, string(s.Mode), string(s.Status), s.StartedAt.UTC().Format(timeFmt))
	return err
}

func (r *Repo) FinishSession(ctx context.Context, id string, status domain.Status, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		update sessions set status = ?, finished_at = ? where id = ?`,
		string(status), at.UTC().Format(timeFmt), id)
	return err
}

// --- Runs -------------------------------------------------------------------

func (r *Repo) CreateRun(ctx context.Context, run domain.Run) error {
	_, err := r.db.ExecContext(ctx, `
		insert into runs (id, session_id, stage, agent, model, status, started_at)
		values (?, ?, ?, ?, ?, ?, ?)`,
		run.ID, run.SessionID, string(run.Stage), nullIfEmpty(run.Agent),
		nullIfEmpty(run.Model), string(run.Status), run.StartedAt.UTC().Format(timeFmt))
	return err
}

func (r *Repo) FinishRun(ctx context.Context, id string, status domain.Status, at time.Time, exit int) error {
	_, err := r.db.ExecContext(ctx, `
		update runs set status = ?, finished_at = ?, exit_code = ?,
		  latency_ms = cast((julianday(?) - julianday(started_at)) * 86400000 as integer)
		where id = ?`,
		string(status), at.UTC().Format(timeFmt), exit, at.UTC().Format(timeFmt), id)
	return err
}

func (r *Repo) ListRecentSessions(ctx context.Context, limit int) ([]domain.Session, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
		select id, project_path, mode, status, started_at, finished_at
		from sessions order by started_at desc limit ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Session
	for rows.Next() {
		var s domain.Session
		var startedAt string
		var finishedAt sql.NullString
		if err := rows.Scan(&s.ID, &s.ProjectPath, &s.Mode, &s.Status, &startedAt, &finishedAt); err != nil {
			return nil, err
		}
		s.StartedAt, _ = time.Parse(timeFmt, startedAt)
		if finishedAt.Valid {
			t, _ := time.Parse(timeFmt, finishedAt.String)
			s.FinishedAt = &t
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// --- Agent certifications ---------------------------------------------------

func (r *Repo) WriteAgentCertification(ctx context.Context, c domain.AgentCertification) error {
	_, err := r.db.ExecContext(ctx, `
		insert into agent_certifications
		  (id, agent_id, cli_version, adapter_version, score, status, details_json, certified_at)
		values (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.AgentID, nullIfEmpty(c.CLIVersion), nullIfEmpty(c.AdapterVersion),
		c.Score, c.Status, c.DetailsJSON, c.CertifiedAt.UTC().Format(timeFmt))
	return err
}

func (r *Repo) LatestAgentCertification(ctx context.Context, agentID string) (domain.AgentCertification, error) {
	row := r.db.QueryRowContext(ctx, `
		select id, agent_id, cli_version, adapter_version, score, status, details_json, certified_at
		from agent_certifications where agent_id = ?
		order by certified_at desc limit 1`, agentID)
	var c domain.AgentCertification
	var cli, adapter sql.NullString
	var certifiedAt string
	if err := row.Scan(&c.ID, &c.AgentID, &cli, &adapter, &c.Score, &c.Status, &c.DetailsJSON, &certifiedAt); err != nil {
		return domain.AgentCertification{}, err
	}
	if cli.Valid {
		c.CLIVersion = cli.String
	}
	if adapter.Valid {
		c.AdapterVersion = adapter.String
	}
	c.CertifiedAt, _ = time.Parse(timeFmt, certifiedAt)
	return c, nil
}

// --- Sensor results ---------------------------------------------------------

func (r *Repo) WriteSensorResult(ctx context.Context, runID, sensor, status string, durationMs int64, outputPath string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		insert into sensor_results (run_id, sensor, status, duration_ms, output_path, created_at)
		values (?, ?, ?, ?, ?, ?)`,
		runID, sensor, status, durationMs, nullIfEmpty(outputPath), at.UTC().Format(timeFmt))
	return err
}

func (r *Repo) ListSensorResults(ctx context.Context, runID string) ([]domain.SensorResult, error) {
	rows, err := r.db.QueryContext(ctx, `
		select id, run_id, sensor, status, duration_ms, output_path, created_at
		from sensor_results where run_id = ? order by id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.SensorResult
	for rows.Next() {
		var sr domain.SensorResult
		var status string
		var dur sql.NullInt64
		var outPath sql.NullString
		var createdAt string
		if err := rows.Scan(&sr.ID, &sr.RunID, &sr.Sensor, &status, &dur, &outPath, &createdAt); err != nil {
			return nil, err
		}
		sr.Status = domain.SensorStatus(status)
		if dur.Valid {
			sr.DurationMs = dur.Int64
		}
		if outPath.Valid {
			sr.OutputPath = outPath.String
		}
		sr.CreatedAt, _ = time.Parse(timeFmt, createdAt)
		out = append(out, sr)
	}
	return out, rows.Err()
}

// --- Metrics ----------------------------------------------------------------

func (r *Repo) WriteMetric(ctx context.Context, runID, name string, value float64, unit, tagsJSON string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		insert into metrics (run_id, name, value, unit, tags, created_at)
		values (?, ?, ?, ?, ?, ?)`,
		runID, name, value, nullIfEmpty(unit), nullIfEmpty(tagsJSON), at.UTC().Format(timeFmt))
	return err
}

// UpdateRunCostAndTokens records adapter usage on the run row, called by
// the fallback executor after each attempt.
func (r *Repo) UpdateRunCostAndTokens(ctx context.Context, runID string, in, cached, out, reasoning int, costUSD float64, agent, model, fallbackFrom string, errorType string) error {
	_, err := r.db.ExecContext(ctx, `
		update runs set
		  input_tokens = ?, cached_input_tokens = ?, output_tokens = ?,
		  reasoning_tokens = ?, estimated_cost_usd = ?,
		  agent = ?, model = ?, fallback_from = ?, error_type = ?
		where id = ?`,
		in, cached, out, reasoning, costUSD,
		nullIfEmpty(agent), nullIfEmpty(model), nullIfEmpty(fallbackFrom), nullIfEmpty(errorType),
		runID)
	return err
}
