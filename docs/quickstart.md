# Quickstart

```bash
cd your-project
harness init       # creates .harness/ with SQLite + default config
harness doctor     # probes toolchain and agent CLIs
harness logs       # tails .harness/logs/events.jsonl (last 50 entries)
```

That is everything Phase 1 implements. Any other command will print:

```
command <name> is not yet implemented (planned for Phase N)
```

See `architecture.md` for the phase plan.
