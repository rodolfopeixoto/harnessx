> **Legado.** Substituído por [docs/TUTORIAL.md](../TUTORIAL.md). Mantido aqui só para referência histórica.

# Tutorial — Build the same Todoist across Python, Rails, Go, Rust, Ruby + React

A single hands-on walk that ships the **same Todoist domain** through
every backend stack HarnessX scaffolds today, then wires a React
frontend that talks to all of them via OpenAPI-style endpoints. The
goal: prove the loop (scaffold → `/drive` → spec → failing tests →
impl → gate → commit) works identically regardless of stack, and
give you muscle memory for the commands.

You will end with **six repos** under `~/dev/todoist-multi/`:

| dir              | scaffold            | runtime           | gate                   |
|------------------|---------------------|-------------------|------------------------|
| `api-python`     | `python-ecommerce`  | FastAPI :8000     | ruff + pytest + bandit |
| `api-rails`      | `rails`             | Rails 7 API :3000 | rubocop + rspec        |
| `api-go`         | `go`                | net/http :8001    | go vet + go test       |
| `api-rust`       | `rust`              | Axum :8002        | cargo clippy + test    |
| `api-ruby`       | `ruby` (Sinatra)    | Sinatra :4567     | rubocop + rspec        |
| `web-react`      | `react`             | Vite :5173        | vitest + eslint        |

Estimated time: 4–6 h with `--with-deps`. Requires macOS or Linux,
Python 3.11+, Ruby 3.2+, Go 1.22+, Rust 1.78+, Node 20+, git, and one
of `claude` / `codex` / `kimi` configured.

> Tip: run `harness onboarding` first. It prints what is missing
> with the exact install command per tool.

---

## Phase 0 · Prep

```bash
brew install rodolfopeixoto/harnessx/harness   # or scripts/install.sh
harness --version                              # v0.143.0 or newer
harness onboarding                             # → green for every row
harness use claude                             # pin once for the whole walk
mkdir -p ~/dev/todoist-multi && cd ~/dev/todoist-multi
```

A shared backlog for every backend. Save it once and reuse:

```bash
cat > /tmp/todoist-features.md <<'EOF'
# Todoist backlog — same prompts ship through every stack

- add JWT auth POST /auth/register and POST /auth/login with bcrypt
- add Lists CRUD scoped per authenticated user
- add Tasks under a List with title, due_at, priority, done
- add Tag CRUD plus many-to-many task↔tag link
- add recurring rule daily | weekly | weekday on Task and decrement on done
EOF
```

`harness drive --features /tmp/todoist-features.md --continue-on-fail`
chews the whole backlog in one shot per backend.

---

## Phase 1 · Python (FastAPI) backend

```bash
harness new python-ecommerce ./api-python --yes --with-deps
cd api-python
harness ci --install-missing               # bandit/mypy/pip-audit into .venv
harness chat --auto-gate
```

Inside the chat:

```text
[dev|claude ✓]> /drive --features /tmp/todoist-features.md --continue-on-fail
```

For each feature you'll see:

```text
drive 1/5  spec    — harness feature
drive 2/5  test-emit — cheap chain writes failing tests
  routing through gemini (cheap_review)
  ✓ wrote 1124 bytes via gemini
drive 3/5  test    — expecting red bar
  ✓ tests red as expected
drive 4/5  impl    — harness do attempt 1/3 (implementation chain)
  [agent] calling claude (implementation, oneshot · API-billed)…
  │ ● Write app/auth.py
  │ ● Edit app/main.py
  │ ✓ Added /auth router
drive 5/5  gate    — harness ci
  ✓ green
  ✓ committed feat: add JWT auth POST /auth/register and POST
```

Smoke:

```bash
harness dev &
curl -s -X POST :8000/auth/register \
  -H 'content-type: application/json' \
  -d '{"email":"a@b.c","password":"hunter2"}' | jq
pkill -f uvicorn
cd ..
```

---

## Phase 2 · Rails 7 API backend

```bash
harness new rails ./api-rails --yes --with-deps
cd api-rails
bundle install                              # if scaffold post-step skipped
harness ci                                  # rubocop + rspec
harness chat --auto-gate
```

Same backlog, different idioms (ActiveRecord migrations,
ApplicationController, RSpec request specs):

```text
[dev|claude ✓]> /drive --features /tmp/todoist-features.md --continue-on-fail
```

The cheap chain emits `spec/requests/auth_spec.rb`, the
implementation chain wires `app/controllers/auth_controller.rb`,
adds migrations under `db/migrate/`, and updates `config/routes.rb`.

Smoke:

```bash
bin/rails db:migrate
harness dev &                               # rails server -p 3000
curl -s -X POST :3000/auth/register …
pkill -f rails
cd ..
```

---

## Phase 3 · Go backend (net/http)

