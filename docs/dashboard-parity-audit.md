# Dashboard parity audit

Cross-reference between the operator-supplied
`harness-cli-ui-handoff-2.zip` Figma export and the React app shipped
under `web/dashboard/src/pages/`.

Date: 2026-06-16.

## Pages in the handoff zip (18 screens, design-only)

```
Home Projects Run Plan Design Roadmap Agents Capabilities Sensors
Context Memory Resources Reports Stakeholder Settings Onboarding
Command Catalog
```

## Pages in repo (React)

```
ActiveRun Agents Backup Capabilities Catalog Cleanup Command
Containers Context Design Home Images Install Memory Onboarding Plan
Projects Reports Resources Roadmap Run RunDetail Runtime Secrets
Sensors SessionDetail Sessions Settings Stakeholder Stub
```

29 pages — superset of the handoff zip plus operational surfaces
(Backup, Cleanup, Containers, Images, Install, Runtime, Secrets,
Sessions).

## Per-screen status

| Screen | Handoff | Repo | Status |
|---|---|---|---|
| Home | ✓ | ✓ | covered |
| Projects | ✓ | ✓ | covered |
| Run | ✓ | ✓ | covered |
| Plan | ✓ | ✓ | covered |
| Design | ✓ | ✓ | covered |
| Roadmap | ✓ | ✓ | covered |
| Agents | ✓ | ✓ | covered |
| Capabilities | ✓ | ✓ | covered |
| Sensors | ✓ | ✓ | covered |
| Context | ✓ | ✓ | covered |
| Memory | ✓ | ✓ | covered |
| Resources | ✓ | ✓ | covered |
| Reports | ✓ | ✓ | covered |
| Stakeholder | ✓ | ✓ | covered |
| Settings | ✓ | ✓ | covered |
| Onboarding | ✓ | ✓ | covered |
| Command | ✓ | ✓ | covered |
| Catalog | ✓ | ✓ | covered |
| ActiveRun | — | ✓ | extra (operational) |
| Backup | — | ✓ | extra |
| Cleanup | — | ✓ | extra |
| Containers | — | ✓ | extra |
| Images | — | ✓ | extra |
| Install | — | ✓ | extra |
| Runtime | — | ✓ | extra |
| Secrets | — | ✓ | extra |
| Sessions | — | ✓ | extra |
| RunDetail | — | ✓ | extra |
| SessionDetail | — | ✓ | extra |
| Stub | — | ✓ | placeholder |

## Decision

**Repo already exceeds zip parity in coverage** (18/18 handoff screens
present plus 11 extra operational pages). Remaining work is visual
fidelity — not a missing-screen problem.

## Visual fidelity gap (deferred)

The zip ships a specific design system (typography, colour palette,
spacing, navigation chrome). Repo uses a working theme but not the
zip's. Visual port is descoped from the v0.x line; tracked in
`docs/dashboard-scope.md` (v0.48).

## Next

Visual fidelity work re-opens when:

- An operator commits to driving 80%+ of their workflow through the
  UI
- A 2-week sprint can be funded for the port
- The handoff zip is updated to match the v0.32+ surface
  (`harness do`, `harness loop`, scaffolds, prune, JSON output)

Until then: CLI is the canonical interface; dashboard is a read-only
companion view (per `docs/dashboard-scope.md`).
