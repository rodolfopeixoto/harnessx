# Coverage Push Plan

Current overall coverage: **49.5%** (Go) as of 2026-06-16. Target:
**90%** across Go core. Plan trades fixture work for coverage gain
across five releases.

## Current per-package state

### Tier A — 100% (5)
`internal/agents`, `internal/platform/budget`, `internal/platform/clock`,
`internal/platform/tokens`.

### Tier B — 90–99% (5)
`internal/health` (93), `internal/intent` (93), `internal/audit` (91),
`internal/taskgraph` (91), `internal/platform/hashing` (90).

### Tier C — 70–89% (33)
Most app + adapter packages.

### Tier D — 50–69% (15)
`internal/sensors` (69), `internal/doctor` (68), `internal/context`
(63), `internal/ui` (52), `internal/execution` (52),
`internal/workflow` (48), `internal/devloop` (49), `internal/backup`
(49), `internal/adapters/lsp` (46), `internal/adapters/http` (41), …

### Tier E — <40% (6, critical)

| Pkg | % | Reason |
|---|---|---|
| `cmd/harness` | 5.7 | 51 cobra files, almost no e2e harness |
| `internal/agents/interactive` | 7 | needs pty/tmux/iterm |
| `internal/app/agentcmd` | 10 | depends on cmd wiring |
| `internal/runtime/containers` | 27 | needs Docker |
| `internal/update` | 29 | network + tarballs |
| `internal/secrets` | 40 | keychain/file-store IO |

## Top 15 untested functions

Ranked by uncovered statements × tractability (E easy / M medium fs+sqlite / H hard docker/pty/network).

| # | Pkg | Function | Stmts | Diff | Batch |
|---|---|---|---|---|---|
| 1 | `optimize` | `Render` | 7 | E | v0.72 |
| 2 | `optimize` | `Snapshot` | 6 | E | v0.72 |
| 3 | `index` | `mergeDeps` | 6 | E | v0.72 |
| 4 | `sensors` | `gitTopologyChanged` | 6 | M | v0.72 |
| 5 | `workflow` | `recordTelemetry` | 6 | M | v0.73 |
| 6 | `workflow` | `Execute` | 6 | M | v0.73 |
| 7 | `agentcmd` | `Run` | 6 | M | v0.73 |
| 8 | `auditrun` | `finalizeRun` | 6 | M | v0.73 |
| 9 | `execution` | `runHooks` | 6 | M | v0.73 |
| 10 | `devloop` | `Step` + `Watch` | 14 | M | v0.74 |
| 11 | `adapters/http` | `eventsHandler`+`subscribeHandler` | 12 | M | v0.74 |
| 12 | `agents/yaml` | `Render` | 6 | E | v0.74 |
| 13 | `cmd/harness` | `runMetricsExport` | 13 | E | v0.75 |
| 14 | `cmd/harness` | `runStackApply`/`runStackList` | 23 | M | v0.75 |
| 15 | `cmd/harness` | `runWorkflowRun` + `runDo` | 21 | M | v0.76 |

## 5-release coverage push plan

- **v0.72 — pure-render foundation** (+3.5pp → 53%). `optimize.Render`,
  `optimize.Snapshot`, `index.mergeDeps`, `sensors.gitTopologyChanged`.
  Fixtures: 2 golden JSON, 1 git repo with two commits. Test LOC ~220.

- **v0.73 — workflow + auditrun** (+4pp → 57%). `workflow.Execute`,
  `workflow.recordTelemetry`, `agentcmd.Run`, `auditrun.finalizeRun`,
  `execution.runHooks`. Reuses sqlite fixture builder; introduces
  `internal/app/workflow/testfakes/`. Test LOC ~450.

- **v0.74 — I/O surface: devloop + HTTP + yaml render** (+3pp → 60%).
  `devloop.Watch`/`Step`, `adapters/http.eventsHandler` +
  `subscribeHandler` (httptest SSE), `agents/yaml.Render` (golden
  YAML). Test LOC ~380.

- **v0.75 — CLI smoke I** (+5pp → 65%). `runMetricsExport`,
  `runStackApply`, `runStackList`, `runStackDiff`. Pattern follows
  `cmd_init_test.go` / `cmd_scaffold_test.go`. Test LOC ~520.

- **v0.76 — CLI smoke II: do + workflow** (+3pp → 68%). `runDo`,
  `planDo`, `runWorkflowRun`, `runWorkflowList`. Stub agents via
  `internal/agents` fakes. Test LOC ~480.

## Shell + UI gaps

### Shell harness
3 shell suites cover ~11 CLI surfaces. **~42 cobra subcommands have
no shell smoke**. Priority: `do`, `run`, `workflow`, `stack`,
`backup snapshot/restore`.

### Vitest UI
14 vitest tests cover api types + 13 pages. **28 React pages + 10 DS
primitives untested**. Priority: `Home`, `ActiveRun`, `Plan`,
`Catalog`, `DataExplorer`, `InspectorPanel`, `Shell`.

## Hard-to-cover (deferred)

- `cmd_backup.*` (rclone)
- `agents/interactive/{iterm2,tmux,pty}` (need real terminal)
- `runtime/containers/run.go` (Docker daemon)
- `adapters/lsp/stdio_client.go` (LSP child process)
- `update/*` (HTTPS releases)

These hit ~25% of remaining uncovered statements. Reaching 90%
requires either Docker-in-CI, mock substitution at the package
boundary, or accepting them as integration-only.
