# HarnessX — Manual Testing Tutorial (v0.4)

This tutorial walks through every user-facing surface of HarnessX with **real CLIs** — no fakes, no stubs. By the end you will have:

- A globally installed `harness` binary.
- A working test project under `~/dev/harness-tutorial/`.
- Verified end-to-end flow with Claude (and optionally Codex / Gemini / Kimi).
- A passing checklist covering doctor, agents, sensors, hooks, MCP, autonomy gates, runs, metrics, dashboard, and the audit bundle.

If a step depends on an agent CLI being installed and authenticated, it says so explicitly.

> **Conventions in this tutorial**
> - Lines starting with `$` are commands you run.
> - Lines starting with `>` are expected output highlights (truncated).
> - `[needs: claude]` markers tell you which agent CLI must be installed and logged in.

---

## 0. Prerequisites

| Tool | Why | Install |
|---|---|---|
| `git` | required everywhere | system |
| `bash`, `curl` | install script | system |
| `go ≥ 1.23` | build from source | https://go.dev/doc/install |
| `node ≥ 18` + `npm` | dashboard build | https://nodejs.org |
| `claude` | run real Claude agent | https://docs.claude.com/en/docs/claude-code/quickstart |
| `codex` | run real Codex agent | https://github.com/openai/codex |
| `gemini` | run real Gemini agent | https://github.com/google-gemini/gemini-cli |
| `kimi` | run real Kimi agent | https://platform.moonshot.cn |
| `pdftotext` (poppler) | `--pdf` input | `brew install poppler` / `apt install poppler-utils` |

You can complete sections 1–4, 7, 11–14 with **zero agent CLIs installed** (uses the bundled `fake-real` adapter). Sections 5, 6, 8 require at least one real CLI.

---

## 1. Install HarnessX

### Global install (recommended)

```bash
$ curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh | bash
$ harness --version
> harness vX.Y.Z (commit …, built …)
```

The script downloads the matching tarball from GitHub releases, verifies SHA-256, and installs to `${HARNESS_PREFIX:-/usr/local/bin}/harness`.

### Build from source (developer flow)

```bash
$ git clone https://github.com/rodolfopeixoto/harnessx.git
$ cd harnessx && make build
$ sudo cp bin/harness /usr/local/bin/
$ harness --version
```

### Paths to know

| Path | Purpose |
|---|---|
| `~/.harness/registry.sqlite` | cross-project workspace (project list, autopilot queue, audit) |
| `<project>/.harness/db/harness.sqlite` | per-project sessions, runs, sensors, memory |
| `<project>/.harness/runs/<run_id>/` | per-run artifacts (diff.patch, report.md, meta.json, stdout/stderr) |
| `<project>/.harness/worktrees/<run_id>/` | git worktree where the agent operates |
| `<project>/.harness/config/{autonomy,routes,agents/}.yaml` | per-project config |

Override registry root: `export HARNESS_HOME=/path/to/registry`.

---

## 2. Doctor + agent login mapping

```bash
$ harness doctor
> claude        ok    Claude Code 2.x.x
> codex         missing  binary not on PATH
> gemini        ok    Gemini CLI 0.x.x
> kimi          ok    Kimi CLI 0.x.x
> go            ok    1.24.0
> node          ok    20.x
```

For every agent CLI you want to use, log in **once** using its own auth flow. HarnessX detects auth state and tells you what to run; it does **not** wrap the login itself.

| Agent | Healthcheck | Login command | Notes |
|---|---|---|---|
| **Claude** | `claude --version` | `claude login` | Opens browser for Anthropic OAuth. After completing, `claude --print "ping"` should return text. |
| **Codex** | `codex --version` | `codex auth login` | OpenAI OAuth. Confirm with `codex chat "ping"`. |
| **Gemini** | `gemini --version` | `gemini auth login` | Google OAuth. Confirm with `gemini -p "ping"`. |
| **Kimi** | `kimi --version` | `kimi login --provider moonshot` | API key or OAuth depending on plan. |

Verify each:

```bash
$ harness agent certify claude    # [needs: claude]
> score: 71/100  ✓ healthcheck  ✓ timeout  ✓ cancellation  ✗ simple_prompt
```

If `simple_prompt` fails with `signal: killed` or `unauthorized`, run the login command from the table.

Plain output for CI / scripts:

```bash
$ harness doctor --plain
```

---

## 3. Bootstrap a test project

```bash
$ mkdir -p ~/dev/harness-tutorial && cd $_
$ git init -q
$ git config user.email "you@example.com"
$ git config user.name "you"
$ echo "# Tutorial" > README.md
$ git add -A && git commit -m "seed"

$ harness init
$ harness project add . --slug tutorial
$ harness project list
> SLUG       PATH                                STATUS
> tutorial   /Users/you/dev/harness-tutorial     active
```

`harness init` writes `.harness/` skeleton (db, default config, runs/, worktrees/).

