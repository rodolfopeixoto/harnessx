# Tutorial — Build an e-commerce API by chatting with HarnessX

You build a small **FastAPI e-commerce backend** (products → cart →
checkout) by **talking to a pinned agent inside `harness chat`**.
Every paragraph below is one chunk you can paste and watch happen.

This is the narrative version of the workflow.
For raw command reference jump to [`COMMANDS.md`](COMMANDS.md). For the
paper mapping see [`PAPER-MAPPING.md`](PAPER-MAPPING.md).

The harness chat REPL (v0.116.0+) treats:

- **plain text** → chat directly with the pinned agent (streams reads,
  writes, diffs);
- **`/exec <prompt>`** (alias `/do`) → deterministic harness plan
  (`do` → `lint` → `test` → `ci`);
- **`/ship <prompt>`** → `harness ship` (branch + spec + loop + commit);
- **`/ci` `/test` `/lint`** → run the gate directly;
- **`!<shell>`** → escape to a shell command in the project root.

You never leave the chat window. CTRL-C aborts the current turn (the spinner clears, you land back at
the prompt); `/exit` or CTRL-D ends the session.

---

## 0. Install once

```bash
brew tap rodolfopeixoto/harnessx https://github.com/rodolfopeixoto/harnessx
brew install harness
harness doctor          # flags missing python/uv/git/agent CLI
```

Pick at least one agent CLI: `claude`, `codex`, `gemini`, `kimi`, or
`ollama`. The tutorial below pins **claude** — substitute freely.

---

## 1. Bootstrap the project

```bash
cd ~/dev
harness new python-ecommerce ./shop-api --yes --with-deps
cd shop-api
```

What just happened (one line each):

- `git init` on `main`;
- `.harness/` created (config + db + logs);
- 20 scaffold files written: `app/main.py`, routers, models, storage,
  tests for `/healthz`;
- `requirements.txt` installed into `.venv` via `uv`;
- pre-push hook installed; baseline commit `chore: scaffold baseline`.

Sanity check the floor:

```bash
harness lint    # ruff
harness test    # pytest — 9 tests pass
harness ci      # full sensor gate — green
```

> If `harness new` errors with "refusing nested target", you are
> standing inside a folder of the same name. `cd ..` and rerun.

---

## 2. Pin an agent and open chat

```bash
harness use claude          # pins claude as the active adapter
harness chat --goal dev     # auto-picks up the pin; --adapter still works
```

Pass `--auto-gate` if you want `harness ci` to run after every agent
turn — handy for tight feedback during the cart/checkout iterations
below. You can toggle it later from inside chat with `/auto-gate on`
or `/auto-gate off`.

You land in:

```
chat: claude wired — plain text streams to agent; /exec for harness plan
harness chat — session 01..., goal=dev
plain text → talk to agent · /exec → plan+run · !<cmd> → shell · /help · /exit

[dev|claude ✓]>
```

The badge `|claude ✓` means the periodic health probe pinged the
adapter successfully.

---

## 3. Add the product catalogue (talking to claude)

At the prompt, type the feature in your own words:

```
[dev|claude ✓]> add a real product catalogue: hard-code 3 products in app/storage.py
with name, price_cents, stock. Update GET /products to return them.
Add pytest in tests/test_products.py covering 3-item response and 200
status. Keep the existing /healthz green.
```

You will watch a live stream like:

```
  [agent] calling claude…
  │ ● Reading app/storage.py
  │ ● Reading app/routers/products.py
  │ ● Writing app/storage.py
  │ ● Writing tests/test_products.py
  │ ● Diff: 2 files changed, +18 -1
  ✓ claude done in 22s (in=4310 out=812 ~$0.0254)
```

When it stops, verify in the same chat:

```
[dev|claude ✓]> /test
[dev|claude ✓]> /ci
```

Both must come back green. If they fail, **stay in the chat** and
describe the failure in plain text — claude will read the failing
output and patch.

---

## 4. Cart (multi-turn iteration)

Same chat, next feature. Notice every line is conversational — no
shell escaping, no flag soup:

```
[dev|claude ✓]> implement POST /cart/{user_id} and GET /cart/{user_id}
backed by an in-memory dict in app/storage.py. POST takes
{product_id, quantity}; GET returns the running total. Cover both with
pytest in tests/test_cart.py including missing-product 404.
```

Watch the diff stream. Then gate it:

```
[dev|claude ✓]> /ci
```

If the gate breaks, debug conversationally:

```
[dev|claude ✓]> the 404 path returns 500; trace it and fix
```

claude will read the traceback, propose the patch, and re-run.

---

## 5. Checkout + a one-shot ship

For the third feature use the **`/ship`** slash command. That switches
to `harness ship` under the hood: a branch is cut, a spec is written,
`harness do` runs in a bounded loop, and the green result is committed
with a Conventional Commit subject.

```
[dev|claude ✓]> /ship implement POST /checkout that clears the cart for
                 a user and returns the total; tests in tests/test_checkout.py
```

Stream looks like:

```
  $ harness ship implement POST /checkout … --yes
  ship: agent=claude
  ship: branch feature/implement-post-checkout ← develop
  ship: feature spec — harness feature implement POST /checkout … --yes
  ship: do attempt 1 — harness do … --autonomy ask --yes
    [agent] calling claude…
    | ● Reading app/routers/checkout.py
    | ● Writing tests/test_checkout.py
  ship: ci attempt 1 — harness ci
  ✓ ship: green
```

