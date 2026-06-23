> **Legado.** Substituído por [docs/TUTORIAL.md](../TUTORIAL.md). Mantido aqui só para referência histórica.

# Tutorial — Ship a real Todoist clone with HarnessX

End-to-end build of a productivity SaaS: FastAPI backend, React +
Vite + Vitest frontend, JWT auth, lists, tasks, due dates, tags,
completion, recurring rules, audit log. Every feature ships via
`harness drive` — spec → failing test → implementation → gate →
conventional commit — so the surface area is bounded and the spend
per feature is visible.

You should finish with two repos (`todoist-api`, `todoist-web`),
35–45 committed features across them, full pytest + vitest coverage,
and a one-command `harness backup snapshot` you can hand to a
teammate.

Estimated time: 2–3 h with `--with-deps`. Requires macOS or Linux,
Python 3.11+, Node 20+, git, and one of `claude` / `codex` / `kimi`
configured.

---

## 0 · Install + sanity

```bash
brew install rodolfopeixoto/harnessx/harness   # or scripts/install.sh
harness --version                              # v0.140.0 or newer
harness doctor                                 # flags missing CLIs
```

If something old lingers:

```bash
harness update --force                         # reinstall same tag
```

---

## 1 · Scaffold both projects

```bash
mkdir -p ~/dev/todoist && cd ~/dev/todoist

harness new python-ecommerce ./todoist-api --yes --with-deps
harness new react            ./todoist-web --yes --with-deps
```

The python-ecommerce scaffold already ships the dev gate
(`bandit`, `mypy`, `pip-audit`, `ruff`, `pytest`), so on a fresh
clone:

```bash
cd todoist-api
harness ci                  # → 7 green, 0 failed, 4 skipped (optional venv tools)
harness ci --install-missing  # auto-installs bandit/mypy/pip-audit
harness ci                  # → 11 green, 0 failed
```

`todoist-web`:

```bash
cd ../todoist-web
harness ci                  # node sensors: vitest, eslint, prettier
```

Pin an agent for both projects (you can switch per-feature with
`/use`):

```bash
harness use claude          # writes .harness/config/active.yaml
```

---

## 2 · Open chat against the backend

```bash
cd ../todoist-api
harness chat --auto-gate    # ci runs after every agent turn
```

```text
chat: auto-pinned claude from .harness/config/active.yaml
chat: claude wired — plain text streams to agent; /exec for harness plan
harness chat — session 01KX… , goal=dev
plain text → talk to agent · /exec → plan+run · !<cmd> → shell · /help · /exit
multi-line: end line with \  ·  or wrap with """ … """  ·  / lists slashes
```

Confirm the routing decision is visible:

```text
[dev|claude ✓]> /cost
session 01KX…: 0 chat turns
  (no recorded usage)
```

---

## 3 · Feature 1 · `/auth` register + login (JWT)

`/drive` opens the spec, the cheap_review chain writes failing
pytests, the implementation chain fills them in, the gate decides
green/red, and the green run commits.

Paste this whole block (heredoc handles the newlines, or just
Cmd+V — bracketed paste mode is on):

```text
[dev|claude ✓]> """
                /drive add JWT auth at POST /auth/register and
                POST /auth/login. Persist users in storage with
                bcrypt-hashed password. Issue HS256 token with 24h
                expiry signed by JWT_SECRET (env). Return 409 on
                duplicate email, 401 on bad credentials. Cover
                happy + error paths with pytest in
                tests/test_auth.py.
                """
```

Expected stream:

```text
drive 1/5  spec  — harness feature
drive 2/5  test-emit  — cheap chain writes failing tests
  routing through gemini (cheap_review)
  ✓ wrote 1124 bytes via gemini
drive 3/5  test  — expecting red bar
  ✓ tests red as expected
drive 4/5  impl  — harness do attempt 1/3 (implementation chain)
  [agent] calling claude (implementation, oneshot · API-billed)…
  │ ● Read app/models.py
  │ ● Read app/main.py
  │ ● Write app/auth.py
  │ ● Edit app/main.py
  │ ● Edit app/models.py
  │ ● Write app/security.py
  │ ✓ Added /auth router with HS256 JWT
drive 5/5  gate  — harness ci
  11 passed, 0 failed
  ✓ green
  ✓ committed feat: add JWT auth at POST /auth/register and POST
```

Verify:

