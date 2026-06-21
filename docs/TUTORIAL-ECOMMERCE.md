# Tutorial — Build a real e-commerce backend with HarnessX

A complete, file-by-file walk that takes a fresh laptop to a green
FastAPI + Vitest-tested React e-commerce app. Every step lists the
files HarnessX touches, the agent it calls (cheap vs expensive),
and what to expect on screen. No skipped phases, no magic — the
commands here are the same the smoke script asserts on every
release.

Estimated time: 30–60 min with `--with-deps`, depending on your
network. Requires macOS or Linux, Python 3.11+, Node 20+, git,
and at least one agent CLI configured (`claude`, `codex`, `kimi`,
`gemini`, or `ollama`).

## 0 · Install

```bash
brew tap rodolfopeixoto/harnessx https://github.com/rodolfopeixoto/harnessx
brew install harness
harness doctor       # flags missing python/node/uv/git/agent CLIs
harness --version    # v0.132.0 or newer
```

If you have an existing install:

```bash
harness update            # latest stable
harness update --force    # reinstall the current tag if a botched cp
                          # left the binary out of date
```

## 1 · Scaffold the FastAPI backend

```bash
cd ~/dev
harness new python-ecommerce ./shop-api --yes --with-deps
cd shop-api
```

What landed on disk:

```text
shop-api/
├── .harness/                 # config + session DB + audit log
├── .venv/                    # ruff + pytest + fastapi (uv-installed)
├── app/
│   ├── main.py               # FastAPI app + /healthz
│   ├── models.py             # Pydantic Product / Cart / Checkout
│   ├── storage.py            # thread-safe in-memory store
│   └── routers/
│       ├── products.py       # GET /products, GET /products/{id}
│       ├── cart.py           # GET/POST /cart/{user}
│       └── checkout.py       # POST /checkout
├── tests/
│   ├── conftest.py
│   ├── test_healthz.py
│   ├── test_products.py
│   ├── test_cart.py
│   └── test_checkout.py
├── pyproject.toml · README.md · requirements.txt · ruff.toml · Makefile
└── pre-push hook installed (runs `harness ci`)
```

Sanity check the floor:

```bash
harness lint     # → All checks passed!
harness test     # → 9 passed
harness ci       # → 7 sensors green, 4 skipped (bandit/mypy/pip-audit missing locally)
```

## 2 · Open the chat and pin an agent

```bash
harness use claude         # pin (optional — auto-pin still works)
harness chat --goal dev    # readline TAB + ↑/↓ history active in TTY mode
```

The header line tells you what is wired:

```text
chat: auto-pinned claude from .harness/config/active.yaml
chat: claude wired — plain text streams to agent; /exec for harness plan

[dev|claude ✓]>
```

You will see one of two header forms when an agent turn runs:

```text
  [agent] calling claude (implementation)…      # plain text → impl chain
  [agent] calling gemini (cheap_review)…        # /recap → cheap chain
```

The `(task)` suffix shows the router decision. `/cost` after the
session lists every adapter touched so you can audit the spend.

## 3 · Feature 1 — Product stock via `/drive`

`/drive` is the spec-driven, test-first slash. It writes a spec to
`.harness/artifacts/specs/`, drops a failing pytest placeholder,
calls the implementation chain only after the red bar is real, and
gates the result with `harness ci`. On green it commits with a
Conventional Commit subject.

```text
[dev|claude ✓]> /drive add a stock field to Product (int >= 0) and
                seed every SKU with a stock count
```

What appears on screen:

```text
drive: "add a stock field to Product (int >= 0) and seed every SKU…" (slug=add-a-stock-field-to-product-int…)
drive: 1/5 — harness feature (spec)
  spec written: .harness/artifacts/specs/01KX…SPEC.md
drive: 2/5 — test-emit (cheap chain)
drive: routing test-emit through gemini (cheap_review)
drive: tests written at tests/test_drive_add-a-stock-field-to-product….py
drive: 3/5 — harness test (expect red)
  test_drive_add_a_stock_field_to_product FAILED
drive: tests red as expected
drive: 4/5 — harness do attempt 1/3 (implementation chain)
  [agent] calling claude…
  │ ● Read app/models.py
  │ ● Edit app/models.py
  │ ● Edit app/storage.py
  │ ● Write tests/test_drive_add-a-stock-field-to-product….py
  │ ✓ Added stock field with ge=0; seeded 25/10/50 per SKU.
drive: 5/5 — harness ci
  9 passed, 0 failed
drive: ✓ green
drive: ✓ committed feat: add a stock field to Product (int >= 0) and seed…
```

Verify on disk:

```bash
git log --oneline -1            # the new commit
git diff HEAD~1 app/models.py   # +stock: int = Field(ge=0)
git diff HEAD~1 app/storage.py  # +stock=25/10/50 per SKU
ls tests/test_drive*            # the test file the cycle wrote
```

## 4 · Feature 2 — Cart accumulation iteration

The cart router exists; this iteration adds a "total includes tax"
behaviour to show how the same `/drive` flow handles edits, not
just additions.

```text
[dev|claude ✓]> /drive include a fixed 10% tax in cart.total_cents
                and expose the breakdown on Cart.tax_cents
```

Watch the same five steps. The cheap chain writes a failing test
that asserts `cart.tax_cents == round(subtotal * 0.10)`. The
implementation chain edits `app/models.py` (new field) and
`app/storage.py` (tax math) until `harness ci` is green.

Now run it for real:

