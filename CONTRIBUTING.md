# Contributing to HarnessX

HarnessX is open source. The bar is high, the surface area is broad, and
the only way the project stays maintainable is if every contributor
follows the rules below. Read them before you open a PR.

## TL;DR

1. Pick (or open) an issue. Wait for `accepted` before sinking a weekend.
2. Branch from `develop` using GitFlow (`feature/<short-name>` etc.).
3. **Install the local CI hook once: `make install-hooks`.** Every
   `git push` then runs `make ci` (vet + race tests + build + every
   phase e2e) — the push is blocked until it is green.
4. `make ci` green is the contract. **There is no GitHub Actions
   pipeline** — local CI is the entire CI.
5. Conventional Commits (enforced by `commit-msg` hook). Squash-friendly.
   One concern per PR.
6. No hardcoded constants — share via `internal/platform/constants`.
7. User-facing strings via `internal/platform/i18n` (English bundle first).
8. Tests at the interface seam. Phase boundaries are real.

---

## Local setup

```bash
git clone https://github.com/ropeixoto/harnessx
cd harnessx
make install-hooks                # pre-commit + commit-msg + pre-push gates
make check                        # vet + race tests + build
make e2e                          # scripts/e2e-phase1.sh
make dashboard-install            # one-time React deps
make dashboard-test               # Vitest smoke
```

`make install-hooks` wires `scripts/git-hooks/*` into `.git/hooks/`:

| Hook         | What |
|--------------|------|
| `pre-commit` | `gofmt -l` + `go vet` against staged files (sub-3 s). |
| `commit-msg` | Conventional Commits regex. |
| `pre-push`   | `make ci` (full vet + tests + every phase e2e). Blocks the push if red. |

Bypass an emergency push: `HARNESS_SKIP_CI=1 git push`. Don't make it a
habit — `main` and `develop` branches reject unverified pushes via the
remote's branch protection, not just the hook.

### Branch protection (GitHub repo admins)

Local-only CI/CD means there are no GitHub Actions status checks the
remote can wait on. Configure the equivalent protection manually:

1. **Settings → Branches → Add classic branch protection rule**.
2. Pattern: `main` (repeat for `develop`).
3. Required checks: `Require a pull request before merging`,
   `Require approvals: 1`, `Dismiss stale approvals when new commits
   are pushed`.
4. **Restrict who can push to matching branches** → maintainers only.
5. `Require linear history` (no merge commits) so PRs land via squash.
6. **Do not** enable "Require status checks to pass" — there are none.
   Reviewers gate merges based on the PR checklist + the contributor's
   `make ci` evidence (paste output / link).

## Branching — GitFlow

We use [GitFlow](https://nvie.com/posts/a-successful-git-branching-model/)
with two long-lived branches:

| Branch    | Purpose                                       |
|-----------|-----------------------------------------------|
| `main`    | Released versions. Tagged. Protected.         |
| `develop` | Integration branch. PRs land here.            |

Short-lived branches:

| Prefix       | When                                         | Merged into          |
|--------------|----------------------------------------------|----------------------|
| `feature/*`  | New capability, sensor, agent adapter, doc.  | `develop`            |
| `fix/*`      | Bug fix that targets the next release.       | `develop`            |
| `hotfix/*`   | Production-blocking issue on a tagged release. | `main` + back-merge `develop` |
| `release/*`  | Stabilising a tagged release.                | `main` + back-merge `develop` |
| `chore/*`    | Tooling, CI, dependency bumps, infra.        | `develop`            |

Open the PR against `develop` unless the change is a `hotfix/*` or
`release/*`.

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/) only.
Examples:

```
feat(agents): add ruby-lsp adapter
fix(sensors): rewrite forbidden-commands regex without negative lookahead
docs(architecture): document hardening pass 9
chore(deps): bump charmbracelet/bubbletea
```

Subject ≤ 50 chars. Body wraps at 72.

## Code rules (non-negotiable)

**Constants** — every magic value lives in
`internal/platform/constants/constants.go`. If the same number/string
shows up in two packages, extract it.

**i18n** — every user-facing string goes through `i18n.T("key")`. Add
the English copy to `internal/platform/i18n/locales/en.json` first;
community translations land as PRs adding `locales/<lang>.json`.

**No CGO** — `modernc.org/sqlite` only. PRs that introduce CGO are
closed unless they ship a `--no-cgo` build path that keeps single-binary
distribution intact.

**No hardcoded paths** — use `constants.HarnessDir`, `constants.DBFilename`,
`paths.FindProjectRoot`, etc.

**Phase boundaries** — `HARNESSX-MASTER-PLAN.md §9` lists the phase each
package belongs to. A bug fix in Phase 6 must not silently add Phase 8
dashboard work. Call out cross-phase changes in the PR description.

**Comments** — none, unless they explain a non-obvious business rule,
security constraint, performance tradeoff, external system quirk, or
temporary workaround. Don't restate code.

**Clean Architecture** — `domain` imports nothing. `app` imports `domain`
+ interfaces. `adapters` implement those interfaces. Tests substitute
fakes at the seam.

**Tests** — at the interface seam, not behind a private function.
Prefer table-driven with one `t.Run` per case.

## Reviewing PRs

Every PR gets ≥ 1 maintainer review (see `.github/CODEOWNERS`). When you
review:

- Check the pre-merge checklist in the PR template.
- Verify the change respects phase boundaries.
- Run `make check` locally for non-trivial diffs.
- Use suggested-change blocks for nits; comments for design questions.
- Approve only when the PR could ship as-is.

## Adding a new agent adapter

1. Copy a bundled template (`harness agent discover <bin> > .harness/config/agents/<id>.yaml`).
2. Implement `Healthcheck` + run-args + `output.*` JSONPath + `failure_detection`.
3. `harness agent certify <id>` until `passed`.
4. PR adds the YAML under `internal/app/agentcmd/bundled/` so it ships
   with the binary, updates `templates/agents/`, and extends `docs/agents.md`.

## Adding a new sensor

1. Pure-Go scanner → new struct implementing `Sensor`; add to
   `sensors.Catalog`. Shell-backed → `ShellSensor` with `OptionalTool:true`
   so missing binaries skip cleanly.
2. Test the failing + passing cases.
3. Update `docs/sensors.md` with the sensor id + when it applies.

## Adding a new LSP client

1. Add a thin wrapper under `internal/adapters/lsp/<server>.go` returning
   `NewStdio(binary, args, "lang", "languageId", root)`.
2. Add the (binary, manifest-file) tuple to `context.autoClients`.
3. Test against the live binary when present, skip otherwise.

## Releases (local-only)

There is no hosted CI/CD. The release flow runs on the maintainer's
machine:

```bash
git checkout main
git pull --ff-only
git tag v0.1.0
VERSION=v0.1.0 make cd          # dashboard-install + dashboard-build + ci + release
# dist/ now holds harness-{darwin,linux}-{amd64,arm64}.tar.gz + .sha256
```

`make release` cross-builds with CGO disabled into `dist/` and writes
SHA-256 sums. Upload the tarballs + sums to whichever distribution
target you prefer (GitHub Releases UI, S3 bucket, internal mirror).

`scripts/install.sh` resolves the latest GitHub release tag and consumes
those tarballs — point `HARNESS_REPO` at a fork or mirror if you host
elsewhere.

## Community

- Bugs / features / questions → GitHub Issues (templates auto-load).
- Open-ended ideas → GitHub Discussions.
- Security → see `SECURITY.md`.
- Code of conduct → `CODE_OF_CONDUCT.md`.

Reactions on issues count as votes. Maintainers triage by reaction
count plus impact.