```bash
git log --oneline -1
git diff HEAD~1 --stat            # app/auth.py, app/security.py, …
.venv/bin/pytest tests/test_auth.py -v   # everything green
```

---

## 4 · Feature 2 · `/lists` CRUD scoped per user

```text
[dev|claude ✓]> """
                /drive add CRUD for Lists scoped per authenticated
                user: POST/GET /lists, GET/PUT/DELETE /lists/{id}.
                A List has id (ulid), owner_id, name, created_at.
                Other users get 404 on someone else's list. Tests
                cover the auth scope in tests/test_lists.py.
                """
```

The cheap chain emits ~150 LoC of pytest exercising owner
isolation. The implementation chain wires the router, the in-memory
storage, and the per-user query. `harness ci` re-runs the full
suite, and the commit lands as `feat: add CRUD for Lists scoped
per user`.

---

## 5 · Feature 3 · `/tasks` with due dates + priorities

```text
[dev|claude ✓]> """
                /drive add Tasks under a List: POST /lists/{id}/tasks,
                GET /lists/{id}/tasks, PATCH /tasks/{tid}, DELETE
                /tasks/{tid}. Task has id, list_id, title,
                description, priority (1..4), due_at (ISO 8601 or
                null), done (bool), created_at. Listing supports
                ?include_done=true|false (default false). Tests
                cover priority bounds + future due_at + done filter.
                """
```

Watch the file tree:

```bash
tree app tests
# app/
# ├── auth.py
# ├── main.py
# ├── models.py
# ├── routers/
# │   ├── auth.py
# │   ├── lists.py
# │   └── tasks.py
# ├── security.py
# └── storage.py
# tests/
# ├── conftest.py
# ├── test_auth.py
# ├── test_lists.py
# └── test_tasks.py
```

---

## 6 · Feature 4 · `/tags` many-to-many

```text
[dev|claude ✓]> """
                /drive add Tag CRUD plus many-to-many task↔tag link:
                POST /tags, GET /tags, POST /tasks/{tid}/tags
                {tag_id}, DELETE /tasks/{tid}/tags/{tag_id},
                GET /tags/{id}/tasks. Tests cover detach + cascade
                on tag delete.
                """
```

---

## 7 · Feature 5 · `/recurring` daily/weekly rules

```text
[dev|claude ✓]> """
                /drive add recurring Tasks: PATCH /tasks/{tid}
                {recurrence: "daily" | "weekly" | "weekday" | null}.
                When a task marked done has a recurrence, the storage
                schedules the next occurrence with due_at adjusted
                forward by the rule. Tests cover the three rules and
                that completing twice produces two future tasks.
                """
```

---

## 8 · Frontend: list + task UI with Vitest

Switch repos:

```bash
cd ../todoist-web
harness chat --auto-gate
```

```text
[dev|claude ✓]> """
                /drive render src/components/Lists.tsx that calls
                http://localhost:8000/lists (Authorization: Bearer
                <jwt> from localStorage). Click a list to load its
                tasks into src/components/TaskList.tsx with an
                inline "+ new task" form. Vitest covers loading +
                error + empty states using msw mocks in
                src/test/mocks/.
                """
```

The frontend has its own gate: `vitest`, `eslint`, `prettier`. The
cheap chain emits the vitest skeleton + msw handlers; the
implementation chain wires the components.

Two more iterations:

```text
[dev|claude ✓]> /drive add a TaskRow component with checkbox,
                priority badge, due-date pill, and tag chips.
                Vitest covers each prop combination.
```

```text
[dev|claude ✓]> /drive add src/auth/AuthGate.tsx that wraps the
                router. Unauthenticated users land on a Login form
                that POSTs /auth/login, stores the JWT in
                localStorage, and redirects. Vitest covers the
                redirect.
```

---

## 9 · Run the whole thing

Two terminals:

```bash
# terminal 1 — backend
cd ~/dev/todoist/todoist-api
source .venv/bin/activate
JWT_SECRET=$(openssl rand -hex 32) harness dev
# uvicorn on :8000
```

```bash
# terminal 2 — frontend
cd ~/dev/todoist/todoist-web
harness dev
# vite on :5173
open http://localhost:5173
```

Register, log in, add a list, add tasks, complete one, mark
another recurring. Network tab shows the JWT travelling on every
call.

---

## 10 · Audit the run

