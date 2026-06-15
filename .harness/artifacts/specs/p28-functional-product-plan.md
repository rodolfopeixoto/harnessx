# P28 — functional handoff dashboard + scanners

## Current diagnosis

P27 closed Home/Projects/Catalog with real backend wiring and audit-required selectors. 9/17 handoff screens still render `StubPage` (Command/Plan/ActiveRun/Context/Resources/Cleanup/Reports/Stakeholder/Onboarding). Visual diff stays `null`. MCP/Hook deterministic scanners exist as backend `catalog/kinds` discovery but no CLI surface and no dashboard tab.

## What P28 PR delivers (this branch, honest scope)

1. `internal/mcpscan` — deterministic MCP config scanner across `.harness/mcp/**`, `**/*mcp*.{json,yml,yaml}`, `.claude/**`, `.codex/**`, `.gemini/**`, `.kimi/**`. Typed `McpServer`/`McpConfigFile` + `Scan(root)`.
2. `internal/hookscan` — deterministic hook scanner across `.harness/hooks/**`, `scripts/git-hooks/*`, `.claude/hooks/**`, `.codex/hooks/**`. Typed `Hook` + `Scan(root)`.
3. CLI: `harness mcp scan|list`, `harness hook scan|list`. Both honour `--json` for machine output.
4. HTTP: `GET /api/mcps`, `GET /api/hooks`, `GET /api/cleanup/scan`. Wired in `internal/adapters/http`.
5. Cleanup page rewritten with `DataExplorer` over `/api/cleanup/scan` + risk badges + PathCell + plan-only banner (no apply from UI without policy file).
6. 4 new DS components: `AutocompleteSearch`, `ConfigDiffPreview`, `WorkflowTimeline`, `StatefulActionButton`.
7. Audit hardened: `/sensors`, `/agents`, `/catalog` now require richer selectors. `/cleanup` requires `cleanup-explorer + cleanup-plan-banner`. Placeholder pages fail.
8. Visual diff scaffold: `internal/auditrun/visualdiff` package + `harness stack audit-reference` command that uses Playwright to capture each handoff `screen-*.jsx` into `reference/screenshots/`. Pixelmatch comparison wired into runner — `visual-diff.json` non-null when `AUDIT_VISUAL=1`.
9. Tests for `mcpscan`, `hookscan`, `cleanupcmd` API handler.
10. Updated `e2e-phase24` to validate `cleanup-explorer` + MCP/Hook scan endpoints.

## Honest deferrals (P29+)

- Full Inspector tabs per entity (Sensor 7 / Agent 7 / MCP 7 / Hook 6 / Memory 5 / Context 7).
- Rich content for: Command, Plan, ActiveRun, Context, Resources, Reports, Stakeholder, Onboarding (still `StubPage` after this PR — audit only requires `nav-X` for them).
- Backend persistence for ActionService (currently localStorage only).
- Settings Global vs Project split.
- Real `/design` state machine + feature toggle rules.
- ConfigDiffPreview wired into actual install/apply flows (component lands here, callers come in P29).
- Sample data fixtures for every screen.
- Cross-tab event sync.

## Implementation order

1. Spec (this file).
2. Go scanners: `internal/mcpscan`, `internal/hookscan`.
3. CLI: `harness mcp`, `harness hook`.
4. HTTP: `/api/mcps`, `/api/hooks`, `/api/cleanup/scan`.
5. DS components.
6. Cleanup page.
7. Audit feature-map hardening.
8. Visual diff scaffold + reference capture command.
9. Tests + e2e + commit + push.

## Rollback plan

Single branch, single PR. Revert via `git revert <merge-sha>` on `develop`.
