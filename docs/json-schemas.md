# JSON Schemas

Stable JSON contracts emitted by HarnessX. Every payload carries a
top-level `schema_version` (since v0.43) so consumers can refuse
versions they do not know.

Current schema_version: **1**

## `harness route show "<prompt>" --json`

```json
{
  "schema_version": 1,
  "prompt": "scaffold python and add a /healthz endpoint",
  "steps": [
    {
      "index": 1,
      "kind": "scaffold",
      "tags": ["scaffold", "code"],
      "routing": "deterministic:scaffold:python",
      "adapter_id": "",
      "prompt": "scaffold python",
      "confidence": 1.0,
      "lang": "python"
    },
    {
      "index": 2,
      "kind": "code",
      "tags": ["code"],
      "routing": "adapter:claude",
      "adapter_id": "claude",
      "prompt": "add a /healthz endpoint",
      "confidence": 0.7,
      "lang": ""
    }
  ]
}
```

### Field semantics

| Field | Type | Notes |
|---|---|---|
| `schema_version` | int | Bump on breaking change. Today: 1. |
| `prompt` | string | Original prompt. |
| `steps[].index` | int | 1-based. |
| `steps[].kind` | string | One of `scaffold lint test format secrets code refactor docs review image vision search data shell generic`. |
| `steps[].tags` | []string | Controlled vocabulary. Used by router. |
| `steps[].routing` | string | `deterministic:scaffold:<lang>`, `deterministic:sensor:<kind>`, `adapter:<id>`, or `none (...)`. |
| `steps[].adapter_id` | string | Empty for deterministic routings. |
| `steps[].prompt` | string | Sub-prompt for this step. |
| `steps[].confidence` | float | 0..1 — task classification confidence. <0.5 prints a warning in the human view. |
| `steps[].lang` | string | Populated for `kind=scaffold` (e.g. `python`). |

## `harness do "<prompt>" --yes --json`

Implies `--yes`. Logs route to stderr; JSON lands on stdout.

```json
{
  "schema_version": 1,
  "prompt": "scaffold python and add a /healthz endpoint",
  "report_path": "/path/to/.harness/runs/_do/do-20260616-104500.md",
  "steps": [
    {"index": 1, "kind": "scaffold", "...": "..."}
  ],
  "results": [
    "scaffold-dry",
    "workflow-status:applied"
  ]
}
```

### Additional fields

| Field | Type | Notes |
|---|---|---|
| `report_path` | string | Absolute path to the markdown report. Empty on early failure. |
| `results` | []string | One per step. Values: `scaffold-dry`, `sensor-hint`, `workflow-status:<status>`, `skipped`, `error: <message>`. |

## Consumer pattern

```bash
plan=$(harness route show "$PROMPT" --json)
ver=$(echo "$plan" | jq -r '.schema_version')
if [ "$ver" != "1" ]; then
  echo "harness JSON schema $ver not supported by this client" >&2
  exit 1
fi
echo "$plan" | jq '.steps[] | select(.routing | startswith("adapter:"))'
```

## Compatibility rules

- Additive fields **do not** bump `schema_version`. Consumers must
  ignore unknown fields.
- Renames, deletions, type changes **do** bump `schema_version`.
- The current major (`1`) covers `route show --json` (v0.39+) and
  `do --json` (v0.40+) plus the version field (v0.43+).
- Run reports under `.harness/runs/<id>/meta.json` follow the Go
  `execution.Result` struct shape; treat as internal until a future
  release moves them under this contract.

## Future

- v0.46+: `harness loop --json` (same schema with `attempts` array)
- v0.47+: `harness scaffold list --json`
- v0.48+: Run JSON contract under `meta.json` enrolled in this doc
