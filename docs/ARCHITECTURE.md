# HarnessX architecture

This document explains how the HarnessX runtime is organised and why.
Everything here is anchored on the survey paper
**"Code as Agent Harness"** (Ning et al., UIUC / Meta / Stanford,
arXiv 2605.18747v1, May 2026, 102 pp). Section references below
(`§N.M`) point into that paper. For the §-by-§ implementation map see
[`PAPER-MAPPING.md`](PAPER-MAPPING.md).

---

## 1. Mental model

HarnessX is a **code-centric agent harness** in the paper's sense:

- **Code is executable** — every agent action ends up running a
  command, a sensor, or an LLM adapter call.
- **Code is inspectable** — every action emits a structured event into
  `.harness/logs/events.jsonl`, and every artefact lands in
  `.harness/artifacts/`.
- **Code is stateful** — projects keep durable state in SQLite, the
  blackboard, the plan contracts, and the memory store.
- **Code is governed** — every mutation that changes the harness
  itself flows through deterministic sensors and HITL gates.

These four properties define the three layers below.

---

## 2. Layered architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  §5 Domain surfaces — code assistants, ops, ads, research, ... │
│  harness new │ ship │ chat │ dashboard │ backup │ doctor       │
├─────────────────────────────────────────────────────────────────┤
│  §4 Scaling — multi-agent orchestration over code              │
│  orchestrate roles + topology + blackboard                     │
├─────────────────────────────────────────────────────────────────┤
│  §3 Mechanisms — planning, memory, tools, PEV loop, AHE        │
│  router │ memory │ devloop │ sensors │ evolve │ configwiz      │
├─────────────────────────────────────────────────────────────────┤
│  §2 Interface — code for reasoning, acting, environment        │
│  agents adapters │ scaffold │ projectcfg │ context providers   │
├─────────────────────────────────────────────────────────────────┤
│  Platform — sqlite store, logger, paths, ids, config           │
└─────────────────────────────────────────────────────────────────┘
```

### 2.1 Platform (`internal/platform/*`, `internal/adapters/sqlite`)

- `paths` — locates the project root and `.harness/` layout.
- `config` — loads `harness.yaml` and resolves embedded paths.
- `ids` — ULIDs for every run, session, mutation, plan, blackboard.
- `tokens`, `budget`, `clock`, `i18n` — shared primitives.
- `adapters/sqlite` — single-binary database (no CGO,
  `modernc.org/sqlite`). Schema in `migrations/0001_init.sql`. Stores
  sessions, runs, sensor outcomes, memories, skills, settings.
- `adapters/logger` — append-only JSONL with rotation.

### 2.2 §2 Interface — code for reasoning, acting, environment

- `internal/agents/` — the adapter contract (`AgentAdapter`,
  `Capabilities`, `Usage`, `FailureType`). Concrete adapters live
  under `internal/agents/http`, `internal/agents/yaml`,
  `internal/agents/interactive`, `internal/agents/fake`. The HTTP
  adapter consumes a `Spec` declared in YAML so adding a provider
  (Ollama, Moonshot, Anthropic, OpenAI…) is data-driven.
- `internal/app/agentcmd/bundled/*.yaml` — 12 ready-to-use specs:
  Claude (CLI + interactive + API), Codex, Gemini (CLI + API), Kimi,
  Moonshot, Minimax, OpenAI, Anthropic API, Ollama (local), Fake
  (testing).
- `internal/scaffoldpkg/` — language scaffolds (Go, Python, Rails,
  React, Ruby, Rust). Each scaffold writes a Makefile, `.gitignore`,
  source, and tests. The scaffolded `project.yaml` records the
  canonical `test`/`lint`/`run` commands so the wrappers can resolve
  them without recomputing.
- `internal/projectcfg/` — manages `.harness/config/project.yaml` and
  detects the stack from manifest files (`go.mod`, `pyproject.toml`,
  …). The wrappers (`harness test|lint|dev|bench|profile`) read this
  file first, then fall back to per-stack defaults.
- `internal/context/` — the deterministic context-pack builder. It
  orders providers (Memory → Git → Ripgrep → TestMap → LSP), ranks
  files by relevance, and caches by task + profile + HEAD hash so
  repeated runs do not pay the cost twice.

### 2.3 §3 Mechanisms — planning, memory, tools, control, AHE

- `internal/router/` — task → adapter chain over the agent registry.
  Each chain has a primary and an ordered fallback list. Failure
  classification (`agents.FailureType.IsRecoverable`) drives the
  fallback chain: rate-limit, context-limit, transient, timeout move
  on; auth and permanent failures abort the chain.
- `internal/intentplan/` — the **JSON plan schema** consumed by
  `harness chat`, `harness ship`, and `harness orchestrate`. Steps
  are typed (`harness`, `shell`, `wait`). Every `harness` step must
  use a command from the goal palette; this keeps LLM-emitted plans
  inside a safe, deterministic dispatch table.
- `internal/repl/` — the REPL behind `harness chat`. The session goal
  selects the palette; the planner produces a typed plan; the
  executor dispatches deterministically. A `Planner` is injectable
  (default deterministic; LLM-backed via `NewLLMPlanner` when an
  adapter is supplied) so the same loop can be tested without an LLM.
- `internal/memory/` — evidence-gated promotion with the paper's five
  kinds: `Working`, `Semantic`, `Experiential`, `LongTerm`,
  `MultiAgent`. Promotions require a run id, confidence ≥ 0.4, and
  must not contain secrets (same regex catalogue as the secrets
  sensor).
- `internal/devloop/` — the Plan-Execute-Verify loop. Generates a
  diff, runs the project sensors, detects regression vs. the
  baseline, and feeds compact failure context back into the next
  attempt. Bounded by `--max-attempts` and `--budget-usd`.
- `internal/sensors/` — the deterministic sensor catalogue:
  - Universal: `forbidden_files`, `forbidden_commands`,
    `secrets_scan`, `changed_files`, `performance_budget`.
  - Stack rule pack: Go, Python, Ruby/Rails, Node/React, Rust,
    Docker — about 24 sensors total.
  - `go_coverage_gate` — auto-registered for Go projects (paper
    §3.4.4); wraps `coverage.ParseGoCover` with a configurable
    threshold (default 90%).
  - `plan_scope` — auto-registered when `.harness/config/plan.yaml`
    pins an active plan (paper §3.4.2 contract enforcement). Wraps
    `internal/sensors/planscope`.
  - `internal/sensors/commentscan` — Go AST scan that flags
    narrative comments outside SPDX, package docs, or godoc on
    exported symbols.
- `internal/plancontract/` — parses the `PLAN-<id>.md` artefacts
  produced by `harness plan write`. Sections: intent, files,
  invariants, validation, rollback, risk. Used by
  `harness plan check`, the `plan_scope` sensor, and
  `harness ship --plan <id>`.
- `internal/evolve/` — the Evolution Agent for paper §3.5:
  - `Diagnose` clusters `events.jsonl` failures by signature.
  - `Replay` scores a candidate trace against the diagnosis.
  - `RunSandbox` is the **real** A/B replay: it invokes both the
    baseline and the candidate `harness` binary against an isolated
    workspace, reports the failure delta and whether the candidate
    improved.
  - `Propose` and `Promote` write to `mutations.jsonl`; promotion
    refuses to write without `--hitl` (paper §3.5.3).
- `internal/configwiz/` — the wizard behind `harness config`. Every
  mutation appends to `config-mutations.jsonl` with full before/after
  state so reverts are auditable.
- `internal/customrules/` — loads `.harness/rules/*.yaml` so projects
  can ship structure-grounded invariants alongside the bundled
  catalogue (paper §3.1.2).

### 2.4 §4 Scaling — multi-agent orchestration over code

- `internal/orchestrate/` — flow loader, validator, and executor.
  Flows are YAML documents declaring `topology: chain | cyclic`,
  `max_cycles`, and `steps`. Each step declares a `role`
  (`manager | planner | coder | reviewer | tester`) and either a
  `command`, a shell call, or an `adapter` id.
  - Shell steps run as subprocesses inside the project root.
  - Adapter steps go through `NewAdapterRunner`, which looks up the
    adapter in the registry, builds a role-aware prompt with the
    most recent blackboard entries, and writes the adapter response
    back to the blackboard.
  - Every run writes `.harness/artifacts/runs/<id>/blackboard.json` —
    the paper's "file-only shared substrate" (§4.3.1).

### 2.5 §5 Domain surfaces and supporting commands

- `cmd/harness/cmd_new.go` — bootstrap.
- `cmd/harness/cmd_ship.go` — SDLC driver.
- `cmd/harness/cmd_chat.go` — REPL.
- `cmd/harness/cmd_dashboard.go` — React dashboard launcher.
- `cmd/harness/cmd_backup.go` — portable run-state tarballs.
- `cmd/harness/cmd_doctor.go`, `cmd_runtime.go`, `cmd_containers.go`,
  `cmd_metrics.go`, `cmd_audit*.go`, `cmd_secret.go` — operations
  surface.

---

## 3. Where everything lives on disk

```
.harness/
├── config/
│   ├── harness.yaml            project config
│   ├── routes.yaml             router overrides
│   ├── project.yaml            per-stack command catalogue
│   ├── plan.yaml               active plan pin (drives plan_scope)
│   └── agents/*.yaml           project adapter overrides
├── artifacts/
│   ├── plans/PLAN-<id>.md      plan-as-contract
│   ├── runs/<id>/
│   │   ├── blackboard.json     orchestrate
│   │   ├── do.md               do report
│   │   └── ...
│   ├── sensors/<run>/...       sensor bundles
│   └── specs/<id>.md           feature specs
├── db/harness.sqlite           SQLite store
├── logs/
│   ├── events.jsonl            append-only telemetry
│   ├── mutations.jsonl         evolve audit log
│   └── config-mutations.jsonl
├── hooks/                      pre-tool-use, post-tool-use, session
├── rules/*.yaml                user-defined custom rules
├── orchestrations/*.yaml       user-defined orchestration flows
├── sessions/<id>.jsonl         chat REPL turns
└── runs/<id>/                  per-run scratch + report.md
```

---

## 4. Data flow for the four canonical entry points

### 4.1 `harness ship "<prompt>" [--plan <id>]`

1. Verify git tree clean. If `--plan` is set, load the contract and
   reuse the intent if the prompt is empty.
2. Branch `feature/<slug>` from `--base`.
3. Subprocess `harness feature "<prompt>" --yes` to materialise a
   spec.
4. Loop up to `--max-attempts`:
   - `harness do "<prompt>" --autonomy <mode>`
   - `harness ci`
   - On HTTP 429 / rate-limit markers, exponential backoff up to
     `--rate-limit-retries`, then router fallback.
   - On green, break the loop.
5. If `--plan <id>`: run `planscope.Check`; refuse commit on
   out-of-scope diffs.
6. Conventional commit. Prompt and plan id end up in the body.

### 4.2 `harness chat --goal dev [--adapter ollama]`

1. Greet, load `intentplan.GoalPalette(goal)`.
2. Read a line. Slash command? Handle. Otherwise treat as prompt.
3. Planner (deterministic or LLM-backed) returns
   `intentplan.Plan`.
4. `intentplan.Execute` dispatches `harness` steps to the same
   binary as subprocesses; `shell` steps run via `/bin/sh -c`;
   `wait` steps sleep. Each step is appended to the session JSONL.
5. Repeat until `/exit`.

### 4.3 `harness orchestrate run review-cycle`

1. Load `.harness/orchestrations/review-cycle.yaml`.
2. Validate roles + topology + step shape.
3. For each step (and for each cycle in cyclic flows):
   - Shell step → subprocess.
   - Adapter step → `NewAdapterRunner` runs the agent with the most
     recent blackboard entries as context.
4. Write the entry to the blackboard JSON.
5. Stop on first failure in chain topology; persist regardless.

### 4.4 `harness evolve sandbox <trace>`

1. Materialise two workspaces (`baseline/`, `candidate/`) under a
   temp root.
2. Copy the trace into `<workspace>/.harness/logs/events.jsonl`.
3. Invoke each binary with `evolve diagnose --json` inside its
   workspace.
4. Parse the failure cluster JSON, compute delta, report
   improvement.

---

## 5. Verification surface

| Gate | Command | What it runs |
|---|---|---|
| Pre-commit hook | `harness lint` | Project linter (resolved per stack) |
| Commit-msg hook | regex | Conventional Commits |
| Pre-push hook | `harness ci` | Every applicable sensor; non-zero on red |
| Full local CI | `make ci` | vet + race tests + coverage gate + every phase e2e |
| Cross-stack regression | `make smoke` | `harness smoke matrix` across 6 stacks |
| Tutorial regression | `make tutorial-replay` | Walks the cheat sheet end-to-end |
| Coverage floor | `harness coverage --threshold 0.9` | `go test -cover ./...` parsed and gated |
| Scope contract | `harness plan check --plan <id>` | `planscope.Check` against working diff |
| Self-evolution | `harness evolve sandbox <trace>` | Real A/B replay across baseline/candidate |

---

## 6. Design constraints

- **No CGO.** `modernc.org/sqlite` only. Single-binary distribution is
  a hard requirement.
- **No comments in scaffolded code.** Python templates ship with the
  `ERA` rule enabled in `ruff.toml`. The `commentscan` sensor enforces
  the same convention for Go.
- **English first.** All code, log messages, and the English bundle
  (`internal/platform/i18n/locales/en.json`) are in English.
- **All user-facing strings through `i18n.T(key)`.**
- **Adapters are data, not code.** New providers ship as YAML specs
  inside `internal/app/agentcmd/bundled/` or under
  `.harness/config/agents/`.
- **Sensors are deterministic.** Inferential gates live outside the
  sensor catalogue (paper §3.4.4).
- **Mutations are governed.** `harness evolve promote` and
  `harness config wizard` both append to audit logs.
- **GitFlow.** Branches off `develop`. `main` is releases only.

---

## 7. Where the paper's open problems are addressed

| Open problem (§5.2) | HarnessX surface |
|---|---|
| §5.2.1 Oracle adequacy | Multi-tier sensor catalogue + coverage gate + scope gate |
| §5.2.2 Semantic verification beyond executable feedback | `harness evolve` clusters telemetry into intent-bearing signatures |
| §5.2.3 Self-evolving harness without regression | `harness evolve sandbox` real A/B replay |
| §5.2.4 Transactional shared program state | `internal/sharedstate` + blackboard JSON |
| §5.2.5 Human-in-the-loop safety | `--hitl` flag + `mutations.jsonl` + autonomy gates |
| §5.2.7 Toward a science of harness engineering | Coverage floor (90% default), paper mapping doc, deterministic smoke matrix |

---

## 8. Pointers

- Paper map: [`PAPER-MAPPING.md`](PAPER-MAPPING.md)
- Command reference: [`COMMANDS.md`](COMMANDS.md)
- Tutorial (Python FastAPI): [`tutorial-python-demo.md`](tutorial-python-demo.md)
- Tutorial (end-to-end e-commerce): [`TUTORIAL-ECOMMERCE.md`](TUTORIAL-ECOMMERCE.md)
- Contributing: [`../CONTRIBUTING.md`](../CONTRIBUTING.md)
