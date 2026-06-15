# HarnessX — Unified Manual Tutorial (v0.28)

End-to-end walkthrough that exercises every CLI surface against a
brand-new **FastAPI / Python** sample app. Python (not Go) is the
target on purpose: HarnessX is written in Go, so driving a Python
project proves the tool is language-independent.

Each section ends with **Expected output** so you can diff your
terminal against the doc. Every command in this tutorial is current
as of v0.28.0 — no flag mismatches, no skipped sections.

> Older tutorials in `docs/tutorial-v0.4-manual.md`,
> `docs/tutorial-v0.11-manual.md`, `docs/tutorial-v0.27-manual.md`
> still apply for historical commands; this file replaces them as the
> primary walkthrough.

---

## 0. Prerequisites

| Tool | Need | Install hint |
|---|---|---|
| `harness` ≥ 0.28.0 | always | `brew install harnessx` or `harness update` |
| `python3` ≥ 3.11 | sample app | system Python or pyenv |
| `pipx` | install ruff / pytest cleanly | `brew install pipx` |
| `git` | gitflow demo | system git |
| `claude` (Claude Code CLI) | section 4+ | `npm i -g @anthropic-ai/claude-code` |
| `tmux` | optional (claude-interactive tmux strategy) | `brew install tmux` |

```bash
harness --version
# expect: 0.28.0
```

Sample-app working dir:

```bash
mkdir -p ~/dev/harness-fastapi-demo && cd ~/dev/harness-fastapi-demo
git init && git checkout -b main
```

---

## 1. Initialise + register

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
```

Confirm the new auto-scaffolded hook:

```bash
ls .harness/hooks/
cat .harness/hooks/pre-tool-use.sh | head -10
```

The script is a permissive `exit 0` stub with comments listing the
installable bundled templates. Empty directory ⇒ no phantom blocks.

---

## 2. `harness list` — composite read-only view

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
ID           NAME              CERT  EXP  SOURCE
claude       Claude Code       —          bundled:claude.yaml
anthropic-api Anthropic API    —          bundled:anthropic-api.yaml
claude-interactive Claude Code (interactive REPL)  —  ★  bundled:claude-interactive.yaml
fake-real    fake-real         —          bundled:fake-real.yaml
```

Single read-only call; no LLM invoked. The `★` flags experimental
adapters.

---

## 3. Doctor + agents

```bash
harness doctor
harness install claude
harness agent install claude
claude login                                # one-time OAuth
harness agent certify claude --simple-timeout 180s
```

**Expected:** doctor reports your toolchain, certify shows
`✓ ready — claude is usable end-to-end.`

---

## 4. Billing primer

```bash
harness help billing
```

You decide which Anthropic bucket each run hits:

| Adapter | Bucket |
|---|---|
| `--agent claude` | Agent SDK monthly credit ($20–$200/mo as of 2026-06-15) |
| `--agent claude-interactive` (experimental ★) | Pro/Max subscription |
| `--agent anthropic-api` | pay-as-you-go API |

Full details: `docs/anthropic-billing.md`.

---

## 5. Bootstrap the FastAPI app

```bash
harness feature "scaffold a FastAPI app with a /healthz endpoint that returns 200, plus pytest tests, plus a requirements.txt with fastapi + uvicorn + pytest" \
  --agent claude --apply --autonomy safe_execute --budget-usd 0.50
```

**Expected:** `Execute: ... status=applied files=N cost=$0.0xxx`.
Verify:

```bash
ls
# app.py  requirements.txt  tests/  README.md
```

Install + run the test suite:

```bash
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
pytest -q
uvicorn app:app --reload &
curl -sf http://localhost:8000/healthz
kill %1
```

Commit baseline:

```bash
git add . && git commit -m "feat(api): scaffold via harness"
```

---

## 6. Iterate

```bash
harness feature "add /metrics endpoint that returns a JSON dict with requests_total counter, plus a test" \
  --agent claude --apply --autonomy safe_execute --budget-usd 0.30

harness bugfix "/healthz returns 500 when the env var ENV is missing — set a default of 'dev'" \
  --agent claude --apply --autonomy safe_execute --budget-usd 0.30

pytest -q
git add . && git commit -m "feat(api): /metrics + healthz default env"
```

---

## 7. Sensors + checks

```bash
harness sensor list
harness sensor run secrets_scan --root .
harness check
```

The new `--root` flag pins the project root for `sensor run`. `check`
runs every applicable sensor.

---

## 8. Autonomy + budget gates

