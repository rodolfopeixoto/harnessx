# Tutorial — Build a small e-commerce API end-to-end with HarnessX

In this tutorial you will build a small **FastAPI e-commerce backend**
from zero while exercising every layer of HarnessX. Each step cites the
paper section it implements so you finish the walkthrough understanding
both the tool and the *"Code as Agent Harness"* framework
(arXiv 2605.18747).

The application supports:

- `GET /healthz` — liveness
- `GET /products` — product catalogue
- `POST /cart` — add to cart
- `GET /cart/{user_id}` — read cart
- `POST /checkout` — checkout an existing cart

Estimated time: 60 min. Requires: macOS or Linux, Python 3.11+, git,
network for adapter calls (optional — most steps are deterministic).

> Reference material:
> - [`PAPER-MAPPING.md`](PAPER-MAPPING.md) — paper § → command
> - [`COMMANDS.md`](COMMANDS.md) — every command in one page
> - [`ARCHITECTURE.md`](ARCHITECTURE.md) — runtime architecture

---

## 0. Install

```bash
brew tap rodolfopeixoto/harnessx https://github.com/rodolfopeixoto/harnessx
brew install harness
harness version
harness doctor
```

`harness doctor` flags any missing dependency. Install before moving on
or the per-stack sensors will be skipped.

---

## 1. Bootstrap the project (§2.3 + §5.1.1)

One command does git init, `.harness/`, the Python scaffold, and the
git hooks.

```bash
harness new python ./shop-api --yes --with-deps
cd shop-api
```

You now have:

- `app.py`, `tests/test_app.py` — FastAPI hello world.
- `requirements.txt`, `pyproject.toml`, `ruff.toml` — toolchain.
- `Makefile` — convenience targets.
- `.harness/` — runtime: config, db, hooks, logs.
- `.harness/config/project.yaml` — canonical
  `test`/`lint`/`run`/`bench` commands, used by the wrappers.
- `.git/hooks/pre-push` — runs `harness ci` before every push.

Verify the wrappers resolve the right commands without you remembering
them:

```bash
harness test                # .venv/bin/pytest -q
harness lint                # .venv/bin/ruff check .
harness dev &               # uvicorn on :8000
curl http://localhost:8000/healthz
kill %1
```

---

## 2. Inspect the harness surface (§3.3 + §3.4.4)

```bash
harness scaffold list       # bundled stacks
harness sensor list         # every sensor that applies to this project
harness routes              # task → adapter chain table
harness agent list          # registered adapters
harness flow list           # bundled flows
harness orchestrate list    # user-defined orchestration flows
```

The sensors include the universal pack (`forbidden_files`,
`secrets_scan`, …) plus the Python stack pack (`py_ruff`, `py_pytest`,
`py_mypy`, `py_bandit`, `py_pip_audit`). Run them to baseline:

```bash
harness ci
```

The first run may report missing optional tools (mypy, bandit). Install
them with `pip install mypy bandit` to take the catalogue to green.

---

## 3. Write the plan contract (§3.4.2)

Before any LLM call, pin the scope deterministically:

```bash
harness plan write "build product catalogue, cart, and checkout endpoints" \
  --file "app.py" \
  --file "tests/test_app.py" \
  --invariant "GET /healthz still returns 200" \
  --invariant "ruff stays green" \
  --validate "harness test" \
  --validate "harness lint" \
  --validate "harness ci" \
  --rollback "git revert HEAD" \
  --risk medium
```

This writes `.harness/artifacts/plans/PLAN-<ulid>.md`. Capture the id
into a variable (the simplest way is to derive it from the file name):

```bash
PLAN_ID=$(ls -t .harness/artifacts/plans/PLAN-*.md | head -1 \
  | xargs basename | sed 's/^PLAN-//; s/\.md$//')
echo "${PLAN_ID}"
```

Pin it as the active contract so the `plan_scope` sensor enforces it:

```bash
mkdir -p .harness/config
printf "active_plan_id: %s\n" "${PLAN_ID}" > .harness/config/plan.yaml

harness sensor list | grep plan_scope     # confirm it now applies
harness plan check --plan "${PLAN_ID}"    # baseline: empty diff -> green
```

> Both `PLAN-<id>`, `<id>`, or `PLAN-<id>.md` are accepted by the
> `--plan` flag.

---

## 4. Pick a model (§3.5.3)

By default, HarnessX uses the bundled router map. Override per task
with the audited config wizard:

```bash
harness config show
harness config set --task implementation --primary kimi --fallback gemini,claude --budget 0.5
harness config show
cat .harness/logs/config-mutations.jsonl   # before / after recorded
```

Prefer the deterministic planner first; flip to the LLM planner when
you want a richer plan:

