# P112 — Multi-agent critic loop

## Context

Paper § 4.2 multi-agent review. After a task writes a diff, route the
diff to a critic adapter (different `strengths`) and capture verdict.
v0.32 picks one adapter per task; this release adds the critic
follow-up.

## What ships

- `internal/critic/critic.go` — `Request{Diff, OriginalPrompt,
  AdapterID}`, `Verdict{Score, Concerns []string, Suggestions
  []string}`, `Critique(ctx, req, registry) (Verdict, error)`.
- Critic adapter picked by `router.Pick(["review", "critic"])`;
  falls back to `["review"]` then any.
- Pure-data layer; wiring into `cmd_do.go` comes later.

## Critical files

| Path | Change |
|---|---|
| `internal/critic/critic.go` (new) | Critique entry + types |
| `internal/critic/critic_test.go` (new) | router fallback + Verdict parsing |

## Reuse, do not reinvent

- `router.Match`/`Pick` for adapter selection
- `agents.AgentAdapter.Run` for the LLM call

## Verification

- `make lint` 0 issues
- `go test ./internal/critic/...` ≥ 90% (fake adapter via existing pattern)

## Risks + mitigation

| Risk | Mitigation |
|---|---|
| Critic adds latency | Optional behind `--critic` flag in cmd_do later |
| Critic disagrees with applier | Verdict is advisory; cmd_do reports both |

## Acceptance

- `Critique` returns Verdict from fake adapter
- Router fallback `[review,critic] → [review] → any` works
- Verdict.Concerns/Suggestions populated from adapter output (line-split heuristic)
