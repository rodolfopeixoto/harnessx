# P109 — Transactional shared state across agents

## Context

Paper § 5.2.4: "transactional shared program state" — multiple
agents writing into the same project artifacts need read/write sets +
version dependencies, not free-form text handoff.

Today `harness do` prepends a "Past steps in this run" block (v0.38).
That's good as context but does not stop task#3 from clobbering task#1's
edits to the same file with stale assumptions.

Add `.harness/runs/<id>/shared.json` recording per-task
`{read_set, write_set, assumptions, version}`. A pre-execute checker
blocks a later task whose `read_set` overlaps an earlier task's
`write_set` without re-reading.

## What ships

- `internal/sharedstate/sharedstate.go` with `Snapshot`, `Task`,
  `Conflict`, `Detect`.
- Serialization to `.harness/runs/<id>/shared.json`.
- `harness do` will call `sharedstate.Detect` between tasks (wiring
  in a follow-up release; this release ships the type + tests +
  store).

## Critical files

| Path | Change |
|---|---|
| `internal/sharedstate/sharedstate.go` (new) | type defs + Detect |
| `internal/sharedstate/sharedstate_test.go` (new) | conflict cases |

## Reuse, do not reinvent

- `ids.New()` for snapshot ids.
- JSON write pattern mirrors `internal/devloop/resume.go` to keep
  consistency.

## Verification

- `make lint` 0 issues.
- `go test ./internal/sharedstate/...` — round-trip + detect.
- Coverage ≥ 90%.

## Risks + mitigation

| Risk | Mitigation |
|---|---|
| read/write sets explode for large diffs | Sets store globs / path prefixes, not every line |
| False positives stop legitimate refactors | Detect returns advisory `Conflict` records; caller decides whether to block |

## Acceptance

- `Detect` flags overlap of `task2.read_set` ∩ `task1.write_set`
  when `task2.assumptions.version < task1.version`.
- Returns nil when assumptions are current.
- Round-trip JSON works.
