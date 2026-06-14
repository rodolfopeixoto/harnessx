-- HarnessX schema. Created from spec §23. Phase 1 only writes to
-- sessions + runs; later phases populate the remaining tables.

create table if not exists sessions (
  id                  text primary key,
  project_path        text not null,
  mode                text not null,
  status              text not null,
  started_at          text not null,
  finished_at         text,
  total_cost_usd      real default 0,
  total_latency_ms    integer default 0,
  total_input_tokens  integer default 0,
  total_output_tokens integer default 0
);
create index if not exists idx_sessions_project_path on sessions(project_path);
create index if not exists idx_sessions_started_at   on sessions(started_at);

create table if not exists runs (
  id                  text primary key,
  session_id          text not null references sessions(id) on delete cascade,
  stage               text not null,
  agent               text,
  model               text,
  status              text not null,
  prompt_hash         text,
  context_hash        text,
  started_at          text not null,
  finished_at         text,
  latency_ms          integer,
  input_tokens        integer,
  cached_input_tokens integer,
  output_tokens       integer,
  reasoning_tokens    integer,
  estimated_cost_usd  real,
  exit_code           integer,
  fallback_from       text,
  error_type          text
);
create index if not exists idx_runs_session_id on runs(session_id);
create index if not exists idx_runs_status     on runs(status);
create index if not exists idx_runs_agent      on runs(agent);
create index if not exists idx_runs_started_at on runs(started_at);

create table if not exists sensor_results (
  id          integer primary key autoincrement,
  run_id      text not null references runs(id) on delete cascade,
  sensor      text not null,
  status      text not null,
  duration_ms integer,
  output_path text,
  created_at  text not null
);
create index if not exists idx_sensor_results_run_id on sensor_results(run_id);
create index if not exists idx_sensor_results_sensor on sensor_results(sensor);

create table if not exists metrics (
  id         integer primary key autoincrement,
  run_id     text not null references runs(id) on delete cascade,
  name       text not null,
  value      real not null,
  unit       text,
  tags       text,
  created_at text not null
);
create index if not exists idx_metrics_run_id_name on metrics(run_id, name);

create table if not exists memories (
  id              text primary key,
  scope           text not null,
  kind            text not null,
  content         text not null,
  evidence_run_id text,
  confidence      real not null,
  created_at      text not null,
  updated_at      text not null
);
create index if not exists idx_memories_scope_kind on memories(scope, kind);

create table if not exists skill_versions (
  id            text primary key,
  skill_name    text not null,
  version       integer not null,
  content_hash  text not null,
  score         real,
  accepted      integer not null,
  created_at    text not null
);
create index if not exists idx_skill_versions_name on skill_versions(skill_name);

create table if not exists agent_certifications (
  id              text primary key,
  agent_id        text not null,
  cli_version     text,
  adapter_version text,
  score           integer not null,
  status          text not null,
  details_json    text not null,
  certified_at    text not null
);
create index if not exists idx_agent_certs_agent on agent_certifications(agent_id);

create table if not exists artifacts (
  id            text primary key,
  session_id    text,
  run_id        text,
  kind          text not null,
  path          text not null,
  content_hash  text not null,
  metadata_json text,
  created_at    text not null
);
create index if not exists idx_artifacts_session_id   on artifacts(session_id);
create index if not exists idx_artifacts_run_id       on artifacts(run_id);
create index if not exists idx_artifacts_content_hash on artifacts(content_hash);
