# HarnessX Architecture

End-to-end map of the code that makes `harness do "..."`, `harness
loop`, `harness feature`, and friends work.

## Layered, dependency-inward

```
cmd/harness          CLI entrypoint (Cobra)
internal/app         use-case orchestration (init, workflow, agentcmd, ...)
internal/domain      pure types: Session, Run, Sensor, Agent, ContextPack
internal/adapters    SQLite repo, JSONL logger, exec probe
internal/platform    config, paths, ids, hashing, clock, budget
internal/agents      adapter abstractions (yaml, http, interactive, fake)
internal/execution   executor + hooks + sensors + diff capture + prune
internal/intent      rule-based intent classifier
internal/router      adapter selection by strengths matching
internal/taskgraph   prompt decomposition into typed tasks
internal/devloop     deterministic agent → lint/test → retry
internal/scaffoldpkg deterministic language scaffolds (5 langs)
internal/sensors     deterministic checks (lint, secrets, perf, ...)
internal/recall      bag-of-words search over past run reports
internal/scm         minimal git wrapper (HasRepo, Init, CurrentBranch)
```

## harness do pipeline (paper Layer 3: composability)

```
prompt
  │
  ▼
internal/taskgraph.Decompose
  │  rule-based splitter, 14 task kinds
  ▼
internal/router.Match / Pick
  │  strengths intersection scoring; ties broken by adapter id
  ▼
cmd_do.executeStep per task
  │  ├─ deterministic → internal/scaffoldpkg.Apply (no LLM)
  │  └─ adapter → internal/app/workflow.Feature → executor
  │
  ▼
handoff block prepended to each later task
  │  cmd_do.handoff: "Past steps in this run\n..."
  ▼
.harness/runs/_do/do-<ts>.md  (markdown report)
stdout → JSON when --json (schema_version=1, v0.43)
```

## harness loop pipeline (paper: verifiability + regression-free)

```
baseline lint + test  (devloop.runShell × 2)
  │
  ▼
for attempt in 1..max-attempts:
    workflow.Feature(prompt)        ← agent writes diff
    runShell(lint) + runShell(test)
    │
    ├─ both pass → exit
    │
    └─ fail → devloop.checkRegression(baseline, attempt)
                ↓
              devloop.Canonicalise(original, attempt)
                ↓
              prompt = canonical block + original
```

## Adapter abstraction

```
internal/agents/types.go
  AgentAdapter interface { ID, Name, Capabilities, Healthcheck,
                           Run, ParseUsage, ClassifyFailure }
  Capabilities { Text Vision Files Diff MCP MaxContextTokens
                 Strengths Models LoginCommand AuthDocURL ... }

Implementations:
  internal/agents/yaml        CLI binary wrap (claude, codex, ...)
  internal/agents/http        REST adapter (anthropic-api, ...)
  internal/agents/interactive PTY / tmux / iterm2 REPL driver (v0.27)
  internal/agents/fake        deterministic test fixture
```

Adapters declare `strengths` from the controlled vocabulary
(`code refactor reasoning search docs tests image vision audio data
sql shell review`). `internal/router` scores against this vocabulary;
unknown strengths are ignored (forward compatible).

## State + persistence

```
~/.local/share/harness (linux)
~/Library/Application Support/harness (macOS)
  registry.sqlite             cross-project workspace registry

per-project:
  .harness/
    config/harness.yaml         init template
    db/harness.sqlite           session + run telemetry
    logs/events.jsonl           rotating log
    hooks/pre-tool-use.sh       scaffolded by init (v0.28)
    artifacts/specs/<id>.md     spec template
    artifacts/plans/<id>.md     plan template
    runs/<id>/
      meta.json                 Result struct
      diff.patch                captured diff
      report.md                 canonical run report (v0.28+)
      stdout / stderr / sensors/ hooks/
    runs/_loop/loop-<ts>.md     devloop report (v0.31)
    runs/_do/do-<ts>.md         harness do report (v0.32)
```

## Templates as embed.FS (all-Go, no runtime fetch)

```
internal/hookpkg/templates/       5 hook scripts
internal/mcppkg/templates/        7 MCP server configs
internal/skillpkg/templates/      4 skill snippets
internal/install/manifests/       17 install manifests
internal/scaffoldpkg/templates/   5 language scaffolds (v0.30)
internal/app/agentcmd/bundled/    11 agent YAMLs
```

All consumed via `//go:embed`. No template touches the network at
runtime. Updates ship as new releases.

## Observability

- `harness logs --follow` tails `.harness/logs/events.jsonl`
- `harness metrics --since 1d|7d` aggregates `.harness/runs/*/meta.json`
- `harness audit --kind <kind>` filters JSONL by event type
- `harness memory recall "<query>"` (v0.33) bag-of-words search over
  past run reports

## Quality gates

- `make lint` golangci-lint with gocognit ≤ 25, gocyclo ≤ 15
- `go test ./...` unit + integration
- `make release` 6-platform tarballs + sha256
- `scripts/gen-brew-formula.sh <tag>` regenerates Formula/harness.rb
- `HARNESS_SKIP_CI=1` bypasses local pre-push gate for releases only

## Where to look for what

| Question | Read |
|---|---|
| how does `harness do` route? | `internal/router/strengths.go` + `internal/taskgraph/taskgraph.go` |
| how is a scaffold applied? | `internal/scaffoldpkg/pkg.go::Apply` |
| how is the LLM actually invoked? | `internal/execution/executor.go::invokeAdapter` |
| how is the canonical error built? | `internal/devloop/loop.go::Canonicalise` |
| how does a sensor declare confidence? | `internal/sensors/types.go::Result` |
| where do reports land? | `internal/execution/executor.go::writeReport` |
| how does brew tap work? | `Formula/harness.rb` + `scripts/gen-brew-formula.sh` |
