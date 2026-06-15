# Changelog

Format: [phase] short summary, then bullet list of concrete additions.
Newest milestones at the top. Dates are when the milestone landed in repo.

## 2026-06-15 — v0.6.0 — Doctor probe fix + harness install (P36)

- **Probe parser** in `internal/adapters/execprobe` now captures stdout+stderr, runs an optional `VersionRegex` override per probe, and treats non-zero exits as success when a semver match is extracted. Fixes the "present, version probe failed" warning for `go` and `gemini`.
- **`harness install <tool>`** with sub-commands `list` and `show`. Reads bundled YAML manifests under `internal/install/manifests/*.yaml` and runs the first viable per-platform strategy (brew, apt, dnf, pacman, go_install, npm_global, cargo_install, pip_user). Flags: `--dry-run`, `--upgrade`.
- **16 bundled manifests**: gopls, ripgrep, syft, ruby-lsp, solargraph, pyright, basedpyright, rust-analyzer, tsserver, gemini, claude, codex, kimi, golangci-lint, govulncheck, gitleaks.
- **`harness doctor`** appends a `Recommended installs` section listing actionable `→ harness install <name>` lines for every ⚠/✗ probe that has a bundled manifest.

## 2026-06-15 — v0.5.0 — Release channels (stable/beta/develop) + harness help topics (P35)

- **`harness update --channel stable|beta|develop`**: stable picks newest non-prerelease, beta includes pre-releases (`-beta*`, `-rc*`), develop builds from source. Aliases: `harness upgrade`, `harness self-update`.
- **`harness update status`**: compares current binary against the channel's latest, reports up-to-date / update available / current-is-newer.
- **`harness update channels`**: lists every release per channel newest-first with publish date and HTML URL.
- **`harness help <topic>`**: in-CLI tutorials for quickstart / agents / sensors / hooks / autonomy / mcp / update / input / tracker — no doc lookup required.
- **`internal/update`**: stand-alone channel resolver + version comparator (handles `v` prefix, `-beta`, `-rc` suffixes); unit-tested with deterministic release fixtures.

## 2026-06-15 — v0.4.0 — Real agentic execution + MCP/hook integration + workflow improvements (P31–P34)

- **P31 Executor**: real agentic loop wired end-to-end — worktree (git or copy fallback), adapter.Run, stdout/stderr capture, diff capture (unified patch + stat + JSON), sensors bridge, autonomy gate, apply via `git apply --3way` or `waiting_approval`, persisted `meta.json` + `report.md` per run under `.harness/runs/<id>/`.
- **P31 fake-agent**: deterministic Claude-shaped JSON adapter for tests and the e2e smoke (`cmd/fake-agent`, `templates/agents/fake-real.yaml`).
- **P31 run CLI**: `harness runs list|inspect|report|sensors|approve|discard`; `harness execute` direct path; `harness feature/bugfix/run --agent <id> --apply --autonomy <level>` routes through the executor.
- **P32 MCP injection**: when adapter capability `mcp=true`, executor merges `mcpscan.Scan` output into `runs/<id>/mcp-config.json` and appends `--mcp-config <path>` via `AgentRequest.ExtraArgs`.
- **P32 Hook dispatch**: pre-tool-use and post-tool-use hooks fire around `adapter.Run`; non-zero pre-hook routes to `autonomy_denied` unless level is `full_project_loop`.
- **P32 `harness mcp install`**: writes `.harness/mcp/<name>.json` with `--command/--url/--transport/--yes`.
- **P34 trivial fast path**: `intent.Complexity {Trivial, Standard, Complex}` heuristic; `workflow.askAgent` skips spec+plan+worktree+diff for question-style prompts and writes the answer to `report.md`.
- **P34 prompt enhancement**: `internal/promptenh` deterministically composes skill prefix + context summary + task; persisted as `enhancement.json` per run.
- **P34 cost auto-routing**: `router.PickModel(adapter, complexity)` picks cheap / default / deep model from the adapter's `models` map.
- **P34 autonomy policy**: `.harness/config/autonomy.yaml` supports `allow_paths|deny_paths|allow_commands|deny_commands`; deny wins, allow-list non-match downgrades to `require_approval`.
- **P34 multi-input flags**: `--prompt-file`, `--pdf` (via local `pdftotext`), `--image` (base64 attachment for vision-capable adapters).
- **P34 auth login hints**: adapter YAML `auth: { login_command, doc_url }` surfaced on certify auth failure (`run: claude login | docs: ...`).
- **P34 tracker CLI**: `harness metrics --since 1d|7d|30d|all` aggregates per-run state; `harness audit --kind <k> --limit N` reads the audit JSONL.
- **P34 update**: `harness update` self-updates from latest GitHub release (verifies SHA-256, swaps binary in place); `harness update --channel develop` builds from source.
- **Tutorial**: `docs/tutorial-v0.4-manual.md` walks every surface end-to-end in English with the login matrix per agent CLI.

