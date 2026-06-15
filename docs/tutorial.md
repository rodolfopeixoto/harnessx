# HarnessX — Master Tutorial (v0.20)

End-to-end walkthrough for every surface shipped between v0.4 and v0.20.
Per-OS, real CLIs, no fakes. Tick the 18-row validation checklist at the
end to declare the install green.

> Stay on `0.x.x` — no v1.0 yet. Dog-food first.

---

## 0. Install

| OS | Command |
|---|---|
| macOS / Linux (recommended) | `curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh \| bash` |
| macOS / Linux (Homebrew) | `brew tap rodolfopeixoto/tap && brew install harnessx` |
| Windows | see `docs/install.md` (manual unzip from the GitHub release `.zip` artifacts) |

Already installed (v0.6+):

```bash
harness update                  # stable
harness update --channel beta
harness update status
```

Shell completion (auto-detect):

```bash
harness completion install
```

---

## 1. Doctor + tool install

```bash
harness doctor                  # probes every required + recommended binary
harness doctor --fix --dry-run  # plan to install every missing tool
harness doctor --fix            # run the plan
```

Per-tool override:

```bash
harness install list                  # 17 bundled manifests
harness install gopls
harness install --upgrade golangci-lint
harness install rclone                # required for harness backup
```

Strategies per platform: `brew → apt → dnf → pacman → go install → npm -g → cargo install → pip --user`. First viable wins.

---

## 2. Container runtime selection

```bash
harness runtime list                  # detected with ★ on selected
harness runtime info
harness runtime set docker            # override auto-pick
HARNESS_RUNTIME=podman harness ...    # one-shot env override
```

Preference order (auto-pick):

| OS | Order |
|---|---|
| macOS | apple_container → docker → orbstack → podman → colima |
| Linux | docker → podman → orbstack → colima |

`apple_container` is auto-disabled when its CLI cannot run `container list --format json`; fallback is `docker`.

### Container + image ops

```bash
harness containers list [--all]
harness containers kill <id>
HARNESS_CONTAINERS_I_UNDERSTAND=1 harness containers prune --stopped --older-than 720h
harness images list
HARNESS_CONTAINERS_I_UNDERSTAND=1 harness images prune --older-than 720h
```

Two-key safety: prune always needs interactive `yes` or the env var.

---

## 3. Cross-platform secrets

```bash
harness secret info                       # backends active per OS
harness secret list                       # names per backend
harness secret set ANTHROPIC_API_KEY      # hidden stdin prompt
echo "v" | harness secret set X --from-file /dev/stdin
harness secret set X --from-env MY_VAR
harness secret get X                      # redacted by default
harness secret get X --reveal
harness secret unset X
```

Backend priority: env > Keychain (macOS) / Secret Service (Linux) > AES-GCM encrypted file fallback.

---

## 4. Agents

### CLI adapters (login via the CLI itself)

```bash
harness install claude            # or codex / gemini / kimi
harness agent login claude        # prints `claude login` + docs URL
claude login                      # do it
harness agent certify claude
```

### API adapters (no CLI required, just an API key)

```bash
harness agent install anthropic-api   # bundled: anthropic / openai / gemini-api / moonshot / minimax
harness agent login anthropic-api --from-env ANTHROPIC_API_KEY
harness agent certify anthropic-api
```

---

## 5. MCP servers

```bash
harness mcp templates              # 7 bundled: filesystem, github, postgres, sqlite, brave-search, fetch, memory
harness mcp install filesystem     # auto-fill from template
harness mcp install github         # needs GITHUB_PERSONAL_ACCESS_TOKEN
harness mcp install my-server --command /bin/x --transport stdio
harness mcp scan                   # what's wired
```

When the chosen agent declares `capabilities.mcp: true`, the Executor merges every `.harness/mcp/*.json` into a single `runs/<id>/mcp-config.json` and passes `--mcp-config <path>` to the adapter automatically.

---

## 6. Hooks

```bash
harness hook templates             # 5 bundled
harness hook install pre-tool-use-lint
harness hook install post-tool-use-test
harness hook scan                  # confirm event + status
```

Failing pre-tool-use hook routes the run to `autonomy_denied` under `safe_execute`; `full_project_loop` bypasses with explicit operator opt-in.

---

## 7. Autonomy + policy

```bash
harness autonomy get
harness autonomy set --level safe_execute    # default
```

Per-project `.harness/config/autonomy.yaml`:

```yaml
level: safe_execute
allow_paths: ["src/**", "docs/**"]
deny_paths:  ["secrets/**", ".env*", "infra/prod/**"]
allow_commands: ["go test", "npm test"]
deny_commands:  ["rm -rf /", "git push --force"]
```

