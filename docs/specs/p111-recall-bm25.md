# P111 — Long-term memory: BM25 + embedding interface

## Context

`internal/recall` uses bag-of-words (v0.33). Paper § 3.2 long-term
memory should support stronger retrieval. Add BM25 as the default
algorithm (no LLM dep) plus a `Scorer` interface so an embedding
backend can drop in later without touching callers.

## What ships

- `internal/recall/scorer.go`: `Scorer` interface
  `Score(query, doc string) float64`.
- `internal/recall/bm25.go`: BM25 scorer (k1=1.5, b=0.75 defaults,
  tunable via `BM25.Tune(k1, b)`).
- `internal/recall/recall.go::Recall` accepts optional `scorer`;
  defaults to bag-of-words to preserve behaviour.
- Embedding backend stays out of scope (interface only); `recall.go`
  documents how to plug one in.

## Critical files

| Path | Change |
|---|---|
| `internal/recall/scorer.go` (new) | interface + default factories |
| `internal/recall/bm25.go` (new) | BM25 implementation |
| `internal/recall/bm25_test.go` (new) | per-term IDF + ranking |

## Reuse, do not reinvent

- `recall.tokenise` already exists — BM25 reuses it.
- `recall.stopWords` already exists — BM25 inherits stop list.

## Verification

- `make lint` 0 issues
- `go test ./internal/recall/...` — BM25 ranking matches expectation
  (longer doc with same term hits scores lower)
- Coverage ≥ 90% for new files

## Risks + mitigation

| Risk | Mitigation |
|---|---|
| BM25 introduces dep | None — pure Go math, no external pkg |
| Default behaviour change breaks callers | `Recall` keeps existing default; opt-in BM25 via new `RecallWith(scorer)` |

## Acceptance

- BM25 scores >0 for matching terms, 0 for none
- Length normalization works (longer doc same terms < shorter doc)
- Interface allows mock scorer in tests
