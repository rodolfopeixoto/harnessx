# P31 Final Report

## Summary

Real agentic execution loop is live behind `harness execute` + `harness runs`.
End-to-end smoke against a fixture project proves the chain: agent
invocation → diff capture → autonomy gate → apply to project root, with
waiting_approval gating Dockerfile-class changes under SafeExecute.

## What was broken

- `harness feature/bugfix --yes` invoked the router but never captured a
  diff, never ran sensors, never gated on autonomy, never wrote a
  per-run artifact tree. Cost reported $0 because the round trip didn't
  surface real changes.
- No worktree manager existed. The agent ran against the project root.
- `harness ask` only built context, never asked an agent.
- MCP / hook scanners produced inventories but their results never
  reached the adapter.

## What now works

- `internal/execution.DefaultExecutor` drives the full loop:
  worktree → adapter.Run → stdout/stderr capture → diff capture
  (unified patch + stat + JSON file list) → sensors → ClassifyRisk →
  autonomy.Gate → apply or waiting_approval, all persisted as
  `meta.json` + `report.md` under `.harness/runs/<run_id>/`.
- `cmd/fake-agent` deterministic adapter exercises the loop in tests and
  the E2E smoke without depending on Claude auth.
- `harness execute <prompt> --agent <id> [--apply|--plan-only]
  [--autonomy ...]` is the direct CLI path into the loop. Flags map 1:1
  onto `execution.Request`.
- `harness runs list|inspect|report|sensors|approve|discard` reads the
  persisted runs and supports human-in-the-loop approval for
  waiting_approval states.
- Report explicitly records MCPs and hooks detected on the project and
  states they are not yet injected — P32 work.

## Commands implemented

```
harness execute <prompt> [--agent ...] [--apply|--plan-only]
                         [--autonomy ...] [--budget-usd N]
harness runs list [--json]
harness runs inspect <run-id> [--json]
harness runs report <run-id>
harness runs sensors <run-id>
harness runs approve <run-id>
harness runs discard <run-id>
```

`harness feature/bugfix/ask` still take the v0.3.0 path; migrating them
onto `execution.DefaultExecutor` is the only P31 task deferred (#105 on
the todo list — small wiring change, no contract change).

## Real execution path

1. Resolve runID, mkdir `.harness/runs/<id>/`
2. Scan MCPs / hooks → record as detected-not-active
3. If `--plan-only`: write plan-only report + meta and exit
4. Create isolation workspace:
   - git project → `git worktree add -b harness/run/<id> .harness/worktrees/<id>`
   - non-git → controlled copy
5. Invoke `agents.AgentAdapter.Run` with stdin prompt + worktree as CWD
6. Persist `stdout.log` + `stderr.log`; parse usage if JSON path present
7. If failure: status `agent_failed`, cleanup worktree, write report
8. Stage worktree changes (`git add -A`), capture:
   - `diff.patch` (`git diff --cached`)
   - `diff-stat.txt` (`git diff --cached --stat`)
   - `changed-files.json`
9. If no diff and mode = feature/bugfix → status `no_changes`, non-zero
10. Run configured sensors against the worktree path
11. `ClassifyRisk(changed)`; `GateApply(level, risk, sensors)`
12. Branches by decision: Allow + `--apply` → `git apply --3way --index`
    against project root, cleanup worktree, status `applied`.
    Approval / no `--apply` → status `waiting_approval`. Deny → status
    `autonomy_denied`.
13. Failed blocking sensor downgrades non-denied statuses to
    `sensor_failed`.
14. Write `report.md` + `meta.json`. Return Result.

## Agent adapter behavior

Re-uses `internal/agents/yaml` adapter unchanged. The Executor talks the
existing `AgentAdapter` interface. Anything you can describe in the YAML
spec (Claude, Codex, Gemini, Kimi, fake-real) plugs in without code
changes.

## Worktree behavior

- Git-backed worktrees by default, fallback to controlled copy (skips
  `.harness/` and `.git/`) for non-git roots.
- Cleanup removes the worktree dir + the harness branch, idempotent on
  re-invocation.
- `harness runs discard` reuses Cleanup for the human-in-the-loop path.

## Diff behavior

- Worktree path: `git add -A` then `git diff --cached` for a clean
  unified patch + stat artifact.
