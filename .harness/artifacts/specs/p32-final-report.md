# P32 Final Report — MCP + Hook real integration

## Summary

P31 loop now consumes the project's MCP discovery and hook scanners. The
adapter receives a merged MCP config when it advertises `mcp: true`, and
hook scripts under `.harness/hooks/{pre,post}-tool-use.*` actually fire
around the agent invocation, blocking the run unless autonomy =
`full_project_loop`.

## What now works (proven against /tmp/p31-e2e fixture)

| Capability | Proof |
|---|---|
| `harness mcp install <name> --command <bin> --yes` | writes `.harness/mcp/<name>.json` |
| `harness mcp scan` | shows installed server |
| Executor builds `mcp-config.json` per run when adapter mcp=true | `runs/<id>/mcp-config.json` written, `mcpServers: {name: {command, transport}}` shape |
| `--mcp-config <path>` appended to adapter args | via `AgentRequest.ExtraArgs` (no spec template change) |
| `pre-tool-use` hook fires before agent | run report Hooks table shows row + duration |
| `post-tool-use` hook fires after sensors | run report Hooks table shows both rows |
| Failing pre-hook blocks under `safe_execute` | status `autonomy_denied`, file NOT created |
| `full_project_loop` bypasses hook block | status `applied`, file created |

## Wiring

- `internal/agents/types.go::AgentRequest.ExtraArgs` — new field for
  runtime arg injection without templating the YAML spec.
- `internal/agents/yaml/adapter.go::Run` — appends `req.ExtraArgs` after
  template substitution.
- `internal/execution/mcp.go::BuildMCPConfig` — reads `mcpscan.Scan`,
  merges into a single Claude/Codex/Gemini-shaped envelope, writes to
  `runs/<id>/mcp-config.json`, returns the path.
- `internal/execution/hooks.go::DispatchHooks` — runs every hook whose
  `event` field matches the lifecycle phase, captures exit code +
  duration, surfaces via `Result.Hooks`.
- `internal/execution/executor.go::Execute` —
  - injects MCP config when adapter capability `mcp = true`
  - dispatches pre-tool-use hooks before adapter.Run
  - dispatches post-tool-use hooks after sensors
  - non-zero pre-hook denies the run unless autonomy is
    full_project_loop

## CLI surface added

```
harness mcp install <name> [--command bin] [--url addr]
                            [--transport stdio|http] [--yes]
```

`harness feature/bugfix/run` already pick this up because Executor reads
the adapter's capability map. Set `capabilities.mcp: true` in the
agent's YAML spec and the next run gets the injected config + hook
dispatch automatically.

## Real verified scenarios

```bash
# install + scan
harness mcp install filesystem --command npx --yes
harness mcp scan
# SOURCE NAME TRANSPORT RISK PATH
# harness root stdio low .harness/mcp/filesystem.json

# successful pre+post hook + injected MCP
harness feature "create test.md with content: x" --agent fake-real --apply
# report shows: Injected 1 MCP server(s) via .../mcp-config.json
# report shows: Hooks table with pre-tool-use Exit=0

# failing pre-hook blocks safe_execute
echo 'exit 1' >> .harness/hooks/pre-tool-use.sh
harness feature "..." --agent fake-real --apply --autonomy safe_execute
# → status=autonomy_denied, file not created

# full_project_loop bypasses
harness feature "..." --agent fake-real --apply --autonomy full_project_loop
# → status=applied, file created
```

## Limits + honest gaps remaining for v1.0

1. **Adapter spec opt-in** — claude.yaml / codex.yaml / gemini.yaml /
   kimi.yaml still have `capabilities.mcp` from the existing handoff
   (Claude=true, Codex=true, Gemini=true). Flipping them to receive the
   real config requires no code change but does require the user to
   confirm those CLIs accept `--mcp-config <path>` (Claude does;
   Codex/Gemini exact flag may differ — adapter spec can override
   `ExtraArgs` shape via a future `mcp.injection_template` field).
2. **Hook scoping** — every matching hook fires for every run. No
   per-mode / per-agent filtering yet. Add `events: [feature, bugfix]`
   to the hook manifest when needed.
3. **Skill loading** — `internal/skills` exists; Executor doesn't yet
   prefix prompts with skill snippets routed by intent. Trivial to add
   when router output exposes the selected skill list.
4. **Real Claude end-to-end** — `harness execute "..." --agent claude
   --apply --autonomy safe_execute` runs against the real Claude CLI
   but the diff/`--print` output parsing may need adapter spec tuning
   (`claude.yaml::failure_detection` already covers auth/timeout). To
   prove a real Claude run end-to-end the user needs an authenticated
   Claude session (`claude login`) and a small fixture prompt.

## P33 deferred work (Inspector + SSE + memory promote UI)

Not delivered this iteration:

- Inspector tabs per kind (Sensor 7 / Agent 7 / MCP 7 / Hook 6 /
  Memory 5 / Context 7) — React work, 3-4 days. Concrete files:
  `web/dashboard/src/ds/InspectorPanel.tsx` extension + per-kind tab
  builders under `web/dashboard/src/inspector/`.
- ActionService backend persistence — replace
  `web/dashboard/src/lib/actions.ts` localStorage with `POST
  /api/audit-events` + read from `internal/audit` FileSink.
- SSE on `/api/events?run_id=<id>` for Active Run page — tail
  `events.jsonl` per run; reuse `internal/audit` writer.
- Memory promote/retire UI — backend already at
  `internal/memory.Repo.Promote`; surface buttons on Memory page.

These are all bounded UI work. Each is one PR. None blocks the agentic
loop being usable today via CLI.

## Status today

| Surface | Status |
|---|---|
| `harness feature/bugfix/run --agent <id> [--apply] [--autonomy ...]` | real loop |
| MCP discovery + install + injection into adapter | real |
| Hook pre/post dispatch + autonomy gate | real |
| Worktree + diff + sensors + autonomy gate + apply/approve | real |
| `harness runs list/inspect/report/sensors/approve/discard` | real |
| Dashboard rich pages + audit | green from P29 |
| Inspector tabs / SSE / memory promote UI | **P33 deferred** |
| Real Claude/Codex/Gemini end-to-end | adapter spec ready, needs user CLI auth confirmation |

## Branches

- `feature/p31-real-agentic-execution` — P31 + P32 commits

Commits added this turn:

- `feat(workflow): wire feature/bugfix onto execution.DefaultExecutor`
- `feat(execution): MCP config injection + hook pre/post dispatch + harness mcp install`
- `docs(execution): P32 final report`
