# HarnessX — Unified Manual Tutorial (v0.29)

End-to-end walkthrough that exercises every CLI surface against a
brand-new **FastAPI / Python** sample app. Python (not Go) is the
target on purpose: HarnessX is written in Go, so driving a Python
project proves the tool is language-independent.

Each section ends with **Expected output** so you can diff your
terminal against the doc. Every command in this tutorial is current
as of v0.29.0 — no flag mismatches, no skipped sections.

> Older tutorials (`v0.4`, `v0.11`, `v0.27`, `v0.28`) carry a redirect
> notice at the top; this file replaces them as the primary
> walkthrough.

---

## 0. Install / reinstall / uninstall

### A — Install via Homebrew (recommended, auto-upgrades)

```bash
brew tap rodolfopeixoto/harnessx https://github.com/rodolfopeixoto/harnessx.git
brew install harness                  # ← v0.29 renamed harnessx → harness
harness version                       # expect: 0.29.0
```

### B — Install via curl (no brew dependency)

```bash
curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh | bash
harness version
```

### C — Upgrade

```bash
# brew
brew update && brew upgrade harness

# curl (re-fetches latest release)
curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh | bash
```

### D — Uninstall (new in v0.29)

```bash
harness uninstall project             # wipe ./.harness/ only
harness uninstall global              # wipe cross-project registry
harness uninstall all --yes           # wipe project + global + binary
brew uninstall harness                # if installed via brew
brew untap rodolfopeixoto/harnessx
```

`uninstall all` prints `sudo rm <path>` when the binary lives in a
non-writable dir; copy-paste and run.

---

## 1. Prerequisites

| Tool | Need | Install hint |
|---|---|---|
| `harness` ≥ 0.29.0 | always | section 0 |
| `python3` ≥ 3.11 | sample app | system Python or pyenv |
| `pipx` | install ruff / pytest cleanly | `brew install pipx` |
| `git` | gitflow demo | system git |
| `claude` (Claude Code CLI) | section 5+ | `npm i -g @anthropic-ai/claude-code` |
| `tmux` | optional (claude-interactive tmux strategy) | `brew install tmux` |

Sample-app working dir:

```bash
mkdir -p ~/dev/harness-fastapi-demo && cd ~/dev/harness-fastapi-demo
git init && git checkout -b main
```

---

## 2. Initialise + register

```bash
harness init
harness project add . --slug fastapi-demo
```

**Expected:**

```
harness: initialised .../.harness
  config:  .../.harness/config/harness.yaml
  db:      .../.harness/db/harness.sqlite
  log:     .../.harness/logs/events.jsonl

registered harness-fastapi-demo
  slug: fastapi-demo
  root: .../harness-fastapi-demo
  db:   .../harness-fastapi-demo/.harness/db/harness.sqlite
```

> Re-running `project add . --slug other-name` now updates the slug
> (v0.28.1 fix). Previously the second add silently kept the
> original slug.

Confirm hook scaffold:

```bash
cat .harness/hooks/pre-tool-use.sh | head -8
```

`exit 0` stub with commented list of installable templates.

---

## 3. `harness list` — composite read-only view

```bash
harness list
```

**Expected:**

```
## projects
  SLUG          NAME                  STATUS  LAST SEEN         ROOT
* fastapi-demo  harness-fastapi-demo  active  2026-06-15 22:30  .../harness-fastapi-demo

## recent runs
  (no runs yet)

## agents
ID                 NAME                            CERT  EXP  SOURCE
claude             Claude Code                     —          bundled:claude.yaml
anthropic-api      Anthropic API                   —          bundled:anthropic-api.yaml
claude-interactive Claude Code (interactive REPL)  —     ★    bundled:claude-interactive.yaml
fake               fake                            —          bundled:fake.yaml
```

No LLM call. `★` flags experimental adapters.

---

## 4. Doctor + agents

```bash
harness doctor
harness install claude
harness agent install claude
claude login                          # one-time OAuth
harness agent certify claude --simple-timeout 180s
```

**Expected:** doctor reports your toolchain, certify shows
`✓ ready — claude is usable end-to-end.`

---

## 5. Billing primer

```bash
harness help billing
```

Three Anthropic buckets, one adapter per bucket:

| Adapter | Bucket |
|---|---|
| `--agent claude` | Agent SDK monthly credit ($20–$200/mo as of 2026-06-15) |
| `--agent claude-interactive` (experimental ★) | Pro/Max subscription |
| `--agent anthropic-api` | pay-as-you-go API |

Full details: `docs/anthropic-billing.md`.

---

## 6. Bootstrap the FastAPI app

```bash
harness feature "scaffold a FastAPI app with a /healthz endpoint that returns 200, plus pytest tests, plus a requirements.txt with fastapi + uvicorn + pytest" \
  --agent claude --apply --autonomy safe_execute --budget-usd 0.50
```

**Expected:** `Execute: ... status=applied files=N cost=$0.0xxx`.

Verify:

```bash
ls
# app.py  requirements.txt  tests/  README.md

python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
pytest -q
uvicorn app:app --reload &
curl -sf http://localhost:8000/healthz
kill %1

git add . && git commit -m "feat(api): scaffold via harness"
```

---

## 7. Iterate

```bash
harness feature "add /metrics endpoint that returns a JSON dict with requests_total counter, plus a test" \
  --agent claude --apply --autonomy safe_execute --budget-usd 0.30

harness bugfix "/healthz returns 500 when the env var ENV is missing — set a default of 'dev'" \
  --agent claude --apply --autonomy safe_execute --budget-usd 0.30

pytest -q
git add . && git commit -m "feat(api): /metrics + healthz default env"
```

