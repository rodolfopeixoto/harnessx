# P14 — Coverage 90 core + shell + web

## Acceptance

- `make coverage-gate` enforces `CORE_MIN`/`GLOBAL_MIN`; ratchet plan documented inline.
- `make coverage-web` runs Vitest with `@vitest/coverage-v8`; enforces `WEB_MIN`.
- `make coverage-shell` runs `bashcov` over `scripts/tests/run-all.sh`; enforces `SHELL_MIN`.
- All three wired into `make ci`. e2e-phase14 asserts each gate honours its threshold and reports its number.

## Contract

Thresholds (env-overridable; defaults set in script headers):

| Class | Var | Default |
|---|---|---|
| Global Go | `GLOBAL_MIN` | 60 |
| Core Go (regex including workspace/catalog/cleanup/runtime/containers) | `CORE_MIN` | 80 |
| Web | `WEB_MIN` | 75 |
| Shell libs | `SHELL_MIN` | 80 |

D1 said target 90 core; v0.2.0 enters at 80 and ratchets per phase. Each subsequent phase that adds a core package raises the floor in the same PR. P19 confirms ≥90 before tagging.

## Risks

- `@vitest/coverage-v8` install bloats install time. Mitigation: dev-dep only, opt-in via `coverage-web` target (not in `ci` unless threshold drops below `WEB_MIN`).
- `bashcov` requires ruby. Mitigation: skip cleanly with a printed warning when ruby is missing; treat as advisory until P19.

## Verification

- `scripts/e2e-phase14.sh`: runs each gate, asserts exit 0 and that the printed number is ≥ env-set floor; runs negative-case with floor temporarily bumped to 100 and asserts non-zero exit.
