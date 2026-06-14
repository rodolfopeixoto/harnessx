# P18 — Autonomy · Autopilot · Audit Center · Health Score

## Acceptance

- 5 autonomy levels enumerated as constants (Manual, Plan-and-Ask, Safe-Execute, Full-Project-Loop, Scheduled-Maintenance) with explicit allow/deny matrix.
- `internal/autonomy.Gate(level, op) (Decision, error)` returns allow|deny|require_approval.
- `internal/autopilot.Queue` persists pending cross-project tasks (in-memory store wired by HTTP for v0.2.0; durable repo deferred).
- `internal/audit` writes structured events sourced from cleanup/catalog/workflow.
- `internal/health.Score(project)` returns deterministic 0-100 with reasoned breakdown.
- CLI: `harness autonomy {get,set}`, `harness health show <slug?>`, `harness audit tail`.
- HTTP: `/api/autonomy/{get,set}`, `/api/health[/slug]`, `/api/audit`, `/api/autopilot`.

## Verification

- `internal/{autonomy,autopilot,audit,health}` ≥ 80% each.
- `scripts/e2e-phase18.sh`: walk autonomy levels asserting Gate decisions; enqueue audit events; query health for fixture project.
