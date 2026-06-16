# P108 — Change-contract harness optimiser

## Context

Paper § 5.2.3: "every proposed edit should carry a change contract:
which component is modified, which failure mode it targets, what
improvement it predicts, which invariants it must preserve, which
evaluation can falsify it, and how it can be rolled back."

HarnessX `internal/optimize/` today only emits container/Dockerfile
suggestions. Promote it to a harness-mutation engine: every proposed
edit to `.harness/config/*.yaml`, router strengths, autonomy policy,
or sensor set ships with a `ChangeContract` and is auditable, canary-
verifiable, rollback-able.

## What ships

- New `internal/optimize/contract.go` with `ChangeContract` struct.
- `harness optimize propose` writes proposals to
  `.harness/runs/_optimize/<ts>/proposals.json`.
- `harness optimize apply <proposal-id> --canary` runs the
  falsifier_test in a temp worktree; only promotes when it stays
  green; otherwise runs rollback_cmd.
- `harness optimize rollback <proposal-id>` re-runs rollback_cmd
  manually.

## Critical files

| Path | Change |
|---|---|
| `internal/optimize/contract.go` (new) | `ChangeContract` data structure + JSON marshal |
| `internal/optimize/contract_test.go` (new) | round-trip + validate |
| `cmd/harness/cmd_optimize.go` | wire `propose|apply|rollback` subcommands |
| `internal/optimize/apply.go` (new) | `Apply(c ChangeContract, canary bool) error` |

## Reuse, do not reinvent

- `internal/optimize/snapshot.go::Snapshot` already captures runtime
  state — reuse as the pre/post telemetry input for contracts.
- `internal/optimize/compare.go::Compare` already diffs snapshots —
  contract's falsifier_test calls it.
- `internal/scm/git.go::Init` and existing worktree code under
  `internal/execution/Manager` for canary isolation.

## Verification

- `make lint` 0 issues.
- `go test ./internal/optimize/...` — round-trip + canary path.
- Smoke: hand-craft a proposal JSON, `harness optimize apply --canary`,
  observe canary passes/fails and rollback fires.

## Risks + mitigation

| Risk | Mitigation |
|---|---|
| Apply mutates state unsafely | `--canary` mandatory at first; `rollback_cmd` required field; refuse Apply when falsifier_test absent |
| Proposal JSON grows complex | Schema versioned; v1 is intentionally narrow (4 component kinds) |
| Operators forget to run rollback | `harness optimize status` lists in-flight proposals + their rollback commands |

## Acceptance

- `ChangeContract` round-trips through JSON without loss.
- `Apply` refuses contracts without `rollback_cmd` or
  `falsifier_test`.
- Coverage `internal/optimize/contract.go` ≥ 90%.
