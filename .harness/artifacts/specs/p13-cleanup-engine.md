# P13 — Cleanup Engine

## Acceptance

- `harness cleanup scan [<root>] [--kind <kind>]` lists detected candidates grouped by risk (low/medium/high). Report-only; no writes.
- `harness cleanup apply` requires policy match OR interactive `y` per group OR `HARNESS_CLEANUP_I_UNDERSTAND=1`. Every applied delete writes `audit_events` row with content hash.
- `harness cleanup policy show|init` reads/scaffolds `.harness/cleanup/policy.yaml`.
- Detectors: git worktrees, package caches (npm/go/pnpm/cargo/pip), abandoned `.harness/` dirs, VM leftovers (vagrant, parallels, vmware), Claude leftovers (`~/.claude/*` orphans), large files (>50 MB > 30 days untouched), docker containers + volumes.
- `GET /api/cleanup/scan` returns findings JSON.
- `POST /api/cleanup/apply` requires explicit `{"policy_path":"..."}` body.

## Contract

`internal/cleanup/`:
- `Detector` interface: `Detect(ctx, root) ([]Finding, error)`.
- `Finding{Kind, Path, Risk, Reason, SizeBytes, LastTouched}`.
- `Policy` schema versioned, never trusts unknown keys.
- `Executor.Apply(ctx, finding, policy) (Outcome, error)` re-checks policy + writes audit row.
- `internal/runtime/containers`: deterministic `docker compose up|down|health|verify-clean` shared with stack tour (P20).

## Risks

- False-positive delete. Mitigation: D5 two-key; high-risk findings require explicit `allowlist:` entry naming each path glob; default policy is empty.
- Containers leaked by tests. Mitigation: `runtime/containers.VerifyClean` enforces `docker ps -a -q` empty at test teardown.

## Verification

- `internal/cleanup` ≥ 92%, `internal/cleanup/detectors` ≥ 90%, `internal/runtime/containers` ≥ 85% with `docker` fake.
- `scripts/e2e-phase13.sh`: tmp tree with stale worktree + oversized cache + abandoned `.harness/`; scan reports all; `apply --policy fixture.yaml` deletes only allowlisted, audits each, leaves rest; `apply` without policy + without `--yes` exits 3.
