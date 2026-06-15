# HarnessX Roadmap — Completed Phases (v0.4 → v0.21)

Source of truth for what shipped, when, and where to look.

| Phase | Version | Ship date | Theme |
|---|---|---|---|
| P31 | v0.4.0 | 2026-06-15 | Real agentic execution loop (worktree → adapter → diff → sensors → autonomy gate → apply) |
| P32 | v0.4.0 | 2026-06-15 | MCP config injection + hook pre/post dispatch + `harness mcp install` |
| P34 | v0.4.0 | 2026-06-15 | Trivial fast path + prompt enhance + cost auto-route + autonomy policy + multi-input + auth hints + tracker |
| P35 | v0.5.0 | 2026-06-15 | Release channels (stable/beta/develop) + self-update + `harness help` topics |
| P36 | v0.6.0 | 2026-06-15 | Doctor probe parsing fix + `harness install` (16 manifests) |
| P37 | v0.7.0 | 2026-06-15 | Container runtime selection (apple/docker/podman/orbstack/colima) + `harness containers` |
| P38 | v0.8.0 | 2026-06-15 | API agent adapters (anthropic/openai/gemini/moonshot/minimax) + cross-platform secret store |
| P39 | v0.9.0 | 2026-06-15 | Sandboxed agent execution + `harness images` |
| P40 | v0.10.0 | 2026-06-15 | Clean-code sweep: split BuildAPIMap, snapshotValue, CONTRIBUTING.md policy |
| P41 | v0.11.0 | 2026-06-15 | install.sh smoke + completion + tutorial v0.11 |
| P42 | v0.12.0 | 2026-06-15 | Dashboard read-only APIs (runtime/containers/install/secrets/images) |
| P43 | v0.13.0 | 2026-06-15 | Portable backup + sync via rclone |
| P44 | v0.14.0 | 2026-06-15 | Dashboard UI pages for the new surfaces |
| P45 | v0.15.0 | 2026-06-15 | Windows binaries + Homebrew formula generator |
| P46 | v0.16.0 | 2026-06-15 | Dog-food fixes (apple container fallback, secret API shape) |
| P47 | v0.17.0 | 2026-06-15 | Quality-of-life batch (backup config show, completion install, better errors) |
| P48 | v0.18.0 | 2026-06-15 | `doctor --fix` + harness worktree cleanup detector |
| P49 | v0.19.0 | 2026-06-15 | Bundled MCP templates (filesystem/github/postgres/sqlite/brave/fetch/memory) |
| P50 | v0.20.0 | 2026-06-15 | Bundled hook templates (lint/secrets/noforce/test/audit) |
| P51-P55 | v0.21.0 | 2026-06-15 | Master tutorial + SSE Active Run events + skill templates + security pass |

## Cumulative CLI surface (one command per capability)

```
# install
harness install <tool>                # 17 manifests
harness install rclone                # required for backup
harness doctor --fix                  # bulk install everything missing

# runtime + containers + images
harness runtime list|select|set|info
harness containers list|kill|prune
harness images list|prune

# secrets + agents
harness secret list|set|get|unset|info
harness agent install <id>            # 9 bundled (5 api + 4 cli)
harness agent login <id>
harness agent certify <id>

# mcp + hooks + skills
harness mcp templates
harness mcp install <name>            # 7 bundled
harness hook templates
harness hook install <name>           # 5 bundled
harness skill templates
harness skill install <name>          # 4 bundled

# execution
harness feature/bugfix/run <prompt> --agent <id> --apply --autonomy <level>
harness execute <prompt> --sandbox container --sandbox-image <img>
harness runs list|inspect|approve|discard|report|sensors

# tracker + cleanup
harness metrics --since 1d|7d|30d|all
harness audit --kind <k>
harness cleanup scan|apply

# backup
harness backup snapshot|restore|list|sync
harness backup remote add|remotes
harness backup config show|set-default-remote

# update + help + completion
harness update --channel stable|beta|develop
harness update status|channels
harness help <topic>
harness completion install
```

