# Tutorial — Build TaskHive, a full-stack SaaS, with harnessx

> Aurora, senior technical writer.
> Companion tutorial to `docs/GETTING-STARTED.md`, `docs/WORKFLOW.md`,
> `docs/PAPER-IMPLEMENTATION.md`, and `docs/spec-driven-development.md`.

---

## Prologue

This tutorial is the long one. By the end you will have shipped
**TaskHive**, a small but honest task-management SaaS — Go backend, React
frontend, JWT auth, project/task CRUD with permissions, Docker Compose
for local dev, Playwright E2E, and a cost report you can defend to
your finance team. Along the way you will exercise practically every
subsystem in harnessx: spec generation, planning, multi-agent routing,
context engineering (cache-aware layout, salience reorder, BM25 rerank,
LLMLingua compression), sensors, dashboard, orchestration, and the
`ship` gate.

**Who this is for.** A professional developer who has read
`docs/GETTING-STARTED.md` at least once and now wants to see the whole
thing exercised in anger, not in a hello-world.

**Prerequisites.**

- Go 1.25 or newer (`go version`).
- Node 20 or newer (`node -v`).
- `harness` v0.170.1 or newer on `PATH` (`harness --version`).
- Git configured with your user.
- One agent API key on your shell (`ANTHROPIC_API_KEY` is the assumed
  default; the router will fall back to Codex, Kimi or the `fake`
  adapter as configured).

**Estimated time.** 4-6 hours of wall clock, most of it spent reading
diffs and running tests, not typing. The agent calls themselves are
seconds each.

**Estimated agent cost.** USD 2-5 end-to-end on a Sonnet+Haiku+Codex
mix. This tutorial breaks that down chapter by chapter and shows you
how to verify with `harness cost report --breakdown`.

**Why TaskHive.** It is small enough to finish but large enough to
force the interesting decisions: authentication, ownership-scoped
CRUD, joins across two tables, an SPA that talks to that API, and E2E
that spans both. If harnessx cannot save you tokens on this shape of
work it cannot save you tokens on anything.

**What you will have at the end.** A working repo, a green
`harness ci`, a signed release tag, and — more importantly — a mental
model of when to reach for which agent, when to skip the LLM entirely,
and how to read a cost report as a first-class artefact.

---

## Chapter 1 — Setup and the first sensor pass

Create the project directory and initialise the vault.

```bash
mkdir -p ~/dev/taskhive && cd ~/dev/taskhive
git init -q
harness init
```

`harness init` writes `.harness/config/harness.yaml`, seeds the empty
sqlite state under `.harness/db/`, prints the paths it created, and
installs the local pre-push hook. Nothing hits the network yet.

Now build the JSON index maps that back the context engine:

```bash
harness project index
```

The maps live under `.harness/project/`. Each one exists so that
`internal/context.Build` (`internal/context/build.go`) can answer a
different question without re-walking the tree:

| Map | Answers |
|---|---|
| `profile.json` | detected stacks + toolchain versions |
| `commands.json` | canonical lint / test / build commands per stack |
| `dependencies.json` | direct + transitive dependency graph |
| `architecture.json` | package layout + import edges |
| `test-map.json` | file-to-test mapping for the ci gate |
| `api-map.json` | HTTP handlers, CLI entry points, RPC surfaces |
| `design-system.json` | UI tokens, components, screens (if a frontend) |
| `performance-budget.json` | perf snapshot budget the sensor gate checks |

Verify your toolchain:

```bash
harness doctor
```

`doctor` checks Go, Node, git, the writable state directory, and each
adapter's binary. Fix anything red before continuing.

List agents:

```bash
harness agent list
```

You should see the bundled adapter set: `claude`, `claude-interactive`,
`codex`, `kimi`, `antigravity` (formerly Gemini, see commit `06e0534`),
`ollama`, `fake`, `fake-real`, plus the direct-API bindings
(`anthropic-api`, `openai-api`, `gemini-api`, `moonshot-api`,
`minimax-api`). Only `claude`, `codex`, `gemini`, and `fake-real` are
accepted as `--agent` values on the `feature` / `run` / `bugfix`
commands today; run `harness agent list` to see the full inventory.
Pin a default:

```bash
harness use claude
```

(Use `harness use fake` if you want to do this tutorial fully offline;
you will see the workflow but not the LLM output.)

