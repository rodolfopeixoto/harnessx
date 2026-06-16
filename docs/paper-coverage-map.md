# Paper Coverage Map — arXiv 2605.18747 vs HarnessX

Last refreshed 2026-06-16. Cross-reference of every concrete concept
from "Code as Agent Harness" against shipped surface, with current
unit/integration coverage and the next-release gap-closer.

Coverage % comes from `go test -cover ./...` baseline 2026-06-16
(total 49.5%). Per-package figures cited from
[`docs/v1-readiness.md`](v1-readiness.md) where available.

## Layer 1 — Harness Interface

| Concept | Harness file / cmd | Status | Cov |
|---|---|---|---|
| Code for Reasoning (PoT / PAL) | `internal/promptenh/`, `internal/intent/intent.go` | ◐ | 86% / 93% |
| Code for Acting (intent → ops) | `cmd/harness/cmd_do.go`, `internal/execution/executor.go` | ✓ | n/a / 52% |
| Code for Environment (state) | `internal/sensors/runner.go`, `internal/context/builder.go` | ✓ | 69% / 63% |
| Unified substrate | `internal/agents/registry.go`, `internal/adapters/` | ✓ | 100% / mixed |

## Layer 2 — Mechanisms

| Concept | Harness | Status | Cov |
|---|---|---|---|
| Planning (decomposition) | `internal/taskgraph/`, `internal/plan/` | ✓ | 91% / 84% |
| Re-plan at runtime | `internal/devloop/loop.go::Canonicalise` | ◐ | 49% |
| Working memory | `internal/context/builder.go` | ✓ | 63% |
| Long-term memory | `internal/recall/recall.go` (`harness memory recall`) | ◐ | 86% |
| Multi-agent / shared memory | `cmd_do.go::handoff` | ◐ | n/a |
| Tool use / MCP | `internal/mcppkg/`, `internal/execution/mcp.go` | ✓ | 81% |
| Plan→Execute→Verify loop | `internal/devloop/loop.go` (`harness loop`) | ✓ | 49% |
| Iterative debug / canonical error | `devloop.Canonicalise` | ✓ | 49% |
| Adaptive harness optimisation | `internal/optimize/optimize.go` | ✗ | 73% (limited scope) |
| Telemetry / observability | `cmd/harness/cmd_metrics.go`, SSE | ◐ | n/a |

## Layer 3 — Multi-agent coordination

| Concept | Harness | Status | Cov |
|---|---|---|---|
| Per-task routing | `internal/router/strengths.go` | ✓ | 76% |
| Shared code substrate | `internal/scm/git.go`, `.harness/runs/` | ◐ | 73% |
| Multi-agent review | `cmd_do.go` + sensors | ◐ | n/a |
| Execution feedback sync | v0.38 handoff block | ◐ | n/a |

## Principles

| Principle | Surface | Status |
|---|---|---|
| Executability | `internal/scaffoldpkg/` (5 langs, byte-identical), `harness do --deterministic` | ✓ |
| Verifiability | `internal/sensors/runner.go`, `harness loop` baseline | ✓ |
| Statefulness | sqlite registry + `.harness/runs/` + recall | ✓ |
| Composability | router + scaffold + sensor + hook + mcp | ✓ |

## Open challenges

| Challenge | Surface today | Status |
|---|---|---|
| Harness-level evaluation | `cmd_metrics.go` (cost only) | ✗ |
| Verification with incomplete feedback | `sensors.Result.Confidence` (field exists) | ◐ |
| Regression-free self-evolution | devloop baseline within one loop | ◐ |
| Transactional shared state | handoff text block | ✗ |
| Human-in-the-loop safety as harness state | `internal/autonomy/` tiers + per-path policy | ◐ |
| Multimodal code-harness | `harness do --image` adds vision tag | ◐ |
| Science of harness engineering | `docs/v1-readiness.md`, `architecture.md` | ◐ |
| Long-horizon checkpoint/resume | bounded `--max-attempts` only | ◐ |

## Top 5 gaps — prioritized

1. **v0.72: Trajectory-efficiency metrics** — extend `execution.Result`
   with `trajectory_efficiency{tool_calls, tokens, edits, wall_ms}`,
   `verification_strength{coverage, oracle_count}`,
   `recovery{retries, regressions}`, `replayability{events_complete}`.
   New `harness metrics trajectory --since`.

2. **v0.73: Evidence-bundle verifiers** — every sensor returns
   `Result{Confidence, Scope, Verified[], Unverified[], Risks[]}`.
   `devloop` refuses green when `Confidence < 0.5` and `Unverified`
   non-empty.

3. **v0.74: Long-horizon checkpoint/resume** — `devloop` persists
   per-attempt `state.json`; new `harness loop resume <run-id>`.

4. **v0.75: Change-contract harness optimiser** — promote
   `internal/optimize/` to harness-mutation engine with
   `{component, target_failure, predicted_improvement, invariants,
   falsifier_test, rollback_cmd}`. Apply via
   `harness optimize apply --canary`.

5. **v0.76: Transactional shared state across agents** —
   `.harness/runs/<id>/shared.json` with per-task
   `{read_set, write_set, assumptions, version}`. Conflict detector
   blocks stale-snapshot tasks.

## Anchor passages from the paper

> "code as agent harness: code as the executable and inspectable
> medium through which agents reason, maintain state, and expose
> feedback" (§ Introduction)

> "harness mechanisms are not isolated add-on modules, but coordinated
> control surfaces" (§3)

> "every proposed edit should carry a change contract: which component
> is modified, which failure mode it targets, what improvement it
> predicts, which invariants it must preserve, which evaluation can
> falsify it, and how it can be rolled back" (§5.2.3)
