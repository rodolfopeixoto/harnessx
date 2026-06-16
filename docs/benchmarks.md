# Benchmarks baseline

Performance budgets per P104 Phase E. Pre-push gate compares against
these floors; regression > 10% blocks push.

Refresh procedure:

```bash
make bench          # raw go test -bench against ./internal/...
make profile-mem    # heap pprof -> dist/profiles/mem.pprof
make profile-cpu    # cpu pprof  -> dist/profiles/cpu.pprof
```

## Command latency budgets

| Command | Budget | Measured how |
|---|---|---|
| `harness route show <prompt>` | < 500 ms | manual `time` |
| `harness scaffold apply python` | < 100 ms (no LLM, FS only) | manual `time` |
| `harness do <prompt>` (plan phase only) | < 1 s | trajectory `WallMs` |
| `harness loop` per-attempt overhead (no LLM) | < 200 ms | devloop checkpoint metric |
| `harness flow apply` per phase overhead | < 50 ms | flowpkg metric |
| `harness sensor run secrets_scan` | < 300 ms typical repo | sensor `WallMs` |
| `harness audit replay <run>` | < 1 s per 10k events | replay log |

## Go bench baseline (per-op alloc/ns)

Filled in by `make bench` output. Placeholder rows seed the table;
real numbers land in PR that introduces the bench file.

| Package | Bench | ns/op | allocs/op | B/op |
|---|---|---|---|---|
| `internal/router` | `BenchmarkPick` | — | — | — |
| `internal/scaffoldpkg` | `BenchmarkApply` | — | — | — |
| `internal/recall` | `BenchmarkBM25Search` | — | — | — |
| `internal/sensors` | `BenchmarkSecretsScan` | — | — | — |
| `internal/flowpkg` | `BenchmarkApply` | — | — | — |

## Heap snapshots

Capture via `profile.Snapshot()` at entry + exit of hot paths.
Allowed regression window: ±10% on `AllocBytes`.

| Scenario | Baseline AllocBytes | Notes |
|---|---|---|
| `harness do` cold start | — | first `internal/profile` integration |
| `harness loop` × 10 attempts | — | check for leaks across attempts |
| `harness flow apply rails-api` | — | full deterministic + LLM mix |

## Memory leak gate

`make profile-mem` runs hot scenarios 100×. Heap diff between run 1
and run 100 must stay within ±10% per scenario. Drift > 10% = leak,
blocks v1.0.0 cut.

## Update cadence

- Refresh baseline numbers each minor release that touches a hot
  path (router, devloop, executor, scaffoldpkg, sensors).
- Reset baseline when intentional change lands (document rationale
  in the spec PR).
