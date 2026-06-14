# P11 — Workspace Hub (multi-project foundation)

## Acceptance

- `harness project add <path>` registers a project in `~/.harness/registry.sqlite`; idempotent.
- `harness project list [--archived]` shows registered projects with status, last-seen, db-path.
- `harness project switch <slug|path>` records the active project (per-user, not per-shell).
- `harness project archive <slug|path>` flips `archived_at`; archived projects hidden by default `list`.
- `harness project scan [<root>]` discovers `.harness/` dirs under `<root>` (default: parent of cwd) and offers to register the unregistered ones.
- `harness project current` prints the resolved project (precedence: `--project` flag → `HARNESS_PROJECT` env → cwd walk-up → active project from registry).
- `GET /api/workspace/projects` returns `[{id,slug,name,root,status,last_seen_at,...}]`.
- `POST /api/workspace/projects` registers a path; `POST /api/workspace/switch` sets active; `GET /api/workspace/current` returns active.
- v0.1.0 single-project commands keep working when only one project (or none) is registered: cwd walk-up resolution is preserved as the final fallback.

## Contract (frozen this phase)

### Registry schema v1 (`internal/workspace/migrations/0001_registry.sql`)

```sql
create table if not exists projects (
  id text primary key,                 -- ULID
  slug text not null unique,           -- kebab-case from path basename, unique
  display_name text not null,
  root_path text not null unique,      -- absolute path
  db_path text not null,               -- absolute path to per-project sqlite
  added_at text not null,              -- RFC3339 UTC
  last_seen_at text,
  archived_at text,
  schema_version integer not null default 1
);

create table if not exists active_project (
  singleton integer primary key check (singleton = 1),
  project_id text references projects(id)
);

create table if not exists registry_meta (
  key text primary key,
  value text not null
);
create index if not exists idx_projects_status on projects(archived_at);
create index if not exists idx_projects_last_seen on projects(last_seen_at);
```

### Resolver precedence

1. `--project <slug|path>` global flag.
2. `HARNESS_PROJECT` env var.
3. `cwd` walk-up via existing `paths.FindProjectRoot()` (unchanged from v0.1.0).
4. `active_project` row from registry.
5. Hard error: `no project resolved; run 'harness project add <path>' first`.

### Concurrency

- Single `~/.harness/registry.lock` file held via `flock(2)` (advisory) around every registry write transaction.
- Registry DB always opened with `_pragma=journal_mode(WAL)`.

## Risks

- **Double-register race** — two `harness project add` calls on same path. Mitigation: `slug` is `UNIQUE`; `INSERT … ON CONFLICT DO UPDATE SET last_seen_at = excluded.last_seen_at`.
- **Leftover rows** — registered project folder deleted out of band. Mitigation: `project scan` reconciles + `project archive --missing` flag prunes; never auto-delete.
- **Registry corruption** — single point of failure for cross-project queries. Mitigation: registry is recreatable from filesystem (scan rebuilds); per-project data lives independently in each project's own DB.

## Non-goals (deferred)

- Project Health score (Phase 18).
- Stale detection (Phase 16).
- Cleanup of project leftovers (Phase 13).
- Dashboard `/workspace` UI (Phase 16).

## Verification

- `internal/workspace/` ≥ 95% line coverage (table tests + concurrent-add via two goroutines + resolver precedence cases).
- `scripts/e2e-phase11.sh` builds binary, creates two tmp projects, runs full CRUD flow, curls `/api/workspace/projects`.