```bash
harness new go ./api-go --yes --with-deps
cd api-go
harness ci                                  # go vet + go test
harness chat --auto-gate
[dev|claude ✓]> /drive --features /tmp/todoist-features.md --continue-on-fail
```

Cheap chain emits table-driven `*_test.go`. Implementation chain
wires `internal/auth/`, `internal/lists/`, `internal/tasks/`,
each behind `http.ServeMux`. `harness ci` runs `go test -race
./...` and `go vet`.

Smoke:

```bash
harness dev &                               # net/http :8001
curl -s :8001/healthz
pkill -f "api-go"
cd ..
```

---

## Phase 4 · Rust backend (Axum)

```bash
harness new rust ./api-rust --yes --with-deps
cd api-rust
harness ci                                  # cargo clippy + cargo test
harness chat --auto-gate
[dev|claude ✓]> /drive --features /tmp/todoist-features.md --continue-on-fail
```

Cheap chain emits `tests/integration_*.rs` using `reqwest::Client`
against the running Axum app. Implementation chain wires
`src/auth.rs`, `src/lists.rs`, etc. Migrations are SQLx-less in the
default scaffold; sub `sqlx` if you want persistence past process
death. The `harness ci` gate runs `cargo clippy -- -D warnings`
and `cargo test`.

Smoke:

```bash
cargo run --release &                       # axum :8002
curl -s :8002/healthz
pkill -f api-rust
cd ..
```

---

## Phase 5 · Ruby (Sinatra) backend

```bash
harness new ruby ./api-ruby --yes --with-deps
cd api-ruby
bundle install                              # if scaffold post-step skipped
harness ci                                  # rubocop + rspec
harness chat --auto-gate
[dev|claude ✓]> /drive --features /tmp/todoist-features.md --continue-on-fail
```

Cheap chain emits `spec/auth_spec.rb` using `Rack::Test`.
Implementation chain wires the Sinatra app under `app.rb` plus
`lib/` helpers.

Smoke:

```bash
harness dev &                               # sinatra :4567
curl -s :4567/healthz
pkill -f rackup
cd ..
```

---

## Phase 6 · React frontend (Vite + Vitest)

```bash
harness new react ./web-react --yes --with-deps
cd web-react
harness ci                                  # vitest + eslint + tsc
harness chat --auto-gate
```

Inline-frontend backlog. Paste it (bracketed paste handles the
newlines):

```text
[dev|claude ✓]> """
                /drive add src/api/client.ts that takes the API base
                URL from VITE_API_BASE env (default http://localhost:8000)
                and exposes register, login, listLists, createList,
                listTasks, createTask, completeTask functions. The JWT
                lives in localStorage as "todoist.jwt" and is sent as
                Authorization: Bearer on every authed call.
                """
[dev|claude ✓]> /drive render src/auth/LoginForm.tsx with email +
                password fields. On submit call client.login and
                redirect to /lists on success. Vitest covers the
                success + 401 paths using msw mocks.
[dev|claude ✓]> /drive render src/lists/Lists.tsx with the user's
                lists, "+ new list" inline form, and a click handler
                that loads tasks into src/tasks/TaskBoard.tsx. Vitest
                covers loading + empty + error using msw.
[dev|claude ✓]> /drive render src/tasks/TaskRow.tsx with checkbox,
                title, priority badge, due-date pill, tag chips, and
                done/complete toggle. Vitest covers prop combinations.
```

Switch the backend per session via env:

```bash
VITE_API_BASE=http://localhost:8000 harness dev   # python
VITE_API_BASE=http://localhost:3000 harness dev   # rails
VITE_API_BASE=http://localhost:8001 harness dev   # go
VITE_API_BASE=http://localhost:8002 harness dev   # rust
VITE_API_BASE=http://localhost:4567 harness dev   # ruby
```

Same React app, five compatible APIs. Visual confirmation of the
contract.

---

## Phase 7 · Cross-stack acceptance

A single shell script you can run against any backend:

```bash
cat > /tmp/todoist-smoke.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
API="${1:?usage: todoist-smoke.sh <base-url>}"
EMAIL="walk-$(date +%s)@example.test"

reg=$(curl -fsS -X POST "$API/auth/register" \
  -H 'content-type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"hunter2\"}")
JWT=$(echo "$reg" | jq -r .token)

LIST=$(curl -fsS -X POST "$API/lists" \
  -H "authorization: Bearer $JWT" \
  -H 'content-type: application/json' \
  -d '{"name":"Today"}' | jq -r .id)

curl -fsS -X POST "$API/lists/$LIST/tasks" \
  -H "authorization: Bearer $JWT" \
  -H 'content-type: application/json' \
  -d '{"title":"Ship multi-stack tutorial","priority":1}' | jq

echo "ok: $API"
EOF
chmod +x /tmp/todoist-smoke.sh
```

