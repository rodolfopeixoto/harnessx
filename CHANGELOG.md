# Changelog

Format: [phase] short summary, then bullet list of concrete additions.
Newest milestones at the top. Dates are when the milestone landed in repo.

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
