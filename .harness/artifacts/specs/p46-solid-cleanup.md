# P46 — SOLID Cleanup (partial)

## Scope

Split the two biggest god files flagged by `harness audit-solid` on `develop`:

- `internal/specflow/specflow.go` (was 501 loc > 400)
- `internal/workspace/registry.go` (was 468 loc > 400)

## What moved

### `internal/specflow/`
- Persistence helpers (`Save`, `metadataHeader`, `EditViaEditor`, `runEditor`,
  `renderBaseline`) → `internal/specflow/persist.go`
- Small utilities (`answerOr`, `formatAnswers`, `pickBody`, `slugify`,
  `unifiedLines`) → `internal/specflow/helpers.go`

### `internal/workspace/`
- Scanner + defaults (`byID`, `bySlug`, `byRoot`, `scanOne`, `rowScanner`,
  `scanProject`, `parseTimePtr`, `defaultProjectDBPath`) →
  `internal/workspace/scan.go`
- Slug utilities (`Slugify`, `isSlugRune`) → `internal/workspace/slug.go`

No API changes. All exports unchanged. Tests pass unmodified.

## Result

`harness audit-solid` violations: **22 → 20** (2 god files removed).

## Backlog (waived — separate PRs)

Remaining 20 violations flagged, split across 15+ files. Each is its
own refactor because splitting them safely requires domain reading
and coverage retention:

| File | Metric | Value | Notes |
|---|---|---|---|
| `internal/repl/repl.go` | loc | 1829 | Massive — needs staged split across commands vs render vs state |
| `internal/execution/executor.go` | loc | 513 | Split by mode (worktree vs copy vs sandbox) |
| `internal/adapters/lsp/stdio_client.go` | loc+fan-out | 475/17 | Split proto vs framing vs io |
| `cmd/harness/cmd_onboarding.go` | loc | 689 | Wizard steps → sub-files |
| `cmd/harness/cmd_drive.go` | loc+fan-out | 509/18 | Split by CRUD op |
| `internal/auditrun/runner.go` | loc | 486 | Split by phase (build vs execute vs report) |
| `internal/runtime/containers/runtime.go` | loc | 466 | Split by runtime (docker vs podman) |
| `internal/app/learncmd/learncmd.go` | loc | 431 | Split promote vs demote |
| `cmd/harness/cmd_backup.go` | loc | 406 | Split by provider |
| `internal/agents/yaml/adapter.go` | loc | 405 | Split parse vs run |
| `cmd/harness/cmd_do.go` | loc+fan-out | 402/17 | Split by phase |
| `cmd/harness/cmd_mcphook.go` | loc | 401 | Split by hook kind |
| `internal/app/agentcmd/agentcmd.go` | fan-out | 22 | Split adapter helpers |
| `internal/app/sensorcmd/sensorcmd.go` | fan-out | 17 | Split by output format |
| `cmd/harness/cmd_new.go` | fan-out | 16 | Split by stack |
| `cmd/harness/cmd_ship.go` | fan-out | 16 | Split loop vs commit |

## Waiver rationale

`make audit-solid` is **not part of `make ci`** (advisory only). Each
remaining split needs its own review to avoid breaking behaviour;
bundling them here would produce an unreviewable diff.

## References

- Ran on `develop@fc3f90a` (post toolchain backport).
- Baseline noted in PR body for the follow-up work.
