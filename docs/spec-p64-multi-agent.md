# SDD: P64 — Multi-agent routing + composability (v0.32)

## Why

"Code as Agent Harness" (arXiv 2605.18747) identifies four principles
for production agent infrastructure: **executability, verifiability,
statefulness, composability**. HarnessX already covers the first three:
sensors run real lint/test, `harness loop` (v0.31) canonicalises
failures back to the LLM, every run is persisted under
`.harness/runs/`. The fourth — composability — is missing: each
`harness feature` run pins a single adapter, and the operator must
pre-decide whether `claude` or `codex` or `gemini` is best for the
task. The paper's Layer 3 (multi-agent coordination via shared code
artifacts) is unimplemented.

Operator request: "send one prompt, harness picks the best CLI per
sub-task — claude for code, codex for refactor, gemini for image."

## What we ship

### 1. Adapter capability matrix (already declared, now used)

Every bundled adapter YAML already carries `strengths: []` (claude
declares `[code, reasoning]`, gemini declares `[vision, code]`, etc.).
Nothing reads them today. Promote `strengths` to a first-class router
input: `internal/router/strengths.go::Match(task, registry)` returns
the ranked adapter list per task type. Strengths get a controlled
vocabulary so matching is deterministic:

```
code | refactor | reasoning | search | docs | tests
image | vision | audio | data | sql | shell | review
```

Adapters declare any subset. The router scores `intersection(task.tags,
adapter.strengths) / len(task.tags)`; ties broken by adapter cost
(cheapest wins).

### 2. Task decomposer (`internal/taskgraph`)

One prompt → list of `Task{Kind, Tags, Prompt, DependsOn}`. Two
strategies:

- **Rule-based** (default, deterministic, no LLM): regex/keyword
  matcher. `"image of X"` → `Task{Kind: ImageGen, Tags: [image]}`.
  `"refactor Y"` → `Task{Kind: Refactor, Tags: [refactor, code]}`.
  ~20 rules.
- **LLM-based** (opt-in, `--decompose=llm`): cheap model (Haiku /
  Gemini Flash) returns task graph JSON via the existing adapter
  contract. Used when the rule matcher returns
  `Kind: Generic` confidence < 0.5.

### 3. `harness do "<prompt>"` (new top-level command)

```
harness do "scaffold a FastAPI app, add a /healthz endpoint, generate \
  an OpenGraph banner image"
```

Pipeline:

1. `taskgraph.Decompose(prompt)` → 3 tasks
2. For each task, `router.Pick(task)` chooses adapter via the
   strengths matrix
3. Show the plan (task table) and ask for confirmation unless
   `--yes`
4. Execute tasks in order, passing previous outputs as context to
   later tasks (the "shared code artifacts" from the paper)
5. One report at `.harness/runs/<id>/do.md` listing each step,
   chosen adapter, cost, status

Flags: `--yes`, `--deterministic` (forces scaffold/sensor where
matchable, fails task otherwise), `--budget-usd`, `--max-tasks`,
`--decompose=rules|llm`.

### 4. Deterministic-first toggle (paper Principle: executability)

For task kinds that already have a deterministic implementation
(`scaffold`, `lint`, `test`, `format`, `secrets-scan`), the router
ALWAYS prefers the deterministic path. LLM only invoked when no
deterministic match exists. `--no-deterministic` forces LLM-only.

Examples:
- `"scaffold python app"` → `scaffoldpkg.Apply("python")`, no LLM
- `"run all lint checks"` → `sensorcmd.Run(all)`, no LLM
- `"add a Redis client"` → no deterministic match → LLM

### 5. Regression-aware devloop (paper open challenge: regression-free)

`harness loop` (v0.31) currently retries until lint+test pass. Add a
**baseline capture** before the first attempt: run tests once, record
which currently pass. After each attempt, fail the loop if a
previously-passing test now fails (regression), even if other tests
went green. The canonical error block names which test regressed.

### 6. Cross-session memory (paper Layer 2)

`internal/memory/index.go`: SQLite-backed FTS over past
`.harness/runs/*/report.md` plus the prompt and tags. New
`harness memory recall "<query>"` shows the 5 most similar past runs.
`harness do` automatically prepends a "Past similar work" block when
the top match score > 0.6.