---

## 4. Plan-only smoke (no agent required)

Prove the deterministic pipeline before involving an agent:

```bash
$ harness feature "add a /healthz endpoint that returns 200" --plan-only --agent fake-real
> Detected intent: feature (confidence 0.50)
> Spec written: .harness/artifacts/specs/<id>.yaml
> Plan written: .harness/artifacts/plans/<id>.yaml
> Report written: .harness/artifacts/reports/<id>.md

$ cat .harness/runs/$(ls -t .harness/runs/ | head -1)/report.md
```

The report includes the spec + plan + the chosen agent + an empty diff (plan-only does not invoke the agent).

---

## 5. Real Claude run end-to-end [needs: claude]

Now exercise the full agentic loop with the real Claude CLI:

```bash
$ harness feature "create HELLO.md with content: hi from harness" \
    --agent claude --apply --autonomy safe_execute --budget-usd 0.50
> Execute: run=run_… status=applied files=1 cost=$0.00…
>   diff: .harness/runs/<id>/diff.patch
>   report: .harness/runs/<id>/report.md
$ cat HELLO.md
> hi from harness
```

What happened end-to-end:

1. `harness feature` classified intent + wrote spec + plan.
2. Executor created `.harness/worktrees/<id>/` as a git worktree from HEAD.
3. Claude was invoked inside that worktree, stdout/stderr captured.
4. `git diff --cached` produced `diff.patch` + `changed-files.json`.
5. Risk classifier flagged the change as low (plain `.md`); autonomy `safe_execute` allowed apply.
6. `git apply --3way --index` merged the patch into the project root.
7. Worktree was cleaned up.

**Troubleshooting:**

- `status=agent_failed, error_type=auth` → `claude login`, retry.
- `status=no_changes` → Claude refused or misunderstood; rephrase the prompt.
- `status=agent_failed, error_type=timeout` → bump `--budget-usd` and rerun; check Claude has network.

---

## 6. Waiting-approval path (high-risk class) [needs: claude]

```bash
$ harness feature "create a Dockerfile with content: FROM scratch" \
    --agent claude --apply --autonomy safe_execute
> Execute: run=run_… status=waiting_approval files=1 cost=$0.00…
>   next: harness runs approve run_… | harness runs discard run_…

$ harness runs list
$ harness runs inspect <run-id>
$ harness runs approve <run-id>     # applies after human OK
> Applied. Worktree removed.

# OR, if you decide against it:
$ harness runs discard <run-id>
```

The risk classifier flags `Dockerfile`, `go.mod`, `package.json`, lockfiles, `.env*`, `migrations/`, `.github/workflows/`, `secrets/` and a few more as high-risk; under `safe_execute` these always route to `waiting_approval` regardless of `--apply`.

---

## 7. Sensors

```bash
$ harness sensor list
$ harness sensor run --root .
> forbidden_files       passed
> forbidden_commands    passed
> secrets_scan          passed
> changed_files         passed

$ harness runs sensors <run-id>
```

Sensor outcomes are persisted under `.harness/runs/<run-id>/sensors/`. A failing blocking sensor downgrades the run to `status=sensor_failed` and blocks apply.

---

## 8. MCP install + injection [needs: claude or codex]

```bash
$ harness mcp install filesystem --command npx --yes
> Wrote .harness/mcp/filesystem.json

$ harness mcp scan
> SOURCE   NAME         TRANSPORT  RISK  PATH
> harness  filesystem   stdio      low   .harness/mcp/filesystem.json
```

When the next run uses an adapter with `capabilities.mcp: true` (claude / codex by default), the Executor merges every discovered MCP into `<run>/mcp-config.json` and appends `--mcp-config <path>` to the agent invocation. The run report's `## MCP` section confirms it:

```
## MCP
Injected 1 MCP server(s) via .harness/runs/<id>/mcp-config.json
```

---

## 9. Hooks pre/post

Create a pre-tool-use hook:

```bash
$ mkdir -p .harness/hooks
$ cat > .harness/hooks/pre-tool-use.sh <<'SH'
#!/bin/bash
echo "pre-hook run=$HARNESS_RUN_ID agent=$HARNESS_AGENT" >&2
exit 0
SH
$ chmod +x .harness/hooks/pre-tool-use.sh
$ harness hook scan
> NAME           EVENT          STATUS   PATH
> pre-tool-use   pre-tool-use   enabled  .harness/hooks/pre-tool-use.sh
```

Run any feature and check the report:

```bash
$ harness feature "create x.md with content: y" --agent fake-real --apply
$ grep -A 4 "^## Hooks" .harness/runs/$(ls -t .harness/runs | head -1)/report.md
```

Test **blocking** behavior — failing pre-hook denies the run under `safe_execute`:

```bash
$ cat > .harness/hooks/pre-tool-use.sh <<'SH'
#!/bin/bash
exit 1
SH
$ harness feature "create blocked.md with content: x" --agent fake-real --apply --autonomy safe_execute
> Execute: run=… status=autonomy_denied
$ test -f blocked.md && echo present || echo absent
> absent
```

