# Decomposer scope decision

Status: **`--decompose=llm` deferred indefinitely**. The rule-based
decomposer in `internal/taskgraph` stays as the only decomposer
through the v0.x line. This file pins the rationale.

## What ships

- `internal/taskgraph.Decompose(prompt, Options{})` — pure regex +
  keyword classifier
- 14 task kinds, ~20 rules, deterministic, <1ms per prompt
- Returns `Confidence` 0..1; `harness do` warns on any task with
  `confidence < 0.5`
- The `Options.UseLLM` flag is reserved for a future implementation
  but currently no code path reads it

## What does not ship

`--decompose=llm` would route the prompt through a cheap LLM
(claude-haiku, gemini-flash) when the rule decomposer returns
`KindGeneric` with low confidence. Reasons it stays deferred:

1. **Cost surprise**: every `harness do` would gain a hidden LLM call
   even when the user expects deterministic routing. Hard to budget.
2. **Round-trip latency**: the rule decomposer is <1ms; an LLM call
   is 1–10s. A `route show` that pretends to be a dry-run but
   silently hits the network breaks operator expectations.
3. **Determinism regression**: same prompt may produce different task
   graphs across runs, undermining the v0.32 composability guarantee.
4. **Rule decomposer is good enough**: real dog-food prompts hit a
   high-confidence rule (`scaffold X`, `add Y`, `generate Z`,
   `refactor W`). When they do not, the operator can re-phrase or
   pass `--max-tasks 1` to force a single LLM call through the
   workflow.

## When to reconsider

Re-open when one of these is true:

- Three operators report rule-decomposer mis-classifies their typical
  prompts and re-phrasing is not viable.
- Rule additions hit diminishing returns (>50 rules and accuracy
  plateaus).
- A free local classifier (small model on-device) drops the
  latency/cost objections.

Until then: rules win.

## Mitigations for low-confidence prompts today

- `harness route show "<prompt>"` to preview before paying
- The CONF column in the plan flags any task with `confidence < 0.5`
- Pass `--max-tasks 1` to skip decomposition and run one workflow
  call against the picked adapter
- Pass `--deterministic=false` to force adapter routing even when a
  scaffold matches
