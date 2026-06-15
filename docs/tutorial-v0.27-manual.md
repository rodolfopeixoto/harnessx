> **Superseded by `docs/tutorial-v0.28-manual.md`.** The v0.28 manual
> covers every surface end-to-end against a real FastAPI sample with
> all UX defects from v0.27 fixed.

# HarnessX — Manual Testing Tutorial (v0.27)

End-to-end walkthrough for everything shipped between v0.12 and v0.27.
If upgrading from older releases, `docs/tutorial-v0.4-manual.md` and
`docs/tutorial-v0.11-manual.md` still apply; this file layers the new
surfaces on top.

What you will exercise:

- `claude-interactive` adapter (v0.27, P61) — subscription-billed
  Claude Code via PTY / tmux / iTerm2
- Anthropic billing model (v0.26, P60) — three streams, which adapter
  hits which bucket
- `harness agent certify --simple-timeout` (v0.25, P59) and the new
  per-check `what:` / `fix:` lines (v0.24, P58)
- Probe runtime-error guard (v0.23, P57)
- Bundled adapters: `claude`, `anthropic-api`, `claude-interactive`
- `harness agent list` with the new `EXP` column
- `harness help billing`

---

## 0. Prerequisites

| Tool | Need |
|---|---|
| git, bash, curl | always |
| `claude` (Claude Code CLI ≥ recent) | every section involving Claude |
| `tmux` | only for the tmux strategy |
| iTerm2 + macOS | only for the iterm2 strategy |
| go ≥ 1.23 | only if building from source |

```bash
harness --version            # expect 0.27.0
harness help                 # confirms topic list now includes billing
```

If `harness --version` shows older, update first:

```bash
harness update --channel stable
```

---

## 1. Billing primer (read once, decide once)

```bash
harness help billing
```

You should see three buckets and the adapter mapping:

```
--agent claude               -> Agent SDK monthly credit ($20-$200, as of 2026-06-15)
--agent claude-interactive   -> Pro/Max subscription bucket (experimental)
--agent anthropic-api        -> pay-as-you-go API
```

Full breakdown: `docs/anthropic-billing.md`.

Pick the bucket you want to spend from. The rest of this tutorial
assumes you want to test all three.

---

## 2. Install and certify the stable Agent SDK adapter

```bash
harness agent install claude
claude login                                   # one-time OAuth
harness agent certify claude --simple-timeout 180s
```

Expected: status `ready` or `usable`. Each check now prints `what:`
(one-sentence description) and, on failure, a `fix:` line with the
exact next command. If you see `signal: killed`, the remediation now
suggests `--simple-timeout 180s` plus a manual smoke command.

```bash
harness feature "create HELLO-sdk.md with content: hi via Agent SDK" \
  --agent claude --apply --autonomy safe_execute
```

Anthropic console → Agent SDK usage should tick up.

---

## 3. Install and certify the pay-as-you-go API adapter

```bash
harness agent install anthropic-api
harness secret set anthropic_api_key           # paste API key
harness agent login anthropic-api --from-env ANTHROPIC_API_KEY
harness agent certify anthropic-api --simple-timeout 180s
harness feature "create HELLO-api.md with content: hi via API key" \
  --agent anthropic-api --apply --autonomy safe_execute
```

Anthropic console → API usage should tick up. Agent SDK credit
unchanged.

---

## 4. Install the subscription-billed interactive adapter (experimental)

This is the new v0.27 surface. Drives the Claude Code REPL
programmatically.

```bash
harness agent install claude-interactive
harness agent list
```

You should see a row like:

```
ID                NAME                          CERT       EXP  SOURCE
claude-interactive Claude Code (interactive REPL) —         ★    bundled:claude-interactive.yaml
```

The `★` in the EXP column flags experimental status.

### 4a. PTY strategy (default)

No extra config required.

```bash
harness agent certify claude-interactive --simple-timeout 180s
```

Final line of the output should read:

```
experimental — REPL surface is undocumented; may break on Claude Code upgrades
```

Real run:

```bash
harness feature "create HELLO-sub.md with content: hi via subscription" \
  --agent claude-interactive --apply --autonomy safe_execute
```

Verify in the Anthropic console:

- Subscription usage tick goes up
- Agent SDK credit unchanged
- API usage unchanged

### 4b. tmux strategy (opt-in)

```bash
which tmux || brew install tmux
harness agent add claude-interactive           # copies bundled YAML to .harness/config/agents/
$EDITOR .harness/config/agents/claude-interactive.yaml
# change:  strategy: pty   →   strategy: tmux
harness agent certify claude-interactive --simple-timeout 180s
harness feature "create HELLO-tmux.md with content: hi via tmux" \
  --agent claude-interactive --apply --autonomy safe_execute
tmux ls                                        # session "harness-claude-interactive" should appear
```

### 4c. iTerm2 strategy (macOS only)

```bash
# .harness/config/agents/claude-interactive.yaml
# change:  strategy: tmux   →   strategy: iterm2
harness agent certify claude-interactive --simple-timeout 180s
harness feature "create HELLO-iterm.md with content: hi via iTerm2" \
  --agent claude-interactive --apply --autonomy safe_execute
```

A new iTerm2 window opens, the REPL launches, the prompt is keyed in,
the response is captured via osascript. Non-darwin systems will see a
clear `iterm2 strategy is macOS-only` error from the validator.

---

## 5. Switch between adapters per call

```bash
harness feature "explore: list files in src/" --agent claude --autonomy plan_only
harness feature "explore: list files in src/" --agent anthropic-api --autonomy plan_only
harness feature "explore: list files in src/" --agent claude-interactive --autonomy plan_only
```

Same prompt, three buckets billed. Check `harness metrics --since 1d`:

```bash
harness metrics --since 1d
```

`claude-interactive` rows show `mode: estimated` for token counts (the
REPL emits no usage block; counts are estimated from prompt + output
length).

---

## 6. Budget caps still apply

```bash
harness feature "expensive prompt" --agent anthropic-api \
  --apply --budget-usd 0.10
```

When the per-run cap exceeds, the executor reports `budget_exceeded`
and stops further calls. Works identically across all three adapters.

---

## 7. Verify the probe runtime-error guard

The doctor probe no longer mistakes a runtime error path like
`cannot find GOROOT directory: /Users/.../go1.19.2` for a valid
semver. Trigger a real-world misconfiguration to confirm:

```bash
GOROOT=/tmp/does-not-exist go version || true
harness doctor                                 # the Go row should NOT show a check + version
```

---

## 8. Cleanup

```bash
# project YAML overrides created above
rm -f .harness/config/agents/claude-interactive.yaml

# tmux session (if you used the tmux strategy)
tmux kill-session -t harness-claude-interactive 2>/dev/null || true
```

---

## What you proved

- All three Anthropic billing streams are reachable from HarnessX, one
  adapter per stream.
- `claude-interactive` ships, certifies, and runs against the
  subscription bucket with three orchestration strategies.
- Experimental status is visible at install (`★`) and at certify (final
  disclaimer line).
- Certify output now self-explains each check (`what:`) and tells you
  exactly how to fix failures (`fix:`).
- Budget caps + metrics + audit logging stay consistent across
  adapters.

Next: read `docs/anthropic-billing.md` for the per-plan credit table
and pricing-page cross-check link, and `harness help billing` for the
in-CLI summary.
