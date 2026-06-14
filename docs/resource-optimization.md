# Resource optimisation

Reduce memory, CPU, I/O, image size, bundle size, dependency weight, and log noise **without degrading behaviour, security, observability, debug capability, or developer experience.**

Core rule (do not violate):

> Remove only when there is evidence a dependency/log/file/tool is unnecessary for runtime, build, test, security, observability, recovery, or debugging. When uncertain, keep and document as `"kept for operational safety"`.

## Cycles (spec §21)

| Cycle | What | Command |
|---|---|---|
| A — measurement | Baseline snapshot | `harness perf-snapshot --label baseline --report` |
| B — image audit | Static Dockerfile analysis | `harness image-audit` |
| C — dependency audit | Classify + flag removal candidates conservatively | `harness dependency-audit` |
| D — log audit | Noisy call-site detection | `harness log-audit` |
| E — runtime audit | (deferred — needs Docker stats) | — |
| F — performance budget | Sensor-enforced | (Phase 9.5) |
| G — report | Before / after / delta / status | `harness perf-compare` |
| meta | Run A→G in order | `harness optimize resources` |

## Snapshot

`harness perf-snapshot` captures:

- Project metadata (name, detected stacks).
- Dockerfile metrics (base image, stages, RUN/COPY counts, USER, HEALTHCHECK, cache cleanup, latest-tag use).
- Dependency totals per ecosystem + removal candidates + kept-for-operational-safety entries.
- Noisy log call sites (console.log/.debug/.info, puts, println!/print!, fmt.Println/Printf, print).
- Disk usage (`.harness/` + project bytes, with `node_modules`/`.git`/`target`/`dist`/`build` excluded).

The result is written to `.harness/artifacts/perf/<ts>-<id>.json`.

## Compare

`harness perf-compare [from] [to]` diffs two snapshots. With no args it picks the two most recent. Each numeric metric carries a status: `improved` (delta < 0), `regressed` (delta > 0), or `unchanged`.

Report lands at `.harness/artifacts/reports/perf-compare-<id>.md`.

## Dockerfile findings

| id | severity | meaning |
|---|---|---|
| `docker.latest_tag` | warn | base image uses `:latest` or no tag |
| `docker.no_user` | warn | no USER directive — container runs as root |
| `docker.no_healthcheck` | info | no HEALTHCHECK declared |
| `docker.no_cache_cleanup` | warn | RUN steps install packages but never clear caches |
| `docker.single_stage_heavy` | info | single-stage build with many COPYs |

## Conservative dependency classification

`removalCandidate` only flags **obvious dev-tool duplicates** in the runtime ecosystem section. `keepReason` documents observability (`@sentry/node`, `prometheus`, etc.), security (`helmet`, `argon2`, etc.), and debugging (`pry`, `delve`, etc.) as kept for operational safety.

## Anti-patterns

- Do not remove a dependency the moment it looks unused. Sit on the candidate list for a release cycle before pulling the trigger.
- Do not remove a log line because it's noisy at startup. Move it behind an env flag instead.
- Do not delete security or observability tooling to shrink an image. The image is cheaper than an outage.
- Do not optimise without a baseline. Capture Cycle A first, every time.