```bash
cd ~/dev/todoist/todoist-api
harness chat list                 # one row per feature
harness session show feat-tasks   # the labelled session
harness session export feat-auth > shareable.json
tree .harness/artifacts/specs     # one spec per /drive
tree .harness/artifacts/runs      # blackboard + diff per agent invocation
harness audit tail --limit 50     # cross-cmd event timeline
```

`/cost` inside any chat session shows the per-adapter spend table
so you can see which chain handled what:

```text
session 01KX…: 14 chat turns
  ADAPTER      TASK             TURNS       IN      OUT       COST
  claude       implementation       8    34122     2814 $   0.8412
  gemini       cheap_review         5     2310      940 $   0.0182
  ───────────────────────────────────────────────────────────────
  TOTAL                            13    36432     3754 $   0.8594
```

---

## 11 · Backup the project state

```bash
harness backup quickstart              # prints the 3-step recipe
harness backup remote add gdrive --provider drive --interactive
harness backup config set-default-remote gdrive
harness backup snapshot                # ships .harness/ + specs + sessions
```

Secrets stay out by default. To include them: `--include-secrets
HARNESS_BACKUP_I_UNDERSTAND_SECRETS=1`, and route through an
`rclone crypt` overlay.

---

## 12 · What you have on disk

- ~12 conventional commits on `todoist-api` (auth, lists, tasks,
  tags, recurring, plus the cheap-chain tests for each).
- ~6 conventional commits on `todoist-web` (lists, tasks, auth
  gate, task row).
- `.harness/artifacts/specs/*.md` — one spec per `/drive`,
  reviewable as PR context.
- `.harness/artifacts/runs/*/diff.patch` + `report.md` per agent
  invocation, including the worktree blackboard.
- `.harness/sessions/<id>.jsonl` + `.meta.json` for every chat
  session, replayable with `harness chat --replay <label>`.
- Pre-push hook on both repos runs `harness ci`.

---

## 13 · Cheat sheet inside the chat

| You type                          | What happens                                    |
|-----------------------------------|-------------------------------------------------|
| `add /tags endpoint`              | streams to the implementation chain              |
| `/drive add /tags endpoint`       | spec → failing test (cheap chain) → impl → ci   |
| `/exec add /tags endpoint`        | deterministic harness plan                      |
| `/ship add /tags endpoint`        | branch + spec + loop + commit (auto allow-dirty)|
| `/ci`, `/test`, `/lint`           | run the gate                                    |
| `/agents`, `/use <id>`            | inspect / switch adapter                        |
| `/cost`                           | per-adapter token + USD table                   |
| `/timeline`                       | every turn with cost                            |
| `/recap`                          | cheap-chain summary                             |
| `/save <name>`, `/branch <name>`  | label session + git branch                      |
| `/save-prompt <n>`, `/prompt <n>` | template capture + replay                       |
| `/clear`                          | drop conversation history                       |
| `/auto-gate on` / `off`           | toggle ci after each agent turn                 |
| `!<shell>`                        | shell command from project root                 |
| `"""` … `"""`                     | multi-line heredoc                              |
| Cmd+V (paste)                     | bracketed paste mode glues into one prompt      |
| `/` (alone) or `/?`               | grouped slash menu                              |
| `/<TAB>`                          | readline autocompletes                          |
| ↑ / ↓                             | readline history (resizes survive SIGWINCH)     |

---

## 14 · Trouble?

- **`harness ci` skips `bandit`/`mypy`/`pip-audit`** → run
  `harness ci --install-missing` once; venv keeps them.
- **`/drive` aborts with "agent produced no changes"** → the
  prompt was ambiguous or truncated. Re-paste with `"""` heredoc
  or read the spec under `.harness/artifacts/specs/` and refine.
- **`/cost` looks small but the bill is bigger** → the bundled
  oneshot adapters (`claude --print`, `codex exec`, `gemini -p`)
  charge against your API key. Use `claude-interactive` /
  `kimi` for plan/local tokens — billing mode shows in the chat
  header.
- **Pasting still splits in two** → your terminal lacks bracketed
  paste support. Wrap the prompt with `"""` `…` `"""`.
- **Typed `/drv`, got a suggestion** → `/drive` is one letter
  away. The REPL prints `did you mean /drive?`.

That is the whole loop: scaffold once, `/drive` per feature, gate
runs on every turn, commits land green, frontend wires through the
same loop. Build a Todoist or whatever else fits the same shape.
