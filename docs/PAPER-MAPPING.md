# Paper "Code as Agent Harness" → HarnessX implementation map

Single source of truth mapping each paper section to the code, command,
and test path that implements it. Update this file whenever you add a
new feature that materialises a paper concept.

Citation: Ning et al. (UIUC / Meta / Stanford), arXiv 2605.18747v1,
May 2026. 102 pp.

---

## §2 Harness Interface — Code for Reasoning, Acting, Environment

| Paper concept | Command surface | Package | Tests |
|---|---|---|---|
| §2.1 Code for Reasoning (PoT, PAL, CodeAct) | `harness do`, `harness ask`, `harness plan` | `internal/app/workflow`, `internal/devloop` | `internal/devloop/loop_test.go` |
| §2.1.3 Iterative code-grounded reasoning | `harness loop`, `harness ship` | `internal/devloop`, `cmd/harness/cmd_ship.go` | `cmd/harness/cmd_ship_test.go` |
| §2.2 Code for Acting | adapter contract `internal/agents.AgentAdapter` | `internal/agents/{http,yaml,interactive}` | `internal/agents/*_test.go` |
| §2.3 Code for Environment (repository state) | `harness scaffold apply`, `harness new` | `internal/scaffoldpkg`, `internal/projectcfg` | `internal/scaffoldpkg/pkg_test.go`, `internal/projectcfg/projectcfg_test.go` |
| §2.3.3 Code-grounded evaluation environments | `harness smoke matrix` | `internal/app/smokecmd` | `internal/app/smokecmd/smokecmd_test.go` |

## §3 Harness Mechanisms

### §3.1 Planning

| Concept | Surface | Package | Tests |
|---|---|---|---|
| §3.1.1 Linear decomposition | `harness do` | `internal/app/workflow`, `internal/taskgraph` | `internal/taskgraph/*_test.go` |
| §3.1.2 Structure-grounded planning (custom rules) | `.harness/rules/*.yaml` loader | `internal/customrules` | `internal/customrules/customrules_test.go` |
| §3.1.4 Orchestration-based planning | `harness chat`, `harness orchestrate`, `harness ship` | `internal/repl`, `internal/orchestrate`, `internal/intentplan` | `internal/repl/repl_test.go`, `internal/orchestrate/orchestrate_test.go`, `internal/intentplan/intentplan_test.go` |

### §3.2 Memory taxonomy

| Paper kind | Code constant | Surface | Tests |
|---|---|---|---|
| Working | `memory.KindWorking` | `harness memory list --kind working` | `internal/memory/kinds_test.go` |
| Semantic | `memory.KindSemantic` | `harness memory list --kind semantic` | `internal/memory/kinds_test.go` |
| Experiential | `memory.KindExperiential` | `harness memory list --kind experiential` | `internal/memory/kinds_test.go` |
| Long-term | `memory.KindLongTerm` | `harness memory list --kind long_term` | `internal/memory/kinds_test.go` |
| Multi-agent | `memory.KindMultiAgent` | `harness memory list --kind multi_agent` | `internal/memory/kinds_test.go` |
| §3.2.6 Context compaction | DefaultProviders pipeline | `internal/context` | `internal/context/*_test.go` |

### §3.3 Tool Use

| Concept | Surface | Package |
|---|---|---|
| Function-oriented + Env-interaction + Verification-driven | adapters + sandboxed runtime | `internal/agents`, `internal/runtime` |
| Workflow-orchestration | `harness orchestrate run` | `internal/orchestrate` |

### §3.4 Plan-Execute-Verify Loop

| Concept | Surface | Package | Tests |
|---|---|---|---|
| §3.4.1 PEV loop | `harness ship`, `harness loop` | `internal/devloop`, `cmd/harness/cmd_ship.go` | `cmd/harness/cmd_ship_test.go` |
| §3.4.2 Planning as contract formation | `harness plan write` → `.harness/artifacts/plans/PLAN-<id>.md` | `cmd/harness/cmd_plan_write.go`, `internal/plancontract` | `internal/plancontract/plancontract_test.go` |
| §3.4.2 Contract enforcement | `harness plan check`, `harness ship --plan <id>` (pre-commit scope check) | `internal/sensors/planscope` | `internal/sensors/planscope/planscope_test.go` |
| §3.4.3 Sandboxed + permissioned execution | `internal/runtime/containers`, autonomy gates | `internal/runtime`, `internal/autonomy` | `internal/runtime/containers/*_test.go` |
| §3.4.4 Verification through deterministic sensors | `harness ci`, `harness check`, sensor catalog | `internal/sensors` (universal + stack + coverage + planscope + commentscan) | `internal/sensors/*_test.go` |