## 2026-06-15 — v0.3.0 — Stack Audit + Handoff parity + MCP/Hook scanners (P23–P29)

- **P23 stack audit**: deterministic visual + functional audit pipeline.
- **P24 audit bundle**: single zip + BUNDLE_INDEX.md for LLM consumption.
- **P25 audit hardening**: pass-rate 23% → 100%.
- **P26 handoff gap report**: 11 new routes, nav 7 → 17, audit features 10 → 22.
- **P27 Home/Projects/Catalog rich content**: MetricCard, PathCell, TerminalReflection, ActionService.
- **P28 MCP/Hook scanners + rich Cleanup page**: deterministic discovery + CLI + HTTP + UI.
- **P29 finalise handoff**: 8 remaining stub pages rewritten; Catalog MCP/Hook tabs wired.

Final audit: 34/34 passed, 0 console_error, 0 selector_missing, 0 layout_collapsed.

## 2026-06-14 — v0.2.0 — Workspace Hub + Capabilities + Stack Tour (P11–P22)

- **P11 Workspace Hub** — Multi-project registry at `~/.harness/registry.sqlite`;
  `harness project add|list|switch|archive|scan|forget|import|stale`; HTTP
  `/api/workspace/{projects,switch,current,import,stale[/slug]}`. Resolver
  precedence: --project flag > HARNESS_PROJECT env > cwd walk-up > active row.
- **P12 Capabilities Center** — 8 kinds (agent/mcp/hook/sensor/skill/context/
  resource/plugin) with deterministic discovery, Plan→approval→Apply pipeline
  (path-traversal guard + stage-then-rename), bundled manifests, HTTP
  `/api/catalog/{kinds,items,plan}`, CLI `harness catalog`.
- **P13 Cleanup Engine** — Scanner + detectors (worktrees, caches, abandoned
  .harness, large files, VM leftovers, Claude leftovers, containers via
  runtime/containers.Lister). D5 two-key rule: policy match OR interactive y
  OR HARNESS_CLEANUP_I_UNDERSTAND=1. Every delete writes audit_event with
  sha256.
- **P14 Coverage gates** — `coverage-gate.sh` (go), `coverage-web.sh` (vitest
  + @vitest/coverage-v8), `coverage-shell.sh` (bashcov). CORE regex covers
  every new pkg. Entry thresholds 50/55 with documented ratchet to 90 by v0.3.
- **P15 Design System** — `web/dashboard/src/ds/`: tokens + strings + Badge,
  Card, EmptyState, Tabs, InspectorPanel, DataExplorer, Shell. App wired via
  Shell with nav as data. 11 DS unit tests.
- **P16 Import Wizard + Stale Detection** — Shared `importwiz.Plan` powering
  both CLI + future UI; `internal/stale` fingerprints package files into
  `.harness/project/fingerprints.json`; Detect reports changes by kind.
