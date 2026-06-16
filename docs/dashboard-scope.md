# Dashboard scope decision

Status: **descoped from v0.x roadmap**. Dashboard parity with the
operator-supplied `harness-cli-ui-handoff-2.zip` (Figma prototype) is
intentionally **not** a v0.x deliverable. This file pins the rationale
and the contract that survives.

## What ships

- `harness dashboard --addr :7373`
- Serves built-in HTML by default
- When `web/dashboard/dist/` exists (`make dashboard-build`), serves
  the React UI from disk

Pages currently in the React app (`web/dashboard/src/pages/`):

```
Home Projects Run Plan Design Roadmap Agents Capabilities Sensors
Context Memory Resources Runtime Containers Images Install Secrets
Backup Cleanup Reports Stakeholder Settings ActiveRun Sessions
```

23+ pages — exceeds the 18-screen handoff zip.

## What does not ship (and why)

The handoff zip is a Figma export with 26 `.jsx` files designed to
specify visual + interaction behaviour. Porting it 1:1 means:

- Re-implementing 18 screens against the actual `/api/*` endpoints
- Reconciling 5+ screens absent from the zip (ActiveRun, Backup,
  Cleanup, Containers, Images, Install, Runtime, Secrets, Sessions)
  with the design language
- Building a shared component library (`ds/`) the zip references but
  does not export

Operator-week estimate: 4–8 weeks of focused frontend work to reach
1:1 parity. HarnessX value proposition is the CLI + agent pipeline,
not the dashboard. The CLI surface already covers every operator
workflow.

## Contract that survives

- `harness dashboard --addr :PORT` always works
- The REST API at `/api/*` is the stable contract for any future UI
  (zip-derived or otherwise); see `docs/json-schemas.md` for related
  output contracts
- React build is opt-in via `make dashboard-build`; binary stays small
  when not built

## If the dashboard becomes critical

Re-open the decision when:

- An operator commits to driving 80%+ of their workflow through the
  UI (today every operator runs CLI)
- A 2-week sprint can be funded for the port
- The handoff zip is updated to match the v0.32+ surface
  (`harness do`, `harness loop`, scaffolds, prune, JSON output)

Until then: CLI is the canonical interface. Dashboard is a read-only
companion view.