```bash
harness chat --goal dev
> /plan add /products endpoint returning a static catalogue
> /exit

harness chat --goal dev --adapter ollama --model llama3.1:8b
> add /products endpoint with pytest coverage
> /exit
```

Each `harness chat` session lands at
`.harness/sessions/<ulid>.jsonl` with every turn (input, plan, exec
result).

---

## 5. Ship the first feature (§3.4 + §3.4.2)

`harness ship` refuses to run on a dirty tree because the loop needs a
known baseline. Commit the scaffold first:

```bash
git add -A
git -c user.email=you@example.com -c user.name="You" \
    commit -m "chore: initial scaffold"
```

Now ship the feature end-to-end:

```bash
harness ship "implement GET /products returning a static catalogue with pytest" \
  --plan "${PLAN_ID}" \
  --autonomy ask \
  --max-attempts 3 \
  --rate-limit-retries 3
```

Under the hood `ship`:

1. Branches `feature/implement-get-products...` from `develop` (or
   `main` if `develop` is absent).
2. Calls `harness feature` to write a spec.
3. Loops `harness do` + `harness ci` until green or attempts
   exhausted, with exponential backoff on `429` markers.
4. Runs `harness plan check --plan ${PLAN_ID}` before committing.
5. On success, emits a Conventional Commit referencing the plan.

If `harness ci` stays red, the loop emits failure context back into
the next `harness do` attempt — paper §3.4 PEV loop.

Inspect the run:

```bash
git log --oneline -1            # see the conventional commit
harness audit tail --limit 30   # event timeline (may be empty until you run a do/ship that touches sensors)
ls .harness/runs/                # per-run scratch dirs
```

---

## 6. Add the cart endpoints with an orchestration flow (§4.1)

For richer changes (multiple roles, critic + repair), use a flow:

```bash
mkdir -p .harness/orchestrations
cat > .harness/orchestrations/cart-cycle.yaml <<'EOF'
name: cart-cycle
description: Build /cart endpoints with planner -> coder -> tester chain
topology: chain
steps:
  - role: planner
    command: ["harness", "plan", "write", "implement POST /cart and GET /cart/{user_id}",
              "--file", "app.py", "--file", "tests/test_app.py",
              "--invariant", "existing endpoints unchanged",
              "--validate", "harness ci", "--risk", "medium"]
  - role: coder
    command: ["harness", "do",
              "implement POST /cart and GET /cart/{user_id} backed by an in-memory dict",
              "--autonomy", "safe_execute", "--apply"]
  - role: tester
    command: ["harness", "test"]
  - role: reviewer
    command: ["harness", "ci"]
EOF

harness orchestrate show cart-cycle
harness orchestrate run cart-cycle
cat .harness/artifacts/runs/*/blackboard.json | head -40
```

The blackboard JSON is the paper's "file-only shared substrate"
(§4.3.1). Each role's stdout becomes context for the next role.

