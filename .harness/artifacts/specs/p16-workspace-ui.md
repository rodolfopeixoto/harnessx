# P16 — Workspace UI · Import Wizard · Stale Detection

## Acceptance

- Dashboard `/workspace` lists projects via DataExplorer + InspectorPanel.
- `harness project import [--path] [--yes]` walks: folder → stack detect → registry add → first index → recommendation. Same logic as future UI wizard.
- `internal/stale.Detect(root, fingerprintPath)` returns `[]StaleEntry{path, kind, hash_before, hash_now}` for `package.json`, `Dockerfile`, `routes.yaml`, `package-lock.json`, `go.mod`.
- HTTP: `GET /api/workspace/projects/:slug/stale`, `POST /api/workspace/import {path, name?, yes?}`.

## Contract

- `internal/importwiz` + `internal/stale` own logic; CLI + UI are renderers calling shared `Plan(opts) → []Step`.
- Stale fingerprints persisted at `.harness/project/fingerprints.json` (project-local, opt-in).

## Risks

- Auto-importing into the registry without explicit `--yes` would surprise the user. Default: CLI prints the plan and prompts; `--yes` skips.

## Verification

- `internal/importwiz`, `internal/stale` ≥ 80%. UI vitest covers Workspace page render + import wizard happy path. `scripts/e2e-phase16.sh` does end-to-end import + stale detect via HTTP.
