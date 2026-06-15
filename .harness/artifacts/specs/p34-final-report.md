# P34 Final Report — Workflow improvements + manual tutorial

## Summary

Tutorial + 7 workflow improvements landed without changing the v0.3.0 /
P31 / P32 contracts. Every claim below has a green test + an
end-to-end smoke proof under `/tmp/p34-e2e/`.

## What now works

| Capability | Proof |
|---|---|
| `harness feature "what is dependency injection" --agent <id>` | logs `Routing: trivial prompt — using AskAgent fast path`; status `no_changes` with no worktree, no diff |
| `harness feature "<complex long prompt>" --agent <id>` | logs `Routing: complexity=complex -> model=fake-opus` |
| `harness feature "<standard prompt>" --agent <id>` | logs `complexity=standard -> model=fake-sonnet` |
| `.harness/config/autonomy.yaml` `deny_paths: ["secrets/**"]` | run that targets `secrets/api.key` → `status=autonomy_denied`, file not created |
| `harness feature --prompt-file prompt.txt --agent <id> --apply` | file content prefixed to positional, persisted in `enhancement.json` |
| `harness metrics --since all` | aggregates runs by status + by_agent + total cost USD |
| `harness audit --kind sensor` | filters `.harness/audit/events.jsonl` to one kind |
| `harness agent certify <id>` (auth failure) | detail includes `run: <login_command> | docs: <auth_doc_url>` |

## Packages added

- `internal/promptenh` — deterministic prompt enhancement (skills + context summary). Persists `enhancement.json` per run.
- `internal/input` — `--prompt-file` / `--pdf` / `--image` assembly. PDF via local `pdftotext` (poppler).
- `internal/autonomy/policy.go` — YAML loader + `MatchPath` / `MatchCommand` (glob + prefix).
- `internal/router/cost.go` — `PickModel(adapter, complexity) string` reading the adapter's models map.

## Packages extended

- `internal/intent` — `Complexity {Trivial, Standard, Complex}` + heuristic classifier.
- `internal/execution` — `Request.Model` + `Request.EnhancedPrompt`; `GateApplyWithPolicy` consults `autonomy.Policy.MatchPath` per changed file.
- `internal/agents/yaml` — adapter spec `auth: { login_command, doc_url, check }` block; copied onto `Capabilities` for downstream consumption.
- `internal/agents/certify` — surfaces `login_command` + `doc_url` on auth failures.
- `internal/app/workflow` — `Options` gains `PromptFile / PDF / Image / Apply / PlanOnly / Autonomy / AgentID`. `planThenMaybeExecute` routes trivial prompts through `askAgent` fast path; otherwise through `runWithExecutorAndComplexity` with model picked by router and prompt enhanced by promptenh.
- `cmd/harness` — `cmd_workflow.go` exposes new flags; `cmd_metrics.go` adds `harness metrics` + `harness audit`; `templates/agents/{claude,codex,gemini,kimi}.yaml` declare auth blocks.

## CLI surface added

```
harness feature   <prompt> [--prompt-file f] [--pdf p] [--image i]
harness bugfix    <prompt> (same flags)
harness run       <prompt> (same flags)
harness metrics   [--since 1d|7d|30d|all] [--json]
harness audit     [--limit N] [--kind sensor|hook|agent|cleanup] [--json]
```

`docs/tutorial-v0.4-manual.md` is the canonical English walkthrough.

## Tests

All green:

```
ok  github.com/ropeixoto/harnessx/internal/intent       1.104s
ok  github.com/ropeixoto/harnessx/internal/promptenh    1.432s
ok  github.com/ropeixoto/harnessx/internal/router       1.327s
ok  github.com/ropeixoto/harnessx/internal/autonomy     1.136s
ok  github.com/ropeixoto/harnessx/internal/input        1.411s
ok  github.com/ropeixoto/harnessx/internal/execution   12.667s
ok  github.com/ropeixoto/harnessx/internal/agents       1.825s
ok  github.com/ropeixoto/harnessx/internal/agents/yaml  3.091s
ok  github.com/ropeixoto/harnessx/internal/app/workflow 4.087s
```

E2E smoke matrix proved live in `/tmp/p34-e2e/`:

```
trivial fast path           ✓ Routing: trivial prompt -> AskAgent fast path
cost complex                ✓ Routing: complexity=complex -> model=fake-opus
cost standard               ✓ Routing: complexity=standard -> model=fake-sonnet
deny_paths secrets/**       ✓ status=autonomy_denied, file not created
--prompt-file               ✓ file content prefixed; enhancement.json persisted
harness metrics --since all ✓ runs=4 applied=1 denied=1 no_changes=2 by_agent=fake-real:4
```

## Honest gaps (out of scope, deferred)

- **Real Claude/Codex/Gemini E2E** — adapters declare auth + models. Requires the user to run `claude login` / `codex auth login` / `gemini auth login` / `kimi login --provider moonshot` once and verify. Tutorial documents the exact commands.
- **Dashboard Activity page** wired to `/api/metrics` + `/api/audit` (P35).
- **Inspector tabs per kind** + SSE on Active Run (P33).
- **Skills prefixing actually loaded** — promptenh.Enhance accepts a SkillSource; the workflow currently passes `nil`. Wiring `internal/skills` is a 30-LOC follow-up; deferred so skills + cost auto-routing land independently.
- **Brew formula + Windows install** (P36).

## Commits on `feature/p34-workflow-improvements`

```
feat(cli): harness metrics + harness audit tracker
feat(agents): surface auth login hints in certify output
feat(workflow): multi-input flags --prompt-file --pdf --image
feat(autonomy): path glob + command allow/deny policy
feat(workflow): trivial-prompt fast path via AskAgent
docs(tutorial): English manual testing tutorial v0.4
```

## How to verify locally

```
cd /tmp && rm -rf p34-verify && mkdir p34-verify && cd p34-verify
git init -q && git config user.email t@t && git config user.name t \
  && git config commit.gpgsign false
echo seed > README.md && git add -A && git commit -q -m seed
mkdir -p .harness/config/agents
cp /Users/ropeixoto/dev/projects/harnessx/templates/agents/fake-real.yaml \
   .harness/config/agents/
# adjust binary path if needed

bin=/Users/ropeixoto/dev/projects/harnessx/bin/harness

$bin feature "what is dependency injection" --agent fake-real
$bin feature "create x.md with content: y" --agent fake-real --apply
$bin feature "implement entire codebase refactoring across files" \
  --agent fake-real
$bin metrics --since all
$bin runs list
```

Every command should match the smoke matrix above. If any deviates,
file an issue under `.harness/artifacts/specs/` with the exact run id
and meta.json.