- **P17 Command Palette ⌘K** — `internal/palette` (Score: exact 100 > prefix
  80 > contains 60 > fuzzy 20+), sources for Projects/Capabilities/Builtin
  commands. CLI `harness palette search`, HTTP `/api/palette`, React modal
  with arrow-key navigation.
- **P18 Autonomy + Health** — 5 levels (Manual/PlanAndAsk/SafeExecute/
  FullProjectLoop/ScheduledMaintenance) with declarative Gate matrix.
  Deterministic 10-subsystem health score. CLI `harness autonomy get|set`,
  `harness health show`. HTTP `/api/autonomy`, `/api/health/score`.
- **P20 Stack Tour** — `harness stack tour [--dashboard --keep]` deterministic
  walkthrough (workspace add → catalog install → cleanup scan → autonomy
  gate → health score → optional /api/health probe). `harness stack status`
  reports dashboard online/offline. Self-cleaning by default.
- **P21 Role Grid** — `web/dashboard/src/auth/{roles.ts,RoleContext.tsx}`
  (anonymous|operator|admin), role-grid.test.tsx walks every page × every
  role. Vitest total: 15 files / 67 tests.
- **P22 Container lifecycle harness** — `internal/runtime/containers` shared
  by Cleanup + Stack Tour: typed docker Lister, Compose Up/Down, HealthProbe
  with bounded backoff, VerifyClean asserting `docker ps -a` empty.

Install: `curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh | bash`
(now accepts `--dry-run` and `--prefix`).

## 2026-06-14 — GitFlow hardening sweep (G1–G7)

- G1: GitFlow — `main` (releases) + `develop` (integration);
  feature/fix/chore/release/hotfix prefixes documented in
  `CONTRIBUTING.md`.
- G2: God-file refactor — `internal/app/workflow/workflow.go` 517→215 LOC
  (split into `telemetry.go`, `helpers.go`, `execute.go`);
  `internal/adapters/http/server.go` 559→104 LOC (split into
  `handlers.go`, `static.go`, `helpers.go`). No file > 260 LOC.
- G3: Go coverage push — `registry_test.go` (agents 0→100%),
  `defaults_test.go` (router 59→70%), `snapshot_test.go` (optimize
  52→73%), `build_test.go` (design 56→77%), `builder_more_test.go`
  (context 42→54%). Global coverage 47.9%.
- G4: React component tests — Sessions, SessionDetail, RunDetail,
  Sensors, Agents, Memory, Design, Roadmap, Settings. 20 tests passing.
- G5: Shell test harness — `scripts/lib/assert.sh` + 4 suites in
  `scripts/tests/`; wired into `make ci` via `make test-sh`.
- G6: Coverage gate ratcheted — `GLOBAL_MIN` 35→40, `CORE_MIN` 35→50.
  Lowest core: `internal/context` 53.7%.
- G7: Final sweep — `make ci` green (lint + tests + shell harness +
  coverage gate + 10 e2e phases); `make security` clean (vet, gitleaks,
  harness security-audit).

## 2026-06-14 — Senior-engineering polish (B1–B12 + lint clean + local CI)

- Lint: full `.golangci.yml` (v2) with errcheck, govet, staticcheck,
  ineffassign, unused, misspell, revive, nolintlint, gocyclo (15),
  gocognit (25), gosec, unconvert, gocritic, dupl. `golangci-lint run`
  reports **0 issues**.
- Coverage gate: `scripts/coverage-gate.sh` + `scripts/coverage-aggregate.py`
  enforce global ≥ 35% / core ≥ 35% with a documented ratchet plan.
- Supply-chain: `make licenses` runs `go-licenses` → `THIRD_PARTY_LICENSES.md`
  + `NOTICE` + CSV; blocks AGPL/GPL/LGPL/SSPL/EUPL. `make sbom` produces
  CycloneDX 1.5 JSON via syft or stdlib-Python fallback.
