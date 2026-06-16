# HarnessX

Local-first adaptive runtime for **agentic software engineering**.
Orchestrates multiple coding CLIs (Claude Code, Codex, Gemini, Kimi)
through a pluggable adapter system, gates work with deterministic
sensors, persists evidence in SQLite, and surfaces results via TUI +
local React dashboard.

> **Status:** v0.100 (Phase D). v1.0.0 deferred until operator
> manual testing + community feedback validates each surface.
> See [docs/specs/](docs/specs/) for the per-release spec log and
> [CHANGELOG.md](CHANGELOG.md) for milestones.

## Why HarnessX

Three-layer harness inspired by *"Code as Agent Harness"*
([arXiv 2605.18747](https://arxiv.org/abs/2605.18747)). Each layer
enforces one paper principle:

| Layer | Principle | What it does |
|---|---|---|
| **L1 — Executability** | Make agents actually run | Adapters + router + scaffolds; one prompt → routed CLI execution |
| **L2 — Verifiability** | Trust nothing without evidence | Sensors (secrets/lint/tests), audit log, change-contract optimizer |
| **L3 — Composability** | Compose runs into workflows | `harness do`, `harness loop`, `harness flow` — task graphs + replan + critic |

See [ARCHITECTURE.md](ARCHITECTURE.md) for dataflow diagrams,
persistence layout, and the embed.FS template inventory.

---

## Install

### macOS (Homebrew)

```bash
brew tap rodolfopeixoto/harnessx https://github.com/rodolfopeixoto/harnessx
brew install harness
```

Tap lives in this same repo under `Formula/harness.rb` — auto-refreshed
every release with new SHAs across all 6 platforms.

### Linux / macOS (one-line installer)

```bash
curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh | bash
```

### From source (Go 1.21+)

```bash
git clone https://github.com/rodolfopeixoto/harnessx
cd harnessx
make build
./bin/harness version
```

### Pre-built binaries

Every release ships 6 platforms (`make release`):
darwin-amd64, darwin-arm64, linux-amd64, linux-arm64,
windows-amd64, windows-arm64. Each with `.sha256` companion.
Download from the [Releases page](https://github.com/rodolfopeixoto/harnessx/releases).

---

## Quickstart

```bash
# Init a project (writes .harness/ + hooks + first scaffold prompt)
harness init

# Install the pre-push gate (lint + tests + coverage + security)
harness install-git-hooks

# One-shot: route prompt to best adapter + critic + sensors
harness do "add a /healthz endpoint with a test"

# Continuous loop: rerun on save until sensors green
harness loop

# Bundle scaffolds + first feature in one go
harness flow init rails-api          # software flow
harness flow init meta-ads-campaign  # non-software flow
```

---

## Features (organised by layer)

### Layer 1 — Executability

| Command | What it does |
|---|---|
| `harness init` | Bootstraps `.harness/`, hooks, default policy |
| `harness do <prompt>` | Routes prompt → best adapter → critic → sensors |
| `harness scaffold list / apply <name>` | Deterministic templates (python, go-cli, react-spa, …) |
| `harness hookpkg list / install` | Bundled git hooks (pre-commit, gitleaks, etc.) |
| `harness mcppkg list / install` | Bundled MCP servers (filesystem, git, sqlite) |
| `harness skill list / apply` | Skill manifests |
| `harness route show <prompt>` | Preview routing decision (strengths match) |
| `harness execute <task>` | Run a task off the queue |

### Layer 2 — Verifiability

| Command | What it does |
|---|---|
| `harness sensor list / run <id>` | Deterministic checks (secrets, lint, tests, multimodal) |
| `harness audit tail / replay` | Append-only event log under `.harness/audit/events.jsonl` |
| `harness optimize propose / apply --canary` | Change-contract harness mutations (predicted improvement + falsifier test + rollback cmd) |
| `harness metrics show` | Trajectory + verification + recovery + replayability |
| `harness autonomy show / suggest` | Per-path policy + history-mined proposals |
| `harness secret get/set/list` | Encrypted secret backend |
| `harness backup save / restore` | Portable run-state snapshots |

### Layer 3 — Composability

| Command | What it does |
|---|---|
| `harness do <prompt>` | Plan → execute → critic → sensors → store |
| `harness loop` | Watcher rerun until green; checkpoint + resume via `harness loop resume <id>` |
| `harness flow list / show / apply / init` | Bundled flows: rails-api, python-fastapi, go-cli, meta-ads-campaign, content-pipeline, release-notes |
| `harness workflow run` | Multi-step orchestration |
| `harness stack show` | Active runs + sessions |

### Cross-cutting

| Command | What it does |
|---|---|
| `harness dashboard` | Boots local React UI on `http://localhost:5173` |
| `harness palette` | Fuzzy command palette TUI |
| `harness doctor` | Diagnose env + dependencies |
| `harness install-git-hooks` | Wires `scripts/git/pre-push.sh` as `.git/hooks/pre-push` |
| `harness update` | Self-update via release channel |
| `harness uninstall` | Remove harness + state |
| `harness version` | Build version + commit + date |

Full reference: [docs/cli-reference.md](docs/cli-reference.md).

---

## Make targets (contributor surface)

`harness` is the **user** CLI. `make` is the **contributor** surface.

| Target | What it does |
|---|---|
| `make build` | Build `bin/harness` with ldflags-stamped version |
| `make test` | Race tests + coverage report |
| `make test-short` | Skip slow integration tests |
| `make lint` | golangci-lint (file LOC ≤ 400, gocognit ≤ 25, gocyclo ≤ 15) |
| `make fmt / tidy / vet` | Formatting + module hygiene |
| `make check` | vet + race tests + build |
| `make ci` | Full local gate: lint + coverage-gate + tests + e2e |
| `make cd` | ci + dashboard build + security + licenses + sbom + release |
| `make coverage` | Per-package coverage report |
| `make coverage-gate` | Enforce floor (current: 58 core / 52 global) |
| `make bench` | Run benchmarks across `./internal/...` |
| `make profile-mem` | Dump heap pprof to `dist/profiles/mem.pprof` |
| `make profile-cpu` | Dump cpu pprof to `dist/profiles/cpu.pprof` |
| `make security` | govulncheck + gitleaks |
| `make licenses / sbom` | License inventory + SPDX SBOM |
| `make release` | Cross-compile 6 platforms into `dist/` + SHA-256 sums |
| `make e2e / e2e-all` | Shell + Playwright E2E suites |
| `make dashboard-install / dashboard-dev / dashboard-build / dashboard-test` | React dashboard lifecycle |
| `make install-hooks / uninstall-hooks` | Git hook wiring |

---

## Architecture (at a glance)

```
                      ┌───────────────────────────────────┐
                      │   cmd/harness (cobra commands)    │
                      └─────────────┬─────────────────────┘
                                    │
   ┌────────────────────────────────┼────────────────────────────────┐
   │                                │                                │
┌──▼──────────┐              ┌──────▼──────┐                ┌────────▼────────┐
│  L3: flow   │              │  L3: do     │                │  L3: loop       │
│  flowpkg    │              │  cmd_do     │                │  devloop        │
└──┬──────────┘              └──────┬──────┘                └────────┬────────┘
   │                                │                                │
   └──────────────┬─────────────────┼──────────────┬─────────────────┘
                  │                 │              │
              ┌───▼─────┐      ┌────▼────┐    ┌────▼─────┐
              │ router  │      │ critic  │    │ sensors  │
              │ adapters│      │ adapter │    │ multimd  │
              └───┬─────┘      └────┬────┘    └────┬─────┘
                  │                 │              │
              ┌───▼─────────────────▼──────────────▼─────┐
              │  execution + audit + recall + sharedstate │
              └────────────────────┬──────────────────────┘
                                   │
                          ┌────────▼─────────┐
                          │  .harness/ (FS)  │
                          │  SQLite + JSONL  │
                          └──────────────────┘
```

Full diagrams + dataflow per command: [ARCHITECTURE.md](ARCHITECTURE.md).

---

## Insights (what this harness believes)

1. **Deterministic-first.** Templates, sensors, gates run before any
   LLM call. The LLM only fills gaps deterministic code cannot.
2. **Evidence over assertion.** No green signal without a sensor
   bundle (`Confidence`, `Verified`, `Unverified`, `Risks`). Devloop
   refuses green when `conf < 0.5 ∧ unverified ≠ ∅`.
3. **Change contracts.** Every harness mutation declares
   `{component, target_failure, predicted_improvement, invariants,
   falsifier_test, rollback_cmd}` — and rolls back on canary failure.
4. **Replayable runs.** Audit log under `.harness/audit/events.jsonl`
   is append-only; `harness audit replay` reconstructs any past run.
5. **Local-first.** No cloud, no telemetry, no SaaS. State lives
   under `.harness/` per repo. Backups portable via `harness backup`.
6. **Multi-adapter routing.** `router.Pick(strengths)` selects the
   right CLI per task type. Critic loop re-routes diffs to a
   different adapter (`tags:["review","critic"]`).
7. **Domain-agnostic flows.** Same harness drives `rails-api` AND
   `meta-ads-campaign`. Phases: `deterministic|llm|sensor`.

---

## Paper followed

**"Code as Agent Harness"** — arXiv 2605.18747.

Maps to repo:

| Paper section | Repo surface |
|---|---|
| § Executability | `internal/execution`, `internal/adapters`, `internal/scaffoldpkg` |
| § Verifiability | `internal/sensors`, `internal/audit`, `internal/optimize` |
| § Statefulness | `internal/sharedstate`, `internal/recall`, `internal/devloop` (checkpoint/resume) |
| § Composability | `internal/flowpkg`, `cmd/harness/cmd_do.go`, `internal/critic` |
| § Trajectory metrics | `internal/execution/types.go::Trajectory` |
| § Long-term memory | `internal/recall/bm25.go` (+ optional embeddings) |
| § Multi-agent critic | `internal/critic/critic.go` |
| § Human-in-the-loop | `internal/autonomy/{approvals,suggest}.go` |
| § Multimodal grounding | `internal/multimodal/grounding.go` |

Gap closure log: [docs/paper-coverage-map.md](docs/paper-coverage-map.md).

---

## GitFlow + contribution rules

- Branches: `feature/p<NN>-slug` → `develop` → `main`
- Every release ships a spec in `docs/specs/p<NN>-*.md` **first**
- Conventional Commits enforced by pre-commit
- Pre-push gate (`scripts/git/pre-push.sh`) blocks: lint failures,
  red tests, coverage below floor, direct push to main/develop
- Escape hatch: `HARNESS_SKIP_CI=1 git push` (CI re-runs on PR)
- No god files (file LOC ≤ 400), no useless comments, English in
  code, brew formula refresh per release

Full rules: [CONTRIBUTING.md](CONTRIBUTING.md).

---

## Release status

| Release | What landed |
|---|---|
| v0.72–v0.81 | Phase A: 10 paper-gap closures (trajectory, evidence, checkpoint, change contracts, sharedstate, replan, BM25, critic, autonomy, multimodal) |
| v0.82–v0.91 | Phase B: coverage 50% → 58% core; gate hardened |
| v0.92–v0.95 | Phase C: flowpkg + 6 bundled flows + `harness flow init` |
| v0.96 | Phase D: dashboard parity audit (18/18 handoff screens covered) |
| v0.97 | Phase D: ARCHITECTURE.md |
| v0.98 | Phase E: pprof helpers + `make profile-mem/cpu` |
| v0.99 | Phase D: version bump (README rewrite slipped) |
| v0.100 | Phase D: full README rewrite (this doc) |

**v1.0.0:** deferred indefinitely. Stays on `0.xx.xx` until operator
explicitly green-lights cut. No timeline.

---

## License

MIT. See [LICENSE](LICENSE).
