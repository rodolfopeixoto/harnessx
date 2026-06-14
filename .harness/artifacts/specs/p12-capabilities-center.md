# P12 — Capabilities Center

## Acceptance

- `harness catalog list [--kind <kind>] [--project <ref>]` shows discovered + installed capabilities across all kinds.
- `harness catalog show <kind> <name>` prints details + config preview.
- `harness catalog install <kind> <name> [--yes] [--dry-run]` plans a `[]FileOp`, prints unified diff, requires approval or `--yes`. With `--dry-run` writes nothing.
- `harness catalog configure <kind> <name>` opens-or-creates the config file at the spec'd path.
- `harness catalog remove <kind> <name>` deletes the installed config + registry row.
- `GET /api/catalog/kinds`, `/api/catalog/items?kind=...`, `/api/catalog/item?kind&name`, `POST /api/catalog/plan`, `POST /api/catalog/install`.
- Project DB migration `0002_capabilities.sql` adds `installed_capabilities(kind,name,version,source,installed_at,config_path,content_hash)`.

## Contract

### Domain (`internal/domain/capability.go`)

```go
type CapabilityKind string
const (
  KindAgent    CapabilityKind = "agent"
  KindMCP      CapabilityKind = "mcp"
  KindHook     CapabilityKind = "hook"
  KindSensor   CapabilityKind = "sensor"
  KindSkill    CapabilityKind = "skill"
  KindContext  CapabilityKind = "context"
  KindResource CapabilityKind = "resource"
  KindPlugin   CapabilityKind = "plugin"
)

type Capability struct {
  Kind        CapabilityKind
  Name        string
  Version     string
  Source      string      // bundled | user | external
  Status      string      // detected | installed | configured | enabled | failed
  ManifestPath string
  ConfigPath  string
  Description string
}

type FileOp struct {
  Op   string // create | overwrite | append | delete
  Path string
  Body []byte
}
```

### Per-kind plug-in

`internal/catalog/kinds/<kind>.go` implements:

```go
type Kind interface {
  Kind() domain.CapabilityKind
  Discover(ctx, root string) ([]domain.Capability, error)  // filesystem-only, never LLM
  Plan(ctx, root, name string) ([]domain.FileOp, error)
}
```

Plans are pure data; commits go through `catalog.Apply(root, ops)` which stages to a temp dir + atomic rename.

### Approval flow

`catalog.Install` returns `ExitUserDeny` (3) when stdin is closed and `--yes` is absent. Interactive prompt is rendered by the cmd layer, not the package.

## Risks

- **Kind sprawl → anaemic interface.** Mitigation: per-kind file caps at ~200 LOC; shared validation in `catalog/discovery.go`.
- **Partial writes.** Mitigation: stage-then-rename via temp dir, every Plan is all-or-nothing.
- **Discovery false positives.** Mitigation: deterministic globs + manifest schema; LLM never narrows the list.

## Non-goals

- Catalog UI (Phase 17).
- Cross-project install (this phase only installs to current project; multi-project install is queued by Phase 18 autopilot).

## Verification

- `internal/catalog` ≥ 92% (golden Plan output per kind).
- `internal/catalog/kinds` ≥ 85% per kind (fixtures under `internal/catalog/testdata/`).
- `scripts/e2e-phase12.sh` installs one of every kind into a tmp project and asserts files on disk + registry rows.