`harness sensor list` will show the default sensor set (secret scan,
forbidden files, license drift). None have fired yet — the repo is
empty.

Finally, drop a router file so the multi-agent router in
`internal/orchestrate/router.go` has something to obey:

```yaml
# .harness/config/routes.yaml
budget_usd: 5.00
default: claude
routes:
  - match: { intent: spec }
    agent: claude
    fallback: [codex, kimi]
  - match: { intent: implement }
    agent: codex
    fallback: [claude]
  - match: { intent: test }
    agent: kimi
    fallback: [codex]
  - match: { intent: docs }
    agent: claude-haiku
    fallback: [kimi]
```

`budget_usd` is a hard cap: once cumulative cost for the session
crosses it, every subsequent plan is refused with `budget exceeded`
(see `internal/costtrack/`). Fallback chains are tried in order on any
adapter-level failure, not on quality — quality gating happens in the
critic, later.

**Cost so far:** USD 0.00. Nothing has called an LLM.

---

## Chapter 2 — Spec-driven backend init

We do not open an editor. We write a spec.

```bash
harness feature "Go HTTP API scaffold with Chi router, \
modernc.org/sqlite persistence, JWT HS256 auth, structured slog" \
  --agent claude --yes
```

The generated spec lands at `.harness/artifacts/specs/<id>.md`. Open
it. You will find: user story, acceptance criteria, non-goals,
proposed file layout, test seams, and a small ADR-style block on
sqlite versus Postgres.

Why spec-first? Because the plan you review in a moment is derived
from the spec, and the diff you apply is derived from the plan. If the
spec is wrong the rest is wrong quietly. Two minutes of reading here
saves an hour of re-rolling later.

Inspect the plan by opening the file `harness feature` printed above
(look for `Plan written:` in the run log). The plan file lists:
files it will create, files it will modify (none, this is greenfield),
approximate token budget, chosen adapter, and the fallback chain that
would fire on failure. If anything looks wrong, delete the run
artefacts under `.harness/runs/` and re-invoke `harness feature`.

Execute directly from the feature command (`harness feature --agent
<id> --yes` streams the patch and applies it after the sensor gate).
For a spec-then-execute flow without applying, drop `--yes` and the
CLI will pause after writing the plan. `harness run <prompt>` is the
lower-level equivalent that skips spec enrichment.

Now the sanity gate:

```bash
harness check
```

`check` runs `go vet`, `go test ./... -race`, and `staticcheck`. First
pass should be green. If not, `harness bugfix` is the right next call,
not manual editing (Chapter 14).

**Cost delta:** ~USD 0.15. Sonnet writes the spec, Sonnet writes the
plan, cached prompt-prefix on the second call keeps the second bill
low.

---

## Chapter 3 — Feature: authentication (JWT)

```bash
harness feature "add POST /auth/register and POST /auth/login \
using bcrypt cost 12 and JWT HS256 with a 24h expiry; \
refuse duplicate emails; return 400 on weak passwords" \
  --agent claude --yes
```

Why Claude here rather than Codex? Auth code has a long tail of bad
defaults (constant-time compare, error-message leakage, JWT `alg=none`
tricks). Sonnet's reasoning trace catches those more often than
Codex's implementation-first style. This is a judgement call, not a
law — see the rubric in Appendix A.

Read the spec, plan, run. Watch the context builder log line:

```
context: pack=18 files reordered by salience (router.go +3, main.go +2)
saved: ~210 tokens vs. lexicographic order
```

That is the salience reorder from `internal/context/salience.go`
paying rent. Files touched by the current spec bubble to the top of
the pack so the LLM's attention lands on them, and — because
Anthropic's cache is prefix-sensitive — the *bottom* of the pack stays
byte-identical between calls, keeping the cache hit alive.

Verify:

```bash
harness check
go test ./internal/auth -run TestJWT -v
```

**Cost delta:** ~USD 0.20.

---

## Chapter 4 — Feature: projects CRUD

```bash
harness feature "add /projects CRUD (POST, GET list, GET one, \
PATCH, DELETE) scoped by the authenticated user; sqlite schema \
with created_at, updated_at; 404 on cross-user access" \
  --agent codex --yes
```

Why Codex? This is implementation-heavy, security-shallow (the auth
middleware from Chapter 3 already gates it), and the shape is
familiar to the model. Codex is roughly a third the cost per output
token of Sonnet and lands the diff cleanly on this kind of shape.

