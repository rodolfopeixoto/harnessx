# HarnessX — Manual Testing Tutorial (v0.11)

End-to-end walkthrough for every surface shipped between v0.6 and v0.10:

- `harness install` (P36) — tools, LSPs, agent CLIs
- `harness runtime` + `harness containers` + `harness images` (P37, P39)
- `harness secret` + API agent adapters (P38)
- `harness execute --sandbox container` (P39)
- `harness update --channel stable|beta|develop` (P35)
- `harness help <topic>` (P35)
- `harness completion bash|zsh|fish`

If you are upgrading from v0.4 / v0.5, every step in
`docs/tutorial-v0.4-manual.md` still applies; this file adds the new
surfaces on top.

---

## 0. Prerequisites

| Tool | Need |
|---|---|
| git, bash, curl | always |
| go ≥ 1.23 | only for `harness update --channel develop` |
| docker / podman / orbstack / colima / apple container | only for `--sandbox container` and `harness containers` |
| `pdftotext` (poppler) | only for `--pdf` |

A platform-matched harness binary is downloaded for every other section.

---

## 1. Install or update

Fresh install:

```bash
curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh | bash
harness version
```

Already installed (v0.6+):

```bash
harness update                  # latest stable
harness update --channel beta   # opt into pre-releases
harness update --tag v0.10.0    # pin a specific tag
harness update --dry-run        # preview without swapping
harness update status           # is something newer?
harness update channels         # all releases per channel
```

Shell completion:

```bash
harness completion bash > /usr/local/etc/bash_completion.d/harness        # mac
harness completion bash > /etc/bash_completion.d/harness                  # linux
harness completion zsh  > "${fpath[1]}/_harness"
harness completion fish > ~/.config/fish/completions/harness.fish
```

Verify each:

```bash
harness <TAB>
```

---

## 2. Doctor + install missing tools

```bash
harness doctor
```

`go` and `gemini` now show ✓ instead of "present, version probe failed".
A `Recommended installs` section lists one-shot install commands for
every ⚠/✗ entry.

```bash
harness install list                  # all bundled tool manifests
harness install --dry-run gopls       # see the chosen strategy
harness install gopls                 # run it
harness install --upgrade golangci-lint
harness install show ripgrep          # resolved plan only
```

Per platform the installer picks: `brew` → `apt` → `dnf` → `pacman` →
`go install` → `npm -g` → `cargo install` → `pip --user`. First viable
strategy whose binary is on PATH wins.

---

## 3. Pick a container runtime (per project)

```bash
harness runtime list                  # detected runtimes with versions
harness runtime select                # interactive picker
harness runtime set docker            # explicit pin
harness runtime info                  # current pick + source
```

The selection lands in `.harness/config/runtime.yaml`. Resolve order:

```
HARNESS_RUNTIME=<id>  >  .harness/config/runtime.yaml  >  auto-detect
```

Preference per platform:

| OS    | Order |
|---|---|
| macOS | apple_container > docker > orbstack > podman > colima |
| linux | docker > podman > orbstack > colima |

### Cross-runtime container ops

```bash
harness containers list               # running
harness containers list --all         # incl. stopped
harness containers kill <id> [<id>]
HARNESS_CONTAINERS_I_UNDERSTAND=1 harness containers prune --stopped
HARNESS_CONTAINERS_I_UNDERSTAND=1 harness containers prune --older-than 720h
```

### Image ops

```bash
harness images list
HARNESS_CONTAINERS_I_UNDERSTAND=1 harness images prune --older-than 720h
```

Two-key safety: prune always requires either interactive `yes` or
`HARNESS_CONTAINERS_I_UNDERSTAND=1`.

---

## 4. Cross-platform secret store

```bash
harness secret info             # which backends are active
harness secret list             # names only (values redacted)
harness secret set demo-key     # prompts via terminal (hidden)
echo "v" | harness secret set demo-key --from-file /dev/stdin
harness secret set demo-key --from-env MY_ENV_VAR
harness secret get demo-key            # redacted
harness secret get demo-key --reveal   # plaintext
harness secret unset demo-key
```

Backend priority per OS:

| OS    | Order |
|---|---|
| macOS | env > Keychain (`security`) > encrypted file |
| linux | env > Secret Service (`secret-tool`) > encrypted file |

Env reads `HARNESS_SECRET_<UPPER>` first, then `<UPPER>` (so
`ANTHROPIC_API_KEY` is honoured without any wrapper).

---

## 5. API-based agents (no CLI required)

5 bundled API adapters added in v0.8: `anthropic-api`, `openai-api`,
`gemini-api`, `moonshot-api`, `minimax-api`.

```bash
harness agent install anthropic-api
harness agent login anthropic-api --from-env ANTHROPIC_API_KEY
harness agent certify anthropic-api
harness feature "create HELLO_API.md with content: hi" \
  --agent anthropic-api --apply --autonomy safe_execute
```

For CLI adapters (`claude`, `codex`, `gemini`, `kimi`),
`harness agent login <id>` prints the official login command + doc URL
instead of trying to wrap it.

---

## 6. Sandboxed agent execution

Run an agent inside the selected container runtime. Worktree is
bind-mounted read-write at `/work`:

```bash
harness execute "create x.md with content: y" \
  --agent fake-real --apply --autonomy safe_execute \
  --sandbox container --sandbox-image alpine:3.20
```

When `--sandbox container` is set:

- `runtime.Resolve` picks the project runtime
- The worktree is mounted at `/work`
- The agent CLI runs as the container's command
- Container is auto-removed on exit
- Stdout/stderr captured into the run dir as usual

Apple Container is not yet wired for `Run`; pin docker or podman
(`harness runtime set docker`) until that lands.

---

## 7. In-CLI help topics

```bash
harness help                  # list topics
harness help quickstart
harness help agents
harness help sensors
harness help hooks
harness help autonomy
harness help mcp
harness help update
harness help input
harness help tracker
```

---

## 8. Validation checklist

Tick locally before declaring v0.11 ready.

- [ ] `harness version` shows v0.11.0 (or later)
- [ ] `harness update status` reports up-to-date OR offers a newer tag
- [ ] `harness completion bash|zsh|fish` produces a non-empty completion
- [ ] `harness doctor` — go + gemini both ✓; `Recommended installs`
      section lists at least one row when something is missing
- [ ] `harness install list` shows ≥ 16 manifests
- [ ] `harness install --dry-run gopls` prints
      `→ dry-run: go_install: [go install …]`
- [ ] `harness runtime list` shows ≥ 1 ✓ runtime
- [ ] `harness runtime set docker` (or another) persists
      `.harness/config/runtime.yaml`
- [ ] `harness containers list` returns a table
- [ ] `harness images list` returns a table
- [ ] `harness secret info` lists ≥ 2 backends
- [ ] `harness secret set demo-key --from-file …` then
      `harness secret unset demo-key` round-trips
- [ ] `harness agent list` shows ≥ 5 API adapters
      (anthropic-api / openai-api / gemini-api / moonshot-api / minimax-api)
- [ ] `harness execute --sandbox container --agent fake-real` runs a
      transient container against the selected runtime
- [ ] `scripts/tests/install_smoke.sh` runs to completion against a
      clean temp prefix

---

## What is **not** yet shipped (transparent gap list)

These are planned for v0.12 and later, not present today:

- Dashboard pages for runtime / containers / install / secrets / images
- Brew formula + tap
- Windows binaries
- Apple Container `Run` (currently stub)
- v1.0.0 release ritual

Updates land via `harness update` on the channel of your choice.