| Level | Low-risk apply | High-risk apply | Hook block bypass |
|---|---|---|---|
| manual | deny | deny | no |
| plan_and_ask | approval | approval | no |
| safe_execute | allow | approval | no |
| full_project_loop | allow | approval | yes |
| scheduled_maintenance | approval | deny | no |

---

## 8. Real agentic run

```bash
mkdir ~/dev/harness-tutorial && cd $_
git init -q && git config user.email t@t && git config user.name t \
  && git config commit.gpgsign false
echo "# init" > README.md && git add -A && git commit -q -m seed

harness init
harness project add . --slug tutorial
harness feature "create HELLO.md with content: hi" \
  --agent claude --apply --autonomy safe_execute --budget-usd 0.50
```

Multi-input:

```bash
harness feature "summarize" --pdf brief.pdf --agent claude --apply
harness feature "redo this layout" --image mockup.png --agent claude
harness feature "..." --prompt-file ./prompt.md --agent claude --apply
```

Sandboxed execution (runs the adapter inside the selected runtime):

```bash
harness execute "..." --agent claude --apply --sandbox container \
  --sandbox-image alpine:3.20
```

---

## 9. Run inspection + tracker

```bash
harness runs list
harness runs inspect <run-id>
harness runs report  <run-id>
harness runs sensors <run-id>
harness runs approve <run-id>    # waiting_approval → applied
harness runs discard <run-id>

harness metrics --since 7d
harness audit --kind sensor --limit 20
```

---

## 10. Cleanup

```bash
harness cleanup scan             # worktrees (git + .harness), caches, abandoned dirs, leftovers, containers, large files
harness cleanup apply --policy .harness/cleanup/policy.yaml
```

Two-key safety: policy match + interactive `y` or `HARNESS_CLEANUP_I_UNDERSTAND=1`.

---

## 11. Portable backup + sync (rclone)

```bash
harness install rclone
rclone config                                      # one-time provider auth
harness backup remote add gdrive --provider drive --interactive
harness backup config set-default-remote gdrive
harness backup config show
harness backup snapshot --tag pre-experiment
harness backup list
harness backup restore <snapshot> --target /tmp/restored
harness backup sync push --dry-run
harness backup sync pull
```

Default config excludes secrets. Opt-in via `--include-secrets` requires `HARNESS_BACKUP_I_UNDERSTAND_SECRETS=1`; route the bucket through an rclone `crypt` overlay.

---

## 12. Dashboard

```bash
harness dashboard --addr :7373
open http://localhost:7373
```

Real-backed pages: Sessions, SessionDetail, RunDetail, Sensors, Agents, Catalog, MCP, Hooks, Runtime, Containers, Images, Install, Secrets, Backup.

---

## 13. Release channels + help topics

```bash
harness update --channel stable|beta|develop
harness update channels
harness update status

harness help                 # list topics
harness help quickstart agents sensors hooks autonomy mcp update input tracker
```

---

## 14. Validation checklist (tick locally)

- [ ] `harness version` reports v0.20.0+
- [ ] `harness update status` is up-to-date
- [ ] `harness doctor` — every desired CLI ✓
- [ ] `harness doctor --fix --dry-run` lists the plan
- [ ] `harness install list` shows ≥ 17 manifests
- [ ] `harness runtime list` has at least one ✓ runtime
- [ ] `harness containers list` returns a table
- [ ] `harness images list` returns a table
- [ ] `harness secret info` shows ≥ 2 backends
- [ ] `harness mcp templates` lists 7 servers
- [ ] `harness hook templates` lists 5 hooks
- [ ] `harness agent list` shows ≥ 9 adapters
- [ ] `harness feature "..." --agent <id> --apply --autonomy safe_execute` produces `status=applied`
- [ ] High-risk file (Dockerfile / go.mod / .env*) routes to `waiting_approval`
- [ ] `harness runs approve <id>` merges into project root
- [ ] `harness backup snapshot --remote <name>` round-trips with `restore`
- [ ] Dashboard `/runtime`, `/install`, `/secrets`, `/backup` all load
- [ ] `scripts/tests/install_smoke.sh` runs to completion

---

## What is **not** shipped yet (transparent gap list)

- v1.0.0 release ritual
- Apple Container `Run` (Available probe disables it until upstream stabilises)
- SSE on `/api/runs/:id/events` for live Active Run page
- Bundled skill snippets (`harness skill install` planned for v0.21)
- Brew tap `rodolfopeixoto/homebrew-tap` requires the operator to create that repo + commit the generated `Formula/harnessx.rb`
- Windows installer (no `install.ps1` yet — manual unzip works)
- Dashboard mutate-actions (kill/prune/install) stay on the CLI

Updates land via `harness update` on the channel of your choice.