### §3.5 Agentic Harness Engineering (AHE)

| Concept | Surface | Package | Tests |
|---|---|---|---|
| §3.5.1 Deep telemetry as substrate | `.harness/logs/events.jsonl` | `internal/adapters/logger` | `internal/adapters/logger/*_test.go` |
| §3.5.2 Evolution Agent (diagnose → propose → replay) | `harness evolve diagnose|propose|replay|sandbox` | `internal/evolve` | `internal/evolve/evolve_test.go`, `internal/evolve/sandbox_test.go` |
| §3.5.3 Governed mutation (HITL) | `harness evolve promote --hitl`, `harness config wizard` (audit log) | `internal/evolve`, `internal/configwiz` | `internal/evolve/extra_test.go`, `internal/configwiz/configwiz_test.go` |

## §4 Scaling — Multi-agent orchestration over code

| Concept | Surface | Package | Tests |
|---|---|---|---|
| §4.1.1 Role specialization (Manager/Planner/Coder/Reviewer/Tester) | `orchestrate.Role*` constants | `internal/orchestrate` | `internal/orchestrate/orchestrate_test.go` |
| §4.1.3 Topology (chain/cyclic) | `orchestrate.Topology*` | `internal/orchestrate` | `internal/orchestrate/orchestrate_extra_test.go` |
| §4.3.1 Shared substrate — file-only blackboard | `.harness/artifacts/runs/<id>/blackboard.json` | `internal/orchestrate.writeBlackboard` | `internal/orchestrate/orchestrate_test.go` |
| Adapter-backed role step | `harness orchestrate run` + `NewAdapterRunner` | `internal/orchestrate/adapter_runner.go` | `internal/orchestrate/adapter_runner_test.go` |

## §5 Emerging fields

### §5.1.1 Code assistants (SDLC participation)

| Surface | Package |
|---|---|
| `harness new <stack>` | `cmd/harness/cmd_new.go` |
| `harness ship "<prompt>"` | `cmd/harness/cmd_ship.go` |
| `harness chat --goal dev` | `cmd/harness/cmd_chat.go`, `internal/repl` |
| `harness test/lint/dev/bench/profile` | `cmd/harness/cmd_wrappers.go`, `internal/projectcfg` |
| Tutorial replay | `scripts/tutorial-replay.sh` + `make tutorial-replay` |

Other §5 application domains (GUI/OS agents, embodied, scientific
discovery, personalization) — not in scope for v1.

### §5.2 Open problems addressed in HarnessX

| Problem | Mitigation |
|---|---|
| §5.2.1 Oracle adequacy | Multi-tier sensor catalog (forbidden, secrets, format, lint, typecheck, test, security, perf, deps, coverage, planscope, commentscan) |
| §5.2.3 Self-evolving harness without regression | `harness evolve sandbox` baseline-vs-candidate replay |
| §5.2.5 HITL safety as harness state | `--hitl` flag + `mutations.jsonl` audit log |

---

## Coverage status

| Package | % | Last updated |
|---|---|---|
| `internal/sensors/commentscan` | 100.0 | F12 |
| `internal/intentplan` | 98.2 | F15 |
| `internal/sensors/coverage` | 95.2 | F16 |
| `internal/sensors/planscope` | 95.2 | F19 |
| `internal/orchestrate` | 95.1 | F8 / F20 |
| `internal/customrules` | 94.4 | F10 |
| `internal/projectcfg` | 94.1 | F11 |
| `internal/plancontract` | 93.4 | F13 |
| `internal/app/smokecmd` | 93.8 | F0 |
| `internal/memory` | 91.7 | F6 |
| `internal/evolve` | 91.1 | F9 / F23 |
| `internal/repl` | 90.3 | F14 / F18 |
| `internal/configwiz` | 90.0 | F3 |

All new packages ≥ 90% coverage (project default threshold).

## Verification surface

- `make ci` — vet + race tests + build + every phase e2e
- `make smoke` — cross-stack CLI smoke matrix (Go, Python, Rails, React, Ruby, Rust)
- `make tutorial-replay` — deterministic walk of `docs/tutorial-python-demo.md`
- `harness coverage --threshold 0.9` — go coverage gate (default 90%)
- `harness ci` — every applicable sensor, exits non-zero on red
- `harness plan check --plan <id>` — scope violation gate per `§3.4.2`
- `harness evolve sandbox <trace.jsonl>` — A/B harness mutation replay
