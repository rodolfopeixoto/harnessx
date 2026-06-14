# P26 — UI Handoff Gap Report

Reference handoff: `docs/design-handoff-v2/harness-cli-ui/project/` (extracted from `harness-cli-ui-handoff-2.zip`).

## 1. Handoff screen inventory (from `screen-*.jsx` + `app.jsx` nav)

| id | file | purpose |
|---|---|---|
| home | screen-home.jsx | Project dashboard: run summary, sensors, agents, cost. |
| onboarding | screen-onboarding.jsx | First-run wizard: pick project, detect stack, certify agents. |
| projects | screen-projects.jsx | Workspace hub: every registered project + switch. |
| command | screen-command.jsx | Natural-language prompt → mode detection → start run. |
| plan | screen-plan.jsx | Spec/plan preview + approval gate. |
| activerun | screen-run.jsx | Live run: stages, agent IO, sensors, diff, cost. |
| design | screen-design.jsx | Design-to-Product workflow (ZIP → React parity → toggles). |
| roadmap | screen-roadmap.jsx | MVP roadmap + feature toggle map. |
| reports | screen-reports.jsx | Report archive, filters by project/period. |
| agents | screen-agents.jsx | Adapter health + certification + fallback chain. |
| sensors | screen-sensors.jsx | Sensor catalog + last results + fix flow. |
| catalog | screen-catalog.jsx | Capabilities Center (MCPs · Hooks · Sensors · Context). |
| context | screen-context.jsx | Context pack inspector + token cost. |
| memory | screen-memory.jsx | Project memory + evidence chain. |
| resource | screen-resource.jsx | Resource optimization: caches/worktrees/images. |
| settings | screen-settings.jsx | Global + project settings tabs. |
| stakeholder | screen-stakeholder.jsx | Non-technical readout: ready/blocked/mock/cost. |

Total handoff screens: **17**.

## 2. Current implementation route inventory (web/dashboard/src/App.tsx)

| route | page component | notes |
|---|---|---|
| / | SessionsPage | List of recent sessions. Maps to "home" approximately. |
| /sessions/:id | SessionDetailPage | Drill-down for one session's runs. |
| /runs/:id | RunDetailPage | Sensors per run. |
| /sensors | SensorsPage | Sensor results list. |
| /agents | AgentsPage | Agent certifications. |
| /design | DesignPage | Reads /api/design (design manifest). |
| /roadmap | RoadmapPage | Reads /api/roadmap. |
| /memory | MemoryPage | Memory entries. |
| /settings | SettingsPage | Health + profile JSON dump. |

Total current routes: **9** (counting detail routes).

## 3. Missing screens (need new route + page component)

| handoff id | proposed route | priority |
|---|---|---|
| onboarding | `/onboarding` | P1 (first-run flow) |
| projects | `/projects` | P0 (workspace switching) |
| command | `/command` | P0 (start a run) |
| plan | `/plan` | P0 (plan approval) |
| activerun | `/run` | P0 (live run telemetry) |
| catalog | `/catalog` | P0 (capabilities) |
| context | `/context` | P1 (context pack inspector) |
| resource | `/resources` | P1 (cleanup visibility) |
| reports | `/reports` | P1 (report archive) |
| stakeholder | `/stakeholder` | P2 (non-technical view) |
| cleanup | `/cleanup` | P1 (cleanup engine UI) |

## 4. Screens implemented partially

| handoff | current | gap |
|---|---|---|
| home | `/` (SessionsPage) | Shows sessions table only. Handoff Home has cost, sensors summary, next-action card. |
| sensors | SensorsPage | Lists rows. Handoff has per-sensor Inspector w/ Overview/Output/Files/History/Configuration/Fix Plan/Audit tabs. |
| agents | AgentsPage | Lists certifications. Handoff has fallback chain, performance, adapter config. |
| design | DesignPage | JSON dump. Handoff has structured page/component/asset/flow maps + parity status. |
| roadmap | RoadmapPage | List of phases. Handoff has feature toggle map + MVP roadmap visual. |
| memory | MemoryPage | List of memories. Handoff has evidence chain + promote/retire/protect actions. |
| settings | SettingsPage | Reads /api/health + /api/profile. Handoff splits Global vs Project tabs. |

