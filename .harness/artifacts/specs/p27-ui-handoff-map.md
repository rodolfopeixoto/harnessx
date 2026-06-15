# P27 — UI Handoff functional parity map

Source handoff: `docs/design-handoff-v2/harness-cli-ui/project/`.

## Handoff → current route map

| handoff file | route | status before P27 | status after P27 |
|---|---|---|---|
| screen-home.jsx | `/` | rendered SessionsPage (table only) | rich Home: workspace summary + cards + recent runs |
| screen-projects.jsx | `/projects` | StubPage | DataExplorer over `/api/workspace/projects` + PathCell |
| screen-command.jsx | `/command` | StubPage | structured prompt form + intent preview |
| screen-plan.jsx | `/plan` | StubPage | placeholder w/ workflow timeline scaffold |
| screen-run.jsx | `/run` | StubPage | timeline scaffold + terminal reflection |
| screen-design.jsx | `/design` | DesignPage (json dump) | feature toggle list (existing) — deferred polish |
| screen-roadmap.jsx | `/roadmap` | RoadmapPage | unchanged this PR |
| screen-agents.jsx | `/agents` | AgentsPage | DataExplorer wrapper added |
| screen-catalog.jsx | `/catalog` | StubPage | DataExplorer over `/api/catalog/items` + kind tabs |
| screen-sensors.jsx | `/sensors` | SensorsPage | DataExplorer wrapper added |
| screen-context.jsx | `/context` | StubPage | metric scaffold (Token cost summary placeholder) |
| screen-memory.jsx | `/memory` | MemoryPage | DataExplorer wrapper added |
| screen-resource.jsx | `/resources` | StubPage | placeholder, kept |
| screen-reports.jsx | `/reports` | StubPage | placeholder, kept |
| screen-settings.jsx | `/settings` | SettingsPage | Global/Project tabs scaffold |
| screen-stakeholder.jsx | `/stakeholder` | StubPage | placeholder, kept |
| screen-onboarding.jsx | `/onboarding` | StubPage | placeholder, kept |

## What P27 PR delivers (this branch)

1. Real content + DataExplorer wired into the 4 highest-value pages: **Home, Projects, Catalog, Sensors**.
2. New `web/dashboard/src/lib/{actions,terminal}.ts` — `ActionService` + `TerminalReflectionService` (frontend-only event log; backend audit feed defers to P28).
3. New `web/dashboard/src/ds/{TerminalReflection,MetricCard,PathCell}.tsx` primitives.
4. Audit feature-map hardened: per-page required selectors (sensors-explorer, projects-explorer, capabilities-tabs, recent-runs, workspace-summary, health-score-card, next-action-card) — placeholder-only pages now fail.
5. API endpoints removed from screenshot feature list — moved into separate `api` category that doesn't capture screenshots.
6. `web/dashboard/src/lib/demo.ts` — visibly labelled sample data when API returns empty (banner: "demo data — `harness project add` for real").
7. Visual diff scaffold: `internal/auditrun/visualdiff/diff.go` + capture command via Playwright. Full pixelmatch deferred to P28 (requires `pngjs` dep).

## Explicitly deferred to P28+

Honest list (rastreado pra próxima iteração):

- Phase 3 full Inspector tabs per entity (Sensor 7 tabs, Agent 7 tabs, MCP 7 tabs, Hook 6 tabs, Memory 5 tabs, Context 7 tabs).
- Phase 5 MCP/Hook deterministic scanners with CLI surface (`harness mcp scan|list|inspect|plan|apply|validate`, `harness hook scan|enable|disable|test`).
- Phase 4 full stateful mutation backend: ActionService writes only to frontend log this PR; persistence + cross-tab sync deferred.
- Phase 7 sample data fixtures for every screen.
- Phase 8 E2E for every interaction.
- Phase 12 Design-to-Product full state machine.
- Phase 13 multi-project Health score + stale detection UI.
- Visual diff real pixelmatch (reference capture command lands here; comparison code lands in P28).

## Audit upgrades shipped

- per-page required selector matrix expanded
- placeholder pages now classified as `selector_missing` instead of `passed`
- API features moved to category=`api` with no screenshot expectation
- `AUDIT_VISUAL=1` calls the visualdiff package; emits `visual-diff.json` non-null
