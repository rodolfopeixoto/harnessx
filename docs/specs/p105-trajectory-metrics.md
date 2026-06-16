# P105 — Trajectory-efficiency metrics (paper § 5.2.1)

## Context

The paper flags "evaluation beyond task completion" as the #1 open
challenge. HarnessX records `EstimatedCostUSD`, `InputTokens`,
`OutputTokens` per run today but nothing about how *efficiently* a
trajectory reached the result. Operators can't answer: "was that
20-attempt loop expensive because the prompt was hard or because the
agent thrashed?".

This release extends `execution.Result` with four metric groups and
exposes them through `harness runs inspect` + a new
`harness metrics trajectory`. Pure data-layer release: no behaviour
change, no LLM call.

## What ships

- `execution.Result` gains four nested structs:
  - `Trajectory{ToolCalls, EditCount, WallMs}` — process effort
  - `Verification{SensorsRun, SensorsPassed, OracleCount}` — assurance
  - `Recovery{Retries, Regressions}` — wobble
  - `Replayability{EventsComplete bool}` — can the run be re-played?
- Executor populates these from already-collected data:
  `len(Hooks)`, `len(ChangedFiles)`, `time.Since(start)`,
  `len(Sensors)`, count of sensors with `Status==passed`, etc.
- New command `harness metrics trajectory [--since 7d] [--json]`:
  aggregates trajectory across all runs and prints a one-line summary
  per run plus totals.
- `harness runs inspect <id>` adds a "Trajectory" section.

## Critical files

| Path | Change |
|---|---|
| `internal/execution/types.go` | add `Trajectory`, `Verification`, `Recovery`, `Replayability` struct types + fields on `Result` |
| `internal/execution/executor.go` | populate the four fields right before `writeMeta` |
| `cmd/harness/cmd_metrics_trajectory.go` (new) | new subcommand under existing `metrics` group |
| `cmd/harness/cmd_metrics.go` | wire the new subcommand |
| `cmd/harness/cmd_run.go::renderInspect` | new "Trajectory" block when fields populated |

## Reuse, do not reinvent

- `execution.ListRuns` returns `[]Result` — aggregator iterates this.
- `internal/execution/executor.go::Execute` already counts hooks +
  changed files + sensor results; only assign to the new fields.
- `cmd/harness/cmd_metrics.go` already groups subcommands; add `trajectory`.

## Verification

- `make lint` 0 issues.
- `go test -cover ./internal/execution/... ./cmd/harness/...` — new
  tests cover every field population path; pkg coverage ≥ 90% for
  the new code.
- Smoke:
  ```
  harness feature "noop" --agent fake-real --apply --autonomy safe_execute
  harness runs inspect $(ls -t .harness/runs | head -1)   # shows Trajectory
  harness metrics trajectory --since 1d --json | jq .
  ```
- Performance: aggregation < 200 ms for 1k runs (bench in
  `internal/execution/prune_bench_test.go` — already exists for prune;
  add `BenchmarkTrajectoryAggregate`).

## Risks + mitigation

| Risk | Mitigation |
|---|---|
| Schema break for downstream consumers | All new fields optional; existing meta.json fields untouched |
| Aggregation slow on large run dirs | Stream + sum; do not load every meta.json at once when --since narrows the window |
| Operators confuse "Trajectory" with "Plan" | help text + tutorial section clarify: trajectory = process metrics; plan = intent |

## Acceptance

- Every new struct field populated on a real run.
- `harness metrics trajectory --json` schema_version=2 (bump from 1).
- `harness runs inspect` shows Trajectory block when fields present,
  hides when zero (backward-compatible with old runs).
- New tests bring `internal/execution` coverage from 51.7% → ≥ 60%.
