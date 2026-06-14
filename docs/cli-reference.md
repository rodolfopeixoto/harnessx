# CLI reference

## Implemented (Phase 1)

| Command | Description |
|---|---|
| `harness init` | Create `.harness/` in the project root. |
| `harness doctor` | Probe toolchain and agent CLIs; render status panel. |
| `harness logs [--tail N]` | Print last N JSONL events (default 50). |
| `harness version` | Print version, commit, build date. |

## Registered but stubbed

The following commands print `not yet implemented (planned for Phase N)`
and exit 2. They are listed so the help surface is stable across phases.

| Command | Planned phase |
|---|---|
| `harness run "<prompt>"` | 6 |
| `harness ask "<question>"` | 6 |
| `harness plan "<prompt>"` | 6 |
| `harness feature "<prompt>"` | 6 |
| `harness bugfix "<prompt>"` | 6 |
| `harness design-to-product "<prompt>"` | 7 |
| `harness optimize resources` | 9 |
| `harness check` | 4 |
| `harness ci` | 4 |
| `harness dashboard` | 8 |
| `harness report --last` | 6 |
| `harness project index\|inspect` | 2 |
| `harness context build\|inspect` | 5 |
| `harness agent list\|add\|discover\|certify` | 3 |
| `harness sensor list\|run` | 4 |
| `harness perf-snapshot\|perf-compare` | 9 |
| `harness image-audit\|dependency-audit\|log-audit\|security-audit` | 9 |
