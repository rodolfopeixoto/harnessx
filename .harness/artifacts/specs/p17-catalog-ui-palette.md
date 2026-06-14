# P17 — Catalog UI · Command Palette

## Acceptance

- `internal/palette` builds an in-process searchable index across: projects, capabilities (every kind), sensors, runs, cleanup findings, settings, commands.
- `GET /api/palette?q=` returns up to N matches grouped by source.
- Dashboard `/catalog` page renders DataExplorer over `/api/catalog/items`; row click opens InspectorPanel with kind details + Install Plan tab.
- ⌘K (Ctrl+K) opens CommandPalette UI; arrow keys navigate; Enter routes.
- `harness palette search <q>` mirrors the HTTP endpoint for terminal use.

## Contract

- `palette.Source` interface: `Name() string` + `Search(ctx, q) []palette.Hit`. Sources registered explicitly.
- `palette.Hit{Source, Kind, Title, Subtitle, RouterPath, Score}`.

## Verification

- `internal/palette` ≥ 85% (table tests across every source).
- React tests cover Catalog page render + CommandPalette open/close/select.
- `scripts/e2e-phase17.sh`: HTTP smoke + CLI search.
