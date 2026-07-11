<!-- privatecheck: allow — spec documenting the guardrail's blocklist. -->
# p47 — Privatize YouTube channel material + guardrail

## Problem

`docs/youtube/` shipped 10 markdown files (10,478 lines, 424 KB) of
personal marketing content for Rodolfo Peixoto's YouTube channel
(SEO plans, thumb briefs, brand/title matrices, roteiros, shorts
factory, playlists growth, engagement growth 2026, série de vídeos).

harnessx is an open-source MIT project. Personal marketing IP does not
belong in the public tree, and a future contributor could easily copy
the pattern and add more.

## Decision

1. Remove `docs/youtube/` from the git tree. Preserve locally at
   `~/harnessx-notes/youtube/` so the author does not lose the work.
2. Add persistent guardrails so the same class of file cannot land
   again by accident.

## Guardrails

### Layer 1 — `.gitignore`

Extend `.gitignore` with `docs/youtube/`, `docs/marketing/`,
`docs/canal/`, `docs/roteiros/`, plus common runtime dirs that are
already personal (`.claude/`, `.harness/config/active.yaml`,
`.harness/runs/`, `.harness/sessions/`).

### Layer 2 — `scripts/git-hooks/pre-commit-privatecheck.sh`

Fails the commit when staged files match any of:

- Path prefixes: `docs/youtube/`, `docs/marketing/`, `docs/canal/`,
  `docs/roteiros/`
- Filename markers: `THUMB`, `SHORTS`, `ENGAGEMENT-GROWTH`,
  `BRAND-*MATRIX`, `SEO-RETENCAO`, `PLAYLISTS-GROWTH`,
  `SERIE-DENTRO-DO-HARNESSX`, `SERIE-PAPERS-NA-PRATICA`
- Content keywords (only in text files, self-file skipped):
  `CTR ≥`, `AVD ≥`, `Subs 30d`, `thumb brief`, `Shorts factory`,
  `canal do YouTube`

Wired from the existing `pre-commit` hook so it runs on every commit,
not only when Go files change. Bypass: `HARNESS_ALLOW_PRIVATE=1`
(audited in review).

### Layer 3 — `CONTRIBUTING.md`

New "Open-source boundary" section documenting the rule and the
bypass.

## Verification

- `bash scripts/install-hooks.sh` copies the new hook into
  `.git/hooks/`.
- Regression test: staging `docs/youtube/test.md` and running
  `git commit` fails with the guardrail message. Confirmed manually
  in this branch.

## Scope boundaries

Only removes personal marketing material. No product code touched.
Existing runtime configs already ignored (`.harness/artifacts/`,
`.harness/db/`, etc.) untouched.
