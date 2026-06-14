# Configuration

`harness init` writes `.harness/config/harness.yaml`. Phase 1 only reads
`database.path` and `logging.*`; later phases add `agents`, `routes`,
`sensors`, and `context`.

```yaml
version: 1

project:
  name: my-project
  root: /abs/path/to/project

database:
  path: .harness/db/harness.sqlite

logging:
  path: .harness/logs/events.jsonl
  rotate_max_bytes: 10485760
```

Defaults are applied when fields are missing, so editing this file is
optional for the Phase 1 workflow.