If Codex fails or times out, the fallback chain from `routes.yaml`
sends the same plan to Claude. You will see this in the run log:

```
adapter=codex status=timeout after=42s
fallback -> claude
```

Check spend so far:

```bash
harness cost report --since 1h
```

The table is agent × intent × USD, sourced from
`internal/costtrack/store.go`. Numbers are computed from adapter-side
token counts, not estimated.

**Cost delta:** ~USD 0.08.

---

## Chapter 5 — Feature: tasks CRUD with permissions

```bash
harness feature "add /tasks CRUD nested under /projects/{id}/tasks; \
task has title, description, status enum (todo|doing|done), \
due_at nullable; enforce project ownership via existing middleware" \
  --agent codex --yes
```

Watch the context log this time:

```
bm25 rerank: promoted internal/auth/project_authorization.go (+0.71)
skipped duplicate middleware generation
```

BM25 rerank (`internal/context/bm25.go`) found the existing project
authorization helper by keyword-matching the spec against the symbol
index, promoted it into the pack, and Codex read it rather than
reimplement it. This is the single most common way harnessx saves
tokens on brownfield work.

**Cost delta:** ~USD 0.10.

---

## Chapter 6 — Frontend scaffold

Scaffold, do not "feature":

```bash
cd ~/dev
harness new react taskhive-web --yes
```

`harness new` takes a positional `<stack> <path>` pair; bundled stacks
are `go`, `python`, `rails`, `react`, `ruby`, and `rust`. Under the
hood it runs `git init` + `harness init` + `harness scaffold apply
<stack>` + `harness install-git-hooks` for you.

Scaffolding is deterministic — a template under `templates/` is copied
and rendered. No LLM is involved. Use scaffold for anything that has a
single obviously-right answer; use `feature` when the shape is
ambiguous.

**Cost delta:** USD 0.00.

---

## Chapter 7 — Frontend feature: auth flow

```bash
cd taskhive-web
harness feature "login form with email+password, \
protected route wrapper reading token from localStorage, \
axios interceptor that refreshes on 401 by redirecting to /login" \
  --agent claude --yes
```

Note the context log:

```
llmlingua: compressed pack 42.3KB -> 25.1KB (-40.6%) at quality=0.92
```

The LLMLingua-2 compressor at `internal/context/compress.go` drops
low-information tokens (repeated import blocks, dead comments) at a
quality threshold you set in `active.yaml`. On React code, which is
verbose, the win is usually 35-45%.

Read the generated code. There should be a `useAuth` hook, a
`RequireAuth` wrapper, and an axios instance in `src/lib/api.ts`.

**Cost delta:** ~USD 0.18.

---

## Chapter 8 — Frontend feature: project and task views

```bash
harness feature "project list page, project detail page showing \
its tasks grouped by status, modal for create/edit project, \
modal for create/edit task, optimistic updates on status change" \
  --agent codex --yes
```

**Cost delta:** ~USD 0.12.

---

## Chapter 9 — Tests

Three calls, three different agents, three different reasons.

Backend unit tests:

```bash
cd ~/dev/taskhive
harness feature "add table-driven tests for every HTTP handler \
using net/http/httptest; cover happy path, auth failure, \
cross-user 404, validation 400" \
  --agent codex --yes
```

Frontend unit tests:

```bash
cd ~/dev/taskhive-web
harness feature "add Vitest tests for useAuth hook, \
project list rendering, and the axios 401 interceptor" \
  --agent kimi --yes
```

Kimi is cheap and quick on well-scoped test generation. It is not the
model to reach for on architecture, but for "given this file, write
its tests" it is fine.

E2E:

```bash
cd ~/dev
harness feature "add Playwright test at ./e2e: \
register -> login -> create project -> add three tasks -> \
mark one done -> assert count on dashboard" \
  --agent claude --yes
```

Playwright specs benefit from Sonnet's stronger reasoning about async
timing and selector stability.

**Cost delta:** ~USD 0.25.

---

## Chapter 10 — Docker Compose and CI

```bash
harness feature "multi-stage Dockerfile for the Go backend \
(builder + distroless final), nginx-alpine for the built frontend, \
docker-compose.yml with mem_limit 512m on backend + 128m on nginx, \
healthchecks on both, shared network, sqlite volume" \
  --agent claude --yes
```