You can drop a role-played LLM step too by setting `adapter: ollama`
instead of `command:` — see [`COMMANDS.md`](COMMANDS.md#harness-orchestrate-listshow-run-name---dry-run---adapter-timeout-duration).

---

## 7. Add the checkout endpoint with the typed REPL (§3.1.4)

Interactive iteration without re-typing flags:

```bash
harness chat --goal dev --adapter ollama
[dev]> /goal dev
[dev]> implement POST /checkout that empties the cart for the user
[dev]> harness test
[dev]> /exit
```

Each prompt is converted into a typed `Plan` JSON whose `steps[].cmd`
references are restricted to the dev palette (`plan, ship, test, lint,
ci, check, scaffold, smoke, memory, evolve, coverage`). LLM-emitted
plans cannot ask for `runtime`, `containers`, or anything off-palette.

---

## 8. Promote learnings to memory (§3.2)

Once the suite stays green, capture the convention as long-term
memory. Pick a real run id from `.harness/runs/`:

```bash
RUN_ID=$(ls -t .harness/runs/ | head -1)
echo "${RUN_ID}"   # must be a ulid; if blank, run `harness ship` or `harness do` first

harness memory promote \
  --scope project --kind semantic \
  --content "cart state lives in app.state.cart_db; do not import a new dict" \
  --run-id "${RUN_ID}" --confidence 0.85

harness memory list --kind semantic
harness memory recall "cart state"
```

> `harness memory promote` rejects an empty / flag-shaped `--run-id`,
> so a typo here surfaces immediately instead of writing a memory
> with the default confidence.

The five paper kinds (`working | semantic | experiential | long_term |
multi_agent`) all show up under `harness memory list --kind`.

---

## 9. Enforce the 90% coverage floor (§3.4.4)

```bash
harness coverage --threshold 0.9         # gate; non-zero on red
```

For a Python project the gate is bundled with `pytest-cov` once you add
it; for the harnessx-go internal you saw earlier the gate is wired to
`go test -cover`.

You can also drop `pytest --cov` into the Makefile and run it through
`harness bench` if you want a single command.

---

## 10. Iterate with `harness loop` (§3.4)

Watch the working tree for changes and re-run the PEV loop:

```bash
harness loop --max-attempts 3 &
LOOP_PID=$!
# in another shell, change app.py
kill -INT $LOOP_PID
harness loop resume <run-id>             # restore state.json
```

---

## 11. Diagnose and evolve the harness (§3.5)

After a few sessions the telemetry log is rich enough to cluster:

```bash
harness evolve diagnose
harness evolve diagnose --json > /tmp/diag.json
```

Propose a mutation candidate (no automatic apply):

```bash
harness evolve propose \
  --component router \
  --description "drop claude from cheap_review fallback" \
  --rationale "always 429 under load" \
  --risk low
```

Replay the candidate against a held-out trace inside an isolated
sandbox (paper §3.5.2):

```bash
# Snapshot the current trace as the held-out set.
cp .harness/logs/events.jsonl /tmp/heldout.jsonl

# A/B replay: baseline vs the same binary as the candidate.
harness evolve sandbox /tmp/heldout.jsonl
```

You can pass `--candidate /path/to/another/harness` when you actually
have a mutated binary to compare. Promotion is HITL-gated (§3.5.3):

```bash
MUT_ID=$(jq -r .mutation.id .harness/logs/mutations.jsonl | head -1)
harness evolve promote ${MUT_ID} --hitl --reason "review traffic improved"
cat .harness/logs/mutations.jsonl
```

---

## 12. Cross-stack regression and tutorial replay

Before pushing anything to a shared branch, exercise the whole CLI
surface and the cheat-sheet replay:

```bash
harness smoke matrix --langs all --step-timeout 180s
make tutorial-replay                     # deterministic walk
harness ci                               # full sensor gate
```

`harness ci` will run `plan_scope` (your active plan is pinned),
`secrets_scan`, lint, pytest, performance budget, and the optional
coverage gate.

---

## 13. Wrap up

```bash
git log --oneline -10
ls .harness/artifacts/plans/
ls .harness/artifacts/runs/
harness backup save
harness backup list
```

Push your work — the `pre-push` hook will re-run `harness ci` and
reject any red state:

```bash
git push -u origin feature/shop-api-cart-checkout
```

---

## Cheat-sheet (one-pager)

```bash
harness new python ./shop-api --yes
cd shop-api

harness plan write "build product catalogue and cart" \
  --file app.py --file tests/test_app.py \
  --invariant "ruff stays green" \
  --validate "harness ci" --risk medium
PLAN_ID=01...

printf "active_plan_id: %s\n" "$PLAN_ID" > .harness/config/plan.yaml

harness config set --task implementation --primary kimi \
  --fallback gemini,claude --budget 0.5

harness ship "implement GET /products with pytest" --plan ${PLAN_ID}
harness orchestrate run cart-cycle
harness chat --goal dev --adapter ollama

harness memory promote --scope project --kind semantic \
  --content "cart lives in app.state.cart_db" \
  --run-id <run> --confidence 0.85

harness coverage --threshold 0.9
harness evolve diagnose
harness evolve sandbox /tmp/heldout.jsonl
harness backup save
harness dashboard
```

You exercised every layer:

- §2 Interface — scaffold, projectcfg, adapters.
- §3.1 Planning — `plan write`, `chat`, `orchestrate`.
- §3.2 Memory — typed promotions.
- §3.3 Tool Use — adapter routing.
- §3.4 PEV — `ship`, `loop`, sensors, plan-as-contract, coverage gate.
- §3.5 AHE — `evolve diagnose / propose / sandbox / promote --hitl`.
- §4 Scaling — orchestrate roles + topology + blackboard.
- §5.1.1 Code assistants — `new`, `ship`, `chat`, wrappers.

---

## Troubleshooting

| Symptom | Fix |
|---|---|
| `harness doctor` red | Install the listed dep; re-run `harness ci` |
| `harness ship` keeps retrying | Inspect `harness audit tail`; adjust `--rate-limit-retries` |
| `harness ci` blocks on `plan_scope` | Run `harness plan check --plan ${PLAN_ID}` to see violations, then either edit the plan (`harness plan write`) or revert out-of-scope files |
| `pre-push` blocks the push | Reproduce locally with `harness ci`; bypass only with `HARNESS_SKIP_CI=1` |
| Dashboard 404 | `make dashboard-install && make dashboard-build && harness dashboard` |

File issues at <https://github.com/rodolfopeixoto/harnessx/issues>.
