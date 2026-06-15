# P29 — finalize handoff

Closes the 8 remaining stub routes from P26-P28 with rich content,
wires MCP/Hook scanners into Catalog tabs, adds visual-diff scaffold.

## Deliverables

1. Rich content + audit-required selectors for: `/command`, `/plan`,
   `/run`, `/context`, `/resources`, `/reports`, `/stakeholder`,
   `/onboarding`.
2. Catalog page reads `/api/mcps` and `/api/hooks` to populate the
   mcp + hook tabs with real scanner output.
3. Each page surfaces a demo banner when API returns empty (visible
   "demo" label).
4. Audit feature map updated to require page-X + at least one
   page-specific selector for every page.
5. Stub.tsx removed (no remaining StubPage usage).

## Deferred to P30

- Pixelmatch visual diff against handoff (handoff prototype is HTML/JS
  + reactive, not directly screenshot-comparable without a captured
  reference fixture set).
- Inspector tabs per entity (Sensor 7/Agent 7/MCP 7/Hook 6/Memory 5/
  Context 7).
- ActionService backend persistence (still localStorage).
- Settings Global vs Project split.
