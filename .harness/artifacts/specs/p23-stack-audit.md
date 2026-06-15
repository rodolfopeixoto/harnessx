# P23 — bin/stack audit — deterministic visual + functional audit

## Acceptance

- `bin/stack audit` (wrapper) runs the full audit pipeline.
- `harness stack audit` is the Cobra command behind it; both produce identical artifacts.
- Detects whether the dashboard is already running; spins it up otherwise; tears it down only when the audit started it.
- Honours `AUDIT_KEEP`, `AUDIT_HEADED`, `AUDIT_ROLE`, `AUDIT_FEATURE`, `AUDIT_MOBILE`, `AUDIT_VISUAL`, `AUDIT_FIX`, `AUDIT_BASE_URL`, `AUDIT_REFERENCE_PATH`, `AUDIT_PREVIOUS_REPORT`.
- Writes artifacts to `tmp/app-audit/<ISO8601>/` with the documented subtree (reference/, current/, diff/, json/, report/, run.log).
- A feature is `passed` only when route opens AND role matches AND content present AND main components present AND main action works AND no critical console error AND no unexpected 4xx/5xx AND layout intact AND screenshot clear AND (when reference exists) visual diff ≤ 5%.
- Generates `audit.pdf` + `audit.html` + `fix-backlog.md` + 7 JSON files (summary, results, visual-diff, layout-metrics, network-errors, console-errors, missing-selectors, feature-map).
- Terminal summary matches the canonical text block.

## Contract

- Go owns the orchestration (Cobra + feature map serializer + result aggregator + HTML renderer + run.log).
- Playwright owns the browser interactions; spec at `web/dashboard/audit/audit.spec.ts` reads `feature-map.json`, drives the browser, writes per-feature results into `json/results.json`.
- PDF rendered by Playwright `page.pdf()` against the generated `audit.html`.
- Visual diff via `pixelmatch` + `pngjs` (added as optional dev-deps).

## Risks

- Playwright + browsers required. Mitigation: graceful degrade — if `playwright` missing, skip visual+functional and emit a `not_implemented` status per feature; report still generates with technical inventory.
- Long runtimes block CI. Mitigation: feature filter via `AUDIT_FEATURE`; mobile + visual opt-in.

## Verification

- `internal/auditrun` ≥ 80% via deterministic golden tests against canned `results.json`.
- `scripts/e2e-phase23.sh` runs the audit in "dry/skip-browser" mode (PLAYWRIGHT_SKIP=1) and asserts every artifact + the expected summary.json shape.
