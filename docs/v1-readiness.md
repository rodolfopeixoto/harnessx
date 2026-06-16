# HarnessX v1.0.0 Readiness Checklist

Tracks what stands between today's 0.4x.0 line and a v1.0.0 cut.

Status legend: ✓ done · ◐ partial · ✗ open

## CLI surface stability

- ✓ `harness init` (+ `--git`, `--all`)
- ✓ `harness scaffold {list,show,apply}` × 5 languages
- ✓ `harness do "<prompt>"` (+ `--deterministic`, `--budget-usd`, `--max-tasks`, `--image`, `--yes`, `--json`)
- ✓ `harness route show "<prompt>"` (+ `--json`)
- ✓ `harness loop "<prompt>"` (+ regression detection, `--max-attempts`, `--budget-usd`, `--lint`, `--test`)
- ✓ `harness feature` / `harness bugfix` / `harness ask` / `harness run`
- ✓ `harness list` (composite)
- ✓ `harness project {add,list,switch,archive,unarchive,prune,scan,forget,import,stale,index,inspect,current}`
- ✓ `harness runs {list,inspect,report,sensors,approve,discard,prune}`
- ✓ `harness agent {install,list,certify,login,...}`
- ✓ `harness sensor {list,run}` + `harness check` + `harness ci`
- ✓ `harness hook {list,scan,install,add,templates}`
- ✓ `harness mcp {install,scan,templates,...}`
- ✓ `harness backup`
- ✓ `harness memory {list,promote,recall}`
- ✓ `harness uninstall {project,global,all}`
- ✓ `harness help <topic>` × 13 topics
- ◐ `harness dashboard` — serves React UI but pages diverge from the design handoff
- ✗ Stable JSON schema versioning (today schemas are tied to release tag; need a top-level `$schemaVersion` field)

## Paper coverage (arXiv 2605.18747 — Code as Agent Harness)

| Layer / principle | Status | Surface |
|---|---|---|
| L1: Code as harness interface | ✓ | adapter abstraction + context builder + sensors |
| L2: Planning | ✓ | spec/plan templates + intent classifier |
| L2: Memory | ✓ | sqlite registry + run logs + `memory recall` (v0.33) |
| L2: Tools / MCP | ✓ | mcp install/scan + executor injects `--mcp-config` |
| L2: Feedback control loops | ✓ | `harness loop` canonical-error retry (v0.31) |
| L3: Multi-agent coordination | ✓ | `harness do` + cross-task handoff (v0.32 + v0.38) |
| Executability | ✓ | sensors + deterministic scaffolds |
| Verifiability | ✓ | baseline-diff regression detection (v0.33) |
| Statefulness | ✓ | sqlite + budget ledger (v0.31) + memory recall |
| Composability | ✓ | per-task strengths matching + handoff |
| Verification with incomplete feedback | ◐ | `sensors.Result.Confidence` field exists (v0.34); not every scanner populates it yet |
| Regression-free improvements | ✓ | baseline capture + flagged regression in loop (v0.33) |
| Consistent state across agents | ✓ | handoff block prepended to every later task (v0.38) |
| Multimodal context | ✓ | `--image` auto-adds `vision` tag (v0.33) |
| Safety / human oversight | ✓ | autonomy levels + per-path policy + confirmation prompts |
| Long-horizon execution | ◐ | bounded by `--max-attempts`; no checkpoint/resume |
| Eval beyond task completion | ✗ | only pass/fail; no skill score / rubric framework |

## Quality gates

- ✓ `make lint` 0 issues (cumulative across every shipped release)
- ✓ `go test ./...` green
- ✓ `gofmt -l` clean
- ✓ 6-platform release matrix (darwin/linux × amd64/arm64 + windows × amd64/arm64)
- ✓ brew formula auto-regen + tap install path
- ✓ curl install.sh path
- ✓ `harness uninstall all` removes every trace
- ◐ Govulncheck + gitleaks last green at v0.21 — re-run pre-v1.0
- ✗ Coverage gate >= 60 % (today partial coverage; need explicit target)
- ✗ Benchmark suite for executor + router

## Operator UX gates

- ✓ Top-level `harness list` composite view (v0.28)
- ✓ Sectioned `harness do` output with low-confidence warning (v0.34)
- ✓ Agent-call heartbeat (v0.37)
- ✓ Per-call budget ledger primitives (v0.31)
- ◐ Workflow `harness feature` output rework with presenter — primitives shipped (v0.31), wiring partial
- ✗ Live spinner (deliberately deferred — `[agent] calling/returned` lines cover the gap without a new dep)
- ✗ Streaming partial output from the agent

## Documentation

- ✓ `docs/tutorial.md` end-to-end (current as of v0.32)
- ✓ `docs/anthropic-billing.md`
- ✓ `docs/spec-p64-multi-agent.md`
- ✓ `docs/v1-readiness.md` (this file)
- ◐ Tutorial needs refresh for v0.33–v0.40 surface (`memory recall`, `runs prune`, `--json`)
- ✗ `docs/architecture.md` — end-to-end diagram + module map
- ✗ `docs/json-schemas.md` — stable JSON contract for IDE plugins

## What blocks v1.0.0 today

1. Stable JSON schema versioning (top-level `schema_version` on every `--json` output).
2. Govulncheck + gitleaks re-run; pin clean baseline.
3. Coverage gate target + benchmark suite.
4. Tutorial refresh covering v0.33–v0.40.
5. `docs/architecture.md` + `docs/json-schemas.md`.
6. Dashboard parity with handoff zip — deferred or descoped before v1.0.
7. Decide: ship `--decompose=llm` fallback (paper L2 long-horizon) before v1.0 or after.

## Proposed cut criteria

Ship v1.0.0 when:

- All ✗ above turn ◐ or ✓.
- Dog-food test against a freshly scaffolded FastAPI project completes end-to-end
  with `harness do` + `harness loop` + `harness memory recall`, every command
  exits 0, output is coherent, costs visible.
- Three operators run the tutorial cold and report zero blockers.
