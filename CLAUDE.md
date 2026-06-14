# CLAUDE.md — Sticky instructions for every Claude session

These rules apply to **every** conversation with Claude (or any LLM) in
this repo. Reading this file before opening a tool call is mandatory.

## Always

- **Phase boundaries are real.** See `HARNESSX-MASTER-PLAN.md §9`. Do not
  ship Phase N+1 code while working on Phase N. Bug fixes stay in the
  smallest scope possible.
- **No hardcoded constants.** Shared values live in
  `internal/platform/constants`. If a magic number / string shows up in
  two packages, extract it.
- **English first.** All code, comments, identifiers, log messages, and
  English bundle (`internal/platform/i18n/locales/en.json`) are in
  English. Community translations land via PR.
- **All user-facing strings via `i18n.T(key)`.** Adding a new string
  means: 1) add to `en.json`, 2) call `i18n.T` from the code.
- **Tests at the interface seam.** New behaviour ships with at least one
  test that exercises the public surface — not the private helper.
- **No CGO.** `modernc.org/sqlite` only. Single-binary distribution is
  a hard requirement.
- **Conventional Commits.** Subject ≤ 50 chars. Body wraps at 72.
  Enforced by the `commit-msg` hook.
- **GitFlow.** Branch from `develop` (`feature/*`, `fix/*`, `chore/*`).
  `main` is releases only. See `CONTRIBUTING.md`.
- **Local CI/CD only — no GitHub Actions.** `make install-hooks` wires
  the pre-push gate that runs `make ci` (vet + race tests + build +
  every phase e2e). Never push without it green.
- **Clean Architecture.** `domain` imports nothing. `app` imports `domain`
  + interfaces. `adapters` implement them.
- **Comments only when they explain *why*.** Non-obvious business rule,
  security constraint, performance tradeoff, external quirk, temporary
  workaround with tracking ref. Otherwise none.

## Never

- Implement features without a spec entry (use `harness feature` to
  generate one when starting fresh).
- Skip tests "to keep the PR small". Tests are part of the change.
- Delete tests to make a sensor pass. Hard-blocked by `forbidden_files`.
- Send the entire repository to an LLM. Use `internal/context.Build`.
- Mutate project memory without an `evidence_run_id` + confidence ≥ 0.4.
- Add `..` in `//go:embed` paths.
- Commit `.env`, secrets, tokens, or anything that matches
  `internal/sensors.secretPatterns`.
- Force-push to `main` or `develop`.

## Workflow

1. Read the issue (or open one). Wait for `accepted` label before
   sinking serious effort.
2. Branch: `git checkout -b feature/<short-name> develop`.
3. Write the spec if a feature: `harness feature "<prompt>" --yes`
   produces `.harness/artifacts/specs/<id>.md`.
4. Implement at the smallest scope that satisfies the spec.
5. `make check` green. `make e2e-all` for any CLI-surface change.
   The `pre-push` hook runs `make ci` automatically; do not bypass it
   except for documented hotfixes (`HARNESS_SKIP_CI=1`).
6. PR against `develop`. Fill in the PR template completely.
7. Address review comments before pushing more commits — squash-friendly.

## When in doubt

- Read `HARNESSX-MASTER-PLAN.md` end-to-end.
- Read the doc in `docs/` matching the area you're touching
  (`agents.md`, `sensors.md`, `skills.md`, `context-engineering.md`,
  `design-to-product.md`, `resource-optimization.md`, `security.md`,
  `dashboard.md`).
- If still unclear: open a `[question]` issue. The answer becomes a docs
  PR by a maintainer.

## House-keeping

- Run `harness ci` before every commit on the branch.
- Run `harness perf-snapshot` before + after any change that could
  affect resource use; attach the delta to the PR.
- Run `harness check` after merging upstream changes.

Drift between this file and the rest of the repo is the #1 risk.
Update `CLAUDE.md` first, then implement.