- Copy path: list every file (simple change set for `--apply`'s
  rsync-like copy).
- Empty diff is non-zero for feature / bugfix modes.

## Sensor behavior

- Bridged through `sensors.Runner`. Per-sensor output dir is
  `.harness/runs/<id>/sensors/`.
- Status mapped 1:1 to `SensorOutcome.Status`.
- Any failed sensor blocks the gate (sensor_failed downgrade).

## Autonomy gate behavior

- Risk classifier high triggers: Dockerfile, lockfiles, package
  manifests, .env*, migrations/, .github/workflows/, secrets/,
  .harness/config/autonomy.yaml or routes.yaml.
- Decision routed through `autonomy.Gate(level, OpExecuteLowRisk|HighRisk)`.
- Deny ⇒ autonomy_denied. Approval ⇒ waiting_approval. Allow + --apply
  ⇒ applied.

## Tests added

`internal/execution/worktree_test.go`:
- `TestPrepareAndCleanup_Worktree` — git worktree create + diff capture
  + idempotent cleanup.
- `TestPrepare_NonGitFallsBackToCopy` — non-git project falls back to
  controlled copy.

`internal/execution/executor_test.go`:
- `TestExecute_FakeAgentProducesDiff` — diff produced, status
  waiting_approval (no --apply).
- `TestExecute_NoChangesIsErrorForFeatureMode` — non-zero error.
- `TestExecute_AgentFailureRecorded` — agent_failed surfaced.
- `TestExecute_ApplyMergesIntoProjectRoot` — applied, file present in
  project root.
- `TestExecute_HighRiskRequiresApprovalUnderSafeExecute` — Dockerfile
  forces waiting_approval even with --apply under safe_execute.

## Commands run

```
go test ./internal/execution/... ./cmd/fake-agent/...   # ok
go build ./...                                          # ok
gofmt -l internal/execution cmd/harness                 # clean
```

E2E smoke under `/tmp/p31-e2e/`:
```
harness execute "create greet.md with content: hello from harnessx" \
  --agent fake-real --apply --autonomy safe_execute
# → status=applied, greet.md present in project root

harness execute "create Dockerfile with content: FROM scratch" \
  --agent fake-real --apply --autonomy safe_execute
# → status=waiting_approval (high risk class blocked auto-apply)

harness runs approve <run-id>
# → Applied. Worktree removed. Dockerfile present in project root.

harness runs list
# → both runs shown as applied
```

## Results

| Scenario | Expected | Actual |
|---|---|---|
| Low-risk + safe_execute + --apply | applied | applied |
| High-risk + safe_execute + --apply | waiting_approval | waiting_approval |
| approve waiting_approval | applied, worktree gone | applied, worktree gone |
| Feature mode, no diff | exit non-zero, no_changes | exit non-zero, no_changes |
| Agent exits non-zero | agent_failed recorded | agent_failed recorded |

## Artifacts (per run, under .harness/runs/<run-id>/)

- `stdout.log` — captured agent stdout
- `stderr.log` — captured agent stderr
- `diff.patch` — unified diff (git mode)
- `diff-stat.txt` — `git diff --cached --stat`
- `changed-files.json` — JSON list of changed paths
- `report.md` — human-readable report
- `meta.json` — full `execution.Result` (the source of truth for
  `harness runs *`)
- `sensors/` — per-sensor output dirs

## Remaining gaps

1. **P31 #105** — migrate `workflow.Feature/Bugfix/Ask` to call
   `execution.DefaultExecutor` instead of the legacy `executeAgents`.
   Backwards-compatible because `harness execute` already proves the
   pipeline. Plain wiring change, no contract change. ~150 LOC + 1 e2e.
2. **P32** — actually inject MCP configs into the adapter invocation
   (pass `--mcp-config` to Claude/Codex/Gemini, expose stdio MCP
   sockets) and dispatch hook events pre/post adapter run. The Executor
   already surfaces detected-not-active so the report doesn't lie.
3. **P33** — Inspector tabs on the dashboard, ActionService backend
   persistence, SSE for the Active Run page, memory promotion UI.

## Branch

`feature/p31-real-agentic-execution`, 2 commits on top of develop:
1. `feat(execution): foundation types, worktree manager, diff capture, fake agent`
2. `feat(execution): Executor, autonomy gate, run inspection CLI, e2e green`