```bash
harness autonomy get

# manual-mode demo: every execute denied
harness feature "delete app.py" --agent claude --apply --autonomy manual
# expect: autonomy_denied

# budget cap demo
harness feature "rewrite the entire test suite" --agent claude --apply --autonomy safe_execute --budget-usd 0.01
# expect: budget_exceeded after first call
```

---

## 9. Deterministic routing demo

Same prompt, three adapters. Compare cost + which billing bucket
moves.

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

## 10. Hooks: managed install

```bash
harness hook templates                 # list all bundled scripts
harness hook add pre-tool-use          # interactive selector (--yes for first match)
harness hook scan
```

**Expected:** picker lists `pre-tool-use-lint`, `pre-tool-use-secrets`,
`pre-tool-use-noforce`; chosen template lands at
`.harness/hooks/pre-tool-use.sh` with `chmod 755`.

Trigger a blocked run to verify the new error message:

```bash
echo 'exit 1' > .harness/hooks/pre-tool-use.sh
chmod +x .harness/hooks/pre-tool-use.sh
harness feature "noop" --agent fake-real --apply --autonomy safe_execute
```

**Expected:**

```
Execute: pre-tool-use blocked by .harness/hooks/pre-tool-use.sh (exit 1)
  → make the script exit 0 to allow, or remove .harness/hooks/pre-tool-use.sh
```

Restore the permissive stub:

```bash
echo 'exit 0' > .harness/hooks/pre-tool-use.sh
```

---

## 11. MCP

```bash
harness mcp templates
harness mcp install filesystem --command npx --yes
harness mcp scan
```

When an adapter declares `capabilities.mcp=true`, the executor merges
discovered MCPs into `runs/<id>/mcp-config.json` and appends
`--mcp-config <path>` to the agent invocation.

---

## 12. Reports + runs

```bash
harness runs list
harness runs inspect $(ls -t .harness/runs | head -1)
cat .harness/runs/$(ls -t .harness/runs | head -1)/report.md
```

v0.28 ships **one** report per run at `.harness/runs/<id>/report.md`.
The old duplicate at `.harness/artifacts/reports/*` is gone (only
user-triggered artefact reports — `harness security-audit`, etc. —
live there now).

```bash
harness report
```

Prints the most recent run report (now read from
`.harness/runs/<id>/report.md`).

---

## 13. Dashboard + logs

```bash
harness dashboard --addr :7373 &
open http://localhost:7373
harness logs --follow
kill %1
```

The dashboard surfaces the Home/Projects/Run/Agents/etc. pages backed
by the same files this tutorial walks through.

---

## 14. Backup + memory

```bash
harness backup config show
harness memory show
```

No remote configured ⇒ commands print a four-line fix recipe (set
default remote, etc.) instead of failing silently.

---

## 15. Cleanup primer

```bash
harness cleanup scan
harness cleanup apply --dry-run
```

v0.28 only ships workspace-level cleanup (worktrees, caches). Project
prune + runs prune (`harness project prune --older-than 30d`, `harness
runs prune --keep-last 20`) land in **v0.29 (P63)**.

---

## 16. Gitflow demo

```bash
git checkout -b feature/health-detail
harness feature "expand /healthz to return uptime_seconds and git_commit" \
  --agent claude --apply --autonomy safe_execute --budget-usd 0.30
pytest -q
git add . && git commit -m "feat(api): healthz detail"
git checkout main && git merge --no-ff feature/health-detail
```

---

## 17. Final state

```bash
harness list
```

Expected: project still shows up, runs table now contains every run
from sections 5–16, agents still listed. End-of-tutorial sanity check.

---

## What you proved

- HarnessX is language-agnostic: full lifecycle on a FastAPI app
  without any Go-specific assumptions.
- `harness list` gives a one-shot read-only summary — no LLM call.
- `--budget-usd` works on every workflow command (`--budget` still
  works as a hidden deprecated alias).
- `harness sensor run --root .` accepts the flag the docs always
  recommended.
- Every run writes **one** canonical report under
  `.harness/runs/<id>/report.md`.
- `harness init` scaffolds `.harness/hooks/pre-tool-use.sh` so an
  empty directory never blocks a run.
- `harness hook add <event>` installs a bundled template
  interactively.
- Hook block messages now name the offending script and the fix.
- Three Anthropic billing buckets reachable, one adapter per bucket,
  experimental status visible in `harness list` and `harness agent
  certify`.

## Next

- `docs/anthropic-billing.md` for the per-plan credit table + pricing
  cross-check.
- `harness help <topic>` for in-CLI tutorials per surface
  (`agents`, `sensors`, `hooks`, `autonomy`, `mcp`, `update`,
  `billing`).
- v0.29 (P63) will add project + runs prune; v0.30 (P64) will add
  markdown-driven runs and real-time streaming attach/detach.