---

## 8. Sensors + checks

```bash
harness sensor list
harness sensor run secrets_scan --root .        # --root accepts relative paths (v0.28)
harness check
```

---

## 9. Autonomy + budget gates

```bash
harness autonomy get

# manual mode denies every execute
harness feature "delete app.py" --agent claude --apply --autonomy manual
# expect: autonomy_denied

# budget cap demo
harness feature "rewrite the entire test suite" --agent claude --apply --autonomy safe_execute --budget-usd 0.01
# expect: budget_exceeded after the first call
```

---

## 10. Deterministic routing demo

```bash
PROMPT="explain in 50 words what /healthz does in this repo"

harness feature "$PROMPT" --agent claude               --autonomy plan_only
harness feature "$PROMPT" --agent anthropic-api        --autonomy plan_only
harness feature "$PROMPT" --agent claude-interactive   --autonomy plan_only

harness metrics --since 1d
```

`claude-interactive` rows show `mode: estimated` (REPL emits no usage
block). Anthropic console: confirm each adapter hit its declared
bucket.

---

## 11. Hooks: managed install

```bash
harness hook templates                # list bundled scripts
harness hook add pre-tool-use         # interactive selector (--yes for first match)
harness hook scan
```

Trigger a blocked run to verify the rich error:

```bash
printf '#!/bin/sh\nexit 1\n' > .harness/hooks/pre-tool-use.sh
chmod +x .harness/hooks/pre-tool-use.sh
harness feature "noop" --agent fake --apply --autonomy safe_execute
```

**Expected:**

```
Execute: pre-tool-use hook blocked: pre-tool-use blocked by .harness/hooks/pre-tool-use.sh (exit 1)
  → make the script exit 0 to allow, or remove .harness/hooks/pre-tool-use.sh
```

Restore:

```bash
printf '#!/bin/sh\nexit 0\n' > .harness/hooks/pre-tool-use.sh
```

---

## 12. MCP

```bash
harness mcp templates
harness mcp install filesystem --command npx --yes
harness mcp scan
```

---

## 13. Reports + runs

```bash
harness runs list
harness runs inspect $(ls -t .harness/runs | head -1)
cat .harness/runs/$(ls -t .harness/runs | head -1)/report.md
harness report                        # prints last run report
```

v0.28+ ships **one** report per run at `.harness/runs/<id>/report.md`.
`.harness/artifacts/reports/*` is reserved for user-triggered artefact
reports (`harness security-audit`, etc.).

---

## 14. Dashboard + logs

```bash
harness dashboard --addr :7373 &
open http://localhost:7373
harness logs --follow
kill %1
```

React UI loads from `web/dashboard/dist` when present (`make
dashboard-build`); else built-in HTML.

---

## 15. Backup + memory

```bash
harness backup config show
harness memory show
```

No remote configured ⇒ prints a four-line fix recipe.

---

## 16. Cleanup primer

```bash
harness cleanup scan
harness cleanup apply --dry-run
```

v0.29 ships workspace-level cleanup only. Project prune + runs prune
(`harness project prune --older-than 30d`, `harness runs prune
--keep-last 20`) land in **v0.30 (P63)**.

---

## 17. Gitflow demo

```bash
git checkout -b feature/health-detail
harness feature "expand /healthz to return uptime_seconds and git_commit" \
  --agent claude --apply --autonomy safe_execute --budget-usd 0.30
pytest -q
git add . && git commit -m "feat(api): healthz detail"
git checkout main && git merge --no-ff feature/health-detail
```

---

## 18. Uninstall (new in v0.29)

When done with the tutorial:

```bash
# project only
harness uninstall project --yes        # wipes ./.harness/

# global registry too (apaga lista de projetos)
harness uninstall global --yes

# nuke everything (project + global + binary)
harness uninstall all --yes
brew uninstall harness                 # if brew install was used
brew untap rodolfopeixoto/harnessx
```

Verify:

```bash
which harness                          # nada
ls ~/Library/Application\ Support/harness 2>/dev/null   # macOS: nada
ls ~/.local/share/harness 2>/dev/null                   # Linux: nada
```

---

## 19. Final state

```bash
harness list                           # before uninstall
```

Project still shows up, runs table contains every run from sections
6–17, agents still listed. End-of-tutorial sanity check.

---

## What you proved

- HarnessX is language-agnostic: full lifecycle on a FastAPI app
  without Go-specific assumptions.
- Two install paths (brew + curl) + one-shot `harness uninstall all`.
- `harness list` gives a one-shot read-only summary — no LLM call.
- `--budget-usd` works on every workflow command.
- `harness sensor run --root .` accepts relative paths.
- Every run writes **one** canonical report.
- `harness init` scaffolds `.harness/hooks/pre-tool-use.sh`.
- `harness hook add <event>` installs a bundled template
  interactively.
- Hook block messages name the script and the fix.
- Three Anthropic billing buckets reachable.
- `harness uninstall project|global|all` removes every trace.

## Next

- `docs/anthropic-billing.md` — per-plan credit table + pricing
  cross-check.
- `harness help <topic>` — in-CLI tutorials per surface.
- v0.30 (P63): project + runs prune.
- v0.31 (P64): markdown-driven runs + real-time streaming attach.