## 5. Components missing from current design system

- InspectorPanel — exists (`ds/InspectorPanel.tsx`) but only used by DS tests. Pages still render bespoke tables.
- DataExplorer — exists; only wired into SessionsPage indirectly.
- AutocompleteSearch — missing.
- FilterBar — missing.
- Pagination — embedded inside DataExplorer; not standalone.
- PathCell — missing.
- EntityStatusBadge — present as `Badge` but no entity-specific variants.
- BulkActionBar — missing.
- SavedViewTabs — missing.
- TerminalReflection — missing.
- ConfigDiffPreview — missing.
- InstallConfigureWizard — missing.

## 6. Interactions that are toast-only / not stateful

- Catalog install/configure (backend OK but UI dashboard page is empty placeholder).
- Cleanup apply (CLI ok, UI absent).
- Memory promote/demote (UI absent).
- Sensor re-run / fix (UI absent).
- Agent certify (UI absent).

## 7. Drawers/panels that break dense workflows

Current dashboard renders pages directly in <main>. No drawer in production code, so the "broken drawer" exists only in the handoff prototype prior to InspectorPanel adoption. Risk: replicating the prototype's narrow Drawer is forbidden — we ship InspectorPanel directly.

## 8. Tables/lists missing search/pagination/autocomplete

| page | search | filter | pagination | autocomplete |
|---|---|---|---|---|
| Sessions | no | no | no | no |
| Sensors | no | no | no | no |
| Agents | no | no | no | no |
| Memory | no | no | no | no |
| Reports | route missing | — | — | — |
| Catalog | route missing | — | — | — |

DataExplorer must wrap every list page.

## 9. APIs/endpoints required by each screen

| screen | needs |
|---|---|
| projects | `GET /api/workspace/projects` (exists), `POST /api/workspace/switch` (exists). |
| command | `POST /api/run/start` (missing). |
| plan | `GET /api/plan/:run_id` (missing). |
| activerun | SSE `/api/events?run_id=` (missing). |
| catalog | `GET /api/catalog/items` (exists), `POST /api/catalog/plan` (exists), `POST /api/catalog/install` (missing). |
| context | `GET /api/context/:project` (missing). |
| memory | `GET /api/memory` (exists), `POST /api/memory/:id/promote` (missing). |
| resource | `GET /api/cleanup/scan` (missing — only CLI). |
| reports | `GET /api/reports` (missing). |
| stakeholder | aggregated `/api/stakeholder/summary` (missing). |

## 10. State models required by each screen

Per spec §8 of P26: explicit lifecycle on `sensor`, `mcp`, `hook`, `memory`, `run`. Currently only `run` (via existing `Status` enum) is modelled.

## 11. Test gaps

- Audit feature-map covers 10 routes; needs all 17 handoff screens + every required interaction (open inspector, search, paginate, install plan, enable hook, promote memory).
- Visual diff vs handoff: not implemented (reference screenshots never captured).
- Role × page × interaction matrix: only role × page render covered (P21).

## 12. Audit gaps

- No reference screenshot capture command.
- No pixelmatch diff in `internal/auditrun`.
- Severity counts in `summary.json` count all feature priorities, not failure priorities — misleading.
- CLI flows don't yet exercise mcp/hook/cleanup/policy subcommands missing CLI surface (mcp scan/list/inspect/plan/apply/validate, hook enable/disable/test).

## 13. Plan for P26 implementation (executed this PR; remaining tracked as P27+)

P26 PR (this branch) ships:
- This gap report.
- Stub pages for every missing route with stable data-testid (`page-<name>`).
- Shell navigation updated to expose every handoff screen.
- Audit feature-map upgraded to all 17 screens.
- `bin/stack audit-reference` command to capture handoff reference screenshots.
- `internal/auditrun/visualdiff` package with pixelmatch-style PNG diff.

Deferred to P27/P28:
- Full Inspector tabs per entity (Sensor/Agent/MCP/Hook/Memory/Context).
- MCP/Hook deterministic scanners (Phase 6/7).
- TerminalReflection + ActionService + AuditEventService (Phase 8/9).
- ConfigDiffPreview + InstallConfigureWizard.
- Stakeholder summary aggregation.
- SSE for activerun.