```bash
harness dev &                   # uvicorn on :8000 with reload
sleep 1
curl -s -X POST localhost:8000/cart/alice \
   -H 'content-type: application/json' \
   -d '{"product_id":"sku-001","quantity":2}' | jq
# {"user_id":"alice","items":[{"product_id":"sku-001","quantity":2}],
#  "total_cents":3296,"tax_cents":300}
pkill -f uvicorn
```

## 5 · Feature 3 — Checkout finalisation

```text
[dev|claude ✓]> /drive POST /checkout must zero out the cart and
                return an order summary including tax_cents
```

After green you have one commit per feature on the current branch.
Run `git log --oneline` — the three Conventional Commits read like
a release note.

## 6 · React frontend with the same loop

`harness new react` + `harness drive` works the same. The cheap
chain emits a Vitest skeleton, the implementation chain wires the
components, and `harness ci` runs the JS gate.

```bash
cd ~/dev
harness new react ./shop-web --yes --with-deps
cd shop-web
harness chat --goal dev
```

```text
[dev|claude ✓]> /drive render a /products list calling
                http://localhost:8000/products with a Vitest test
```

```text
drive: 1/5 — harness feature (spec)
drive: 2/5 — test-emit (cheap chain) → vitest skeleton
drive: 3/5 — npm test (expect red)
drive: 4/5 — harness do (implementation chain)
  │ ● Write src/components/ProductList.tsx
  │ ● Write src/components/ProductList.test.tsx
  │ ● Edit src/App.tsx
drive: 5/5 — harness ci
drive: ✓ committed feat: render /products list calling localhost:8000…
```

Start both processes:

```bash
# terminal 1 — api
cd ~/dev/shop-api && harness dev

# terminal 2 — web
cd ~/dev/shop-web && harness dev   # vite on :5173
open http://localhost:5173
```

## 7 · Inspect what happened

`harness session show <label>` dumps the chat-side view; the
`.harness/artifacts/` tree gives the agentic-side view.

```bash
harness chat list                       # newest first; labels visible
harness session show shop-api           # chat session with cost + turns
tree .harness/artifacts/specs           # one .md per /drive feature
tree .harness/artifacts/runs            # blackboards per orchestrate run
harness audit tail --limit 30           # cross-cmd event timeline
```

To replay a chat session without writing anything, attach with
`--replay <label>`. Mutating slashes (`/drive`, `/ship`, `/ci`, …)
are refused; `/history`, `/agents`, `/cost`, `/diff`, `/timeline`
all keep working so you can walk the past run safely.

## 8 · Cheat sheet inside the chat

| You type                    | What happens                                     |
|-----------------------------|--------------------------------------------------|
| `add a /widgets endpoint`   | streams plain text to the implementation chain   |
| `/drive add /widgets`       | spec → failing tests → impl → ci → commit        |
| `/exec add /widgets`        | deterministic harness plan: do + lint + test + ci|
| `/ship add /widgets`        | branch + spec + loop + commit (auto allow-dirty) |
| `/ci`, `/test`, `/lint`     | run the gate                                     |
| `/plan add /widgets`        | print the plan JSON without executing            |
| `/agents`                   | list registered adapters with the active marker  |
| `/use <id>`                 | switch adapter mid-session                       |
| `/diff`                     | `git diff --stat` + full diff in the project root|
| `/cost`                     | cumulative session token + USD spend             |
| `/budget 0.50` / `off`      | refuse further chat turns once spend > cap       |
| `/save my-feature`          | label this session for `harness chat list`       |
| `/branch feature/cart`      | `git checkout -B …` + auto-label session         |
| `/save-prompt add-endpoint` | capture the last plain text into a template      |
| `/prompt add-endpoint`      | replay a saved template                          |
| `/prompts`                  | list saved templates                             |
| `/recap`                    | cheap chain summary of the session so far        |
| `/timeline`                 | ASCII timeline of every turn (clock, action, $)  |
| `/clear`                    | drop conversation history from the next prompt   |
| `/auto-gate on` / `off`     | toggle `harness ci` after each agent turn        |
| `!<shell cmd>`              | run any shell command in the project root        |
| `/goal ops`                 | switch session goal (dev/ops/research/ads)       |
| `/history`, `/last`         | inspect / replay previous prompts                |
| `/help`, `/exit`            | usual                                            |
| ↑ / ↓                       | scroll input history (readline)                  |
| TAB                         | autocomplete slash commands + adapter ids        |

## 9 · Troubleshooting

- **Agent CLI prompts for an OS dialog** (Docker, OrbStack, keychain):
  upstream CLI behaviour, not HarnessX. Approve it once. `harness use
  codex` (or `--no-adapter`) is the deterministic escape hatch.
- **Chat looks frozen** during a long agent call: spinner + dots
  indicate progress. Ctrl-C cancels the turn cleanly; Ctrl-D exits
  the REPL.
- **`harness ship` rejects a dirty tree**: pass `--allow-dirty`, or
  `!git stash` first.
- **`harness new` refuses target**: you are inside a folder with the
  same basename — `cd ..` and rerun.
- **`harness ci` skips `py_ruff`/`py_pytest`**: activate the venv —
  `source .venv/bin/activate` — or rerun the project with
  `--with-deps`.
- **Adapter badge flips to `⚠`**: the periodic health probe failed —
  run `harness doctor` or re-login the agent CLI.

That is the whole workflow. **Plain text → impl chain. `/drive` →
spec/test/impl/ci. `/recap` → cheap chain. ↑/↓ + TAB on every
prompt.** Everything else is documentation.
