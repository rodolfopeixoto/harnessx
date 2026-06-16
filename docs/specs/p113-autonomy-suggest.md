# P113 — Human-in-the-loop autonomy suggest

## Context

Paper § 5.2.5 — operator approvals + denials should feed back into
future policy. Today autonomy tiers + per-path policy are static. Add
an approvals log and a `suggest` engine that mines it for tightening
proposals.

## What ships

- `internal/autonomy/approvals.go` — `Event{Path, Decision, Reason,
  At}`, `Append(root, e)`, `List(root) []Event`.
- `internal/autonomy/suggest.go` — `Suggest(events) []Proposal`
  with `Proposal{Path, From, To, Reason, EvidenceCount}`.
- Approvals land at `.harness/audit/approvals.jsonl`.

## Critical files

| Path | Change |
|---|---|
| `internal/autonomy/approvals.go` (new) | append-only JSONL store |
| `internal/autonomy/suggest.go` (new) | mining + proposal logic |
| `internal/autonomy/{approvals,suggest}_test.go` | round-trip + tightening rules |

## Reuse, do not reinvent

- `paths.HarnessDir` for the log location
- Existing `Policy` struct fields define From/To

## Verification

- `make lint` 0 issues
- Coverage ≥ 90% for both new files

## Risks + mitigation

| Risk | Mitigation |
|---|---|
| Approvals log unbounded | Pruner in v0.x.y; for now bytes negligible |
| Proposals tighten unfairly | Min `EvidenceCount=5` denials before suggesting upgrade |

## Acceptance

- Append + List round-trip
- 5+ denials on a path → propose `require_approval → deny`
- 5+ approvals on path with deny → propose `deny → require_approval`
