# HarnessX

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21%2B-00ADD8.svg)](https://go.dev)
[![Paper](https://img.shields.io/badge/paper-arXiv%202605.18747-b31b1b.svg)](https://arxiv.org/abs/2605.18747)
[![Coverage (new pkgs)](https://img.shields.io/badge/coverage%20(new%20pkgs)-%E2%89%A590%25-brightgreen.svg)](docs/PAPER-MAPPING.md#coverage-status)
[![Smoke matrix](https://img.shields.io/badge/smoke%20matrix-6%20stacks-green.svg)](docs/COMMANDS.md#harness-smoke-matrix---langs-csvall---keep---json---step-timeout-duration---bin-path)
[![Single binary](https://img.shields.io/badge/single%20binary-no%20CGO-brightgreen.svg)](#install)

> **HarnessX** is a local-first, single-binary runtime for **agentic
> software engineering**. It turns one prompt into a deterministic
> Plan-Execute-Verify loop across the coding CLIs you already use
> (Claude Code, Codex, Gemini, Kimi, Ollama, …), gates every step with
> deterministic sensors, persists the entire trajectory under
> `.harness/`, and supports HITL-governed self-evolution.
>
> The implementation is end-to-end anchored on the survey paper
> **"Code as Agent Harness"** (Ning et al., UIUC / Meta / Stanford,
> [arXiv 2605.18747v1](https://arxiv.org/abs/2605.18747), May 2026).

```bash
brew tap rodolfopeixoto/harnessx https://github.com/rodolfopeixoto/harnessx
brew install harness
harness new python ./shop-api --yes
cd shop-api && harness ship "add /healthz endpoint with pytest"
```

---

## Table of contents

- [Why HarnessX](#why-harnessx)
- [Quick start](#quick-start)
- [Install](#install)
- [The 8 commands you actually need](#the-8-commands-you-actually-need)
- [How it maps to the paper](#how-it-maps-to-the-paper)
- [Documentation](#documentation)
- [Verification surface](#verification-surface)
- [Contributing](#contributing)
- [Community standards](#community-standards)
- [License](#license)
- [Citation](#citation)

---

## Why HarnessX

The paper argues that production agent systems are bottlenecked by
their **harness** — the software layer that surrounds an LLM with
tools, sandboxes, memory, validators, permission boundaries, execution
loops, and feedback channels — not by the base model. HarnessX is a
concrete, open-source implementation of that view.

| Property the paper requires | How HarnessX delivers it |
|---|---|
| **Executable** — every action is code we can run | One binary, single-process, no CGO. Adapters wrap the existing coding CLIs. |
| **Inspectable** — every action emits structured state | `.harness/logs/events.jsonl` append-only telemetry + `.harness/artifacts/` for diffs, blackboards, plans. |
| **Stateful** — state outlives a single invocation | SQLite store + plan-as-contract artefacts + shared blackboard + 5-kind memory taxonomy. |
| **Governed** — mutations are auditable + HITL-gated | `harness evolve promote --hitl` and `harness config wizard` both append to audit logs. |

### What you get out of the box

- **Single-command SDLC** — `harness ship "<prompt>"` branches, writes
  a spec, loops `do`+`ci` with 429-aware fallback, gates against your
  plan contract, and commits using Conventional Commits.
- **Goal-aware REPL** — `harness chat --goal dev|ads|research|ops`
  emits typed JSON plans and dispatches them against a deterministic,
  per-goal palette.
- **Plan-as-contract** — `harness plan write` materialises the paper's
  §3.4.2 contract; `harness plan check` and the `plan_scope` sensor
  enforce it before commits.
- **Multi-agent orchestration** — declare role + topology in
  `.harness/orchestrations/<name>.yaml`; `harness orchestrate run`
  shares a file-only blackboard between roles.
- **Self-evolving harness with guardrails** — `harness evolve
  diagnose|sandbox|propose|promote --hitl` implements paper §3.5 with a
  real A/B replay sandbox.
- **6-stack scaffolds** — Go, Python, Rails, React, Ruby, Rust —
  every one ready for `harness ship` out of the box.
- **Deterministic regression matrix** — `harness smoke matrix`
  exercises every CLI command across every scaffolded stack.

---

## Quick start

```bash
# 1. Bootstrap a Python e-commerce backend
harness new python ./shop-api --yes
cd shop-api

# 2. Write a deterministic plan contract (no LLM call)
harness plan write "build products + cart + checkout endpoints" \
  --file app.py --file tests/test_app.py \
  --invariant "/healthz still returns 200" \
  --validate "harness ci" --risk medium
PLAN_ID=01...

# 3. Pin the contract so sensors enforce it
printf "active_plan_id: %s\n" "$PLAN_ID" > .harness/config/plan.yaml

# 4. Configure the router (audited)
harness config set --task implementation --primary kimi \
  --fallback gemini,claude --budget 0.5

# 5. Ship the first feature end-to-end
harness ship "implement GET /products with pytest" --plan ${PLAN_ID}

# 6. Iterate via the goal-aware REPL
harness chat --goal dev --adapter ollama
```

Full end-to-end walk-through: [docs/TUTORIAL-ECOMMERCE.md](docs/TUTORIAL-ECOMMERCE.md).

---

## Install

### macOS (Homebrew)

```bash
brew tap rodolfopeixoto/harnessx https://github.com/rodolfopeixoto/harnessx
brew install harness
```

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

Every release ships 6 platforms (darwin-amd64, darwin-arm64,
linux-amd64, linux-arm64, windows-amd64, windows-arm64) with `.sha256`
companions. Download from the
[Releases page](https://github.com/rodolfopeixoto/harnessx/releases).

After install, run:

```bash
harness doctor          # diagnose host environment
harness --help          # full command tree
```

---

## The 8 commands you actually need

| Command | Use it when |
|---|---|
| `harness new <stack> <path>` | Bootstrap a new project (git + .harness + scaffold + hooks) |
| `harness ship "<prompt>"` | Branch → spec → do → ci-loop → conventional commit |
| `harness chat --goal dev` | Iterative REPL with a typed plan + deterministic dispatch |
| `harness plan write` / `plan check` | Materialise + enforce the §3.4.2 plan-as-contract |
| `harness ci` | Run every applicable sensor; exits non-zero on red |
| `harness coverage --threshold 0.9` | 90% coverage gate for Go projects |
| `harness orchestrate run <flow>` | Multi-role flow with shared blackboard |
| `harness evolve diagnose|sandbox` | Telemetry-driven harness evolution with HITL gates |

Everything else (`harness scaffold`, `harness memory`, `harness audit`,
`harness backup`, `harness dashboard`, `harness smoke matrix`, …) is
listed in [docs/COMMANDS.md](docs/COMMANDS.md).

---

## How it maps to the paper

Every paper section is implemented in a specific package with a
specific command and a specific test path. The full table lives in
[docs/PAPER-MAPPING.md](docs/PAPER-MAPPING.md); the short version:

| Paper layer | HarnessX surface |
|---|---|
| **§2 Interface** — reasoning / acting / environment | adapters (`internal/agents`), scaffolds (`internal/scaffoldpkg`), context providers (`internal/context`) |
| **§3.1 Planning** — orchestration, structure-grounded, linear | `harness chat`, `harness orchestrate`, `harness do`, `internal/intentplan`, `internal/customrules` |
| **§3.2 Memory** — 5 kinds taxonomy | `harness memory list --kind {working\|semantic\|experiential\|long_term\|multi_agent}` |
| **§3.3 Tool Use** | adapter contract + sandboxed runtime |
| **§3.4 PEV** — Plan-Execute-Verify | `harness ship`, `harness loop`, sensors catalog, `harness plan write/check`, `harness coverage` |
| **§3.4.2 Plan-as-contract** | `harness plan write`, `harness plan check`, `plan_scope` sensor, `harness ship --plan <id>` |
| **§3.4.4 Deterministic sensors** | universal pack + stack pack + `go_coverage_gate` + `plan_scope` + `commentscan` |
| **§3.5 AHE** — Agentic Harness Engineering | `harness evolve diagnose|propose|replay|sandbox|promote --hitl` |
| **§3.5.3 Governed mutation** | `harness evolve promote --hitl` + `harness config wizard` (audit log) |
| **§4.1.1 Roles** — Manager/Planner/Coder/Reviewer/Tester | `internal/orchestrate.Role*` |
| **§4.1.3 Topology** — chain/cyclic | `internal/orchestrate.Topology*` |
| **§4.3.1 Shared substrate** | `.harness/artifacts/runs/<id>/blackboard.json` |
| **§5.1.1 Code assistants** | `harness new`, `harness ship`, `harness chat`, wrappers |
| **§5.2 Open problems** | Coverage gate, scope gate, A/B replay sandbox, HITL governance |

---

## Documentation

| Doc | Purpose |
|---|---|
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Runtime architecture, layered explanation, disk layout, data flow for every entry point |
| [docs/COMMANDS.md](docs/COMMANDS.md) | Every `harness` command on one page (flags, env vars, exit codes, output paths) |
| [docs/PAPER-MAPPING.md](docs/PAPER-MAPPING.md) | Paper § → command → package → test |
| [docs/TUTORIAL-ECOMMERCE.md](docs/TUTORIAL-ECOMMERCE.md) | **Recommended** — build a small FastAPI e-commerce backend end-to-end |
| [docs/tutorial-python-demo.md](docs/tutorial-python-demo.md) | Original Python FastAPI walkthrough |
| [CONTRIBUTING.md](CONTRIBUTING.md) | GitFlow, conventional commits, pre-push gate, coverage floor |
| [SECURITY.md](SECURITY.md) | Coordinated disclosure policy |
| [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) | Contributor Covenant |
| [CHANGELOG.md](CHANGELOG.md) | Per-release history |
| [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md) | Bundled dependency licenses |

---

## Verification surface

HarnessX dogfoods its own gates.

| Gate | Command | Status |
|---|---|---|
| Build | `go build ./...` | green |
| Vet | `go vet ./...` | green |
| Unit tests | `make test` | 86 packages, 0 failures |
| New-package coverage floor (90%) | every new internal package | ✓ 14 / 14 packages ≥ 90% |
| Cross-stack regression | `make smoke` → `harness smoke matrix` | 6 stacks (go, python, rails, react, ruby, rust) × 10 commands = 60 invocations green |
| Tutorial regression | `make tutorial-replay` | green |
| Pre-push hook | `harness ci` | runs every applicable sensor, blocks the push on red |
| Coverage floor | `harness coverage --threshold 0.9` | 90% default |
| Scope contract | `harness plan check --plan <id>` | refuses commit on out-of-scope diffs |
| Self-evolution | `harness evolve sandbox <trace>` | real A/B replay across baseline / candidate binaries |

Run everything locally:

```bash
make ci                 # vet + race tests + build + every phase e2e
make smoke              # cross-stack CLI smoke matrix
make tutorial-replay    # deterministic walk of the tutorial
```

---

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md)
first — it covers GitFlow (`feature/p<NN>-slug` → `develop` → `main`),
Conventional Commits, the pre-push gate, and the coverage floor.

Quick path:

```bash
git clone https://github.com/rodolfopeixoto/harnessx
cd harnessx
make install-hooks      # pre-commit + commit-msg + pre-push gates
make check              # vet + race tests + build
make smoke              # cross-stack CLI matrix
make tutorial-replay    # deterministic walk
```

Then branch from `develop`, open a PR with the
[pull request template](.github/PULL_REQUEST_TEMPLATE.md). The
`pre-push` hook (installed by `make install-hooks`) runs `make ci`
before every push. Escape hatch: `HARNESS_SKIP_CI=1 git push` —
documented hotfixes only.

### Reporting issues

Use the issue templates under
[`.github/ISSUE_TEMPLATE/`](.github/ISSUE_TEMPLATE/). For security
vulnerabilities, follow [SECURITY.md](SECURITY.md) (coordinated
disclosure — do **not** open a public issue).

---

## Community standards

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
By participating, you agree to uphold its terms. Maintainers will
enforce the code of conduct in good faith; escalate via the contact
listed in the document.

Code-ownership for review is defined in
[.github/CODEOWNERS](.github/CODEOWNERS).

---

## License

Released under the [MIT License](LICENSE).

Bundled third-party software keeps its own licenses, listed in
[THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md). HarnessX itself
contains no copyleft code and ships as a single statically-linked
binary.

---

## Citation

If you use HarnessX in academic work, please cite both the project
and the paper it implements:

```bibtex
@software{harnessx,
  author       = {Peixoto, Rodolfo and contributors},
  title        = {HarnessX: a code-as-agent-harness runtime for software engineering},
  year         = {2026},
  url          = {https://github.com/rodolfopeixoto/harnessx}
}

@misc{ning2026codeagentharness,
  title         = {Code as Agent Harness: Toward Executable, Verifiable, and Stateful Agent Systems},
  author        = {Ning, Xuying and Tieu, Katherine and Fu, Dongqi and Wei, Tianxin and others},
  year          = {2026},
  eprint        = {2605.18747},
  archivePrefix = {arXiv},
  primaryClass  = {cs.CL},
  url           = {https://arxiv.org/abs/2605.18747}
}
```