### 7. Multimodal auto-route

`--image foo.png` already accepted by workflow commands. Today it just
attaches the file. After P64:
- Adds `vision` tag to the task
- Router picks the highest-scoring vision adapter (`gemini-api` or
  `claude` with vision capability)
- If no vision-capable adapter installed, prints actionable hint

### 8. `harness route show "<prompt>"`

Dry-run that prints the task graph + chosen adapter per task without
executing. Lets operator inspect routing before paying.

```
$ harness route show "build a CLI to convert CSV to Parquet, generate \
   a hero image, write docs"

task#1  refactor+code   adapter: claude (score 1.00, $)
task#2  image           adapter: gemini-api (score 1.00, $$)
task#3  docs            adapter: claude (score 0.50, $)
estimated: 3 calls, max ~$0.10
```

## Critical files

| Path | Change |
|---|---|
| `internal/taskgraph/` (new) | task structs, rule decomposer, optional LLM decomposer |
| `internal/router/strengths.go` (new) | match task tags against adapter strengths |
| `internal/router/strengths_test.go` (new) | deterministic ranking tests |
| `cmd/harness/cmd_do.go` (new) | `harness do` + `harness route show` |
| `cmd/harness/main.go` | register new commands |
| `internal/devloop/loop.go` | baseline capture + regression check |
| `internal/memory/index.go` (new) | FTS index over past runs |
| `cmd/harness/cmd_memory.go` | already exists; extend with `recall` subcommand |
| `internal/app/agentcmd/bundled/*.yaml` | unify `strengths` vocabulary across all bundled adapters |
| `docs/tutorial.md` | new section: `harness do` + `harness route show` |
| `CHANGELOG.md` | v0.32 entry |
| `internal/version/version.go` | 0.31.0 → 0.32.0 |

## Reuse, do not reinvent

- `agents.Capabilities.Strengths` (already in `internal/agents/types.go`).
  Use as the single source of truth.
- `internal/intent.Classify` — keep for the workflow command; the new
  decomposer is task-level (subtler granularity).
- `internal/scaffoldpkg` — deterministic-first plug-in for scaffold
  tasks.
- `internal/sensors` — deterministic-first plug-in for
  lint/test/secrets-scan tasks.
- `internal/app/workflow.Feature` — each routed task that hits the LLM
  still goes through Feature so spec/plan/sensor gating is consistent.

## Verification

```bash
# foundation
make lint                    # 0 issues
go test ./...                # green
go test ./internal/router/... -count=1      # deterministic routing
go test ./internal/taskgraph/... -count=1   # rule decomposer

# smoke
harness route show "scaffold python and add a /healthz endpoint"
# expect: task#1 scaffold (deterministic, scaffoldpkg.python)
#         task#2 code (adapter: claude)

harness do "scaffold python and add a /healthz endpoint" --yes --budget-usd 0.20
# expect: scaffolder runs (no LLM), then claude is called for /healthz

# regression check
cd python-scaffold && pytest -q     # baseline: green
harness loop "break /healthz on purpose so I can verify regression detection"
# expect: loop refuses to accept attempt that regresses healthz test

# memory
harness memory recall "add healthz endpoint"
# expect: top match references prior do/feature run

# vision routing
harness do --image mockup.png "implement this UI mockup as a React component"
# expect: route picks vision-capable adapter (gemini-api or claude)
```

## Acceptance

- `harness route show` prints task graph with adapter per task in
  <500ms (no LLM).
- `harness do` executes the pipeline, honours `--budget-usd`, writes
  `do.md` with per-step cost + adapter + status.
- `--deterministic` skips LLM for tasks with scaffold/sensor matches.
- `harness loop` rejects attempts that regress previously-passing
  tests; canonical error names the regressed test.
- `harness memory recall` returns ranked past runs; `harness do`
  auto-prepends a "Past similar work" block when applicable.
- Multimodal: `--image` auto-routes to vision-capable adapter.
- All bundled adapters declare strengths from the controlled
  vocabulary.
- `make lint` clean, `go test ./...` green.

## Ships as

v0.32.0 (P64). Branch `feature/p64-multi-agent-routing`.
