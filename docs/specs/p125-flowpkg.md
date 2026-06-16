# P125 — Flow registry foundation

## Context

Phase C of P104 plan. Operators need domain-agnostic deterministic
workflows: rails-api scaffold, meta-ads-campaign builder,
content-pipeline. Each is a sequence of phases mixing deterministic
shell, harness scaffolds, harness do, sensors.

This release lands the foundation: `internal/flowpkg` package +
`harness flow {list,show,apply}` command. Bundled flows land in v0.93.

## What ships

- `internal/flowpkg/flowpkg.go`:
  - `Flow{Name, Description, Phases[]}`
  - `Phase{Name, Kind, Cmd, Prompt, SensorID, Gates[]}` with
    Kind: `deterministic | llm | sensor`
  - `Load(name)`, `List()`, `Apply(flow, opts)`
- `cmd/harness/cmd_flow.go` wires `harness flow list|show|apply`.
- Bundled flows directory `internal/flowpkg/templates/<flow>/flow.yaml`
  empty in this release; v0.93 fills it.

## Critical files

| Path | Change |
|---|---|
| `internal/flowpkg/flowpkg.go` (new) | types + Load + List |
| `internal/flowpkg/apply.go` (new) | Apply orchestrator |
| `internal/flowpkg/flowpkg_test.go` (new) | round-trip + Apply dry-run |
| `cmd/harness/cmd_flow.go` (new) | cobra wiring |
| `cmd/harness/main.go` | register newFlowCmd |

## Reuse, do not reinvent

- `embed.FS` pattern from `scaffoldpkg`, `hookpkg`, `mcppkg`.
- `os/exec` for deterministic shell phases.

## Verification

- `make lint` 0 issues
- `go test ./internal/flowpkg/... ./cmd/harness/...`
- Smoke: `harness flow list` shows the (empty) registry.

## Risks + mitigation

| Risk | Mitigation |
|---|---|
| Bundled flows ship without contracts | Validator rejects flow.yaml lacking required fields |
| Long phases block | Each phase has its own timeout default (60s shell, 300s llm) |

## Acceptance

- `harness flow list` runs without error even when registry is empty
- Load/List round-trip YAML correctly
- Apply respects `--dry-run`
