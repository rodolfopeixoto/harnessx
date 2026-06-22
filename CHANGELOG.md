# Changelog

Format: [phase] short summary, then bullet list of concrete additions.
Newest milestones at the top. Dates are when the milestone landed in repo.

## 2026-06-21 — v0.144.0 — Wave 18: multi-stack Todoist tutorial (F109)

### Docs

- New `docs/TUTORIAL-MULTI-STACK.md` ships the **same Todoist
  domain** (JWT auth, lists, tasks, tags, recurring rules) across
  every backend stack HarnessX scaffolds today — `python-ecommerce`
  (FastAPI), `rails` (Rails 7 API), `go` (net/http), `rust` (Axum),
  `ruby` (Sinatra) — plus a `react` frontend that talks to all five
  via `VITE_API_BASE`. One shared backlog at
  `/tmp/todoist-features.md` is fed to `harness drive --features
  --continue-on-fail` per repo so the same prompts produce the
  same surface across stacks. Ends with a cross-stack
  `/tmp/todoist-smoke.sh` that hits the contract on every base
  URL and a jq one-liner that aggregates per-stack spend from
  `.harness/sessions/*.jsonl`.

## 2026-06-21 — v0.143.0 — Wave 17: inline / popup + drive --features + onboarding + VCR adapter (F108)

### New

- **Inline `/` popup in chat REPL** via chzyer/readline's
  `Listener` interface. `internal/repl/slash_popup.go` watches
  every key, matches the typed slash prefix against the static
  command list, and renders up to 6 candidates under the prompt
  (TAB still completes via the existing PrefixCompleter). Cursor
  is preserved with save/restore escapes, the strip clears as
  soon as the buffer stops looking like a slash. 8 unit tests
  cover prefix match, exact-match suppression, args-tail
  handling, render+clear, idempotent clear, nil-writer safety,
  and the row cap.
- **`harness drive --features <file.md>`** chains a backlog: one
  prompt per non-empty/non-comment line, leading `- ` or `* `
  bullet stripped. `--continue-on-fail` keeps going after a
  failed feature and surfaces the count at the end. Header for
  every feature: `=== feature i/N ===`. 4 unit tests cover file
  parsing + abort-by-default + continue-on-fail.
- **`harness onboarding`** detects the seven system tools (git,
  python3, node, uv, rg, jq, rclone) and the five bundled agent
  CLIs (claude, codex, gemini, kimi, ollama), prints version
  lines for what's installed and per-tool install hints for
  what's not. Picks a suggested adapter (preferring registered
  bundled ones in claude/codex/kimi/gemini/ollama order),
  prints a 5-step next-actions list (`harness use`, `harness
  new`, `harness chat --auto-gate`, `/drive`, tutorial pointer).
  6 unit tests cover the helpers + render shape.
- **`internal/agents/vcr`** wraps any AgentAdapter to record real
  CLI output on first call (fingerprint = SHA-1 of adapter id +
  model + working-dir basename + prompt) and replay it on every
  subsequent call. Three modes: `ModeAuto` (default — replay if
  cassette exists else record), `ModeReplay` (error on missing
  cassette), `ModeRecord` (always re-record). Lets E2E tests
  exercise the real Claude JSON-Lines / Codex exec stream
  without burning tokens. 9 unit tests cover record→replay,
  replay-missing-error, force-rerecord, fingerprint stability +
  invalidation, save/load round-trip, error restoration, and
  delegation of ParseUsage/ClassifyFailure.

## 2026-06-21 — v0.142.0 — Wave 16: Windows build fix (F107)

### Fixed

- **Windows release artifacts came out 0 bytes** since v0.140
  because `internal/repl/prompt.go` referenced `syscall.SIGWINCH`,
  which only exists on Unix. The SIGWINCH watcher is now split
  across `prompt_unix.go` (build tag `!windows`) and a no-op
  `prompt_windows.go`. `make release` builds clean for every
  platform again; the GH release uploads include the windows
  amd64 + arm64 binaries.

## 2026-06-21 — v0.141.0 — Wave 16: Todoist tutorial end-to-end (F106)

### Docs

- New `docs/TUTORIAL-TODOIST.md`: ~600-line walk that scaffolds two
  projects (`todoist-api`, `todoist-web`), drives five backend
  features (JWT auth, list CRUD, tasks with due/priority, tags
  many-to-many, recurring rules) and three frontend features (lists
  + tasks UI, task row, auth gate) entirely through `harness drive`.
  Documents the heredoc + bracketed-paste flow for the multi-line
  prompts every feature uses, shows the expected per-step output,
  ends with a backup snapshot and the cheat sheet covering every
  slash + readline behaviour shipped in waves 11–15.
- Tutorial doubles as the manual acceptance walk for v0.140's real
  cheap-chain test-emit: every `/drive` block ships through the
  cheap_review → red bar → implementation → ci → commit pipeline.

## 2026-06-21 — v0.140.0 — Wave 16: real cheap-chain test-emit + SIGWINCH refresh (F105)

### Changed

- **`harness drive` test-emit now calls the cheap_review router
  chain for real**. The old placeholder failing test was a stopgap
  that left the user wondering whether the LLM ever ran. The new
  path renders a deterministic prompt template asking the cheap
  chain to emit a pytest module covering happy + error paths,
  extracts the python body from the response code fence, validates
  it contains a `def test_…`, and writes it to disk. Falls back to
  the placeholder when no adapter is registered, no chain is
  available, the adapter call fails, or the response has no
  python block — smoke runs untouched, real walks now produce
  real failing tests.
- **chzyer/readline SIGWINCH handler**. The chat REPL now listens
  for terminal resizes and calls `rl.Refresh()` so the prompt and
  history line-wrap correctly when the window changes width.
  Hooked + cleaned up in `readlinePromptReader.startWinch` and
  `Close`.

### Tests

