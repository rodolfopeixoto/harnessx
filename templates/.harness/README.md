# `.harness/`

Per-project runtime directory for HarnessX. Created by `harness init`.

| Path | Purpose | Committed? |
|---|---|---|
| `config/harness.yaml` | Project configuration | yes |
| `db/harness.sqlite` | Local evidence database | no |
| `logs/events.jsonl` | Append-only event log | no |
| `cache/` | LSP, context, image caches | no |
| `artifacts/` | Generated artifacts (reports, diffs) | no |
| `product/` | Design manifests, feature maps (Phase 7) | yes |
| `project/` | Project profile, maps (Phase 2) | yes |

The bundled `.gitignore` excludes the volatile directories.
