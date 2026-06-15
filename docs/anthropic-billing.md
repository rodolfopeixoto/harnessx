# Anthropic billing — what HarnessX consumes and how to control it

Anthropic split billing into three streams starting 2026-06-15. HarnessX
calls the Claude CLI in non-interactive mode, which lands in the
**Agent SDK credit** bucket — separate from the subscription that
powers interactive `claude` and the Anthropic API key billed
pay-as-you-go.

This page tells you which adapter draws from which bucket and how to
keep automation predictable.

---

## The three billing streams

| Stream | What triggers it | How HarnessX hits it |
|---|---|---|
| Subscription usage | Interactive `claude` in terminal / IDE, `claude.ai` chat, Claude Cowork | **Never** (HarnessX is always non-interactive) |
| Agent SDK monthly credit | `claude -p` (or `--print`), Claude Agent SDK, claude-code GitHub Action | Default `--agent claude` adapter (`claude --print --output-format json`) |
| Pay-as-you-go API | Calls with an Anthropic API key (`x-api-key` header) | `--agent anthropic-api` adapter |

The Agent SDK credit by plan (refreshes monthly, per user, does not pool):

| Plan | Credit |
|---|---|
| Pro | $20 |
| Max 5x | $100 |
| Max 20x | $200 |
| Team Standard | $20 |
| Team Premium | $100 |
| Enterprise (usage-based) | $20 |
| Enterprise (seat-based Premium) | $200 |

Source: https://support.anthropic.com — "Use the Claude Agent SDK with your Claude plan".

---

## Pick an adapter intentionally

### A — `anthropic-api` (recommended for automation)

Pay-as-you-go via your Anthropic API key. Predictable, separate from
the subscription, no Agent SDK credit cap.

```bash
harness agent install anthropic-api
harness secret set anthropic_api_key       # paste API key
harness agent login anthropic-api --from-env ANTHROPIC_API_KEY
harness agent certify anthropic-api
harness feature "..." --agent anthropic-api --apply
```

Cost per token published at https://www.anthropic.com/pricing. The
adapter's `cost.input_token_price_per_1m` + `output_token_price_per_1m`
are pre-filled at v0.21 rates; bump them when Anthropic updates pricing.

### B — `claude` CLI with the Agent SDK monthly credit

Opt in at https://console.anthropic.com → plan settings, then keep
using `--agent claude`. Enable usage credits there too so requests don't
just stop when the monthly credit drains.

```bash
claude login                               # one-time interactive OAuth
harness agent certify claude
harness feature "..." --agent claude --apply
```

`harness metrics --since 1d` aggregates per-run cost from the JSON
returned by `claude --print --output-format json`; cross-check against
the Anthropic console under "Agent SDK usage".

### C — Hybrid

`claude` interactively for exploration (subscription).
`anthropic-api` for HarnessX automation (pay-as-you-go).

Practical: install both adapters. Pin per call.

---

## How HarnessX tracks cost

Every run that goes through the executor records:

- `meta.json` per run: `input_tokens`, `output_tokens`,
  `estimated_cost_usd`, `exact_usage_available`.
- `harness metrics --since 1d|7d|30d|all` aggregates across runs.
- `harness audit --kind agent` lists per-call events.

`estimated_cost_usd` uses the adapter's declared prices, not Anthropic's
billing console. When `exact_usage_available=true`, the token counts
come from the provider's own `usage` block; otherwise they are
estimated from prompt + output length.

---

## What to do today

1. Decide which bucket you want to spend from:
   - automation-heavy → API key + `anthropic-api`
   - exploration + a bit of automation → `claude` CLI with Agent SDK
     credit opted in
2. Configure the adapter you picked (above).
3. Set a budget cap per run: `harness feature ... --budget-usd 0.50`.
4. Watch `harness metrics --since 1d` after every session.

When `--budget-usd` exceeds, the executor reports `budget_exceeded` and
stops further calls in the same run.