Run it across every backend:

```bash
/tmp/todoist-smoke.sh http://localhost:8000   # python
/tmp/todoist-smoke.sh http://localhost:3000   # rails
/tmp/todoist-smoke.sh http://localhost:8001   # go
/tmp/todoist-smoke.sh http://localhost:8002   # rust
/tmp/todoist-smoke.sh http://localhost:4567   # ruby
```

All five must return the same shape. If one diverges, the spec is
ambiguous — refine the backlog entry, `harness drive` again, the
gate re-runs and the test catches the regression.

---

## Phase 8 · Audit + share the spend

Per backend:

```bash
cd api-python
harness chat list                  # one row per /drive session
harness session show feat-auth     # the labelled session
harness session export feat-auth > /tmp/auth-python.json
tree .harness/artifacts/specs      # spec per /drive
tree .harness/artifacts/runs       # blackboard + diff per agent call
harness audit tail --limit 50      # cross-cmd event timeline
```

`/cost` table per session:

```text
session 01KX…: 14 chat turns
  ADAPTER      TASK             TURNS       IN      OUT       COST
  claude       implementation       8    34122     2814 $   0.8412
  gemini       cheap_review         5     2310      940 $   0.0182
  ───────────────────────────────────────────────────────────────
  TOTAL                            13    36432     3754 $   0.8594
```

Aggregate per stack via the JSON export + jq:

```bash
for d in api-*; do
  total=$(cat "$d"/.harness/sessions/*.jsonl 2>/dev/null | \
    jq -s '[.[]|.cost_usd//0]|add')
  printf "%-12s  \$%.4f\n" "$d" "${total:-0}"
done
```

---

## Phase 9 · Back up everything

```bash
cd ~/dev/todoist-multi
harness backup quickstart                          # 3-step recipe
harness backup remote add gdrive --provider drive --interactive
harness backup config set-default-remote gdrive
for d in api-* web-react; do
  (cd "$d" && harness backup snapshot)
done
```

---

## Phase 10 · What you have

- Six repos, each with ~5 conventional commits driven by `/drive`.
- ~25 specs under `.harness/artifacts/specs/`.
- ~25 blackboards + diffs under `.harness/artifacts/runs/`.
- Identical Todoist API surface across Python, Rails, Go, Rust,
  Ruby — verifiable with `/tmp/todoist-smoke.sh`.
- One React frontend that talks to all of them via `VITE_API_BASE`.
- Per-session cost ledger + audit timeline + backup snapshots.

---

## Cheat sheet — works in every chat session

| You type                          | What happens                                    |
|-----------------------------------|-------------------------------------------------|
| `/`                               | inline popup of slash candidates                |
| `/?`                              | grouped slash menu                              |
| `/drive <prompt>`                 | spec → failing tests → impl → ci → commit       |
| `/drive --features /tmp/x.md`     | chain a backlog                                 |
| `/ship <prompt>`                  | branch + spec + loop + commit                   |
| `/exec <prompt>`                  | deterministic harness plan                      |
| `/agents`, `/use <id>`            | inspect / switch adapter                        |
| `/cost`                           | per-adapter token + USD table                   |
| `/timeline`                       | every turn with cost                            |
| `/diff`                           | git diff in project root                        |
| `/recap`                          | cheap-chain summary                             |
| `/save <name>`                    | label session                                   |
| `/branch <name>`                  | git checkout -B + auto-label                    |
| `!<shell>`                        | shell command from project root                 |
| `"""` … `"""`                     | multi-line heredoc                              |
| Cmd+V (paste)                     | bracketed paste mode glues into one prompt      |
| ↑ / ↓                             | readline history (resizes survive SIGWINCH)     |

---

## Trouble?

- **A stack's gate stays red after the cycle** → read the spec under
  `.harness/artifacts/specs/<id>.md`, refine the backlog entry,
  re-run `/drive --features` with `--continue-on-fail` so the
  passing features stay committed.
- **`/drive` aborts with "agent produced no changes"** → the prompt
  was truncated or the worktree was empty for that stack. Re-paste
  the full multi-line block with `"""` heredoc.
- **The React app gets 404 on a stack** → the contract drifted.
  Run `/tmp/todoist-smoke.sh` against that backend to see which
  endpoint is missing, then drive the fix targeted at that stack
  only (`harness chat` in that repo, `/drive add …`).
- **`harness ci` skips optional sensors** → `harness ci
  --install-missing` once per repo; the venv/Gemfile/Cargo deps
  keep them around.

That is the whole loop, five stacks deep. Scaffold once per stack,
drive the same backlog through each, point one React UI at them
all, ship with one snapshot. The harness keeps the spec, the
tests, the diff, the gate, and the audit trail per feature so a
future maintainer can replay any step.
