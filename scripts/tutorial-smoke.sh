#!/usr/bin/env bash
# Tutorial smoke — exercises the chat-driven e-commerce flow end-to-end
# in a throwaway directory so the docs/TUTORIAL-ECOMMERCE.md promises
# stay honest. Skips anything that requires a real adapter CLI; the
# focus is the harness command surface itself.
#
# Usage:
#   scripts/tutorial-smoke.sh                  # builds + runs in temp
#   HARNESS_BIN=/usr/local/bin/harness scripts/tutorial-smoke.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${HARNESS_BIN:-}"
if [[ -z "${BIN}" ]]; then
    echo "==> building harness binary"
    (cd "${ROOT}" && go build -o /tmp/harness-tutorial-smoke ./cmd/harness)
    BIN=/tmp/harness-tutorial-smoke
fi

WORK="$(mktemp -d -t harness-tutorial-XXXX)"
echo "==> scratch dir ${WORK}"
trap 'rm -rf "${WORK}"' EXIT

cd "${WORK}"

step() { printf "\n--- %s ---\n" "$*"; }

step "harness new python-ecommerce ./shop-api --yes"
"${BIN}" new python-ecommerce ./shop-api --yes
cd shop-api

step "harness scaffold list shows python-ecommerce"
"${BIN}" scaffold list | grep -F "python-ecommerce"

step "harness ci (default + --fast)"
"${BIN}" ci || true                 # may report no python binaries on system
"${BIN}" ci --fast || true

step "harness chat list (empty)"
"${BIN}" chat list | tee /tmp/chat-list-empty.out
grep -F "no sessions yet" /tmp/chat-list-empty.out

step "chat round-trip without an adapter — exit immediately"
printf '/help\n/exit\n' | "${BIN}" chat --no-adapter | tee /tmp/chat-help.out
grep -F "/save" /tmp/chat-help.out
grep -F "/branch" /tmp/chat-help.out
grep -F "/prompts" /tmp/chat-help.out

step "/save labels and survives chat list"
printf 'first plain prompt\n/save labelled-session\n/exit\n' | "${BIN}" chat --no-adapter >/dev/null
"${BIN}" chat list | tee /tmp/chat-list-labelled.out
grep -F "labelled-session" /tmp/chat-list-labelled.out

step "/save-prompt + /prompts round-trip (skip plain-text execution)"
# Seed a plain-text turn into the session log directly so we exercise
# /save-prompt without paying for the deterministic-planner harness do
# loop (which takes ~3 min on an empty scratch project).
SEED_ID="01ZZSEEDFORPROMPTTUTORIAL00"
mkdir -p .harness/sessions
printf '{"time":"2026-06-19T00:00:00Z","input":"add a /widgets endpoint with pytest","action":"chat"}\n' \
    > ".harness/sessions/${SEED_ID}.jsonl"
printf '{"id":"%s","goal":"dev","label":"seed-prompt"}\n' "${SEED_ID}" \
    > ".harness/sessions/${SEED_ID}.meta.json"
printf '/save-prompt add-widgets\n/prompts\n/exit\n' \
    | "${BIN}" chat --no-adapter --resume seed-prompt | tee /tmp/chat-prompts.out
grep -F "add-widgets" /tmp/chat-prompts.out
test -f .harness/prompts/add-widgets.md
grep -F "add a /widgets endpoint" .harness/prompts/add-widgets.md

step "harness session show resolves a chat session by label"
"${BIN}" session show labelled-session | tee /tmp/session-show.out
grep -F "chat session" /tmp/session-show.out
grep -F "labelled-session" /tmp/session-show.out

step "harness ship rejects dirty tree without --allow-dirty"
echo "scratch" >> README.md
set +e
"${BIN}" ship "scratch change" 2>&1 | tee /tmp/ship-dirty.out
set -e
grep -F "working tree dirty" /tmp/ship-dirty.out
# Confirm --allow-dirty would proceed past the precondition; abort
# immediately with --dry-run so we don't enter the do/ci loop.
"${BIN}" ship "scratch change" --dry-run --allow-dirty 2>&1 \
    | tee /tmp/ship-allow.out >/dev/null
grep -F "ship:" /tmp/ship-allow.out

echo
echo "tutorial-smoke OK"
