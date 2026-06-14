# P15 — Design System port (InspectorPanel · DataExplorer · primitives)

## Acceptance

- `web/dashboard/src/ds/`: tokens.ts + Badge, Card, Banner, Drawer, EmptyState, Toolbar, Tabs, Terminal, InspectorPanel, DataExplorer, Shell.
- Every primitive is typed (`*.tsx`) and has a co-located `*.test.tsx`.
- Existing pages render through DS primitives; behaviour unchanged.
- No hardcoded colours; everything reads `tokens.ts`. No raw text literals — `ds/strings.ts` holds shared copy.

## Contract

- InspectorPanel API: `{title, subtitle, tabs[{id,label,render}], footer, onClose}`. Sticky header + footer; body scrolls. Full-screen below 768px.
- DataExplorer API: `{items, columns, searchKeys, filters, sort, pageSize, autocomplete(query), onInspect, bulkActions, emptyState}`.

## Risks

- Refactor regressions on existing pages. Mitigation: wrap, don't rewrite. Vitest renders every page before + after.

## Verification

- Vitest covers each DS primitive (interaction + render) and every existing page still passes.
- `scripts/e2e-phase15.sh` builds dist, starts dashboard, asserts `/api/health` + at least one DS-rendered route over HTTP.