The generated `docker-compose.yml` will include memory limits — this
tutorial is written under a project CLAUDE.md that requires them
(see `/Users/ropeixoto/dev/projects/harnessx/CLAUDE.md` §5). If yours
comes back without them, add the limits by hand before running.

Bring it up carefully, per the sequence documented in that same
CLAUDE.md:

```bash
docker compose build --no-cache
docker compose up -d
timeout 60s bash -c \
  'until curl -sf http://localhost:8080/healthz > /dev/null; \
   do sleep 2; done' && echo ok
```

Run the E2E against it, then always:

```bash
docker compose down --remove-orphans
```

**Cost delta:** ~USD 0.15.

---

## Chapter 11 — Dashboard and observability

```bash
harness dashboard --addr 127.0.0.1:0
```

The `:0` idiom asks the kernel for an ephemeral port and prints it —
important because a fixed `--addr 127.0.0.1:8090` collides with
whatever else you have running. This was one of the App1/App2 pilot
caveats and the reason `:0` is now the recommended form.

Endpoints (see `cmd/harness/cmd_dashboard.go` and
`internal/dashboardapi/`):

- `/api/health` — process liveness.
- `/api/runs` — every `harness run` invocation with status and cost.
- `/api/agents` — adapter health, last-seen, error rate.
- `/api/sensors` — sensor findings across the last N runs.
- `/api/cost/report` — the same table as the CLI, JSON-shaped.

The React SPA under `web/dashboard/` renders those. Open the URL
printed by the command.

Take a perf baseline while you are here:

```bash
harness perf-snapshot --label chapter-11-baseline
```

**Cost delta:** USD 0.00.

---

## Chapter 12 — Cost analysis

Sum the chapter deltas: 0.15 + 0.20 + 0.08 + 0.10 + 0.18 + 0.12 + 0.25
+ 0.15 = **~USD 1.23** on a Sonnet+Codex+Kimi mix. Add the
dashboard/scaffold zeros and you land in the USD 1.20-1.50 band.

Compare to a same-shape build driven by raw Sonnet with no context
engineering:

```bash
harness benchmark cost taskhive-replay --agent fake --iterations 3
```

Only the `fake` adapter is currently accepted by `benchmark cost`; the
raw-vs-harnessx comparison uses deterministic token counts so results
are reproducible in CI.

The replay uses the same prompts but sends the *full* file tree on
every call, no cache-aware layout, no salience, no BM25, no LLMLingua.
Expected: USD 3.00-5.00, depending on how much the API-side prompt
cache happens to reuse across calls.

Break down the wins:

```bash
harness cost report --breakdown
```

Sample output (numbers are illustrative, order and categories are
real):

```
cache-aware layout      $0.42 saved   (34%)
salience reorder        $0.18 saved   (15%)
bm25 rerank             $0.29 saved   (24%)
llmlingua compression   $0.33 saved   (27%)
```

**When harness earns its keep.** Large repos, iterative work where the
same prefix is sent many times, and multi-agent workflows where the
cheap agent takes the boring diffs. **When it does not.** Green-field
CRUD in a tiny repo where every prompt is short and the API cache
hits anyway. That is fine — measure, do not assume.

---

## Chapter 13 — Multi-agent orchestration

Write a tiny flow:

```yaml
# .harness/orchestrations/deploy-strategy-debate.yaml
name: deploy-strategy-debate
topology: chain
steps:
  - role: planner
    adapter: claude
    command: [claude]
    prompt: "Propose a deploy strategy for TaskHive: \
             fly.io single region vs. render.com vs. self-hosted docker."
  - role: reviewer
    adapter: kimi
    command: [kimi]
    prompt: "List the two weakest points and one blocker \
             for a two-person team."
  - role: manager
    adapter: claude
    command: [claude]
    prompt: "Pick one option and write the ADR."
```

Roles must be one of `manager`, `planner`, `coder`, `reviewer`,
`tester`. Topology is either `chain` (linear) or `cyclic` (requires
`max_cycles`). Run by name (the file lives under
`.harness/orchestrations/`):

```bash
harness orchestrate run deploy-strategy-debate
```

Artefacts land in `.harness/artifacts/orchestrate/<run-id>/`. The ADR
Claude writes at the end is the one you commit into `docs/adr/`.

**Cost delta:** ~USD 0.05.

---

## Chapter 14 — Iterating: bugfix and evolve

