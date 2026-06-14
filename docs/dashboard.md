# Dashboard

Local-first read-only dashboard backed by `internal/adapters/http`. The Go binary serves the API + (optionally) a built React SPA. Without the SPA build, a built-in HTML page proves the API is alive.

## Start

```bash
harness dashboard                       # 127.0.0.1:7373
harness dashboard --addr 127.0.0.1:9000
harness dashboard --open                # also opens default browser
```

## REST surface

Read-only, no auth (loopback bind by default).

| endpoint | what |
|---|---|
| `GET /api/health` | `{ok, root, time}` |
| `GET /api/sessions?limit=100` | recent sessions, newest first |
| `GET /api/sessions/{id}` | runs in this session |
| `GET /api/runs/{id}` | sensor results for this run |
| `GET /api/sensors` | last 500 sensor results across all runs |
| `GET /api/agents` | latest cert per adapter id |
| `GET /api/memory` | last 200 memory entries |
| `GET /api/cost` | total + per-agent USD + tokens |
| `GET /api/logs?tail=N` | last N JSONL events |
| `GET /api/profile` | project profile JSON |
| `GET /api/design` | design manifest JSON |
| `GET /api/roadmap` | MVP roadmap JSON |
| `GET /api/features` | feature map JSON |
| `GET /api/toggles` | toggle map JSON |

## React SPA

```bash
make dashboard-install
make dashboard-build
harness dashboard   # serves web/dashboard/dist automatically
```

Routes:

- `/` Sessions
- `/sessions/:id` Run timeline
- `/runs/:id` Sensor results
- `/sensors` Recent sensor runs across all sessions
- `/agents` Adapter cert status
- `/design` Design manifest
- `/roadmap` MVP roadmap
- `/memory` Project memory
- `/settings` Health + project profile

Every page surfaces loading / empty / error states via `useFetched` + `PanelState`.

## TUI

```bash
harness logs --follow   # Bubble Tea: 750 ms poll, q/esc/Ctrl-C to quit
```

Tail mode: `harness logs --tail 50` (no TUI).

## Security

- Bind to `127.0.0.1` unless network is trusted.
- API is read-only; mutations go through the CLI commands which run through the same telemetry/audit layer.
- The built-in HTML fallback page is safe to serve over a restricted network — it embeds no secrets and uses only relative API paths.