- `TestExtractPythonBodyFromCodeFence` covers \`\`\`python /
  \`\`\`py / bare \`\`\` fences.
- `TestExtractPythonBodyFallsThroughOnBareTest`,
  `TestExtractPythonBodyEmptyWhenNoTest`,
  `TestRenderTestEmitPromptIncludesFeatureSlug`.

## 2026-06-21 — v0.139.0 — Wave 15: slash typo suggest + session recap + adapter quick-pick (F104)

### New

- **Unknown slash → did-you-mean suggestion**. Typing `/drv` no
  longer silently routes to the implementation chain; instead the
  REPL prints `⚠ unknown slash "/drv" — did you mean /drive?`.
  Score combines Levenshtein distance with a common-prefix bonus
  (`dist*2 - prefix`) so `/drv` reliably resolves to `/drive`
  rather than the cheaper-looking `/do`. Falls back to a generic
  `try / for the menu` line when nothing scores well enough.
- **Session recap on `/exit`**. Before printing `bye` the REPL
  prints a 5-line summary: id, label, goal, turn count + chat
  turn count, cumulative tokens, total cost. Lets users see what
  they just spent before the terminal scrolls away. The cost
  cell is styled in success colour.
- **Adapter quick-pick on first-time chat**. When no pin exists
  the chat prelude lists every registered adapter id with an
  index and reminds the user how to make a pin permanent
  (`harness use <id>`), then proceeds with the first-available
  auto-pick as before. No more "what happened, why claude?"

### Tests

- `TestSuggestSlashPicksClosestMatch`,
  `TestSuggestSlashEmptyDoesNothing`,
  `TestSuggestSlashFarAttemptFallsBack`,
  `TestFirstToken`,
  `TestSummariseSessionPrintsRecap`,
  `TestSummariseSessionEmpty`,
  `TestRunUnknownSlashSuggests`,
  `TestRunExitPrintsRecap`.

## 2026-06-21 — v0.138.0 — Wave 14: bracketed paste + ci --install-missing (F103)

### New

- **Bracketed paste mode in the chat REPL TTY**. On enter the REPL
  emits `\x1b[?2004h`, on leave `\x1b[?2004l`. Stdin is wrapped in
  `pasteCoalescingReader` that detects the `\x1b[200~ … \x1b[201~`
  markers and replaces embedded newlines with `\\\n` so the existing
  backslash-continuation logic glues them into one prompt. No more
  "pasted three lines, only the first ran" — the whole clipboard
  arrives as a single input automatically (heredoc still works).
- **`harness ci --install-missing`** auto-installs the optional
  python dev tools (`bandit`, `mypy`, `pip-audit`) into the
  project's `.venv` whenever a sensor reports
  `binary not on PATH`. Uses `uv pip install --python .venv/bin/python`
  first, falls back to `.venv/bin/python -m pip install`. Without the
  flag the gate prints a hint: `hint: 3 optional python tool(s)
  missing (bandit, mypy, pip-audit) — rerun with
  \`harness ci --install-missing\` to fix`.

### Tests

- `paste_test.go` — passthrough, in-paste newline escape, marker
  split across reads, `isPartialPrefix` helper, `Close`.

## 2026-06-21 — v0.137.0 — Wave 13: """ heredoc + adapter billing-mode disclosure (F102)

### New

- **`"""` heredoc multi-line input** in both the bufio and chzyer
  prompters. Type `"""` at the start of a line to enter heredoc
  mode, paste/type as many newlines as needed, close with `"""`
  on its own line. Fixes the v0.136 walk's worst gotcha: pasting
  a two-line `/drive` prompt only sent the first line because
  terminals split clipboard newlines into separate Enters.
- **Adapter billing-mode shown in the chat header**:
  `[agent] calling claude (implementation, oneshot · API-billed)…`
  vs `[agent] calling kimi (cheap_review, interactive · plan/local)…`.
  Lets users see whether the call charges their API key or
  drains their plan/subscription quota before they hit Enter.
  Helper: `adapterBillingMode(id)` table covering claude/codex/
  gemini/anthropic-api/openai-api/gemini-api/moonshot-api/
  minimax-api (oneshot), claude-interactive/kimi/ollama
  (interactive), fake (free).
- **Greet line** announces the multi-line escapes so users
  discover them without reading the tutorial first.

### Tests

- `TestBufioPromptReaderHandlesTripleQuoteHeredoc` — heredoc
  payload captured, marker not leaked.
- `TestAdapterBillingModeKnownIDs` — table covers every bundled
  adapter id plus an unknown fallback.

### Docs

- `docs/TUTORIAL-ECOMMERCE.md` gained "Pasting multi-line prompts"
  and "Adapter billing mode" sections with concrete examples.

## 2026-06-21 — v0.136.0 — Wave 12: REPL polish, slash menu, colored drive output (F101)

### New

- **Type `/` (or `/?`) at the prompt to open the grouped slash
  menu** — chat / gate / agents+cost / memory / session / exit.
  Saves users from memorising the 30 slashes; the existing
  `/help` long-form still works for anyone scripting against it.
- **Chat header + drive output are colourised** via lipgloss
  styles already shipped in `internal/ui`. `[agent] calling X
  (task)…` and `✓ done · in=… out=… ~$…` use accent + info +
  muted styles; `drive 1/5 spec — harness feature` and the
  `committed feat:` line use heading + success. `--plain` /
  `HARNESS_PLAIN=1` still suppresses every ANSI escape.

### Tests

- `TestRunSlashAloneShowsMenu`, `TestPrintSlashMenuQuestionMark`,
  `TestPadRightFillsToWidth` — slash menu happy path + helper.

## 2026-06-21 — v0.135.0 — Wave 12: drive no-op abort, per-model cost, py dev tools (F100)

### Fixed

- **`harness drive` now aborts when the agent produces zero file
  changes** instead of looping the same prompt three times against
  a confused model. Implemented with a pre/post `git status
  --porcelain` snapshot in `driveImplLoop`; the error explicitly
  asks the user to refine the prompt. Caught the v0.133 walk's
  worst spend (3 wasted opus attempts on a truncated "and").
- **Python scaffolds ship the optional security/typecheck tools**
  (`bandit`, `mypy`, `pip-audit`) in `requirements.txt`. The
  `--with-deps` install now wires every dev-loop sensor green on a
  fresh `harness new python(-ecommerce)`.

### Changed

- **`/cost` is now a per-adapter table**. `Turn` gained `AdapterID`
  and `TaskTag` columns (omitempty, backwards compatible), `printCost`
  delegates to `aggregateCost` + `renderCostReport`. Output:
  ```
  session 01KX…: 4 chat turns
    ADAPTER      TASK             TURNS       IN      OUT       COST
    claude       implementation       2     4310      812 $   0.1929
    gemini       cheap_review         1      150       80 $   0.0019
    ───────────────────────────────────────────────────────────────
    TOTAL                             3     4460      892 $   0.1948
  ```
- Three new unit tests cover `aggregateCost` grouping + ordering,
  the rendered table, and the empty-session case.

## 2026-06-21 — v0.134.0 — Wave 11 cleanup: SOLID + constants + coverage (F99)

### Changed

- **Wave 11 code reaudited against CLAUDE.md house rules.** `cmd_drive`,
  `stream_json`, and `prompt` no longer carry WHAT-comments — only the
  rare WHY remains where a future reader would otherwise be surprised
  by intent. Files reorganised for single-responsibility: `runDrive`
  delegates to `driveSpec`, `driveTestEmit`, `driveExpectRedTests`,
  `driveImplLoop`, `driveCommit`; `stream_json` splits `handleEvent` /
  `renderAssistantBlock` / `renderToolUse` / `emitIfPath`; `prompt`
  factors `readBackslashContinuation` and `toPcItems` so both
  prompters share one continuation contract.
- **Magic strings + ints moved to `internal/platform/constants`** under
  new `Chat REPL UI` and `Drive command` sections (gutter prefix,
  thinking frame ms, history suffix, chat result/bash/grep/glob/raw
  caps, recap byte cap, label max, test-file prefix/suffix, commit
  type, slug max, default autonomy, task tags). Files reference the
  shared names; future drift gets caught at the constants file.

### New tests (raise coverage on wave 11 surface)

- `internal/agents/yaml/stream_json_test.go` — tool_use branches
  (Edit/Grep/Glob/unknown), `Flush` partial line, `user` and
  `rate_limit_event` swallow, `system.init` dedup, `jsonString`
  nil/garbage/missing/wrong-type paths, `jsonFormat` aliases,
  `truncForChat` boundary.
- `internal/repl/prompt_test.go` — bufio reader single-line,
  backslash continuation, EOF, `Close`, non-TTY fallback,
  `isTerminal` negative, `chatCompleter` tree, `toPcItems` order,
  `asReadCloser` plain reader.
- `cmd/harness/cmd_drive_test.go` — `sanitisePyIdent` table,
  `truncSubject` edges, `conventionalDriveSubject` length budget,
  `renderPlaceholderTest` content, `writePlaceholderTest` on-disk,
  `runHarnessChild` exit-code propagation, `runGitInDir` error,
  `driveCommit` real git, `runDrive` spec failure + already-green
  short-circuit, `driveExpectRedTests` both branches, `driveImplLoop`
  green + exhaustion, `driveTestEmit` no-agents fallback,
  `newDriveCmd` flag registration.

Coverage on the wave 11 files now reads:

| file                                     | coverage |
|-----------------------------------------|---------:|
| `internal/agents/yaml/stream_json.go`   | 96.6 %  |
| `internal/repl/prompt.go` (bufio path)  | 88.2 %  |
| `cmd/harness/cmd_drive.go` (business)   | 93.1 %  |

The remaining gaps are the readline TTY branch (needs a pty harness)
and the cobra `RunE` closure (covered by tutorial-smoke end-to-end).

## 2026-06-21 — v0.133.0 — Wave 11 tutorial rewrite around `harness drive` (F98)

### Changed

- **`docs/TUTORIAL-ECOMMERCE.md` is now a real end-to-end walk**.
  Three `/drive` features against the FastAPI scaffold (product
  stock, cart tax, checkout finalisation) plus a fourth `/drive`
  feature against a brand-new React scaffold so the same loop is
  proved across stacks. Each step lists the files touched, the
  router decision (cheap vs implementation chain), and the
  expected on-screen output (`│ ● Read app/storage.py`,
  `drive: 5/5 — harness ci`, `drive: ✓ committed feat: …`). Cheat
  sheet covers every slash up to `/drive` + readline ↑/↓ + TAB.

## 2026-06-21 — v0.132.0 — Wave 11 spec-driven, test-first `harness drive` (F97)

### New

- **`harness drive "<prompt>"`** (paper §3.4 PEV + §3.4.2 plan-as-
  contract) chains the deterministic feature spec, a placeholder
  pytest emitter routed through the cheap_review router chain, the
  test-first red bar assertion, the implementation chain via
  `harness do`, and the gate. On green it commits with a
  `feat: <prompt>` conventional subject. The expensive model only
  sees the spec + the test-shaped hole to fill, so per-feature
  token spend stays bounded.
- **`/drive <prompt>` slash command** in the chat REPL invokes the
  same flow. `chat list` / `--replay` already track it via the
  Action label and the mutating-input refuse list.

### Notes

The test-emit step today writes a placeholder failing test so the
loop runs deterministically without an LLM bill (verified by the
existing tutorial-smoke script and the new fake-adapter path).
Future patches will swap in the cheap chain's actual emission once
the cheap adapter prompt template is tuned.

## 2026-06-21 — v0.131.0 — Wave 11 humanised JSON stream + multi-agent routing (F96)

### Fixed

- **Chat REPL dumped raw Claude JSON-Lines** (tools[], mcp_servers[],
  slash_commands[], thinking blocks, tool_use chains, …) into every
  turn. New `internal/agents/yaml/stream_json.go` wraps the live
  writer for JSON-output adapters and emits one humanised line per
  significant event:
    `• session ready` · `⋯ thinking…` · `● Read app/storage.py` ·
    `● Write tests/test_products.py` · `$ pytest -q` ·
    `» Done. 9/9 pass.` · `✓ Done. 9/9 pass.`
  The huge `system.init` envelope is swallowed; tool paths trim to
  their last two segments; unknown shapes fall back to a truncated
  raw-line passthrough so the user never sees nothing. Four unit
  tests cover the happy path, dedup, garbage fallback, and path
  shortening.

### New

- **Multi-agent routing inside chat** (paper §3.5.3 governed
  mutation). `repl.Options.Route(task)` plugs the existing
  `internal/router.Router` (with `router.Defaults`) into every chat
  turn:
    - plain text → `implementation` chain (codex → claude → kimi …)
    - `/recap` → `cheap_review` (gemini → kimi → codex …) so summaries
      do not burn opus tokens.
  The header line now prints `[agent] calling <id> (<task>)…` so the
  user sees which model is doing what. Falls back to the pinned
  adapter when the route resolves to nothing registered. cmd_chat
  wires the registry into the router; existing `/use <id>` override
  still wins over routing for the current turn.

## 2026-06-21 — v0.130.0 — Wave 11 polish: ship --yes, ruff-format scaffold, readline history + TAB (F95)

### Fixed

- **`/ship "..."` from inside chat failed** with `Error: unknown flag:
  --yes` because the REPL appended `--yes` (mirroring `harness do
  --yes`) but `cmd_ship.go` never registered the flag. `--yes` is now
  accepted as a documented no-op for symmetry with the rest of the
  surface.
- **`harness ci` was red on a fresh `python-ecommerce` scaffold**
  because `app/storage.py` had a multi-line `sum(...)` that the
  current ruff (0.15) wants collapsed onto one line. Template
  reformatted. New `make scaffold-fmt` target re-runs `ruff format`
  against every bundled python scaffold so regressions stay out.

### New

- **Chat REPL gets readline ergonomics**. New `internal/repl/prompt.go`
  picks `github.com/chzyer/readline` whenever stdin+stdout are both
  TTYs, falling back to the existing `bufio.NewReader` for piped
  input so `scripts/tutorial-smoke.sh` keeps working unmodified.
  History persists per session at `.harness/sessions/<id>.history`
  (up/down navigates). TAB completes the static slash list, the
  registered adapter ids on `/use`, and saved session labels on
  `/resume`. Ctrl-C clears the current entry instead of killing the
  REPL; Ctrl-D exits.

## 2026-06-19 — v0.129.0 — backup quickstart + tutorial backup section (F94)

### New

- **`harness backup quickstart`** prints a 3-step recipe for adding
  a remote, setting it as default, and snapshotting. When a default
  remote is already configured it shortcuts straight to `snapshot`.
- The `backup: no remote chosen` error from `harness backup
  snapshot` now references `harness backup quickstart` as the first
  fix step.
- Tutorial gained an "Back up the project (optional)" section that
  walks through the rclone-based snapshot flow with the same
  commands the quickstart prints.

## 2026-06-19 — v0.128.0 — Chat /timeline slash command (F93)

### New

- **`/timeline`** prints an at-a-glance ASCII view of every turn in
  the session: index, clock (local time), action label, truncated
  input, and per-turn USD cost when present, plus a cumulative cost
  footer. Resumed sessions render their full prior history.
- Tutorial cheat sheet lists the new slash.

## 2026-06-19 — v0.127.0 — Wave 10 bug pass: no-adapter refuse, update --force, typo suggest (F92)

### Fixed

- **`harness chat --no-adapter` no longer drops plain text into a
  3-minute deterministic-planner loop**. The flag now propagates as
  `Options.NoAdapter`; plain text turns are refused with
  `✗ no adapter wired (chat --no-adapter). use /exec <prompt> or pin
  one with 'harness use <id>'`. Slashes (`/exec`, `/history`, etc.)
  keep working so the no-LLM smoke flow is fast end-to-end.
- **`harness update --force` now also reinstalls the same tag**. Was
  scoped to "allow downgrade" before; the v0.116 walk surfaced the
  need to reinstall after a botched `cp` to `/usr/local/bin/`. Help
  text updated.

### New

- **`repl.SuggestSession`** returns the closest known label or ulid
  via Levenshtein distance (max 3). `harness chat --resume X` /
  `--replay X` now print
  `chat: resume "X": session not found (did you mean "Y"?)` when a
  typo is close to a real session. Four new tests cover the
  Levenshtein helper, the close-match path, and the far-match
  fallthrough.

## 2026-06-19 — v0.126.0 — session show chat fallback + ci --fast + tutorial smoke (F91)

### Changed

- **`harness session show <id|label>` now dispatches to the chat
  store first**. When the argument resolves to a
  `.harness/sessions/<id>.jsonl` (including via `/save` label) the
  command renders a chat-shaped view: label, goal, started, root,
  auto-gate, budget, turn counts, cumulative tokens + USD, and every
  turn input truncated to 80 chars. The sqlite-backed run view is
  used as the fallback so `harness ship`-style ids keep working.

### New

- **`harness ci --fast`** skips the slowest sensors (today:
  `secrets_scan`, which walks the repo with ripgrep). Implemented
  via a denylist on `sensorcmd.RunOptions.Fast` so the auto-gate
  flow inside `harness chat --auto-gate` stays snappy without
  losing the per-step `--fast` opt-out.
- **`scripts/tutorial-smoke.sh`** is a deterministic no-LLM walk of
  the documented chat flow. It boots a throwaway python-ecommerce
  project, asserts `/save` labels appear in `harness chat list`,
  exercises `/save-prompt` + `/prompt` + `/prompts`, verifies
  `harness session show <label>` returns the new chat view, and
  proves `harness ship` rejects a dirty tree without `--allow-dirty`.
  Wired as the new `make tutorial-smoke` Makefile target.

## 2026-06-19 — v0.125.0 — Chat prompt templates + tutorial finalisation (F90)

### New

- **`internal/prompttpl` package** stores reusable chat prompt
  templates under `.harness/prompts/<name>.md`. Public API: `Save`,
  `Load`, `List`, `ValidName`. Names are restricted to lowercase
  alphanumeric plus `_`/`-`, ≤40 chars, leading alnum, so they map
  safely onto file paths and CLI output. Six unit tests cover the
  validation, the round-trip, and the sort order.
- **`/save-prompt <name>`** captures the most recent plain-text
  session turn (skips slash + `!` shell turns) and persists it.
- **`/prompt <name>`** loads the template and feeds it back into
  `handleInput` so the agent receives it like a fresh user prompt.
- **`/prompts`** lists every saved template in the project.
- Mutating-input list extended so `--replay` refuses both new
  template commands.

### Changed

- Tutorial cheat sheet gained the three template commands and a
  worked "reusable prompt templates" example.
- Troubleshooting section clarifies that the OrbStack/Docker dialog
  from the upstream `claude` CLI is a one-time approval; falling
  back to `--no-adapter` is documented as the deterministic escape
  hatch.

## 2026-06-19 — v0.124.0 — Chat SIGINT cancel + positional id/label + /branch (F89)

### New

- **SIGINT mid-turn cancel**. Hitting Ctrl-C during an agent call or a
  `/ci`/`/test`/`/lint`/`/ship` step now cancels just that turn — the
  spinner clears, "✗ interrupted — back to prompt" is printed, and the
  REPL lands on a fresh prompt instead of dying. Implemented via a
  `signalAwareCtx` wrapper that re-arms `signal.Notify(os.Interrupt)`
  for every turn and releases the handler on the way out.
- **`harness chat <id|label>`** positional shortcut for
  `--resume <id|label>`. Same string passed to `--replay` works too.
  Resolution is label-first; an unknown input falls through to the
  raw `LoadSession` error so you still get a clean "session not
  found".
- **`/branch <name>`** runs `git checkout -B <name>` from the project
  root and labels the session with the slash-flattened branch name
  (so `feature/cart` → `feature-cart`) when no label was set yet.
  Mutating-input list updated so `--replay` keeps refusing it.

### Changed

- `repl.ResolveSessionID(root, arg)` is the helper everyone uses for
  the positional argument and the `--resume`/`--replay` flags. Label
  lookup walks `ListSessions`; ambiguity prefers the newest mtime.
- Tutorial cheat sheet lists `/branch`; the resume/replay snippet now
  shows the positional + `<label>` forms; the Ctrl-C behaviour is
  documented inline.

## 2026-06-19 — v0.123.0 — Chat /save /recap + --replay (F88)

### New

- **`/save <name>`** labels the current session with a human-readable
  alias so `harness chat list` reads
  `01KX… ecommerce-cart dev turns=12 last=fix 500 on /cart` instead
  of an opaque ulid. Refuses slashes, leading dots, tabs, and >80
  chars. Persists into a new `.harness/sessions/<id>.meta.json`
  sidecar alongside the JSONL turn log so older readers ignore it
  cleanly.
- **`/recap`** asks the pinned adapter for an ≤8-bullet summary of
  the session so far (built, ran, still open). When no adapter is
  wired the command falls back to a deterministic list of past
  inputs so the slash is never silent. Cost is recorded in the turn
  like any other chat invocation.
- **`harness chat --replay <id>`** opens a session in **read-only**
  mode: plain text + every mutating slash (`/exec`, `/do`, `/ship`,
  `/ci`, `/test`, `/lint`, `/use`, `/budget`, `/auto-gate`,
  `/clear`, `/save`, `/recap`, `/plan`, `/goal`, `/last`, `!shell`)
  is refused with a clear message, while inspection slashes
  (`/history`, `/agents`, `/cost`, `/diff`, `/help`) keep working.
  Mutually exclusive with `--resume`.

### Changed

- `Session` gained `Label` (omitempty JSON) and an unexported
  `ReadOnly` flag. `persist` now writes a `<id>.meta.json` sidecar
  alongside the JSONL so the new fields survive across resumes;
  `LoadSession` ignores a missing sidecar so older session files
  still load.
- `harness chat list` renders the new label column (padded to 20 cols
  with `—` placeholder when the session has no alias).
- `extra_test.go` filters the session dir by `*.jsonl` to ignore the
  new sidecar.

## 2026-06-19 — v0.122.0 — Chat /use /diff /budget + session export (F87)

### New

- **`/use <id>`** switches the pinned adapter mid-session via a new
  `Options.SwitchTo(id) (agents.AgentAdapter, string, error)`
  callback. `cmd_chat` wires it against the registry that was already
  loaded for `/agents`, so swapping between claude/codex/kimi/gemini
  is two keystrokes.
- **`/diff`** runs `git diff --stat HEAD && git diff HEAD` from the
  project root so the user can review what the last few agent turns
  actually wrote without leaving the REPL.
- **`/budget <usd>`** (and `/budget off`) sets a per-session USD cap
  on cumulative chat spend. The next plain-text chat turn is refused
  with a clear message once the cap is exceeded; toggles persist in
  `Session.BudgetUSD` and survive `--resume`.
- **`harness session export <id>`** emits a sharable JSON envelope
  for any `.harness/sessions/<id>.jsonl`: id, goal, started, root,
  turn counts, cumulative `in_tokens`/`out_tokens`/`cost_usd`, the
  three session toggles, and every Turn record. Useful for pasting
  into a code review, bug report, or backup.

### Changed

- `repl.handleInput` now takes `*Options` so `/use` can mutate the
  live Adapter without a registry import in the repl package. All
  other slashes still see a pointer-to-value with the same semantics.
- `Session.BudgetUSD` (omitempty) backs the new `/budget` toggle and
  is loaded by `LoadSession` so `--resume` continues honouring the
  cap. Older session files keep loading because the field is
  zero-valued when absent.
- Tutorial cheat sheet covers `/use`, `/diff`, `/budget`, and the new
  `session export` flow.

## 2026-06-19 — v0.121.0 — Chat /agents /cost /clear + --auto-gate (F86)

### New

- **`/agents` slash command** lists every adapter the registry knows
  and marks the one wired to this session, so you can switch
  contextually instead of leaving the REPL to run `harness agent list`.
- **`/cost` slash command** sums `InTokens`, `OutTokens`, and
  `EstimatedCostUSD` across every chat turn in the current session
  (resumed sessions included). Turn-level usage is now persisted in
  `.harness/sessions/<id>.jsonl` so the count survives a restart.
- **`/clear` slash command** rewinds the Working Memory preamble by
  setting `Session.ContextMark = len(Turns)`. The agent's next plain
  text turn starts fresh while `/history` still shows everything that
  came before.
- **`/auto-gate on|off`** toggles a per-session flag that re-runs
  `harness ci` after every plain-text chat turn. The starting value
  comes from the new `harness chat --auto-gate` flag and persists into
  resumed sessions via `Session.AutoGate`.

### Changed

- `Session` and `Turn` JSON shapes gained `auto_gate`, `context_mark`,
  `in_tokens`, `out_tokens`, `cost_usd`. All fields are
  `omitempty`/backwards compatible — older session files keep loading.
- `harness chat` forwards `--auto-gate` and a snapshot of the
  registered adapter ids into `repl.Options` so the slash commands
  have something to render.
- Tutorial cheat sheet covers the new slash + `--resume`/`chat list`.

## 2026-06-19 — v0.120.0 — Chat auto-pin + resume + list + /ship --allow-dirty (F85)

### Changed

- **`harness chat` now auto-pins an adapter** instead of silently
  degrading to the deterministic planner when `--adapter` is omitted.
  Precedence: `--adapter` flag → `.harness/config/active.yaml` pin
  (set via `harness use <id>`) → first registered adapter from
  `claude / codex / gemini / kimi / ollama`. The picked adapter is
  printed so users know which CLI will run. Pass `--no-adapter` to
  force the deterministic-planner mode.
- **`/ship <prompt>` from inside chat now passes `--allow-dirty`**
  automatically. The chat REPL frequently edits files via a plain-text
  agent turn just before the user types `/ship`, so the previous
  behaviour rejected the very flow the tutorial documents
  ("ship: working tree dirty; commit or stash first").

### New

- **`harness chat --resume <id>`** rehydrates a prior session from
  `.harness/sessions/<id>.jsonl`, replays the turns into the new REPL
  so `/history` and the chat Working Memory preamble both see the
  prior conversation, and continues persisting under the same id.
- **`harness chat list`** subcommand prints every persisted session in
  the current project sorted newest-first, with goal, turn count, and
  truncated last input — the discoverable pair for the new
  `--resume` flag.

## 2026-06-19 — v0.119.0 — Chat spinner + multi-turn conversation context (F84)

### Changed

- **`harness chat` now shows a braille spinner** (`⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`) plus
  the word *thinking…* every 120 ms while the pinned adapter is
  working. `claude` and `codex` buffer their JSON output until the
  call finishes, so v0.116–v0.118 made the terminal look frozen for
  tens of seconds during real walk-throughs. The spinner is suppressed
  under `--plain`/`HARNESS_PLAIN=1` and cleaned up with a carriage
  return + blank pass so subsequent output is not clobbered. Stop is
  idempotent (`sync.Once` + waited `done` channel) so the goroutine
  never outlives the turn.
- **`harness chat` plain-text turns now thread a Working Memory
  preamble** (paper §3.2.1) built from the last 5 non-slash session
  turns plus their outcomes. Without this, every chat turn was a
  blank-slate single-shot call — "fix the 500 in cart" had no idea
  that you had just asked claude to add `/products`. The preamble is
  capped at 5 turns × 1200 bytes of agent reply each so it stays well
  inside any adapter's context budget. Four new tests cover the
  empty-session pass-through, the threading, the cap, and slash-
  command filtering.

## 2026-06-19 — v0.118.0 — Python venv sensors + plan_scope metadata defaults (F83)

### Fixed

- **`harness ci` skipped half the python gate with "binary not on PATH"**
  even after `harness new python-ecommerce --with-deps` installed
  everything into `.venv/bin/`. `ShellSensor.Run` now resolves the
  binary against `<root>/.venv/bin/`, `<root>/venv/bin/`,
  `<root>/node_modules/.bin/`, and only then falls back to
  `exec.LookPath`. The subprocess is launched with `PATH` prepended
  to the same dirs so scripts that re-exec their own binary
  (`ruff` calling `python`, `pytest` finding plugins) still resolve
  inside the project venv. Four new tests cover the lookup, env
  prepend, and end-to-end venv-only invocation.
- **`plan_scope` rejected every project metadata file** during a
  `harness ship` flow, so editing `.gitignore`, `Makefile`,
  `pyproject.toml`, `requirements.txt`, `ruff.toml`, `package.json`,
  `Cargo.toml`, `go.mod`, …  failed the gate even when the user's
  PLAN only declared the feature files. `Contract.InScope` now has
  an `alwaysInScope` allowlist for those files plus everything under
  `.harness/`. It also recognises trailing-slash patterns
  (`tests/`) and recursive globs (`app/**`, `app/...`) without the
  user having to write the literal `filepath.Match` glob form.

## 2026-06-19 — v0.117.0 — audit tail timestamps + do --agent display + orchestrate streaming (F82)

### Fixed

- **`harness audit tail` showed every row as `01-01 00:00:00`**.
  `internal/eventlog` writes JSONL lines with `ts` (RFC3339Nano) while
  `internal/audit.Event` only knew the `occurred_at` field, so every
  decoded row landed with a zero-value timestamp. Added a tolerant
  `Event.UnmarshalJSON` that maps `ts`/`timestamp`/`time` →
  `OccurredAt`, `stage`/`level` → `Kind`, `sensor`/`agent` → `Source`,
  and synthesises `Subject` from `sensor`+`status` when no native
  subject is present. Two new tests cover the run-log shape end to
  end.
- **`harness do --agent <id>` plan rendered router pick, not the
  override**. `planDo` now resolves the active-agent override (via
  `activeagent.ResolveAgentID`) before composing each step's
  `chosen` string, so the printed plan, the `--json` output, and the
  executor all agree on the chosen adapter. When the override is set
  but the router has no match for the task tags, the override still
  wins as long as the adapter is registered.

### Changed

- **`harness orchestrate run` now streams each step's stdout/stderr**
  with a `  [<role>] ` prefix while still recording the full output
  in the blackboard. Before, the user only saw
  `orchestrate: step N role=X` and the entire child output was
  buried in `blackboard.json`. Implemented via `io.MultiWriter` over
  an in-memory `bytes.Buffer` (for the blackboard) and a
  `linePrefixWriter` (for the live console).

## 2026-06-18 — v0.116.0 — Chat REPL talks to agent + ship --allow-dirty + new nested guard (F81)

### Changed

- **`harness chat` plain text now streams directly to the pinned
  adapter** (paper §3.1.4). Before, every input was wrapped into a
  deterministic `do → lint → test → ci` plan, so typing "testando"
  ran a four-step plan instead of starting a conversation. The new
  default sends plain text to `Adapter.Run` with `LiveOut` wired
  to a `│`-prefixed writer, so reads, writes, and diffs appear live
  on screen. The previous behaviour stays one keystroke away under
  `/exec <prompt>` (alias `/do <prompt>`).
- **New slash commands** in chat: `/ship <prompt>` calls
  `harness ship --yes`, `/ci`, `/test`, `/lint` run the gate, and
  `!<shell cmd>` escapes to `sh -c` in the project root. The
  `/help` and greet lines now describe the chat-first contract.

### New

- **`harness ship --allow-dirty`**: opt out of the clean-tree
  precondition. Useful when the chat REPL has just edited files and
  the same diff should land in the ship commit.
- **`harness new` nested-target guard**: aborts with a clear error
  when the requested target is a sub-folder whose basename matches
  the current directory (which produced `shop-api/shop-api/shop-api`
  triple-nested paths in v0.115.0).
- **Tutorial rewrite** (`docs/TUTORIAL-ECOMMERCE.md`): single
  chat-driven narrative covering catalogue → cart → checkout,
  with troubleshooting for the OrbStack/Docker dialog the upstream
  `claude` CLI shows on first run.

## 2026-06-18 — v0.115.0 — Health probe + python-ecommerce scaffold (F76–F80)

### New

- **`internal/agenthealth` package (F76, F79)**: race-free periodic
  adapter health probe with `sync.RWMutex` over the status struct,
  `atomic.Bool` for the running flag, and a `context`-driven loop.
  Eight unit tests, including a concurrent `Snapshot()` + adapter
  flip stress test, all clean under `-race`.
- **`harness chat` adapter health badge**: when chat is started with
  `--adapter X`, a 30 s background probe runs and the REPL prompt
  shows `[dev|claude ✓]>` (yellow `⚠` on degradation). `--plain`
  drops the colour.
- **`python-ecommerce` scaffold (F77)**: new bundled stack
  (`harness new python-ecommerce`). FastAPI app with
  `products`, `cart`, `checkout` routers, Pydantic models,
  in-memory thread-safe storage, and a pytest suite that covers
  `/healthz`, listing/getting products, cart totals, and the
  checkout flow. Uses the same venv hardening from F42.
- **Tutorial (F78)** updated to bootstrap from the e-commerce
  scaffold so the read-along walkthrough has a real backend
  by step 1.

## 2026-06-18 — v0.113.0 — diff preview + chat history + audit tail + ship watch (F66–F70)

### New

- **Colored diff preview after every `harness do` (F66)**: the
  workflow now reads `runDir/diff.patch` + `diff-stat.txt` and prints
  a unified-diff preview (first 40 lines) with colour for `+`, `-`,
  `@@` hunk headers, and the diff file header. Stat block sits
  above it.
- **Chat `/last` and `/history` (F67)**: `/last` replays the previous
  prompt in the session; `/history` lists the last 20 inputs (slash
  commands omitted from `/last` resolution).
- **`harness audit tail [--limit N]` (F68)**: explicit subcommand
  that tails the project event log. Falls back to
  `.harness/logs/events.jsonl` when the legacy `audit/events.jsonl`
  is empty, so events recorded by harness runs show up immediately.
- **`harness ship --watch` (F69)**: re-runs the ship loop whenever a
  tracked project file changes. Polls at `--watch-interval` (default
  3s), skips `.git`/`.harness`/`node_modules`/`vendor`/`target`/
  `dist`/`build`/`.venv`/`__pycache__` while hashing.

## 2026-06-18 — v0.112.0 — Smoke matrix + chat multiline + tutorial polish (F61–F65)

### New

- **Smoke matrix expanded (F61)**: every fresh-project run now also
  exercises `harness use claude`, `harness use`, `harness use --clear`,
  `harness diagnose`, `harness orchestrate list`, `harness agent list`,
  `harness config show`, `harness plan write --help`, `harness ship
  --help`, `harness chat --help`, `harness coverage --help`,
  `harness loop --help`, `harness evolve diagnose`,
  `harness smoke --help`. Catches regressions in the new commands
  before they reach end users.
- **Chat multiline prompts (F63)**: end any `harness chat` input line
  with a trailing `\` and the REPL keeps reading until you submit a
  line without one. Continuation prompt is `[goal]…`.

### Tutorial

- `docs/TUTORIAL-ECOMMERCE.md` shows the live agent stream output
  (the `│ ` prefix the workflow injects) and documents the multi-line
  `harness chat` UX (F64).

## 2026-06-18 — v0.111.0 — Streaming + stack-aware coverage + loop UX (F56–F60)

### New

- **Live agent output (F56)**: `harness ship`, `harness do`, and the
  workflow executor now tee the adapter subprocess stdout/stderr to
  the caller's writer in real time. New `agents.AgentRequest.LiveOut`
  field; YAML adapters use it via `runStreamed`. Operators see the
  Claude / Codex / Gemini / Kimi CLI as it generates output instead
  of waiting for the buffered final message.
- **`harness coverage` stack-aware (F58)**: detects the project stack
  from `.harness/config/project.yaml` or manifest probes and runs the
  right tool:
  - `go` → `go test -cover`
  - `python` → `.venv/bin/pytest --cov`
  - `rust` → `cargo tarpaulin`
  - `ruby`/`rails` → `bundle exec rake coverage`
  - `node` → `npx c8 --lines <threshold> npm test`
  Default threshold is still 90%.
- **`harness loop` zero-arg (F59)**: running `harness loop
  --max-attempts 3` with no prompt is now valid. The default prompt
  is *"iterate on the current diff until lint and tests pass"*.
- **`harness chat` adapter fallback (F57)**: if the pinned adapter is
  missing, falls through `claude → codex → gemini → kimi` and warns.
  Healthcheck runs before the session starts; failures surface as a
  warning so the operator knows why a turn might fail.

## 2026-06-18 — v0.110.0 — Real agent flow (F52–F55)

### New

- **`harness use <adapter-id>`** pins the active LLM adapter for the
  project. Writes `.harness/config/active.yaml`. `harness do`,
  `harness ship`, and `harness chat` pick it up automatically; CLI
  overrides win (`--agent`, `--adapter`). `harness use --clear`
  removes the pin. `harness use` with no arg shows the current pin.
- **`harness do --agent <id>`** flag forces a specific adapter for a
  routed task.
- **`harness ship --agent <id>`** forwards the choice into the
  embedded `harness do` invocation.

### Fixes

- `harness chat` consults the active pin when `--adapter` is empty.
- `harness ship` passes `--yes` to the embedded `harness do` so the
  loop stays non-interactive.

### Tutorial

- `docs/TUTORIAL-ECOMMERCE.md` rewritten to pin a real adapter
  (`harness use claude`) and drop the manual scaffold commit (now
  automatic in F46).

### Packages

- New `internal/activeagent` — load/save/clear/resolve the project
  adapter pin.

## 2026-06-18 — v0.109.0 — Tutorial end-to-end fixes (F46–F51)

Real-walk fixes uncovered while running TUTORIAL-ECOMMERCE.md.

### Fixes

- **`harness new` refuses non-empty target** (F46). Previously rerunning
  `harness new python ./shop-api` inside an existing project created
  `./shop-api/shop-api/shop-api/...`.
- **`harness new` auto-commits the scaffold baseline** (F46). The
  scaffolded tree is staged + committed as `chore: scaffold baseline`
  so `harness ship --plan <id>` no longer trips on
  *working tree dirty*.
- **`forbidden_files` sensor default-excludes virtualenvs, caches, and
  IDE dirs** (F47): `.venv`, `venv`, `__pycache__`, `.pytest_cache`,
  `.ruff_cache`, `.mypy_cache`, `.tox`, `.cache`, `.idea`, `.vscode`,
  `.bundle`, `.gradle`, `.next`, `.nuxt`, `.turbo`, `.parcel-cache`,
  `coverage`, `htmlcov`.
- **`plan_scope` sensor allows `.harness/` by default** (F48). The
  metadata tree never counts as out-of-scope.
- **`harness chat --goal dev` planner runs `harness do` first**, then
  lint + test + ci (F50). Previously the default plan only verified;
  it never wrote any code.
- Tutorial `cart-cycle.yaml` example drops the unknown `--apply` flag
  on `harness do` (F49).

## 2026-06-18 — v0.108.0 — Robust venv + two-agent diagnose/fix + color UI (F42–F45)

### Fixes

- **Robust Python venv install (F42)**: `harness new --with-deps`
  now tries `uv` first, then the highest stable Python found on PATH
  (`python3.13` → `python3.12` → `python3.11` → `python3`), with three
  fallback strategies (`ensurepip`, `--without-pip + get-pip.py`,
  `--without-pip + system pip --prefix`). Works on hosts where
  Python 3.14's `ensurepip` is broken.

### New

- **`harness diagnose`** (F43): runs the bundled diagnosers (missing
  tools, dirty tree, unpinned plan) and writes a JSON diagnosis to
  `.harness/artifacts/diagnoses/`. Two-agent pattern from paper §3.5.2.
- **`harness fix [problem-id] [--all]`** (F43): applies registered
  fixers (`install-tool`, `commit-snapshot`) to the problems surfaced
  by `harness diagnose`. Diagnosers and Fixers are pluggable; bundled
  defaults live in `internal/twoagent`.
- **Color UI (F44)**: `harness new`, project wrappers (test/lint/dev/
  bench/profile), and `harness diagnose` output now use coloured
  markers (`✓ ✗ ⚠ ℹ ·`) via `internal/ui`. `--plain` /
  `HARNESS_PLAIN=1` still flips everything back to ANSI-free output.

### Packages added

- `internal/venvinstall` — Python venv install with strategy fallback.
- `internal/twoagent` — Diagnoser/Fixer contracts, default
  implementations, JSON persistence.

## 2026-06-18 — v0.107.0 — Tutorial walkthrough fixes (F37–F41)

Fixes raised during real end-to-end walkthrough of TUTORIAL-ECOMMERCE.md.

### Fixes

- `harness new --with-deps` now actually runs the scaffold's post_steps
  (`venv` + `pip install`, `bundle install`, `cargo build`, `npm
  install`). Previously the flag existed but did nothing, so
  `harness lint` / `harness test` immediately failed with
  `.venv/bin/ruff: No such file or directory` (F37).
- `harness plan check --plan <id>` accepts bare ulids, `PLAN-<ulid>`,
  and `PLAN-<ulid>.md` interchangeably. Previously a `PLAN-` prefix
  produced an invalid path (no `.md` suffix) and erroed out (F38).
- `harness memory promote` rejects a `--run-id` that looks like a
  flag (starts with `--`). This catches the common shell pitfall where
  an empty variable expansion shifts subsequent args (`--run-id $BLANK
  --confidence 0.85` → `--run-id` got the literal string
  `"--confidence"`, and `--confidence` silently fell back to its
  default) (F39).
- Tutorial rewritten (F40): captures `PLAN_ID` from the artifact
  filename, commits the scaffold before `harness ship`, points at
  `.harness/runs/` for run ids instead of the legacy `_do/` path.



Fixes blocking the e-commerce tutorial walkthrough.

### Fixes

- `harness update` refuses downgrade unless `--force` (F32). Tag display
  normalised with `v` prefix on both `current:` and `target:` lines.
- `harness install <tool>` now sanitises a stale `GOROOT` when invoking
  `go install` (F33). Previously crashed with
  `cannot find GOROOT directory: …gvm/gos/go1.19.2` when the host had
  a removed gvm install lingering in env.
- Workflow no longer prints `Budget: $0.90 / $1.00 remaining` when no
  LLM was actually charged (F34). The line only appears after a paid
  call.
- Version constant moves to `v` prefix (`v0.106.0`) so internal
  comparisons stop swapping prefixed and unprefixed tags.

### Docs cleanup (F35)

- Single tutorial: `docs/TUTORIAL-ECOMMERCE.md`. Removed legacy
  `docs/tutorial-python-demo.md`, `docs/tutorial.md`,
  `docs/paper-coverage-map.md`, `docs/ROADMAP-DONE.md`,
  `docs/v1-readiness.md`, `docs/cli-reference.md`,
  `docs/dashboard-parity-audit.md`, `docs/decomposer-decision.md`,
  `docs/spec-p64-multi-agent.md`, `docs/coverage-plan.md`,
  `docs/quickstart.md`, `docs/overview.md` (superseded by README +
  COMMANDS.md + ARCHITECTURE.md + PAPER-MAPPING.md).
- `README.md` and `docs/ARCHITECTURE.md` documentation tables updated
  to point at the single remaining tutorial.

## 2026-06-17 — v0.105.0 — Paper end-to-end implementation (F0–F27)

End-to-end implementation of "Code as Agent Harness" (arXiv 2605.18747).
Every paper section now has a concrete command, package, and test.
Single PR `feature/F0-paper-end-to-end` against `develop`.

### New commands

- `harness new <stack>` — single-command project bootstrap.
- `harness ship "<prompt>" [--plan <id>]` — SDLC driver: branch →
  spec → do → ci-loop → conventional commit. 429-aware fallback and
  optional plan-as-contract enforcement.
- `harness chat --goal {dev|ads|research|ops} [--adapter <id>]` —
  typed-plan REPL with deterministic dispatch (§3.1.4).
- `harness plan write` / `harness plan check` — plan-as-contract
  artefact + scope enforcement (§3.4.2).
- `harness orchestrate list|show|run <flow>` — multi-role flow with
  chain/cyclic topology and file-only blackboard (§4.1.1 + §4.1.3 +
  §4.3.1). Adapter steps dispatch through `internal/agents`.
- `harness evolve diagnose|propose|replay|sandbox|promote --hitl` —
  Agentic Harness Engineering (§3.5). `sandbox` runs a real A/B replay
  in an isolated workspace.
- `harness config show|set|unset|wizard` — interactive routing wizard
  with audited mutations (§3.5.3).
- `harness coverage --threshold 0.9` — Go coverage gate (§3.4.4).
- `harness smoke matrix [--langs csv|all]` — cross-stack CLI
  regression harness.
- `harness test|lint|dev|bench|profile` — project wrappers reading
  `.harness/config/project.yaml`.
- `harness install-git-hooks --hooks all` — embed pre-commit +
  commit-msg + pre-push.

### New scaffold

- **Rails 7 API** — Gemfile, RSpec, Rubocop, healthz controller.

### New sensors

- `go_coverage_gate` (auto-registered for Go).
- `plan_scope` (auto-registered when `.harness/config/plan.yaml` pins
  an active plan).
- `internal/sensors/commentscan` — Go AST scan flagging narrative
  comments outside SPDX, package docs, godoc on exported symbols.

### Architecture additions

- `internal/intentplan` — typed JSON plan schema with per-goal palette.
- `internal/repl` — REPL behind `harness chat`, injectable planner.
- `internal/projectcfg` — `.harness/config/project.yaml` schema +
  per-stack detection.
- `internal/plancontract` — parses `PLAN-<id>.md` artefacts.
- `internal/orchestrate` — flow loader + executor + adapter runner.
- `internal/evolve` — telemetry clustering + sandbox replay.
- `internal/configwiz` — interactive router config with audit log.
- `internal/customrules` — `.harness/rules/*.yaml` loader.
- `internal/sensors/coverage` — `go test -cover` parser.
- `internal/sensors/planscope` — scope diff vs PLAN contract.

### Documentation

- `docs/COMMANDS.md` — every command reference.
- `docs/ARCHITECTURE.md` — paper-anchored runtime architecture.
- `docs/PAPER-MAPPING.md` — paper § → command → package → test.
- `docs/TUTORIAL-ECOMMERCE.md` — end-to-end e-commerce build.
- `README.md` — rewritten for OSS best practices (badges, TOC,
  citation, contributing, code of conduct, security policy pointers).

### Verification

- 86 internal packages green, 0 failures.
- 14 new internal packages, average coverage 93.7%.
- `harness smoke matrix` green across 6 stacks (go, python, rails,
  react, ruby, rust).
- `make tutorial-replay` green.

### Open follow-ups

- Coverage sensor for non-Go stacks (pytest --cov, cargo tarpaulin,
  c8, simplecov).
- Full LLM-driven planner in `harness chat` with token-budget feedback
  loop.

## 2026-06-16 — v0.41.0 — v1.0 readiness checklist (P73)

- **`docs/v1-readiness.md`**: single-source checklist tracking what stands between v0.4x and a v1.0.0 cut. Maps every shipped surface to the paper principles + open challenges, lists quality / UX / docs gates, and pins the 7 items that block v1.0.0 today (stable JSON schema versioning, govulncheck re-run, coverage gate, tutorial refresh for v0.33–v0.40, architecture + JSON-schema docs, dashboard parity decision, LLM decomposer fallback decision).
- Proposed cut criteria documented: ✗ → ◐/✓ on every blocker, dog-food test end-to-end, three cold-start operator runs report zero blockers.

## 2026-06-16 — v0.40.0 — harness do --json (P72)

- **`harness do "<prompt>" --json`**: emits the executed plan + per-task results as JSON on stdout (implies `--yes`; routes human-facing logs to stderr so consumers can parse stdout cleanly). Schema: `{prompt, report_path, steps:[...], results:[...]}`. Completes the IDE-plugin contract started in v0.39: plugins can plan via `route show --json`, execute via `do --json`, and consume both with the same `jsonStep` schema.

## 2026-06-16 — v0.39.0 — harness route show --json (P71)

- **`harness route show "<prompt>" --json`**: emits the planned task graph as JSON for programmatic consumers (IDE plugins, scripts, CI). Schema: `{prompt, steps:[{index, kind, tags, routing, adapter_id, prompt, confidence, lang}]}`. No LLM call, <500ms, stable schema versioned via the v0.39 tag.
- First building block toward an IDE plugin contract — plugins can ask harness to plan a multi-agent run without executing and surface the chosen adapter per task in the editor.

## 2026-06-15 — v0.38.0 — Cross-task handoff in harness do (P70)

- **`harness do` now prepends a "Past steps in this run" block** to every task after the first. Each task sees the list of previously-routed steps with their adapter + result so a later code task can build on what the earlier scaffold + image task produced. Implements the paper's "shared code artifacts support multi-agent coordination" without a new shared-memory abstraction.
- Block ends with `Do not redo work that already succeeded.` to keep the LLM from regenerating files the deterministic step already wrote.

## 2026-06-15 — v0.37.0 — Agent-call heartbeat (P69)

- **`DefaultExecutor.Status func(string)`** field receives a notice immediately before and after `adapter.Run`. Workflow wires it to `[agent] calling <id>...` / `[agent] <id> returned in <dur>` lines on the same stream as the rest of the output. Closes the "no visual cue when LLM is being called" complaint from v0.27 dog-food testing without taking a new dep (no spinner library).
- Field is optional: when `Status` is nil, executor is silent (no behaviour change for unit tests / library consumers).

## 2026-06-15 — v0.36.0 — Runs + project prune (P68)

- **`harness runs prune --older-than <dur> --keep-last <n>`**: deletes run directories under `.harness/runs/` matching either policy. Default dry-run; pass `--apply` to delete. Reports total bytes freed. Duration accepts `Nd` (days) plus standard Go duration suffixes.
- **`harness project prune --older-than <dur>`**: archives projects whose `last_seen_at` is older than the threshold. Default dry-run; pass `--apply` to archive. Archived projects stay in the registry — use `harness project unarchive` to restore.
- **`internal/execution.PruneCandidates`** + **`DeletePaths`**: shared helpers so both commands share retention logic.
- **`projectcmd.StaleSince`**: returns projects with `last_seen_at` before a threshold.

## 2026-06-15 — v0.35.0 — Help topics for do/loop/scaffold + tutorial polish (P67)

- **`harness help do`**, **`harness help loop`**, **`harness help scaffold`**: new in-CLI tutorial topics covering the v0.30–v0.34 surface. `harness help` lists them in the topic index.
- **`docs/tutorial.md`** gains sections 10b (`harness loop`) and 10c (`harness memory recall`) with full examples + expected output. Version reference bumped to v0.35.0.

## 2026-06-15 — v0.34.0 — Sensor confidence + low-confidence task warning (P66)

- **`sensors.Result.Confidence` (0..1)**: addresses the paper's "verification with incomplete feedback" open challenge. Renderers show `~conf N.NN` for any non-deterministic verdict (between 0 and 1). 0 = unknown / not set; 1.0 = deterministic (default, hidden).
- **`harness do` plan now includes a CONF column** and emits a warning when any task has classification confidence < 0.5: `⚠ one or more tasks have low classification confidence — review before --yes`.

## 2026-06-15 — v0.33.0 — Regression-aware loop + cross-session memory + multimodal auto-route (P65)

- **`harness loop` now captures a baseline** before the first attempt by running `lint_command` + `test_command` once. If a later attempt breaks something the baseline had green, the loop flags it as a regression and the canonical-error block prepends `## Regression detected\n<reason>\n\nFix this before anything else.` Closes the paper's "regression-free improvements" open challenge.
- **`harness memory recall "<query>"`**: bag-of-words search over every `.harness/runs/*/report.md`. Scores by `intersection(query_terms, report_terms) / len(query_terms)`. No LLM, no external index. Lives in `internal/recall`.
- **`harness do --image <path>`**: attaching an image adds the `vision` tag to every task, so the router automatically picks a vision-capable adapter (`gemini`, `claude`) instead of defaulting to text-only.

## 2026-06-15 — v0.32.0 — Multi-agent routing + composability (P64)

Implements Layer 3 of the "Code as Agent Harness" paper (arXiv 2605.18747): composability via deterministic per-task adapter routing.

- **`harness do "<prompt>"`**: decomposes a free-form prompt into typed tasks ("scaffold X and add Y then generate Z" → 3 tasks) and routes each to the best adapter (or to a deterministic implementation when one exists). Per-task report at `.harness/runs/_do/do-<ts>.md`. Flags: `--yes`, `--deterministic` (default on; prefers scaffold/sensor over LLM), `--budget-usd`, `--max-tasks`, `--autonomy`.
- **`harness route show "<prompt>"`**: dry-run that prints the planned task graph + chosen adapter per task without executing. <500ms, no LLM call.
- **`internal/taskgraph`**: rule-based decomposer. Splits by `and`/`then`/`,`/`;`, classifies each clause against ~20 regex rules covering 14 task kinds (scaffold, lint, test, format, secrets, code, refactor, docs, review, image, vision, search, data, shell, generic). Returns `Task{Kind, Tags, Prompt, Lang, Confidence}`.
- **`internal/router`**: deterministic strengths matcher. Scores adapter via `intersection(task.tags, adapter.strengths) / len(task.tags)`; ties broken by adapter id (stable). `router.Pick(tags, registry)` returns the top adapter.
- **Bundled adapter strengths unified** to controlled vocabulary (code, refactor, reasoning, search, docs, tests, image, vision, audio, data, sql, shell, review). Updated: `claude` (code/reasoning/refactor/review/docs), `codex` (code/tests/refactor), `gemini` (code/vision/image/search/docs), `kimi` (code/search/review), `fake` (code/tests).
- **Deterministic-first** (paper principle: executability). Tasks of kind `scaffold`/`lint`/`test`/`secrets` map to `scaffoldpkg`/`sensorcmd` and skip the LLM entirely when `--deterministic` is on (default).

## 2026-06-15 — v0.31.0 — harness loop + budget ledger + presenter primitives (P63 part 2)

- **`harness loop "<prompt>"`**: deterministic dev-loop. Runs `harness feature`, then the project's lint + test commands (auto-detected from scaffold or set via `--lint` / `--test`). On failure, packages the lint/test output as a canonical error block and feeds it back to the LLM as the follow-up prompt. Bounded by `--max-attempts` (default 3, hard cap 10) and `--budget-usd`. Final report at `.harness/runs/_loop/loop-<ts>.md`.
- **`internal/devloop`**: new package with `Run`, `Canonicalise(prompt, attempt)`, auto-detection that maps `requirements.txt → python`, `Cargo.toml → rust`, `Gemfile → ruby`, `package.json → react`, `go.mod → go`.
- **`budget.Guard.ChargeWith(Entry)` + `Guard.Entries()`**: per-call ledger so callers can render a breakdown table. `Entry{Label, USD, Tag, Note}` lets the renderer show `claude-sonnet-4-6  in=2140 out=512 $0.0144 (reported)` instead of a single mystery cost.
- **`internal/ui/workflowview.go`**: presenter primitives (`Phase`, `Status`, `Presenter`) with rich (lipgloss colors) + plain (`[PHASE] ...` grep-friendly) implementations. Auto-picks via `ui.IsPlain()`. Workflow wiring lands incrementally in a follow-up release; primitives ship now so the loop + downstream code can already use them.
- **`ui.IsPlain()`**: public accessor for plain-mode state (previously only `SetPlain` was exported).

## 2026-06-15 — v0.30.0 — init --git + deterministic language scaffolds (P63 part 1)

- **`harness init --git`**: runs `git init -b main` when `.git/` is absent. `--git-branch <name>` to override. Skip noisily if a repo already exists.
- **`harness init --all`**: implies `--git` plus registers the project in the cross-project workspace registry (slug derived from dir basename, override via `--slug`).
- **`harness scaffold {list,show,apply}`**: new top-level command that drops deterministic language scaffolds with **zero LLM calls**. Bundled languages: `python` (FastAPI + pytest + ruff + Makefile), `go` (net/http + table-driven tests + golangci-lint), `ruby` (Sinatra + rspec + rubocop + Rakefile), `rust` (Axum + integration tests + clippy), `react` (Vite + React 18 + TypeScript + Vitest + ESLint). Each scaffold ships a `scaffold.yaml` declaring `required_tools`, `files`, `post_steps`, `lint_command`, `test_command`, `run_command`. Output is byte-identical for the same `(lang, name)`.
- **`harness scaffold apply <lang> --apply --with-git --with-deps`** writes files, optionally initialises git, and runs `post_steps` (venv + pip install, go mod tidy, bundle install, npm install, cargo build).
- **`internal/scm/git.go`**: tiny helper (`HasRepo`, `Init`, `CurrentBranch`) wrapping the only git calls harness needs.
- **`internal/scaffoldpkg/`** mirrors the `hookpkg` / `mcppkg` pattern (embed.FS + List/Load/Apply). New languages = drop a directory under `templates/<lang>/`.
- **Tutorial rewritten** (`docs/tutorial.md`): sections 2 + 3 replace the old "create dir + git init + harness init + project add" steps with the new one-liner flow.
- **Deferred to v0.31 (P63 part 2)**: colored sectioned output + spinner, per-call budget breakdown table, `harness loop` deterministic LLM ⇒ lint/test ⇒ canonical-error-retry loop.



## 2026-06-15 — v0.29.1 — Single canonical docs/tutorial.md (drop versioned manuals)

- **`docs/tutorial.md`** is now the only manual. Per-version files (`tutorial-v0.4-manual.md`, `v0.11`, `v0.27`, `v0.28`, `v0.29`) are deleted. Historical commands live in `git log docs/tutorial.md` for anyone who needs them. Stops the accumulation of redirect notices and stale walkthroughs.

## 2026-06-15 — v0.29.0 — harness uninstall + brew formula renamed harness

- **`harness uninstall project`**: wipes `./.harness/` in the current directory (after confirmation).
- **`harness uninstall global`**: wipes the cross-project registry under `GlobalHarnessDir()` (`$HARNESS_HOME`, `$XDG_DATA_HOME/harness`, `~/Library/Application Support/harness` on macOS, `~/.local/share/harness` on Linux, `%LOCALAPPDATA%/harness` on Windows).
- **`harness uninstall all`**: runs both, then removes the `harness` binary from `$PATH` (falls back to printing `sudo rm <path>` when the install dir isn't writable). When `brew` is detected, prints the matching `brew uninstall harness && brew untap rodolfopeixoto/harnessx` commands.
- **Brew formula renamed `harnessx` → `harness`**: install command is now `brew install harness` (formerly `brew install harnessx`). Tap URL unchanged: `brew tap rodolfopeixoto/harnessx https://github.com/rodolfopeixoto/harnessx.git`.

## 2026-06-15 — v0.28.1 — Slug update honored on project re-add (hotfix)

- **`harness project add <path> --slug <new>`** now updates the slug when the project root is already registered, instead of silently keeping the original slug. Collision against another row still rejects with the existing `slug %q already used by %s` error.

## 2026-06-15 — v0.28.0 — UX fixes from real dog-food + unified tutorial (P62)

- **`harness feature --budget-usd <n>`**: canonical flag name matching the docs (`docs/anthropic-billing.md`, `harness help billing`). `--budget` kept as hidden deprecated alias for one release.
- **`harness list`** (new top-level): composite read-only view of registered projects, last 10 runs, and bundled/installed agents. No LLM call. Replaces the previous behaviour where bare `harness list` triggered the feature-intent classifier.
- **`harness sensor run --root <path>`**: pins the project root for ad-hoc sensor runs. Default cwd, matching what every other workflow command already accepted.
- **Reports unified to one canonical path**: every run now writes a single `.harness/runs/<id>/report.md`. The duplicate `.harness/artifacts/reports/<id>.md` writer is gone (that directory is now reserved for user-triggered artefact reports — `harness security-audit`, `harness report perf`, etc.). `harness report --last` now reads from the runs directory.
- **`harness init` scaffolds `.harness/hooks/pre-tool-use.sh`** as a permissive `exit 0` stub with a comment block listing the bundled templates. Empty hooks directory never blocks a run again.
- **`harness hook add <event>`**: interactive selector that lists every bundled template matching the event and installs the chosen one to `.harness/hooks/<event>.sh` with `chmod 755`. `--yes` installs the first match without prompting.
- **Hook block messages name the script and the fix**: `pre-tool-use blocked by .harness/hooks/pre-tool-use.sh (exit 1)\n  → make the script exit 0 to allow, or remove .harness/hooks/pre-tool-use.sh` replaces the cryptic `pre-tool-use(pre-tool-use)=1`.
- **`docs/tutorial-v0.28-manual.md`**: new unified end-to-end walkthrough against a FastAPI / Python sample app (17 sections, expected output per step). Proves HarnessX is language-independent. Older tutorials (`v0.4`, `v0.11`, `v0.27`) carry a redirect notice at the top.

## 2026-06-15 — v0.27.0 — Interactive Claude Code adapter (P61) [experimental]

- **`--agent claude-interactive`**: drives Claude Code's interactive REPL programmatically so runs draw from the operator's Pro/Max subscription bucket instead of the Agent SDK monthly credit. Three strategies: `pty` (default, via `github.com/creack/pty`), `tmux` (opt-in, uses `send-keys` + `capture-pane`), `iterm2` (macOS opt-in, via `osascript`).
- **`type: interactive` YAML spec**: new top-level adapter type with `interactive:` block (`strategy`, `binary`, `args`, `idle_ms`, `hard_timeout_seconds`, `banner_pattern`, `tmux.session_name`, `iterm2.profile`). Validator rejects unknown strategies and missing `binary`.
- **`experimental: true` flag**: surfaces in `harness agent list` (new `EXP` column with `★`) and as the final line of `harness agent certify` output. REPL surface is undocumented; can break on Claude Code upgrades. `ParseUsage` returns `mode: estimated` because the interactive REPL emits no usage block.
- **Bundled `claude-interactive.yaml`**: defaults to PTY, idle 1500ms, hard timeout 180s. Install with `harness agent install claude-interactive`.
- **Billing doc + `harness help billing`**: every dollar amount now tagged `as of 2026-06-15` plus a "cross-check at anthropic.com/pricing" line; subscription-stream row now points at `claude-interactive`.

## 2026-06-15 — v0.26.0 — Anthropic billing guide (P60)

- **`docs/anthropic-billing.md`**: explains the three Anthropic spending streams (subscription / Agent SDK monthly credit / pay-as-you-go API), maps `--agent claude` to the Agent SDK credit ($20-$200/month) and `--agent anthropic-api` to pay-as-you-go, lists per-plan credit amounts, and gives a workload-based adapter pick.
- **`harness help billing`**: in-CLI summary with the same mapping plus the per-run `--budget-usd` recommendation and `harness metrics --since 1d` tracking hint.
- **Practical advice baked in**: automation-heavy → API key + `anthropic-api`; exploration + a bit of automation → opt in Agent SDK credit at console.anthropic.com.

## 2026-06-15 — v0.25.0 — Certify simple_prompt timeout + agentcmd split (P59)

- **`certify.Options.SimpleTimeout`** (default `90s`): bounds the simple_prompt round-trip independently from the short `PerCheckTimeout` (10s) used by healthcheck / timeout / cancellation probes. Real LLM round-trips take 5-60s; the old 10s ceiling killed Claude before it could answer.
- **`harness agent certify --simple-timeout <duration>`**: override per call. Default 90s.
- **`signal: killed` + `context deadline` remediation** now points the operator at `--simple-timeout 180s` plus the manual smoke command instead of the generic "waiting on interactive auth" message.
- **`internal/app/agentcmd/agentcmd.go` split**: 462 → 323 LOC. Render + remediation + summary helpers moved to `certify_render.go` (140 LOC). Remediation table is now data-driven (`map[checkName]func(detail, adapter) string`) instead of a giant switch.

## 2026-06-15 — v0.24.0 — Certify output UX (P58)

- **`harness agent certify <id>`** now prints, per check: a `what:` line (one-sentence description of what the check proves), the existing `detail:` line, and a `fix:` line on every failure with an actionable next step (login command, `harness install <name>`, manual smoke command).
- **Final one-line summary** at the bottom: `ready / usable / partial / blocked` plus the exact next command to run.
- **simple_prompt failure** maps `signal: killed` to a clear "CLI is waiting on interactive auth" message and prints the login command from the adapter's auth block plus a manual smoke (`echo "ping" | <cli> --print --output-format json`).

## 2026-06-15 — v0.23.0 — Probe runtime-error guard (P57)

- **Doctor probe** now rejects output containing runtime-error markers (`cannot find`, `command not found`, `no such file`, `permission denied`, `unknown command`, `error:`, `fatal:`, `panic:`) even when a semver digit pattern is present in the same string. Fixes the dog-food report where a broken `GOROOT` env (`go: cannot find GOROOT directory: /Users/.../.gvm/gos/go1.19.2`) was reported as ✓ because the regex extracted `1.19.2` from the path. Now flagged as `⚠ present, version probe failed` with the actionable substring in the version slot.
- **Unit tests** added: runtime-error-with-incidental-semver and command-not-found cases.

## 2026-06-15 — v0.22.0 — Roadmap-done index + final polish (P56)

- **`docs/ROADMAP-DONE.md`**: single source of truth for v0.4 → v0.21. Phase table, cumulative CLI surface, dashboard pages, HTTP API, bundled artifacts (17 install / 7 mcp / 5 hook / 4 skill / 9 agent / 5 runtime / 7 cleanup), cross-platform release matrix, quality gates green at v0.21, open list, operator update path.
- **Quality gates verified once more** at the end of the 0.x cycle: `make lint` 0 issues, `go test ./...` green, `govulncheck` no vulns, `gitleaks detect` no leaks.

## 2026-06-15 — v0.21.0 — Master tutorial + SSE + skill templates + security pass (P51-P55)

- **`docs/tutorial.md`**: master walkthrough consolidating v0.4 → v0.20. Per-OS install, 14 sections, 18-row validation checklist, honest gap list.
- **`GET /api/events/runs/<id>`**: Server-Sent Events tail of `.harness/runs/<id>/events.jsonl`. 15s keepalive, 500ms poll, scoped path so it does not collide with the existing `/api/runs/<id>` REST handler.
- **`internal/skillpkg`** + **`harness skill templates|install <name>`**: 4 bundled deterministic skill snippets (`security-rule`, `clean-code`, `go-feature`, `bugfix-loop`) installed via `.harness/skills/<name>.md`.
- **Security pass**: `govulncheck ./...` no vulns; `gitleaks detect` no leaks; `go test ./...` green; lint zero.

## 2026-06-15 — v0.20.0 — Bundled hook templates (P50)

- **`internal/hookpkg`**: 5 bundled hook scripts — `pre-tool-use-lint` (go vet + golangci-lint), `pre-tool-use-secrets` (refuse runs when .env exposes a key/token), `pre-tool-use-noforce` (refuse force-push prompts), `post-tool-use-test` (go test ./... or npm test), `post-tool-use-audit` (one-line log per run).
- **`harness hook templates`**: lists the bundled scripts with the inferred event + description headers.
- **`harness hook install <name>`**: writes `.harness/hooks/<event>.sh` (or `--filename <override>`), `chmod +x`, picked up by `harness hook scan` and the Executor's pre/post dispatch immediately.

## 2026-06-15 — v0.19.0 — Bundled MCP templates (P49)

- **`internal/mcppkg`**: 7 bundled MCP server templates — `filesystem`, `github`, `postgres`, `sqlite`, `brave-search`, `fetch`, `memory`. Each carries transport / command / args / env / docs URL.
- **`harness mcp templates`**: lists available templates with command + description so operators do not have to grep upstream docs.
- **`harness mcp install <name>` auto-fills** from the bundled template (when one matches the name). `--command`, `--url`, `--transport` still override. Result lands at `.harness/mcp/<name>.json` with `args`, `env`, `docs` fields too — the Executor's MCP injection picks it up unchanged.

## 2026-06-15 — v0.18.0 — Doctor --fix + harness worktree cleanup detector (P48)

- **`harness doctor --fix [--dry-run]`**: walks every ⚠/✗ probe that ships a bundled install manifest and runs `harness install <name>` for each. `--dry-run` prints the chosen strategy per tool without executing. Reuses the same `install.NewRegistry()` strategy picker so the per-platform behaviour matches one-shot `harness install`.
- **`internal/cleanup/detectors/harness_worktrees.go`**: new detector picks up orphan `.harness/worktrees/<run-id>/` directories left over when a run was killed mid-flight or the operator skipped `harness runs discard`. Risk = medium by default, high after the stale threshold. Surfaces in `harness cleanup scan` and `GET /api/cleanup/scan`.

## 2026-06-15 — v0.17.0 — Quality-of-life batch (P47)

- **`harness backup config show`**: print the resolved `.harness/config/backup.yaml` (default remote, compression, include + exclude lists) without opening the file.
- **`harness backup config set-default-remote <name>`**: pin the default remote without editing YAML.
- **Better error** on `harness backup list` / snapshot / etc. when no remote is chosen: prints a 4-line fix recipe pointing at `harness backup remotes`, `harness backup remote add`, `harness backup config set-default-remote`, or `--remote <name>`.
- **`harness completion install`**: auto-detect `$SHELL`, write the completion script to the conventional path per OS (`/usr/local/etc/bash_completion.d`, `~/.zsh/completion/_harness`, `~/.config/fish/completions/harness.fish`, etc.). `--shell`, `--dry-run`. Prints the one-line post-install hint per shell.

## 2026-06-15 — v0.16.0 — Dog-food fixes (P46)

- **Apple Container fallback**: `AppleContainer.Available` now runs a `container list --format json` probe in addition to the version check. When the probe fails (the daemon is unhealthy or the CLI flags do not match our shape), `Detect()` returns `docker` as the auto-pick. Resolves the `container list: exit status 1` operators saw when apple_container was the picked runtime but unable to actually list.
- **`/api/secrets/names` shape**: every detected backend appears in the response (env / keychain / encrypted_file on macOS; env / secret_service / encrypted_file on Linux) with `[]` instead of `null` when no secrets are stored. Stable dashboard rendering.
- **Dog-food smoke** in `/tmp/dogfood`: `harness init`, `project add`, `install list`, `runtime info`, `secret info`, `execute --apply`, `dashboard /api/*` all walk green now.

## 2026-06-15 — v0.15.0 — Windows binaries + Homebrew formula generator (P45)

- **Windows binaries**: `make release` now cross-builds `windows/amd64` and `windows/arm64`. Windows artifacts ship as `harness-windows-<arch>.zip` (instead of tar.gz) plus matching `.sha256`. Same size budget enforcement; 18 MiB on amd64, 17 MiB on arm64.
- **`scripts/gen-brew-formula.sh`** emits a Homebrew formula keyed off the release tag and the per-platform sha256 values in `dist/`. Drop the generated `Formula/harnessx.rb` into `rodolfopeixoto/homebrew-tap` and operators get `brew tap rodolfopeixoto/tap && brew install harnessx`.
- **`Formula/harnessx.rb`** committed in this repo as the source of truth (regenerated per release); the tap repo mirrors it.
- **`docs/install.md`** rewritten as a per-OS install guide (install.sh / Homebrew tap / Windows unzip / Scoop bucket template / build-from-source) with verification and update steps.

## 2026-06-15 — v0.14.0 — Dashboard UI pages for runtime / containers / images / install / secrets / backup (P44)

- **`/runtime`** lists detected runtimes with selected ★, plus current binary / version / source.
- **`/containers`** cross-runtime listing with `--all` toggle; mutations stay on the CLI for safety.
- **`/images`** image listing across the selected runtime.
- **`/install`** bundled manifest catalog with category filter + `harness install <name>` hint per row.
- **`/secrets`** names per backend (env / keychain / encrypted_file); values never returned by the API.
- **`/backup`** copy-paste cheatsheet for `harness backup` (no upload from dashboard process).
- **Nav order** reorganised so infrastructure surfaces cluster between Resources and Cleanup.

## 2026-06-15 — v0.13.0 — Portable backup + sync via rclone (P43)

- **`harness backup snapshot|restore|list|sync|remotes|remote add`**: tar.gz snapshots pushed/pulled through any rclone remote (drive, s3, dropbox, onedrive, r2, webdav, crypt). Provider credentials live in rclone; harness never touches them.
- **Default `.harness/config/backup.yaml`** includes `config + artifacts/specs + runs`; excludes `db`, `cache`, `worktrees`, `secrets.enc`, `secret-seed`.
- **Secrets default to excluded**. `--include-secrets` requires `HARNESS_BACKUP_I_UNDERSTAND_SECRETS=1`; recommendation: route the bucket through an rclone `crypt` overlay.
- **Manifest** per snapshot: harness version, OS/arch, included paths, SHA-256 per entry, tag, timestamp.
- **Path-traversal guard** + 500 MiB per-file ceiling on restore. Refuses to write into a non-empty target without `--force`.
- **`internal/install/manifests/rclone.yaml`** + `harness install rclone` per-platform (brew / apt / dnf / pacman).
- **`.harness/artifacts/specs/p43-backup-sync.md`** records the design + safety rules.

## 2026-06-15 — v0.12.0 — Dashboard read-only APIs (P42)

- **`GET /api/runtime`** — currently selected runtime: id, binary, version, source (env|config|auto).
- **`GET /api/runtimes`** — every known runtime with availability + selected flag.
- **`GET /api/containers?all=true`** — cross-runtime container list via the resolved runtime.
- **`GET /api/images`** — container images.
- **`GET /api/install`** — bundled tool manifests with installed status per binary.
- **`GET /api/secrets/names`** — secret names per backend (env / keychain / encrypted_file); values never returned.

Wire-in via `s.registerP42(mux)` next to the existing `registerScans`. UI pages land in P44.

## 2026-06-15 — v0.11.0 — install.sh smoke + completion + tutorial v0.11 (P41)

- **`scripts/tests/install_smoke.sh`**: runs the public installer against a clean `HARNESS_PREFIX` in a temp dir, verifies the resulting binary boots, reports the version, and exercises `harness update status` + `harness --help`.
- **Shell completion verified** for bash, zsh, fish via the existing `harness completion <shell>` command. Tutorial documents the per-shell install path.
- **Tutorial `docs/tutorial-v0.11-manual.md`**: end-to-end walkthrough for every surface shipped between v0.6 and v0.10 (`install`, `runtime`, `containers`, `images`, `secret`, API adapters, `--sandbox container`, channels, help topics). 14-row validation checklist. Honest "what is not shipped yet" list (dashboard pages, brew, Windows, Apple Container `Run`, v1.0).
- **Roadmap refresh**: v1.0.0 deferred; P41–P45 cover the 0.11 → 0.15 cycle so we dog-food each release before declaring 1.0.

## 2026-06-15 — v0.10.0 — Clean-code sweep (P40)

- **Refactor `internal/index/api.go::BuildAPIMap`** (gocognit 57 → under threshold): one helper per stack (`collectRailsRoutes`, `collectNextRoutes`, `collectGoRoutes`) plus `hasStack`, `nextRoutePath`, `sortRoutes`. Behaviour preserved; tests green.
- **Refactor `internal/sensors/budget.go::snapshotValue`** (gocognit 43 → 4): table-driven `snapshotResolvers` map plus `pathValue`, `sumContainerField`, `maxContainerField`, `dockerfileFindings` resolvers.
- **Move `cmd/harness/cmd_update.go` helpers to `internal/update`**: download / sha256 verify / tar extract / replace binary all live behind the public API now (`PlatformTarget`, `TarballURL`, `DownloadFile`, `VerifySha256`, `ExtractTarget`, `ReplaceBinary`). `cmd_update.go` is pure CLI glue.
- **CONTRIBUTING.md** distils the comments + complexity + attribution rules so the next contributor does not reintroduce noise. Includes single-scope commit convention and no-AI-attribution policy.

## 2026-06-15 — v0.9.0 — Sandboxed agent execution + harness images (P39)

- **`Runtime.Run`**: docker-like runtimes (docker, podman, orbstack, colima) gain `Run(ctx, RunSpec) (RunResult, error)` for one-shot containers with bind mounts, env, stdin, auto-remove, timeout. AppleContainer stub returns actionable error pointing at `harness runtime set docker/podman`.
- **`Runtime.ListImages` + `PruneImages`**: cross-runtime image ops with the same two-key safety rule as containers.
- **`harness images list|prune`**: tabular listing + dangling prune (gated by `HARNESS_CONTAINERS_I_UNDERSTAND=1`).
- **Executor sandbox dispatch**: `execution.Request.Sandbox = "host"|"container"` + `SandboxImage` (default `alpine:3.20`). When `container`, the worktree bind-mounts at `/work` and the agent CLI runs inside the runtime via `runInContainer`. `--sandbox container --sandbox-image <img>` on `harness execute`.

## 2026-06-15 — v0.8.0 — API agent adapters + cross-platform secret store (P38)

- **Secret store** at `internal/secrets`: cross-platform backends in priority order — process env (`HARNESS_SECRET_<UPPER>` or `<UPPER>`), macOS Keychain (`security` CLI), Linux Secret Service (`secret-tool`), AES-GCM encrypted file at `~/.harness/secrets.enc`. Best practice: never log secrets, redact in `harness secret get`, encrypt-at-rest in fallback.
- **`harness secret list|set|get|unset|info`** with `--from-env`, `--from-file`, `--reveal`. Set hidden via stdin terminal prompt; reveal opt-in.
- **API adapter type** in `internal/agents/yaml` accepts `type: api` with an `api:` block (endpoint, method, headers, auth.header/scheme/secret_ref/query_param, request_template with `{{prompt}}/{{model}}`, response.final_message + usage JSONPath, timeout, retry).
- **`internal/agents/http.Adapter`** implements `AgentAdapter` via stdlib `net/http`; resolves secrets via `secrets.Store.Resolve`; classifies HTTP failures (401/403→auth, 429→rate, 5xx→transient, 4xx→permanent).
- **5 bundled API adapters**: `anthropic-api`, `openai-api`, `gemini-api`, `moonshot-api`, `minimax-api`.
- **`harness agent login <id>`**: CLI adapters print `claude login` / `codex auth login` / etc.; API adapters store the API key in the secret backend (via stdin or `--from-env`).
- **`harness agent install <id>`** alias for `agent add`.

## 2026-06-15 — v0.7.0 — Container runtime selection + harness containers (P37)

- **Runtime interface** in `internal/runtime/containers/runtime.go`: pluggable abstraction with Docker, Podman, OrbStack, AppleContainer, Colima impls. `Detect(ctx)` returns available runtimes ordered by per-platform preference (macOS: apple_container > docker > orbstack > podman > colima; linux: docker > podman > orbstack > colima).
- **`harness runtime list|select|set|info`**: list detected runtimes with version + selection status, interactive picker, explicit pinning, current info. Persists to `.harness/config/runtime.yaml`.
- **`HARNESS_RUNTIME=<id>`** env override per call.
- **`harness containers list|kill|prune`** cross-runtime. `prune` honours a two-key rule: interactive `yes` OR `HARNESS_CONTAINERS_I_UNDERSTAND=1` for non-interactive flows. Flags: `--all`, `--stopped`, `--older-than 720h`, `--json`.
- **Resolve precedence** for the runtime: env `HARNESS_RUNTIME` > `.harness/config/runtime.yaml` > auto-detect.

## 2026-06-15 — v0.6.0 — Doctor probe fix + harness install (P36)

- **Probe parser** in `internal/adapters/execprobe` now captures stdout+stderr, runs an optional `VersionRegex` override per probe, and treats non-zero exits as success when a semver match is extracted. Fixes the "present, version probe failed" warning for `go` and `gemini`.
- **`harness install <tool>`** with sub-commands `list` and `show`. Reads bundled YAML manifests under `internal/install/manifests/*.yaml` and runs the first viable per-platform strategy (brew, apt, dnf, pacman, go_install, npm_global, cargo_install, pip_user). Flags: `--dry-run`, `--upgrade`.
- **16 bundled manifests**: gopls, ripgrep, syft, ruby-lsp, solargraph, pyright, basedpyright, rust-analyzer, tsserver, gemini, claude, codex, kimi, golangci-lint, govulncheck, gitleaks.
- **`harness doctor`** appends a `Recommended installs` section listing actionable `→ harness install <name>` lines for every ⚠/✗ probe that has a bundled manifest.

## 2026-06-15 — v0.5.0 — Release channels (stable/beta/develop) + harness help topics (P35)

- **`harness update --channel stable|beta|develop`**: stable picks newest non-prerelease, beta includes pre-releases (`-beta*`, `-rc*`), develop builds from source. Aliases: `harness upgrade`, `harness self-update`.
- **`harness update status`**: compares current binary against the channel's latest, reports up-to-date / update available / current-is-newer.
- **`harness update channels`**: lists every release per channel newest-first with publish date and HTML URL.
- **`harness help <topic>`**: in-CLI tutorials for quickstart / agents / sensors / hooks / autonomy / mcp / update / input / tracker — no doc lookup required.
- **`internal/update`**: stand-alone channel resolver + version comparator (handles `v` prefix, `-beta`, `-rc` suffixes); unit-tested with deterministic release fixtures.

## 2026-06-15 — v0.4.0 — Real agentic execution + MCP/hook integration + workflow improvements (P31–P34)

- **P31 Executor**: real agentic loop wired end-to-end — worktree (git or copy fallback), adapter.Run, stdout/stderr capture, diff capture (unified patch + stat + JSON), sensors bridge, autonomy gate, apply via `git apply --3way` or `waiting_approval`, persisted `meta.json` + `report.md` per run under `.harness/runs/<id>/`.
- **P31 fake-agent**: deterministic Claude-shaped JSON adapter for tests and the e2e smoke (`cmd/fake-agent`, `templates/agents/fake-real.yaml`).
- **P31 run CLI**: `harness runs list|inspect|report|sensors|approve|discard`; `harness execute` direct path; `harness feature/bugfix/run --agent <id> --apply --autonomy <level>` routes through the executor.
- **P32 MCP injection**: when adapter capability `mcp=true`, executor merges `mcpscan.Scan` output into `runs/<id>/mcp-config.json` and appends `--mcp-config <path>` via `AgentRequest.ExtraArgs`.
- **P32 Hook dispatch**: pre-tool-use and post-tool-use hooks fire around `adapter.Run`; non-zero pre-hook routes to `autonomy_denied` unless level is `full_project_loop`.
- **P32 `harness mcp install`**: writes `.harness/mcp/<name>.json` with `--command/--url/--transport/--yes`.
- **P34 trivial fast path**: `intent.Complexity {Trivial, Standard, Complex}` heuristic; `workflow.askAgent` skips spec+plan+worktree+diff for question-style prompts and writes the answer to `report.md`.
- **P34 prompt enhancement**: `internal/promptenh` deterministically composes skill prefix + context summary + task; persisted as `enhancement.json` per run.
- **P34 cost auto-routing**: `router.PickModel(adapter, complexity)` picks cheap / default / deep model from the adapter's `models` map.
- **P34 autonomy policy**: `.harness/config/autonomy.yaml` supports `allow_paths|deny_paths|allow_commands|deny_commands`; deny wins, allow-list non-match downgrades to `require_approval`.
- **P34 multi-input flags**: `--prompt-file`, `--pdf` (via local `pdftotext`), `--image` (base64 attachment for vision-capable adapters).
- **P34 auth login hints**: adapter YAML `auth: { login_command, doc_url }` surfaced on certify auth failure (`run: claude login | docs: ...`).
- **P34 tracker CLI**: `harness metrics --since 1d|7d|30d|all` aggregates per-run state; `harness audit --kind <k> --limit N` reads the audit JSONL.
- **P34 update**: `harness update` self-updates from latest GitHub release (verifies SHA-256, swaps binary in place); `harness update --channel develop` builds from source.
- **Tutorial**: `docs/tutorial-v0.4-manual.md` walks every surface end-to-end in English with the login matrix per agent CLI.

## 2026-06-15 — v0.3.0 — Stack Audit + Handoff parity + MCP/Hook scanners (P23–P29)

- **P23 stack audit**: deterministic visual + functional audit pipeline.
- **P24 audit bundle**: single zip + BUNDLE_INDEX.md for LLM consumption.
- **P25 audit hardening**: pass-rate 23% → 100%.
- **P26 handoff gap report**: 11 new routes, nav 7 → 17, audit features 10 → 22.
- **P27 Home/Projects/Catalog rich content**: MetricCard, PathCell, TerminalReflection, ActionService.
- **P28 MCP/Hook scanners + rich Cleanup page**: deterministic discovery + CLI + HTTP + UI.
- **P29 finalise handoff**: 8 remaining stub pages rewritten; Catalog MCP/Hook tabs wired.

Final audit: 34/34 passed, 0 console_error, 0 selector_missing, 0 layout_collapsed.

## 2026-06-14 — v0.2.0 — Workspace Hub + Capabilities + Stack Tour (P11–P22)

- **P11 Workspace Hub** — Multi-project registry at `~/.harness/registry.sqlite`;
  `harness project add|list|switch|archive|scan|forget|import|stale`; HTTP
  `/api/workspace/{projects,switch,current,import,stale[/slug]}`. Resolver
  precedence: --project flag > HARNESS_PROJECT env > cwd walk-up > active row.
- **P12 Capabilities Center** — 8 kinds (agent/mcp/hook/sensor/skill/context/
  resource/plugin) with deterministic discovery, Plan→approval→Apply pipeline
  (path-traversal guard + stage-then-rename), bundled manifests, HTTP
  `/api/catalog/{kinds,items,plan}`, CLI `harness catalog`.
- **P13 Cleanup Engine** — Scanner + detectors (worktrees, caches, abandoned
  .harness, large files, VM leftovers, Claude leftovers, containers via
  runtime/containers.Lister). D5 two-key rule: policy match OR interactive y
  OR HARNESS_CLEANUP_I_UNDERSTAND=1. Every delete writes audit_event with
  sha256.
- **P14 Coverage gates** — `coverage-gate.sh` (go), `coverage-web.sh` (vitest
  + @vitest/coverage-v8), `coverage-shell.sh` (bashcov). CORE regex covers
  every new pkg. Entry thresholds 50/55 with documented ratchet to 90 by v0.3.
- **P15 Design System** — `web/dashboard/src/ds/`: tokens + strings + Badge,
  Card, EmptyState, Tabs, InspectorPanel, DataExplorer, Shell. App wired via
  Shell with nav as data. 11 DS unit tests.
- **P16 Import Wizard + Stale Detection** — Shared `importwiz.Plan` powering
  both CLI + future UI; `internal/stale` fingerprints package files into
  `.harness/project/fingerprints.json`; Detect reports changes by kind.
- **P17 Command Palette ⌘K** — `internal/palette` (Score: exact 100 > prefix
  80 > contains 60 > fuzzy 20+), sources for Projects/Capabilities/Builtin
  commands. CLI `harness palette search`, HTTP `/api/palette`, React modal
  with arrow-key navigation.
- **P18 Autonomy + Health** — 5 levels (Manual/PlanAndAsk/SafeExecute/
  FullProjectLoop/ScheduledMaintenance) with declarative Gate matrix.
  Deterministic 10-subsystem health score. CLI `harness autonomy get|set`,
  `harness health show`. HTTP `/api/autonomy`, `/api/health/score`.
- **P20 Stack Tour** — `harness stack tour [--dashboard --keep]` deterministic
  walkthrough (workspace add → catalog install → cleanup scan → autonomy
  gate → health score → optional /api/health probe). `harness stack status`
  reports dashboard online/offline. Self-cleaning by default.
- **P21 Role Grid** — `web/dashboard/src/auth/{roles.ts,RoleContext.tsx}`
  (anonymous|operator|admin), role-grid.test.tsx walks every page × every
  role. Vitest total: 15 files / 67 tests.
- **P22 Container lifecycle harness** — `internal/runtime/containers` shared
  by Cleanup + Stack Tour: typed docker Lister, Compose Up/Down, HealthProbe
  with bounded backoff, VerifyClean asserting `docker ps -a` empty.

Install: `curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh | bash`
(now accepts `--dry-run` and `--prefix`).

## 2026-06-14 — GitFlow hardening sweep (G1–G7)

- G1: GitFlow — `main` (releases) + `develop` (integration);
  feature/fix/chore/release/hotfix prefixes documented in
  `CONTRIBUTING.md`.
- G2: God-file refactor — `internal/app/workflow/workflow.go` 517→215 LOC
  (split into `telemetry.go`, `helpers.go`, `execute.go`);
  `internal/adapters/http/server.go` 559→104 LOC (split into
  `handlers.go`, `static.go`, `helpers.go`). No file > 260 LOC.
- G3: Go coverage push — `registry_test.go` (agents 0→100%),
  `defaults_test.go` (router 59→70%), `snapshot_test.go` (optimize
  52→73%), `build_test.go` (design 56→77%), `builder_more_test.go`
  (context 42→54%). Global coverage 47.9%.
- G4: React component tests — Sessions, SessionDetail, RunDetail,
  Sensors, Agents, Memory, Design, Roadmap, Settings. 20 tests passing.
- G5: Shell test harness — `scripts/lib/assert.sh` + 4 suites in
  `scripts/tests/`; wired into `make ci` via `make test-sh`.
- G6: Coverage gate ratcheted — `GLOBAL_MIN` 35→40, `CORE_MIN` 35→50.
  Lowest core: `internal/context` 53.7%.
- G7: Final sweep — `make ci` green (lint + tests + shell harness +
  coverage gate + 10 e2e phases); `make security` clean (vet, gitleaks,
  harness security-audit).

## 2026-06-14 — Senior-engineering polish (B1–B12 + lint clean + local CI)

- Lint: full `.golangci.yml` (v2) with errcheck, govet, staticcheck,
  ineffassign, unused, misspell, revive, nolintlint, gocyclo (15),
  gocognit (25), gosec, unconvert, gocritic, dupl. `golangci-lint run`
  reports **0 issues**.
- Coverage gate: `scripts/coverage-gate.sh` + `scripts/coverage-aggregate.py`
  enforce global ≥ 35% / core ≥ 35% with a documented ratchet plan.
- Supply-chain: `make licenses` runs `go-licenses` → `THIRD_PARTY_LICENSES.md`
  + `NOTICE` + CSV; blocks AGPL/GPL/LGPL/SSPL/EUPL. `make sbom` produces
  CycloneDX 1.5 JSON via syft or stdlib-Python fallback.
- Security: `make security` runs `govulncheck` advisory + `harness
  security-audit` (forbidden_files + forbidden_commands + secrets_scan +
  go_vuln). `.harnessignore` excludes the scanners' own pattern literals.
- Doctor: new categories `lsp` + `quality` — probes for gopls, ruby-lsp,
  solargraph, pyright, basedpyright, rust-analyzer, typescript-language-server,
  golangci-lint, govulncheck, go-licenses, gitleaks, syft.
- Cmd split: `cmd/harness/main.go` 720 LOC → 85 LOC; 19 `cmd_*.go` files;
  shared `cwd()` helper.
- Constants + i18n: `internal/platform/constants` central magic values;
  `internal/platform/i18n` embedded `en` + `pt` bundles, `HARNESS_LANG`
  env, fallback chain.
- Hooks: `make install-hooks` writes `pre-commit` (gofmt + go vet),
  `commit-msg` (Conventional Commits regex), `pre-push` (`make ci`).
- Release: `make release` cross-builds darwin/linux × amd64/arm64,
  enforces 40 MiB per-binary budget, emits SHA-256 sums.
- Docs: `COMPLIANCE.md`, `docs/spec-driven-development.md`, mermaid
  diagrams in `README.md`, refreshed `AGENTS.md`, `CLAUDE.md`,
  `CONTRIBUTING.md` (GitFlow + branch-protection note), SBOM script.
- New commands: `harness explain`, `harness session show`, `harness
  artifact ls/cat`, `harness skill list/promote`, `harness spec init`,
  `harness routes`, `harness completion <shell>`, `harness memory
  list/promote`.
- `harness logs --follow` Bubble Tea TUI; embedded React dashboard via
  `//go:embed all:web/dashboard/dist`.
- AutoLSP: any of gopls / ruby-lsp / solargraph / pyright /
  basedpyright / rust-analyzer / typescript-language-server is spawned
  on demand when its binary + manifest are present.


## 2026-06-14 — Hardening 4: gopls LSP client

- `internal/adapters/lsp/gopls.go` full LSP-over-stdio client (Content-Length
  framing, JSON-RPC 2.0, async demux, mutex-guarded writer, atomic IDs).
- `DocumentSymbols` (hierarchical + flat parsers), `Diagnostics` (drains
  `publishDiagnostics` notifications), cached per spec §15 layout.
- `internal/context.AutoLSP(root)` wires gopls automatically when binary is
  present and `go.mod` exists. Missing binary silently falls back to the
  default provider chain.

## 2026-06-14 — Hardening 3: memory CLI

- `harness memory list [--limit] [--scope]` reads the `memories` table.
- `harness memory promote --content … --run-id … --confidence …` enforces
  the spec §11 evidence gate (rejects missing evidence, low confidence,
  sensitive content).

## 2026-06-14 — Hardening 2: interactive confirmation + Cycle F sensor

- Workflow now prompts `Approve plan? [y/N]` when stdin is a TTY and
  `--yes` is not passed. CI/redirected stdin returns `false`.
- New `performance_budget` sensor compares the most recent perf snapshot
  against `.harness/project/performance-budget.json`.

## 2026-06-14 — Hardening 1: docs + workflow exec wiring

- All eight missing spec §31 docs written (agents, sensors, skills,
  context-engineering, design-to-product, resource-optimization, security,
  dashboard).
- Workflow `--execute` now resolves a route → agent chain and runs it via
  `router.Execute`. Cost + tokens + fallback_from persisted on the run row.

## 2026-06-14 — Phase 10: full end-to-end

- `scripts/e2e-phase10.sh` chains every phase against one project.
- `Makefile` `e2e-all` target runs every phase script.

## 2026-06-14 — Phase 9: resource optimization

- Cycles A (snapshot), B (Dockerfile audit), C (deps classifier), D (log
  scanner), G (report) — pure-Go, no Docker dependency.
- 7 commands: `optimize`, `perf-snapshot`, `perf-compare`, `image-audit`,
  `dependency-audit`, `log-audit`, `security-audit`.
- Conservative dep removal per spec §21 core rule; observability/security/
  debugger deps marked `kept_for_operational_safety`.

## 2026-06-14 — Phase 8: dashboard

- 14 read-only REST endpoints under `/api/*`.
- React SPA w/ 9 routes (Sessions, SessionDetail, RunDetail, Sensors,
  Agents, Design, Roadmap, Memory, Settings); loading/empty/error states
  for every panel.
- Bubble Tea `harness logs --follow` TUI (750 ms poll, rotation-aware).
- Built-in HTML fallback page so the dashboard works without `npm install`.

## 2026-06-14 — Phase 7: design-to-product

- Safe ZIP extractor (zip-slip rejection + 200 MiB cap).
- Inventory pages, components, assets, CSS tokens, JS interactions,
  missing states, responsive notes.
- Six product maps: design-manifest, feature-map, toggle-map, roadmap,
  api-contracts, flow-map.
- Image hash + metadata cache (vision-model `detected` field reserved).

## 2026-06-14 — Phase 6: spec + plan workflow

- Rule-based intent classifier (8 modes + explainable reasons).
- §8 spec renderer with safe defaults per mode.
- §9 plan renderer with confirmation status.
- Budget guard, evidence-gated memory promotion, §28 report writer.
- Six use cases (ask/plan/run/feature/bugfix/report) + natural form
  `harness "<prompt>"`.

## 2026-06-14 — Phase 5: context engineering

- Pack builder with 4 default providers (memory, git, ripgrep, test-map).
- LSP provider abstraction + cache key layout per spec §15.
- Token estimator (4-chars heuristic; pluggable).
- Canonical hash over `task + profile + providers + git HEAD`.

## 2026-06-14 — Phase 4: sensors

- 4 universal scanners (`forbidden_files`, `forbidden_commands`,
  `secrets_scan`, `changed_files`).
- Per-stack rule packs (Go, Node/React/Next.js/Vite, Rails, Python, Rust,
  Docker) — OptionalTool=true skips on missing binaries.
- Runner with computational-first ordering + streaming callback.
- `harness sensor list|run`, `harness check`, `harness ci`.

## 2026-06-14 — Phase 3: agent adapter system

- `AgentAdapter` interface + Capabilities + FailureType (IsRecoverable).
- YAML loader w/ template substitution + tiny JSONPath subset.
- Bundled Claude/Codex/Gemini/Kimi/Fake adapter YAMLs.
- Router w/ explainable Select + fallback Execute.
- Certification suite (healthcheck → simple_prompt → timeout → cancellation
  → failure_classification).

## 2026-06-14 — Phase 2: project index

- Stack detector (Rails, React, Next.js, Vite, Node, Go, Python, Rust, Docker).
- 8 maps: profile, commands, dependencies, architecture, test-map, api-map,
  design-system, performance-budget.
- Incremental cache keyed on input fingerprint.

## 2026-06-13 — Phase 1: core CLI

- `harness init` bootstraps `.harness/` + SQLite schema + bootstrap session.
- `harness doctor` Lip Gloss panel with system + tools + agents + project.
- `harness logs --tail` JSONL viewer.
- 8 SQLite tables from spec §23 (sessions, runs, sensor_results, metrics,
  memories, skill_versions, agent_certifications, artifacts).

## 2026-06-13 — Phase 0: scaffold

- Go module, MIT LICENSE, Makefile, golangci.yml, GitHub Actions CI.
- React + Vite + TS dashboard scaffold w/ Vitest smoke test.
- `templates/.harness/` defaults.
- Docs (overview, architecture, install, quickstart, cli-reference,
  configuration, contributing).
