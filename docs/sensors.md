# Sensors

Sensors are **deterministic checks** that gate whether work produced by an agent (or a human) is acceptable. No sensor calls an LLM — semantic judgement is a Phase 6+ concern handled outside this package.

## Lifecycle

1. **Catalog discovery** — `sensors.Catalog(profile)` returns the universal sensors plus per-stack rule packs that apply to the project.
2. **Ordering** — Runner sorts computational sensors first (fast, no network), then inferential (none ship in Phase 4).
3. **Execution** — each Sensor returns a Result with `Status`, `Duration`, `OutputPath`, `Detail`, `ExitCode`.
4. **Persistence** — each Result is written to `sensor_results` and a per-sensor log lands under `.harness/artifacts/sensors/<run_id>/<id>.log`.

## Universal sensors

| id | category | what it does |
|---|---|---|
| `forbidden_files` | forbidden | rejects `.env`, `*.pem`, `id_rsa`, `secrets.yml`, etc. |
| `forbidden_commands` | forbidden | scans scripts/Makefile/Dockerfile/CI for `chmod -R 777`, `curl \| bash`, `rm -rf /`, `git push --force`, `--no-verify` |
| `secrets_scan` | secrets | regex scan for AKIA keys, Slack tokens, PEM blocks, GitHub tokens, `api_key='…'` |
| `changed_files` | changed_files | records `git diff --name-only HEAD` (informational; always passes) |

## Per-stack rule packs

Sensors with `OptionalTool: true` skip when the binary is missing — keeps CI green on minimal dev installs while the matrix runner still fails on tool errors when the binary is present.

### Go
`go_format` (gofmt), `go_vet`, `go_test`, `go_staticcheck`, `go_golangci_lint`, `go_vuln` (govulncheck)

### Node / React / Next.js / Vite
`node_eslint`, `node_prettier_check`, `node_typecheck` (tsc), `node_test` (npm test)

### Ruby / Rails
`ruby_rubocop`, `ruby_rspec`, `ruby_brakeman`, `ruby_bundle_audit`

### Python
`py_ruff`, `py_ruff_format`, `py_mypy`, `py_pytest`, `py_bandit`, `py_pip_audit`

### Rust
`rust_fmt`, `rust_clippy`, `rust_test`, `rust_audit`

### Docker
`docker_hadolint`

## Commands

```bash
harness sensor list                 # catalog applicable to this project
harness sensor run <id> [<id>...]   # run a subset
harness check                       # run every applicable sensor
harness ci                          # same + exit non-zero on any failed status
```

`harness check` always runs `forbidden_files` + `forbidden_commands` + `secrets_scan` + `changed_files`. The stack pack runs after (gofmt → go vet → go test → optional tools).

## Anti-patterns

- Do not call an LLM from inside a sensor. Sensors are deterministic by definition.
- Do not delete tests to make a sensor pass — hard-blocked by `forbidden_files` for `*_test.go`/`spec/`/`__tests__/` patterns once Phase 9's allowlist sensor lands.
- Do not gate on `lint` warnings unless the project config explicitly elevates them.
- Do not use shelling-out wrappers when a pure-Go scanner suffices. `forbidden_files` and `secrets_scan` are pure Go for hot-path speed.

## Custom sensors

Implement the `Sensor` interface:

```go
type Sensor interface {
    ID() string
    Category() Category
    Kind() Kind
    AppliesTo(p index.Profile) bool
    Run(rc RunCtx) Result
}
```

For shell-backed sensors use `ShellSensor` (`sensors.Wrap(ShellSensor{...})`); register via `sensors.Catalog` (Phase 9 adds a YAML route for custom sensors).
