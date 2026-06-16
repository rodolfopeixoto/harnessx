# P114 — Multimodal grounding

## Context

Paper § 5.2.6. `harness do --image` auto-routes to vision adapter
(v0.33). Missing: verify the generated output actually references the
annotated regions of the image. Add a grounding checker.

## What ships

- `internal/multimodal/grounding.go`:
  - `Annotation{Region:{X,Y,W,H}, Label}` describes annotated regions
  - `LoadAnnotations(path)` reads sidecar `.image.json`
  - `CheckGrounding(text, anns) GroundingResult{Hits, Missing}`
- Default sidecar format: `{schema_version:1, annotations:[{region,label}]}`
- No new sensor wired yet — primitive only.

## Critical files

| Path | Change |
|---|---|
| `internal/multimodal/grounding.go` (new) | types + checker + loader |
| `internal/multimodal/grounding_test.go` (new) | label matching + sidecar load |

## Reuse, do not reinvent

- Standard `encoding/json` for sidecar
- `strings.Contains` (case-insensitive lower) for label hit detection

## Verification

- `make lint` 0 issues
- `go test ./internal/multimodal/...` ≥ 90%

## Risks + mitigation

| Risk | Mitigation |
|---|---|
| Label match too lax | Stem labels + check substring; document expectations |
| Sidecar JSON malformed | Loader returns error; sensor caller handles gracefully |

## Acceptance

- LoadAnnotations round-trips JSON
- CheckGrounding reports each label as hit or missing
- 100% labels → no Missing; 0% → all Missing
