# AGENTS.md — Guidance for AI agents working on HarnessX

This file is the operational version of `CLAUDE.md`. Read it once per
session and apply the rules below before touching code.

## Always

- **Phase boundaries are real.** `HARNESSX-MASTER-PLAN.md §9` lists every
  package's phase. A bug fix in Phase 6 must not silently add Phase 8
  dashboard work.
- **No hardcoded constants.** Shared values live in
  `internal/platform/constants`.
- **All user-facing strings via `i18n.T(key)`** — add the English copy to
  `internal/platform/i18n/locales/en.json` first.
- **No CGO.** `modernc.org/sqlite` only.
- **Clean Architecture.** `domain` imports nothing; `app` imports `domain`
  + ports it owns; `adapters` implement those ports.
- **Tests at the seam.** New behaviour ships with a test at the public
  surface, not behind a private helper.
- **Comments only when they explain *why*.** Non-obvious business rule,
  security constraint, performance tradeoff, external quirk, temporary
  workaround. Do not restate what the code says.
- **Conventional Commits enforced by `commit-msg` hook.** Subject ≤ 50
  chars; body wraps at 72.
- **GitFlow.** Branch from `develop`. `main` is releases only.
- **Local CI/CD only — no GitHub Actions.** `make install-hooks` wires
  the pre-push hook that runs `make ci`. Never push without it green.

## Never

- Implement features without a spec entry (use `harness feature` to
  generate one).
- Skip tests "to keep the PR small". Tests are part of the change.
- Delete tests to make a sensor pass — blocked by `forbidden_files`.
- Send the entire repository to an LLM. Use `internal/context.Build`.
- Mutate project memory without an `evidence_run_id` + confidence ≥ 0.4.
- Add `..` in `//go:embed` paths.
- Commit `.env`, secrets, tokens.
- Force-push `main` or `develop`.
- Run heavy long-lived processes (Next.js dev server, Playwright, Docker
  compose) in parallel with unrelated work. Kill child processes and
  `t.Cleanup` tempdirs in tests.

## Workflow (per issue)

1. Read the issue. Wait for the `accepted` label before serious effort.
2. `git checkout -b feature/<short-name> develop` (or `fix/*`, `chore/*`).
3. `harness feature "<prompt>" --yes` produces
   `.harness/artifacts/specs/<id>.md`. Edit until it reflects the truth.
4. Implement the smallest scope that satisfies the spec.
5. `make ci` green locally. `make e2e-all` for any CLI surface change.
6. Open the PR against `develop`. Fill the template completely.
7. Address review comments before pushing more commits.

## Available subsystems

| Area | Package | Doc |
|---|---|---|
| Agents (CLI adapters + cert) | `internal/agents` | `docs/agents.md` |
| Sensors (deterministic gates) | `internal/sensors` | `docs/sensors.md` |
| Context engineering | `internal/context` + `internal/adapters/lsp` | `docs/context-engineering.md` |
| Project index | `internal/index` | spec §13 |
| Spec / plan / report | `internal/spec`, `internal/plan`, `internal/app/reportcmd` | `docs/spec-driven-development.md` |
| Workflow orchestrators | `internal/app/workflow` | spec §7 |
| Design-to-product | `internal/design` | `docs/design-to-product.md` |
| Resource optimisation | `internal/optimize` | `docs/resource-optimization.md` |
| Dashboard (HTTP + React) | `internal/adapters/http` + `web/dashboard` | `docs/dashboard.md` |
| Skills | `internal/skills` | `docs/skills.md` |
| Memory | `internal/memory` | spec §11 |
| Security posture | `internal/sensors` + `internal/memory` | `docs/security.md` |

## When in doubt

- `HARNESSX-MASTER-PLAN.md` end-to-end.
- `CLAUDE.md` if you are an LLM (same rules, faster to read).
- Open a `[question]` issue. The answer becomes a docs PR.

## Resource hygiene (inherited from operator's environment)

- Always use timeouts on subprocess calls. `execprobe.Probe` defaults to
  2 s.
- Kill processes started in tests with `t.Cleanup`. Don't leave dangling
  `next-server`, `playwright`, `cypress`, `node`, or `docker` instances.
- Always run `harness ci` before committing on the branch.
- If a task crosses 10 minutes, take a `harness perf-snapshot` mid-flight
  so the next snapshot has something to compare against.

## Drift control

`AGENTS.md` and `CLAUDE.md` must stay in sync with code reality. If you
add a subsystem, a sensor pack, or a constraint, update this file in the
same PR. Drift between docs and code is the #1 risk.