Simulate a regression: edit `web/src/pages/ProjectDetail.tsx` and
delete the `queryClient.invalidateQueries(['tasks', projectId])` call
after task edit. The list will now go stale after a status change.

Fix it the harness way:

```bash
harness bugfix "task list on project detail page goes stale \
after editing a task status; suspect missing invalidation" \
  --agent codex --yes
```

`bugfix` differs from `feature` in two ways: it does not write a
spec, and it constrains the plan to a single logical fix (see
`internal/intent/bugfix.go`). Codex is a good pick here because the
fix shape is narrow.

Compare with `evolve`:

```bash
harness evolve diagnose
harness evolve propose --note "add regression coverage for stale-list bug"
```

`evolve` is subcommand-based (`diagnose`, `propose`, `promote`,
`replay`, `sandbox`) and drives the harness's own self-improvement
loop from `.harness/logs/events.jsonl`. Run `diagnose` periodically,
not per bug — it clusters recent failures and surfaces the ones worth
turning into sensors or new remedies.

**Cost delta:** ~USD 0.08.

---

## Chapter 15 — Ship

The pre-push gate:

```bash
harness ci
```

This runs `make ci` (vet + race + build + every phase e2e), matching
what the `pre-push` hook installed by `make install-hooks` would run.
Never push without it green.

Ship:

```bash
harness ship
```

`ship` orchestrates the multi-branch dance for you: it verifies you
branched from `develop`, rebases, opens (or updates) the PR, and — on
merge — tags. See `cmd/harness/cmd_ship.go` and `docs/WORKFLOW.md` for
the full sequence.

**Cost delta:** USD 0.00.

---

## Appendix A — Agent choice rubric

| Task type | First choice | Why | Fallback |
|---|---|---|---|
| Planning / spec writing | Claude Sonnet | reasoning depth on trade-offs | Codex |
| Implementation, large diff | Codex | cheap per output token, focused | Claude Sonnet |
| Test generation | Kimi or Codex | cheap, patterned work | Claude Haiku |
| Security-sensitive code | Claude Opus | strongest defaults, catches JWT/crypto footguns | Claude Sonnet |
| Docs / README / ADR | Claude Haiku | cheap, prose is easy | Kimi |
| Mechanical refactor | Codex | pattern matching over reasoning | Claude Sonnet |
| Debug / root-cause | Claude Sonnet | trace reasoning | Codex |
| Explain existing code | Kimi | cheap, summarisation is easy | Claude Haiku |

Rule of thumb: if a mistake would leak data, wake someone up, or
compound silently, spend the Sonnet/Opus token. Otherwise take the
Codex/Kimi discount.

---

## Appendix B — When to skip agents

You do not need an LLM for:

- Renames, formatter runs, import re-sorts — use `harness scaffold` or
  `gofmt`/`prettier`.
- Refactors with a full test suite and a sensor rule already written
  (`internal/customrules/`) — the rule engine will do it.
- One-line fixes where you already know the line. Just edit it.
- Any change where writing the prompt takes longer than writing the
  code.

Agents cost tokens and add non-determinism. Both are fine when you
buy something with them; neither is fine when you do not.

---

## Appendix C — Debugging harnessx itself

- Cold Go build cache: `go clean -cache && harness check`.
- A sensor keeps firing spuriously: `harness sensor run <id>
  --verbose` and inspect its finding record.
- An adapter refuses to run: `harness agent certify <id>` re-runs the
  handshake and prints the exact stderr from the adapter binary.
- Context looks stale after a big rebase: `harness context build "..."
  --force` invalidates `hash.json` and rebuilds all eight index maps.
- Dashboard port collision: always use `--addr 127.0.0.1:0`.
- `--version` prints nothing: you are on an old binary; check
  `which harness` and re-install.

---

## Closing

You now have TaskHive — spec, code, tests, containers, dashboard, ADR,
release tag — for less than the price of a coffee in agent spend, and
you have exercised every layer of harnessx doing it. The confidence
part matters more than the code. You know when to reach for Sonnet
versus Codex, when to write a spec versus a bugfix, when to compress
context versus rerank it, and when to skip the LLM entirely.

**Next steps.** Deploy TaskHive somewhere real. Add users. Watch the
cost report over a week of iteration and compare it to what you would
have paid on raw Sonnet. Then contribute the improvement back:

```bash
harness feature "add the thing you wished harness did" --yes
```

That is the whole loop.

— Aurora