## Cumulative dashboard pages

`/sessions`, `/sessions/:id`, `/runs/:id`, `/sensors`, `/agents`, `/catalog`, `/projects`, `/command`, `/plan`, `/run`, `/design`, `/roadmap`, `/memory`, `/resources`, `/cleanup`, `/reports`, `/stakeholder`, `/onboarding`, `/settings`, **`/runtime`**, **`/containers`**, **`/images`**, **`/install`**, **`/secrets`**, **`/backup`**.

## HTTP API (new in v0.4 → v0.21)

```
GET  /api/runtime           current selected runtime
GET  /api/runtimes          all detected
GET  /api/containers?all=
GET  /api/images
GET  /api/install
GET  /api/secrets/names     stable per-backend shape; values never returned
GET  /api/events/runs/<id>  SSE tail of events.jsonl
```

## Bundled artifacts shipped

| Kind | Count | Files |
|---|---|---|
| Install manifests | 17 | `internal/install/manifests/*.yaml` (gopls, ripgrep, syft, ruby-lsp, solargraph, pyright, basedpyright, rust-analyzer, tsserver, gemini, claude, codex, kimi, golangci-lint, govulncheck, gitleaks, rclone) |
| MCP templates | 7 | `internal/mcppkg/templates/*.yaml` (filesystem, github, postgres, sqlite, brave-search, fetch, memory) |
| Hook templates | 5 | `internal/hookpkg/templates/*.sh` (pre-lint, pre-secrets, pre-noforce, post-test, post-audit) |
| Skill snippets | 4 | `internal/skillpkg/templates/*.md` (security-rule, clean-code, go-feature, bugfix-loop) |
| Agent adapters | 9 | `internal/app/agentcmd/bundled/*.yaml` (claude, codex, gemini, kimi, fake, anthropic-api, openai-api, gemini-api, moonshot-api, minimax-api) |
| Runtime impls | 5 | `internal/runtime/containers/*.go` (Docker, Podman, OrbStack, AppleContainer, Colima) |
| Cleanup detectors | 7 | abandoned, caches, containers, harness_worktrees, large files, leftovers, worktrees |

## Cross-platform release matrix

- darwin/amd64, darwin/arm64 (tar.gz + sha256)
- linux/amd64, linux/arm64 (tar.gz + sha256)
- windows/amd64, windows/arm64 (zip + sha256)
- Homebrew formula auto-generated per tag (`scripts/gen-brew-formula.sh`)
- install.sh smoke test in `scripts/tests/install_smoke.sh`

## Quality gates (all green at v0.21.0)

- `make lint` — 0 issues (gocognit ≤ 25, gocyclo ≤ 15, no errcheck/unused/staticcheck violations)
- `go test ./...` — every package green
- `govulncheck ./...` — no vulnerabilities
- `gitleaks detect` — no secret leaks
- `gofmt -l` — clean

## What is not in v0.21.0 (open list)

- v1.0.0 release ritual — explicitly deferred until several 0.x dog-food cycles pass
- Apple Container `Run`/`ListImages` (Available probe disables when CLI not ready; pin docker instead)
- Brew tap repo `rodolfopeixoto/homebrew-tap` (formula is generated; operator must `gh repo create + commit + push`)
- Windows `install.ps1` (manual unzip works today)
- Dashboard mutate actions (kill/prune/install) — all mutations stay on the CLI
- E2E pixel diff hooked into stack audit (scaffold exists)
- Scheduled / cron backup
- Realtime image-conditioned codegen via vision-capable adapters

Each item is small + scoped; none blocks the v0.21 install + dog-food.

## Operator update path

```bash
harness update                        # latest stable
harness update --channel beta         # opt into pre-releases
harness update --tag v0.21.0          # pin a tag
harness update status                 # is anything newer?
harness update --channel develop      # source build (git + go)
```

First install on a fresh box (any release in the matrix above):

```bash
curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh | bash
```
