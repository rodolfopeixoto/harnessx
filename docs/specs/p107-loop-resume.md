# P107 — Long-horizon checkpoint / resume

## Context

`harness loop` today is bounded by `--max-attempts`. If the operator
cancels mid-loop (Ctrl-C, terminal closed, host reboots) the run is
lost — attempts so far stay in `.harness/runs/_loop/loop-<ts>.md`
but the loop itself cannot be picked up. The paper flags
"long-horizon execution" as an open challenge.

Add a per-attempt checkpoint that lets `harness loop resume <run-id>`
rehydrate prompt, baseline, completed attempts, last canonical error,
remaining budget and resume from attempt N+1.

## What ships

- Per-attempt `state.json` written to `.harness/runs/_loop/<id>/`
  after each attempt. Carries:
  `{run_id, original_prompt, baseline:{lint, test}, attempts:[],
  budget_usd_remaining, max_attempts, agent_id, autonomy, apply}`.
- `harness loop resume <run-id>` command. Loads state, continues
  from attempt `len(attempts)+1`. Honors remaining budget.
- `harness loop list` shows active resumable runs under
  `.harness/runs/_loop/`.
- Per-loop run gains a stable id (ULID) on first invocation, written
  to state.json.
- Original `harness loop "<prompt>"` unchanged for backward compat;
  it transparently allocates an id + writes state.json after each
  attempt.

## Critical files

| Path | Change |
|---|---|
| `internal/devloop/loop.go` | allocate run id, write state.json each attempt |
| `internal/devloop/resume.go` (new) | LoadState, Resume, ListResumable |
| `cmd/harness/cmd_loop.go` | add `loop resume <id>` + `loop list` subcommands |

## Reuse, do not reinvent

- `ids.New()` from `internal/platform/ids/` for run id.
- Existing `paths.HarnessDir(root)` for the runs root.
- `internal/devloop/loop.go::Run` keeps current entry point; only
  changes: take optional `runID` (empty → allocate new), write
  state after each attempt.

## Verification

- `make lint` 0 issues.
- `go test ./internal/devloop/...` — new tests cover state
  serialization round-trip, ListResumable, resume picks up at right
  attempt index.
- Smoke:
  ```
  harness loop "fix x" --agent fake-real --max-attempts 5 --budget-usd 0.10
  # Ctrl-C after attempt 2
  harness loop list                 # shows the run
  harness loop resume <run-id>      # picks up at attempt 3
  ```
- Performance: state.json write < 5 ms (small JSON blob, < 4 KB).

## Risks + mitigation

| Risk | Mitigation |
|---|---|
| state.json grows with attempts | Cap at 10 attempts (loop already does); strip lint/test output to last 80 lines before write (Canonicalise already trims) |
| Resume picks up after operator fixed the failure manually | Resume re-runs baseline + first attempt's verification cycle; if green, exits early |
| State path collides with existing _loop reports | New layout: `.harness/runs/_loop/<id>/{state.json, loop.md}` instead of flat `loop-<ts>.md` |

## Acceptance

- `harness loop resume <id>` runs end-to-end with fake-real adapter.
- `loop list` shows resumable runs sorted newest first.
- New `internal/devloop` coverage rises (resume.go tested at ≥ 80%).
- Backward-compat: existing operators see the same `harness loop`
  behaviour, just with a per-run sub-directory.
