-- HarnessX workspace registry — schema v1.
-- This database lives at the user-scoped HarnessX home and contains ONLY
-- cross-project state. Per-project data stays in each project's own
-- .harness/db/harness.sqlite.

create table if not exists projects (
  id text primary key,
  slug text not null unique,
  display_name text not null,
  root_path text not null unique,
  db_path text not null,
  added_at text not null,
  last_seen_at text,
  archived_at text,
  schema_version integer not null default 1
);

create table if not exists active_project (
  singleton integer primary key check (singleton = 1),
  project_id text references projects(id) on delete set null
);

create table if not exists registry_meta (
  key text primary key,
  value text not null
);

create index if not exists idx_projects_status on projects(archived_at);
create index if not exists idx_projects_last_seen on projects(last_seen_at);