> If you start a `/ship` with uncommitted work in the tree (because
> you were iterating in chat first), add **`--allow-dirty`** to the
> ship command — the dirty diff becomes part of the same commit.

---

## 6. Run the API and curl it

```
[dev|claude ✓]> !.venv/bin/uvicorn app.main:app --reload &
[dev|claude ✓]> !curl -s localhost:8000/products | jq
[dev|claude ✓]> !curl -s -X POST localhost:8000/cart/alice \
                  -H 'content-type: application/json' \
                  -d '{"product_id":"sku-001","quantity":2}' | jq
[dev|claude ✓]> !curl -s -X POST localhost:8000/checkout \
                  -H 'content-type: application/json' \
                  -d '{"user_id":"alice"}' | jq
[dev|claude ✓]> !pkill -f uvicorn
```

The `!` prefix runs the line through `sh -c` from the project root,
so you stay inside the chat session.

---

## 7. Wrap up

```
[dev|claude ✓]> /exit
bye
```

What you have on disk:

- 4 routers, 1 storage layer, ~12 pytest cases — all green;
- one feature branch per `/ship` with a Conventional Commit;
- `.harness/sessions/<id>.jsonl` with every turn, plan, and result
  (replayable with `harness chat --replay <id>` — see roadmap);
- `.harness/artifacts/specs/*.md` for every feature you shipped;
- a pre-push hook that re-runs `harness ci` before any `git push`.

---

## Cheat sheet inside the chat

| You type                    | What happens                                     |
|-----------------------------|--------------------------------------------------|
| `add /products endpoint`    | streams a chat turn to the pinned agent          |
| `/exec add /products`       | deterministic harness plan: do + lint + test + ci|
| `/ship add /products`       | branch + spec + loop + commit (auto `--allow-dirty`)|
| `/ci`, `/test`, `/lint`     | run the gate                                     |
| `/plan add /products`       | print the plan JSON without executing            |
| `/agents`                   | list registered adapters; mark the active one    |
| `/use <id>`                 | switch adapter mid-session (kimi/codex/gemini/…) |
| `/diff`                     | `git diff --stat` + full diff in project root    |
| `/cost`                     | cumulative session token + USD spend             |
| `/budget 0.50` / `off`      | refuse further chat turns once spend > cap       |
| `/save my-feature`          | label this session for `harness chat list`       |
| `/branch feature/cart`      | `git checkout -B …` + auto-label session         |
| `/save-prompt add-endpoint` | capture the last plain text into a template      |
| `/prompt add-endpoint`      | replay a saved template as a new chat input      |
| `/prompts`                  | list saved prompt templates                      |
| `/recap`                    | ask the agent for an ≤8-bullet summary           |
| `/clear`                    | drop conversation history from the next prompt   |
| `/auto-gate on` / `off`     | toggle `harness ci` after each agent turn        |
| `!<shell cmd>`              | run any shell command in the project root        |
| `/goal ops`                 | switch session goal (dev/ops/research/ads)       |
| `/history`, `/last`         | inspect / replay previous prompts                |
| `/help`, `/exit`            | usual                                            |

## Resume an old session

```bash
harness chat list                  # newest first; labels shown as a column
harness chat <id|label>            # positional shortcut for --resume
harness chat --resume <id|label>   # continues writing to the same .jsonl
harness chat --replay <id|label>   # read-only: /history, /agents, /cost, /diff only
```

`<label>` is anything you typed into `/save`. Passing a string that
matches neither a saved label nor a known ulid falls through to
`--resume` and produces a clean "session not found" error.

The previous turns flow back into `/history` and the Working Memory
preamble, so the agent reads the conversation as one thread instead of
a cold start. `/clear` resets that preamble without ending the session.

## Reusable prompt templates

If you find yourself typing the same instruction every time you add a
new endpoint, capture it:

```
[dev|claude ✓]> add a /<name> endpoint with pytest tests and update README
[dev|claude ✓]> /save-prompt add-endpoint
  ✓ prompt "add-endpoint" saved (74 chars)

# later, in a fresh session:
[dev|claude ✓]> /prompt add-endpoint
↻ replaying prompt "add-endpoint"
  [agent] calling claude…
```

Templates live in `.harness/prompts/<name>.md` and ship with the
repo when you commit them. `/prompts` lists every name currently
saved.

## Share a session

```bash
harness session export <id> > session.json
```

Emits one JSON envelope with the goal, started timestamp, every turn,
cumulative tokens/USD, plus the toggles (`auto_gate`, `budget_usd`,
`context_mark`). Drop the file into a code review or a bug report —
the receiver can replay it locally with `jq` or load it back through
`harness chat --resume`.

---

## When something goes sideways

- **Agent CLI prompts for an OS dialog** (Docker, OrbStack, keychain):
  that is the upstream CLI, not HarnessX. Approve it once. From then
  on the same install is fine; if the dialog keeps reappearing
  switch with `harness use codex` (or any other adapter), or run
  `harness chat --no-adapter` to fall back to the deterministic
  planner.
- **`harness ship` rejects a dirty tree**: pass `--allow-dirty`, or
  `!git stash` first.
- **`harness new` refuses target**: you are inside a folder with the
  same basename. `cd ..` and rerun.
- **`harness ci` skips `py_ruff`/`py_pytest` with "binary not on PATH"**:
  activate the venv first — `source .venv/bin/activate`.
- **Adapter health badge flips to `⚠`**: the periodic probe failed.
  Run `harness doctor` or check the CLI is logged in.

That is the whole workflow. Plain text → agent. Slash → harness.
Bang → shell. Everything else is documentation.
