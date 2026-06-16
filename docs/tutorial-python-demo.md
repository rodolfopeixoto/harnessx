# Tutorial — Python FastAPI demo, end-to-end through every harness feature

Goal: build a working Python FastAPI service from zero, driving
every adapter, sensor, flow, hook, MCP, autonomy gate, optimizer,
critic, devloop, recall, sharedstate, backup, and dashboard surface
along the way.

Estimated time: 45–60 min. Requires: macOS or Linux, Python 3.11+,
Docker (optional for runtime container demo), git, network.

> Read the [README](../README.md) first for layer mental model.
> Architecture reference: [ARCHITECTURE.md](../ARCHITECTURE.md).

---

## 0. Install

```bash
brew tap rodolfopeixoto/harnessx https://github.com/rodolfopeixoto/harnessx
brew install harness
harness version          # expect 0.104.0+
harness doctor           # diagnose env (Go optional, Python recommended)
```

If `harness doctor` flags missing deps, install before continuing.

---

## 1. Bootstrap project

```bash
mkdir demo-api && cd demo-api
git init
harness init             # writes .harness/ + policies + first scaffold prompt
harness install-git-hooks # wires pre-push gate (lint + tests + coverage)
ls .harness/             # audit/ policies/ runs/ secrets/
```

**What just happened:**
- `.harness/audit/events.jsonl` opened (append-only replay log)
- `.harness/policies/autonomy.json` seeded with default per-path rules
- `.git/hooks/pre-push` symlinked to `scripts/git/pre-push.sh`

---

## 2. Inspect bundled assets

```bash
harness scaffold list        # python, go-cli, react-spa, …
harness hookpkg list         # pre-commit, gitleaks, conventional-commits
harness mcppkg list          # filesystem, git, sqlite MCP servers
harness skill list           # skill manifests
harness flow list            # rails-api, python-fastapi, meta-ads-campaign, …
harness sensor list          # secrets_scan, lint, tests, multimodal_grounding
harness catalog              # all-in-one inventory view
```

---

## 3. Scaffold Python app (Layer 1 — Executability)

Deterministic, no LLM call:

```bash
harness scaffold apply python --apply --with-git --with-deps
```

Creates:
- `app/__init__.py`, `app/main.py` (FastAPI hello)
- `tests/test_main.py`
- `requirements.txt`, `requirements-dev.txt`
- `.venv/` (deps installed when `--with-deps`)
- `Makefile` with `test`, `lint`, `run` targets

Verify:

```bash
.venv/bin/pytest -q       # expect 1 passed
.venv/bin/uvicorn app.main:app --reload    # in another shell
curl http://localhost:8000/health
```

---

## 4. Route prompt → adapter (Layer 1 — routing)

```bash
harness route show "add a /healthz endpoint with a pytest"
```

Output shows which adapter strengths matched. To force a specific
adapter:

```bash
harness route show "add /healthz" --adapter claude
```

---

## 5. One-shot agentic edit (Layer 3 — `do`)

```bash
harness do "add a GET /healthz endpoint returning {status:ok}; cover with a pytest" \
  --autonomy ask
```

Pipeline:
1. `router.Pick(strengths)` selects best CLI adapter
2. Adapter generates diff
3. **Critic loop** routes diff to a second adapter (`tags:["review","critic"]`)
4. Sensors gate: secrets scan + lint + pytest
5. **Sharedstate** commits read/write set; conflict detector blocks stale
6. Recall stores prompt + outcome (BM25 + optional embeddings)
7. Audit log appends events
8. `--autonomy ask` prompts for approval before each tool

Inspect:

```bash
harness metrics show              # trajectory: ToolCalls, Tokens, Edits, WallMs
harness audit tail --limit 20     # last 20 events
ls .harness/runs/                 # per-run shared.json + do.md + artifacts
```

---

## 6. Continuous loop (Layer 3 — `loop`)

```bash
harness loop                      # watches *.py, reruns on save, gates sensors
# in another shell, touch app/main.py — loop re-attempts
# kill -INT to checkpoint
harness loop resume <run-id>      # restore from .harness/runs/_loop/<id>/state.json
```

If a phase fails because of a missing dep, `taskgraph.Replan` rewrites
the graph mid-flight — visible via `harness audit tail`.

---

## 7. Sensors (Layer 2 — Verifiability)

```bash
harness sensor run secrets_scan
harness sensor run lint
harness sensor run tests
# multimodal grounding (image attached prompt):
harness do "implement the layout shown" --image ui-mock.png
harness sensor run multimodal_grounding --image ui-mock.png
```

Every sensor emits `Result{Confidence, Scope, Verified, Unverified, Risks}`.
Devloop refuses green when `Confidence < 0.5 ∧ Unverified ≠ ∅`.

---

## 8. Autonomy approvals (Layer 2 — Human-in-the-loop)

```bash
harness autonomy show
harness do "rm -rf logs/" --autonomy strict   # denial recorded
harness do "rm -rf logs/" --autonomy ask      # approve prompt
cat .harness/audit/approvals.jsonl            # full history
harness autonomy suggest                      # mines history → per-path policy proposals
```