- Security: `make security` runs `govulncheck` advisory + `harness
  security-audit` (forbidden_files + forbidden_commands + secrets_scan +
  go_vuln). `.harnessignore` excludes the scanners' own pattern literals.
- Doctor: new categories `lsp` + `quality` — probes for gopls, ruby-lsp,
  solargraph, pyright, basedpyright, rust-analyzer, typescript-language-server,
  golangci-lint, govulncheck, go-licenses, gitleaks, syft.
- Cmd split: `cmd/harness/main.go` 720 LOC → 85 LOC; 19 `cmd_*.go` files;
  shared `cwd()` helper.
- Constants + i18n: `internal/platform/constants` central magic values;
  `internal/platform/i18n` embedded `en` + `pt` bundles, `HARNESS_LANG`
  env, fallback chain.
- Hooks: `make install-hooks` writes `pre-commit` (gofmt + go vet),
  `commit-msg` (Conventional Commits regex), `pre-push` (`make ci`).
- Release: `make release` cross-builds darwin/linux × amd64/arm64,
  enforces 40 MiB per-binary budget, emits SHA-256 sums.
- Docs: `COMPLIANCE.md`, `docs/spec-driven-development.md`, mermaid
  diagrams in `README.md`, refreshed `AGENTS.md`, `CLAUDE.md`,
  `CONTRIBUTING.md` (GitFlow + branch-protection note), SBOM script.
- New commands: `harness explain`, `harness session show`, `harness
  artifact ls/cat`, `harness skill list/promote`, `harness spec init`,
  `harness routes`, `harness completion <shell>`, `harness memory
  list/promote`.
- `harness logs --follow` Bubble Tea TUI; embedded React dashboard via
  `//go:embed all:web/dashboard/dist`.
- AutoLSP: any of gopls / ruby-lsp / solargraph / pyright /
  basedpyright / rust-analyzer / typescript-language-server is spawned
  on demand when its binary + manifest are present.


## 2026-06-14 — Hardening 4: gopls LSP client

- `internal/adapters/lsp/gopls.go` full LSP-over-stdio client (Content-Length
  framing, JSON-RPC 2.0, async demux, mutex-guarded writer, atomic IDs).
- `DocumentSymbols` (hierarchical + flat parsers), `Diagnostics` (drains
  `publishDiagnostics` notifications), cached per spec §15 layout.
- `internal/context.AutoLSP(root)` wires gopls automatically when binary is
  present and `go.mod` exists. Missing binary silently falls back to the
  default provider chain.

## 2026-06-14 — Hardening 3: memory CLI

- `harness memory list [--limit] [--scope]` reads the `memories` table.
- `harness memory promote --content … --run-id … --confidence …` enforces
  the spec §11 evidence gate (rejects missing evidence, low confidence,
  sensitive content).

## 2026-06-14 — Hardening 2: interactive confirmation + Cycle F sensor

- Workflow now prompts `Approve plan? [y/N]` when stdin is a TTY and
  `--yes` is not passed. CI/redirected stdin returns `false`.
- New `performance_budget` sensor compares the most recent perf snapshot
  against `.harness/project/performance-budget.json`.

## 2026-06-14 — Hardening 1: docs + workflow exec wiring

- All eight missing spec §31 docs written (agents, sensors, skills,
  context-engineering, design-to-product, resource-optimization, security,
  dashboard).
- Workflow `--execute` now resolves a route → agent chain and runs it via
  `router.Execute`. Cost + tokens + fallback_from persisted on the run row.

## 2026-06-14 — Phase 10: full end-to-end

- `scripts/e2e-phase10.sh` chains every phase against one project.
- `Makefile` `e2e-all` target runs every phase script.

## 2026-06-14 — Phase 9: resource optimization

