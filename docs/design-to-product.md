# Design-to-product

Ingest a Claude Design ZIP (or folder/prototype), produce **React parity first**, then layer in feature toggles, an MVP roadmap, and API contracts.

## Workflow

```
harness design-to-product "<prompt>" [--source <zip|folder>]
```

If `--source` is omitted, the prompt is mined for a path-shaped token (`./sample-design`, `claude-design.zip`, …). Source resolution falls back to recent artifacts under `.harness/`.

Stages (spec §12):

1. Resolve input (folder/ZIP).
2. Safe-extract ZIPs (rejects path traversal + 200 MiB cap).
3. Inventory pages, components, assets, CSS tokens, JS interactions, missing states, responsive notes.
4. Build `design-manifest.json`.
5. Derive `feature-map.json` (status + priority).
6. Project `toggle-map.json` (runtime-facing).
7. Generate `roadmap.json` (MVP 0–4).
8. Draft `api-contracts.json` (only features that need a backend).
9. Generate `flow-map.json` (page → target navigation graph).
10. Hash + cache images under `.harness/cache/images/<sha256>.json`.

## Feature status rules

| Status | When |
|---|---|
| `disabled` | route exists in design, not in scope |
| `static` | content-only page, no interaction |
| `mock` | interactive UI, no backend required |
| `mock` + `backend_required: true` | UI bound to a future API; ships with mocks |
| `api_contract` | endpoint drafted but not implemented |
| `backend_ready` | endpoint implemented + backend tests pass |
| `production_ready` | real E2E passes |

Backend-required heuristic: page path or interactions include `auth`, `signup`, `login`, `checkout`, `payment`, `settings`, `profile`, `admin`, `dashboard`, `api`, or any `onsubmit` handler.

## Outputs

```
.harness/product/
├── design-manifest.json
├── feature-map.json
├── toggle-map.json
├── roadmap.json
├── api-contracts.json
└── flow-map.json
.harness/cache/images/<sha256>.json   # one per image
```

## Anti-patterns

- Do not invent backend rules from the prototype. API contracts are drafted only when the UI clearly needs one.
- Do not blindly copy prototype code. Generate idiomatic React when the implementation phase runs.
- Do not promote a toggle to `production_ready` without real E2E.
- Do not skip the image cache — repeated runs against the same assets must hit cache.
