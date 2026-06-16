# P110 — Re-plan runtime (task-graph rewrite)

## Context

Paper § 3.1: planning is not only pre-run; the harness must
re-plan when feedback surfaces missed preconditions. Today
`devloop.Canonicalise` rewrites only the prompt. We add a structured
re-plan layer over `taskgraph` so the executor can decide:

- inject a new precondition task (e.g. `pip install` before test)
- skip a task whose precondition cannot be met
- split a too-broad task into smaller ones

## What ships

- `internal/taskgraph/replan.go` with:
  - `Reason{Kind, Detail}` (`MissingDep`, `PreconditionFailed`,
    `OutOfScope`, `Split`)
  - `Replan(graph []Task, reason Reason) []Task` — returns a new
    graph
  - helpers: `InjectBefore(graph, idx, t)`, `Skip(graph, idx)`,
    `Split(graph, idx, into []Task)`

## Critical files

| Path | Change |
|---|---|
| `internal/taskgraph/replan.go` (new) | Reason + Replan + helpers |
| `internal/taskgraph/replan_test.go` (new) | every helper + Replan dispatch |

## Reuse, do not reinvent

- `taskgraph.Task` already exists — replan only manipulates slices
- `taskgraph.Kind` + `Tags` reused as-is

## Verification

- `make lint` 0 issues
- `go test ./internal/taskgraph/...` — replan_test ≥ 90% coverage of
  new file

## Risks + mitigation

| Risk | Mitigation |
|---|---|
| Replan loops infinitely | Caller (devloop) caps replan attempts at 2 per attempt |
| Operators lose visibility | Each replan returns a `Reason` recorded in run report |

## Acceptance

- `InjectBefore` puts new task at idx, shifts the rest
- `Skip` removes task at idx, preserves order
- `Split` replaces task at idx with N tasks
- `Replan` dispatches by Reason.Kind correctly