- Cycles A (snapshot), B (Dockerfile audit), C (deps classifier), D (log
  scanner), G (report) — pure-Go, no Docker dependency.
- 7 commands: `optimize`, `perf-snapshot`, `perf-compare`, `image-audit`,
  `dependency-audit`, `log-audit`, `security-audit`.
- Conservative dep removal per spec §21 core rule; observability/security/
  debugger deps marked `kept_for_operational_safety`.

## 2026-06-14 — Phase 8: dashboard

- 14 read-only REST endpoints under `/api/*`.
- React SPA w/ 9 routes (Sessions, SessionDetail, RunDetail, Sensors,
  Agents, Design, Roadmap, Memory, Settings); loading/empty/error states
  for every panel.
- Bubble Tea `harness logs --follow` TUI (750 ms poll, rotation-aware).
- Built-in HTML fallback page so the dashboard works without `npm install`.

## 2026-06-14 — Phase 7: design-to-product

- Safe ZIP extractor (zip-slip rejection + 200 MiB cap).
- Inventory pages, components, assets, CSS tokens, JS interactions,
  missing states, responsive notes.
- Six product maps: design-manifest, feature-map, toggle-map, roadmap,
  api-contracts, flow-map.
- Image hash + metadata cache (vision-model `detected` field reserved).

## 2026-06-14 — Phase 6: spec + plan workflow

- Rule-based intent classifier (8 modes + explainable reasons).
- §8 spec renderer with safe defaults per mode.
- §9 plan renderer with confirmation status.
- Budget guard, evidence-gated memory promotion, §28 report writer.
- Six use cases (ask/plan/run/feature/bugfix/report) + natural form
  `harness "<prompt>"`.

## 2026-06-14 — Phase 5: context engineering

- Pack builder with 4 default providers (memory, git, ripgrep, test-map).
- LSP provider abstraction + cache key layout per spec §15.
- Token estimator (4-chars heuristic; pluggable).
- Canonical hash over `task + profile + providers + git HEAD`.

## 2026-06-14 — Phase 4: sensors

- 4 universal scanners (`forbidden_files`, `forbidden_commands`,
  `secrets_scan`, `changed_files`).
- Per-stack rule packs (Go, Node/React/Next.js/Vite, Rails, Python, Rust,
  Docker) — OptionalTool=true skips on missing binaries.
- Runner with computational-first ordering + streaming callback.
- `harness sensor list|run`, `harness check`, `harness ci`.

## 2026-06-14 — Phase 3: agent adapter system

- `AgentAdapter` interface + Capabilities + FailureType (IsRecoverable).
- YAML loader w/ template substitution + tiny JSONPath subset.
- Bundled Claude/Codex/Gemini/Kimi/Fake adapter YAMLs.
- Router w/ explainable Select + fallback Execute.
- Certification suite (healthcheck → simple_prompt → timeout → cancellation
  → failure_classification).

## 2026-06-14 — Phase 2: project index

- Stack detector (Rails, React, Next.js, Vite, Node, Go, Python, Rust, Docker).
- 8 maps: profile, commands, dependencies, architecture, test-map, api-map,
  design-system, performance-budget.
- Incremental cache keyed on input fingerprint.

## 2026-06-13 — Phase 1: core CLI

- `harness init` bootstraps `.harness/` + SQLite schema + bootstrap session.
- `harness doctor` Lip Gloss panel with system + tools + agents + project.
- `harness logs --tail` JSONL viewer.
- 8 SQLite tables from spec §23 (sessions, runs, sensor_results, metrics,
  memories, skill_versions, agent_certifications, artifacts).

## 2026-06-13 — Phase 0: scaffold

- Go module, MIT LICENSE, Makefile, golangci.yml, GitHub Actions CI.
- React + Vite + TS dashboard scaffold w/ Vitest smoke test.
- `templates/.harness/` defaults.
- Docs (overview, architecture, install, quickstart, cli-reference,
  configuration, contributing).
