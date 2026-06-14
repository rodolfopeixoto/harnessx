# Architecture

## Layered, dependency-inward

```
cmd/harness          CLI entrypoint (Cobra)
internal/app         use-case orchestration (init, doctor, log tailing, …)
internal/domain      pure types: Session, Run, Sensor, Agent, ContextPack
internal/adapters    SQLite repo, JSONL logger, exec probe, (future) LSP, agents
internal/platform    config, paths, ids, hashing, clock
internal/sensors     deterministic checks (Phase 4)
internal/agents      adapter contract + fake + YAML loader (Phase 3)
internal/router      deterministic agent selection (Phase 3+)
internal/context     context pack builder (Phase 5)
internal/index       project profile/maps (Phase 2)
internal/ui          Lip Gloss views, future Bubble Tea TUI
web/dashboard        React + Vite + TS (scaffold; full IA in Phase 8)
```

`domain` imports nothing. `app` imports `domain` plus interfaces it owns.
`adapters` implement those interfaces. Tests substitute fakes at the seam.

## Roadmap

- **Phase 0** — repo scaffolding, Makefile, CI, dashboard scaffold. ✅
- **Phase 1** — core CLI: `init`, `doctor`, `logs`, `version`. ✅
- **Phase 2** — project index.
- **Phase 3** — agent adapters (Claude, Codex, Gemini, Kimi, YAML loader, certification).
- **Phase 4** — sensor system + rule packs.
- **Phase 5** — context engineering + LSP.
- **Phase 6** — spec + plan workflow.
- **Phase 7** — design-to-product workflow.
- **Phase 8** — dashboard.
- **Phase 9** — resource optimization.
- **Phase 10** — full end-to-end flow.

Each later phase ships its own commands, sensors, and tests. Phase 1
registers every command from `cli-reference.md` so the help surface is
stable; unimplemented commands exit with code 2 and a phase pointer.