Bypass with `full_project_loop` (explicit, documented):

```bash
$ harness feature "create allowed.md with content: x" --agent fake-real --apply --autonomy full_project_loop
> Execute: run=… status=applied
```

---

## 10. Autonomy levels

```bash
$ harness autonomy get
> level: safe_execute

$ harness autonomy set --level safe_execute
```

| Level | Low-risk apply | High-risk apply | Hook block bypass |
|---|---|---|---|
| `manual` | deny | deny | no |
| `plan_and_ask` | approval | approval | no |
| `safe_execute` | allow | approval | no |
| `full_project_loop` | allow | approval | **yes** |
| `scheduled_maintenance` | approval | deny | no |

---

## 11. Cleanup + health + tracker

```bash
$ harness cleanup scan                 # report-only
$ harness cleanup apply --policy .harness/cleanup/policy.yaml
$ harness health show                  # per-project deterministic score

# P34 commands (after the v0.4 improvements ship):
$ harness audit --limit 20             # cross-project event timeline
$ harness metrics --since 1d --json    # tokens, cost, sensor pass rate
```

---

## 12. Dashboard

```bash
$ harness dashboard --addr :7373
$ open http://localhost:7373
```

| Route | Status |
|---|---|
| `/sessions`, `/sessions/:id`, `/runs/:id` | real (sqlite-backed) |
| `/sensors`, `/agents`, `/catalog`, `/mcps`, `/hooks` | real |
| `/design`, `/roadmap`, `/memory`, `/resources`, `/reports` | demo / static for now |

The dashboard is read-only; mutations happen via CLI.

---

## 13. Audit bundle for an LLM reviewer

```bash
$ harness stack audit
$ ls tmp/app-audit/tutorial/$(ls tmp/app-audit/tutorial | tail -1)/report/audit-bundle.zip
```

The ZIP includes HTML + PDF + JSON + screenshots and is the recommended hand-off package for an external reviewer.

---

## 14. Validation checklist

Copy this block and tick each item locally before declaring v0.4 ready in your environment.

- [ ] `harness --version` prints a version string
- [ ] `harness doctor` reports every agent CLI you intend to use as `ok`
- [ ] `harness init` + `harness project add .` registers the project
- [ ] `harness feature "..." --plan-only --agent fake-real` writes spec + plan + report
- [ ] `harness feature "..." --agent claude --apply --autonomy safe_execute` produces `status=applied` and the file appears in the project root [needs: claude]
- [ ] High-risk class (Dockerfile / go.mod / `.env*`) routes to `waiting_approval` under `safe_execute`
- [ ] `harness runs approve <id>` merges the worktree into the project root and cleans up
- [ ] `harness runs discard <id>` removes the worktree without applying
- [ ] `harness mcp install filesystem --command npx --yes` writes `.harness/mcp/filesystem.json`
- [ ] Run report `## MCP` section shows "Injected N MCP server(s)" when adapter has `mcp: true`
- [ ] Pre-tool-use hook is visible in the run report `## Hooks` table
- [ ] Failing pre-hook blocks under `safe_execute`, bypassed by `full_project_loop`
- [ ] `harness sensor run` produces a non-empty table
- [ ] `harness runs sensors <id>` lists per-run sensor outcomes
- [ ] `harness autonomy get|set` round-trips
- [ ] `harness cleanup scan` returns findings without deleting anything
- [ ] `harness dashboard --addr :7373` serves on localhost and Sessions / Sensors / Agents pages load real data
- [ ] `harness stack audit` produces `tmp/app-audit/<slug>/<ts>/report/audit-bundle.zip`

---

## Frequently-hit issues

| Symptom | Likely cause | Fix |
|---|---|---|
| `harness: command not found` | install script wrote to a dir not on PATH | `export PATH="/usr/local/bin:$PATH"` or rerun with `HARNESS_PREFIX=$HOME/.local/bin` |
| `git worktree add: not a git repository` | project not initialized | `git init -q && git add -A && git commit -m seed` |
| `status=agent_failed, error_type=auth` | agent CLI not logged in | run the login command from §2 |
| `status=no_changes` for `feature` | agent didn't produce a diff | rephrase the prompt with concrete file paths |
| `git apply` partial failure | conflicting uncommitted changes in project root | commit or stash before `harness runs approve` |
| Dashboard 404 on `/api/...` | stale built assets | `cd web/dashboard && npm ci && npm run build`, restart `harness dashboard` |

---

## Going further

- `docs/agents.md` — adapter YAML schema (write your own).
- `docs/sensors.md` — sensor authoring guide.
- `docs/architecture.md` — package map + control flow.
- `.harness/artifacts/specs/p31-*.md` and `p32-*.md` — design notes for the execution loop and MCP / hook integration.
