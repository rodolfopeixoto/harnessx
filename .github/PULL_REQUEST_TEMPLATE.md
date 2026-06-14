# PR

## What changed
<!-- One-sentence summary. -->

## Why
<!-- User-facing motivation. Link to issue when it exists: Closes #NNN. -->

## How
<!-- Architectural notes. Mention spec section + phase if relevant. -->

## Spec / phase touched
<!-- §17 / Phase 6 / Hardening N — pick one if applicable. -->

## Pre-merge checklist
- [ ] `make install-hooks` was run once on this clone (pre-push gate active).
- [ ] `make ci` is green locally (the pre-push hook ran).
- [ ] `make e2e` (or relevant `scripts/e2e-phase*.sh`) is green.
- [ ] New behaviour ships with tests at the interface seam.
- [ ] User-facing strings go through `internal/platform/i18n` (English first).
- [ ] No hardcoded constants — shared values live in `internal/platform/constants`.
- [ ] Docs touched when behaviour or CLI surface changes (`README`, `docs/`, `CHANGELOG`, `HARNESSX-MASTER-PLAN`).
- [ ] No CGO. No `mattn/go-sqlite3`. No `..` in `//go:embed`. No secrets in commits.
- [ ] Conventional Commits in title (`feat:`, `fix:`, `docs:`, …).

## Phase boundary check
<!-- HarnessX uses strict phase boundaries — see HARNESSX-MASTER-PLAN §9.
     A bug fix in Phase 6 land must not silently add Phase 8 dashboard
     work. Call out any cross-phase change explicitly. -->

## Screenshots / output
<!-- For CLI: paste before/after stdout. For dashboard: GIF or PNG. -->