Apply proposal:

```bash
harness autonomy suggest --apply
```

---

## 9. Change-contract optimizer (Layer 2 — self-improvement)

```bash
harness optimize propose       # list candidate harness mutations
harness optimize apply --canary <id>   # mandatory canary + falsifier_test + rollback_cmd
harness optimize rollback <id>          # if canary regresses
```

---

## 10. Long-term memory (Layer 2 — Statefulness)

```bash
harness memory list
harness memory promote <id> --global   # promote per-project entry to cross-project
harness recall search "fastapi healthz"
```

BM25 backend by default. Optional embeddings:

```bash
export HARNESS_RECALL_EMBEDDINGS=1
harness recall search "endpoint added today"
```

---

## 11. Hooks + MCP + Skills

```bash
harness hookpkg install pre-commit
harness hookpkg install gitleaks
harness mcppkg install filesystem
harness mcppkg install sqlite
harness skill apply python-test-writer
```

After hook install:

```bash
git commit -am "wip: try invalid commit"   # rejected by conventional-commits hook
```

---

## 12. Flows (Layer 3 — Composability)

End-to-end one-shot via flow:

```bash
harness flow list
harness flow show python-fastapi
harness flow init python-fastapi          # dry-run plan
harness flow init python-fastapi --yes    # execute
```

Phases: deterministic (scaffold) → llm (first endpoint) → sensor
(secrets_scan + pytest). Gates block downstream on red.

Non-software flow:

```bash
harness flow show meta-ads-campaign
harness flow init content-pipeline --yes
```

---

## 13. Workflow + stack visibility

```bash
harness workflow run smoke           # runs every workflow under .harness/workflows
harness stack show                   # active runs + sessions + last 5 outcomes
harness session list
harness session show <id>
```

---

## 14. Containers + runtime (optional, needs Docker)

```bash
harness containers list
harness runtime status
harness images list
```

---

## 15. Secrets

```bash
harness secret set API_KEY=sk-test
harness secret get API_KEY
harness secret list                       # encrypted under .harness/secrets/
```

Inject into a `do` run:

```bash
harness do "call the upstream API" --secret API_KEY
```

---

## 16. Backups (portable run-state)

```bash
harness backup save                       # tarball under .harness/backups/
harness backup list
harness backup restore <id>               # round-trip a state for another machine
```

---

## 17. Audit replay (Replayability principle)

```bash
harness audit tail --limit 50
harness audit replay <run-id>             # reconstruct entire run from events.jsonl
```

---

## 18. Dashboard (local React UI)

```bash
harness dashboard                         # boots Vite preview at :5173
open http://localhost:5173
```

Pages: Home, Projects, Run, Plan, Design, Roadmap, Agents,
Capabilities, Sensors, Context, Memory, Resources, Reports,
Stakeholder, Settings, Onboarding, Command, Catalog, plus 11
operational (Backup, Cleanup, Containers, Images, Install, Runtime,
Secrets, Sessions, ActiveRun, RunDetail, SessionDetail).

Run smoke E2E:

```bash
cd web/dashboard && npm install && npm run test:e2e
```

---

## 19. Performance + memory profiling

```bash
make bench                                # go test -bench across ./internal/...
make profile-mem                          # heap pprof → dist/profiles/mem.pprof
make profile-cpu                          # cpu pprof → dist/profiles/cpu.pprof
go tool pprof -http=:8081 dist/profiles/mem.pprof
```

Baseline budgets: [docs/benchmarks.md](benchmarks.md).

---

## 20. SOLID + god-file audit

```bash
make audit-solid                          # scan: file LOC > 400 or imports > 15
harness audit-solid --root .              # same via CLI
```

---

## 21. Update, uninstall, version

```bash
harness update                            # self-update via release channel
harness version                           # version + commit + date
harness uninstall                         # remove harness + state (asks confirm)
```

---

## Demo cheat-sheet (one-pager)

```bash
mkdir demo && cd demo && git init
harness init && harness install-git-hooks
harness flow init python-fastapi --yes
harness do "add /healthz with pytest" --autonomy ask
harness sensor run secrets_scan
harness sensor run tests
harness loop &                            # watch + rerun
harness metrics show
harness audit tail --limit 20
harness backup save
harness dashboard                          # open http://localhost:5173
```

Done. You exercised every layer + every adapter pattern + every
sensor + every persistence surface.

---

## Troubleshooting

| Symptom | Fix |
|---|---|
| `harness doctor` red | Install missing dep (Go, Python, Docker, git) |
| `harness do` denied | Loosen `--autonomy ask` instead of `strict`, or `harness autonomy show` |
| `pre-push` blocks push | Run `make lint && go test ./... && make coverage-gate` locally; or `HARNESS_SKIP_CI=1 git push` (escape hatch) |
| Dashboard 404 | `cd web/dashboard && npm install && npm run build && harness dashboard` |
| Brew formula stale | `brew update && brew upgrade harness` |

Open issues at https://github.com/rodolfopeixoto/harnessx/issues.
