# P20 — Stack Tour

## Acceptance

- `harness stack tour [--dashboard] [--keep]` runs a deterministic walkthrough: temp project → init → workspace register → catalog install (mcp/filesystem) → cleanup scan → autonomy matrix → health score → optional dashboard probe.
- Every step prints a status line + exits non-zero on first failure.
- Cleans up the temp project, dashboard process, and (when `--containers` toggled) container leftovers via `runtime/containers.VerifyClean`.
- `harness stack status` reports whether the dashboard is reachable on the configured addr.
- `scripts/e2e-phase20.sh` runs the tour twice (with + without `--keep`) and asserts cleanup.

## Risks

- Long-running orphan dashboards. Mitigation: tour owns the lifecycle (start in goroutine, defer kill, VerifyClean at the end).

## Verification

- `internal/stack` ≥ 80%, deterministic stub probe in tests.
- `scripts/e2e-phase20.sh` exits 0; second run with `--keep` leaves the registry intact for manual smoke.
