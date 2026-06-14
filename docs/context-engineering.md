# Context engineering

HarnessX never sends a whole repo to an LLM. Every agent call sees a **context pack** assembled from cheap, deterministic providers, hashed for cache reuse, and bounded by file size.

## Provider order (spec §14)

1. `memory` — top-N evidence-backed memories from sqlite.
2. `git` — `status --porcelain` + `diff HEAD`. Changed paths are promoted into `RelevantFiles`.
3. `ripgrep` — keyword search seeded by the task prompt (stop words dropped, max-count per keyword capped).
4. `lsp` — symbols + diagnostics + definitions + references; one cache file per `(repo-hash, language, query-hash)` under `.harness/cache/lsp/`.
5. `test_map` — links from changed source files to suites in `.harness/project/test-map.json`.

LSP is plug-in: the `lsp.Client` interface is in place; per-server clients (gopls, ruby-lsp, pyright, rust-analyzer, typescript-language-server) land as a follow-up.

## Cache

```
.harness/cache/context/<hash>.json
```

`hash` is `sha256(task + project_profile + provider_names + git_HEAD)`. A second `harness context build "<same task>"` returns the cached pack with `cache_hit=true` (build duration 0 ms).

## File enrichment

For each entry under `RelevantFiles`, the Builder records `Bytes`, `SHA256`, and `EstimatedTokens`. Files larger than 256 KiB are kept in the manifest (path + size) but not loaded — the agent should request them on demand.

Token estimation lives in `internal/platform/tokens`. The default is a 4-chars-per-token heuristic; provider-specific tokenizers plug in by satisfying `tokens.Estimator`.

## Commands

```bash
harness context build "<task>" [--force]   # build (or reuse cache)
harness context inspect [<hash>]           # pretty-print newest cached pack
```

## Anti-patterns

- Do not send entire repos to an LLM. Always go through `internal/context.Build`.
- Do not query LSP without cache.
- Do not silently truncate the pack — fail with a clear error when over budget.
- Do not include files outside the project root (the pack is project-local).
