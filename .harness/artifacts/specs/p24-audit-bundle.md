# P24 — audit bundle

## Acceptance

- `harness stack audit` always emits `report/audit-bundle.zip` containing every artifact (HTML/PDF/JSON/screenshots/diff/run.log/fix-backlog) + BUNDLE_INDEX.md.
- New `json/cli-flows.json` captures `harness <cmd>` runs against a temp project — exit code + stdout/stderr + duration.
- New `json/inventory.json` snapshots go test summary (per-package), shell test summary, total file/loc counts.
- New `json/design-reference.json` records which design handoff screens were captured + Playwright spec for them.
- Disabled via `AUDIT_BUNDLE=0`.

## Verification

- e2e-phase23 keeps passing.
- New e2e-phase24 unzips the produced bundle into a tmp dir and asserts the BUNDLE_INDEX.md plus every required artifact is present.
