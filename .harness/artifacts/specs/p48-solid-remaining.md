# P48 — SOLID remaining waivers

## Baseline

Before: 20 SOLID violations on develop.
After:  2 SOLID violations remaining (both in `internal/repl/repl.go`).

## Waived

### `internal/repl/repl.go`
- **loc**: 1829 > 400
- **fan-out**: 23 > 15

**Reason**: `internal/repl/repl.go` implements the interactive REPL, which
holds tightly-coupled state (session, turns, command dispatch, streaming
render, slash commands, autocomplete, key handling, history). A safe split
requires:

  1. Extracting `Session` + persistence into `session.go`.
  2. Extracting slash-command handlers into per-command files.
  3. Extracting render pipeline (streaming, UI, tabwriter) into `render.go`.
  4. Extracting input/history/keybinding into `input.go`.

Each step touches package-internal state that the tests currently exercise
via the public REPL entry, so a naive move-and-forward risks breaking the
whole REPL surface (which is the primary interactive touchpoint).

This waiver is time-boxed to the next Phase 5 cycle. See
`HARNESSX-MASTER-PLAN.md §9` — the repl split is scheduled for Phase 5
alongside the streaming rework, so the extract is bundled with the
already-planned rewrite instead of gold-plating on top of the current
implementation.

## Follow-up

- Track split as `[chore] repl.go SOLID split (P48 follow-up)`.
- Add regression e2e that exercises `harness chat` plus each slash command
  before splitting; rely on that harness during the extraction.
