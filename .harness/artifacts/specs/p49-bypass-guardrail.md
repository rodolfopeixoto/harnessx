# P49 — Bypass guardrail for local CI/CD gates

## Intent

Prevent silent bypass of the local CI and spec-gate hooks. Any commit or
push that sets `HARNESS_SKIP_CI=1` or `HARNESS_SKIP_SPEC_GATE=1` without
supplying `HARNESS_BYPASS_REASON` must be blocked. Audited bypasses are
appended to `.harness/logs/gate-bypass.jsonl` for later review.

## Files in scope

- `scripts/git-hooks/pre-commit`
- `scripts/git-hooks/pre-push`
- `scripts/git-hooks/pre-commit-bypass-check.sh`
- `scripts/git-hooks/pre-push-bypass-check.sh`
- `docs/bypass-guardrail.md`
- `CONTRIBUTING.md` (bypass usage section)

## Invariants

- Guardrail must run **before** the existing `HARNESS_SKIP_CI` early-exit
  so blocking still fires when the user tries to skip local CI.
- Blocking behaviour requires a non-empty `HARNESS_BYPASS_REASON`; when
  present, the bypass is *audit-logged*, not silenced.
- The audit log line is single-line JSON with fields:
  `ts`, `user`, `hooks_skipped`, `reason`, `hook`.
- Scripts must be POSIX-portable Bash (`set -euo pipefail`) — no
  jq/python dependency in the hot path.
- Guardrail never touches the working tree.

## Validation

```sh
HARNESS_SKIP_CI=1 git commit --allow-empty -m "test" # must FAIL
HARNESS_SKIP_CI=1 HARNESS_BYPASS_REASON="hotfix #42" git commit --allow-empty -m "test" # must PASS + log
```

## Risk tier

low — additive check, cannot corrupt content; failure mode is a false
positive commit block (fix by exporting `HARNESS_BYPASS_REASON`).

## Rollback

`git revert` the wiring commit and delete the two new scripts.
