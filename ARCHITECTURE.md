# HarnessX — Architecture

Local-first runtime built on three layers (per arXiv 2605.18747
"Code as Agent Harness"). All state under `.harness/` per repo.

## Layer cake

```
Layer 3 — Composability        cmd/harness/cmd_do.go, cmd_flow.go, cmd_loop.go
                               internal/flowpkg, internal/devloop, internal/critic

Layer 2 — Verifiability        internal/sensors, internal/audit, internal/optimize
                               internal/recall, internal/sharedstate, internal/multimodal

Layer 1 — Executability        internal/execution, internal/adapters, internal/router
                               internal/scaffoldpkg, internal/hookpkg, internal/mcppkg
```

## Dataflow per command

### `harness do <prompt>`
```
prompt → router.Pick(strengths) → adapter.Run → execution.Result
       → critic.Review (diff routed to critic strengths)
       → sensors.Scan (secrets, lint, tests, multimodal)
       → sharedstate.Commit (read/write sets, conflict check)
       → recall.Store (BM25)
       → audit append (events.jsonl)
```

### `harness loop`
```
watcher → devloop.Step → executor → sensors gate
       → checkpoint .harness/runs/_loop/<id>/state.json
       → resume via cmd_loop resume <id>
       → replan via taskgraph.Replan when sensor surfaces missing dep
```

### `harness flow apply <name>`
```
flowpkg.Load → validate phases → for each phase:
   deterministic → runShell (60s default timeout)
   llm          → cmd_do dispatch (router + critic)
   sensor       → sensors.Scan
   gates block downstream when any phase red
```

## Persistence layout

```
.harness/
├── audit/               events.jsonl (replay log)
│   └── approvals.jsonl  autonomy approvals history
├── runs/
│   ├── <run-id>/        per-run shared.json, do.md, artifacts
│   └── _loop/<id>/      checkpoint state.json
├── recall/              BM25 index + optional embeddings
├── memory/              long-term promoted entries
├── secrets/             encrypted backend (default file)
└── policies/            autonomy per-path rules
```

## embed.FS template inventory

| Package | Templates |
|---|---|
| `internal/scaffoldpkg/templates` | python, go-cli, react-spa, … |
| `internal/hookpkg/templates` | pre-commit, pre-push, gitleaks |
| `internal/mcppkg/templates` | filesystem, git, sqlite mcp servers |
| `internal/skillpkg/templates` | skill manifests |
| `internal/flowpkg/templates` | rails-api, python-fastapi, go-cli, meta-ads-campaign, content-pipeline, release-notes |

## Adapter pattern

`internal/adapters/<provider>/` each implement `Adapter` interface.
`internal/router` matches request tags against adapter `Strengths`.
Critic loop calls `router.Pick` again with `{tags:["review","critic"]}`
to route diffs to a different adapter.

## Sensor pattern

`sensors.Result{Confidence, Scope, Verified, Unverified, Risks}`.
Devloop refuses green when `Confidence < 0.5 ∧ len(Unverified) > 0`.

## Hard gates

- `make lint` — file LOC ≤ 400, gocognit ≤ 25, gocyclo ≤ 15
- `make coverage-gate` — bumps each release (current floor 58 core / 52 global)
- `make security` — govulncheck + gitleaks
- `scripts/git/pre-push.sh` — installable via `harness install-git-hooks`

## Layer mapping to paper principles

| Paper principle | Layer | Where |
|---|---|---|
| Executability | 1 | execution + adapters + scaffoldpkg |
| Verifiability | 2 | sensors + audit + optimize change-contracts |
| Statefulness | 2 | sharedstate + recall + devloop checkpoints |
| Composability | 3 | flowpkg + cmd_do + devloop + critic |
