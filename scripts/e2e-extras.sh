#!/usr/bin/env bash
# Extras e2e — covers the post-Phase-10 commands that don't have a
# dedicated phase script: explain, session show, artifact ls/cat, routes,
# spec init, skill list, completion.
set -euo pipefail

unset GOROOT || true

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/bin/harness"

if [[ ! -x "$BIN" ]]; then
  (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/harness)
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cd "$WORK"
git init -q
git -c user.email=t@t -c user.name=t commit --allow-empty -qm init
"$BIN" init >/dev/null
"$BIN" project index >/dev/null
"$BIN" feature "add greet function" --yes >/dev/null

echo "→ explain"
"$BIN" explain "create product search" | tee /tmp/hx-x-ex.txt
grep -q "^Intent: " /tmp/hx-x-ex.txt
grep -q "^Routed task: " /tmp/hx-x-ex.txt
grep -q "^Chain: " /tmp/hx-x-ex.txt

echo "→ routes"
"$BIN" routes implementation | tee /tmp/hx-x-ro.txt
grep -q "implementation" /tmp/hx-x-ro.txt
grep -q "resolved chain:" /tmp/hx-x-ro.txt

echo "→ session show <id>"
SID="$(sqlite3 .harness/db/harness.sqlite 'select id from sessions order by started_at desc limit 1;')"
test -n "$SID"
"$BIN" session show "$SID" | tee /tmp/hx-x-se.txt
grep -q "^session " /tmp/hx-x-se.txt
grep -q "runs:" /tmp/hx-x-se.txt

echo "→ artifact ls"
"$BIN" artifact ls | tee /tmp/hx-x-art.txt
grep -q "MTIME" /tmp/hx-x-art.txt
grep -qE "specs/.+\.md" /tmp/hx-x-art.txt

echo "→ artifact cat"
SPEC_REL="$(ls .harness/artifacts/specs | head -1)"
"$BIN" artifact cat "specs/${SPEC_REL}" > /tmp/hx-x-cat.txt
head -3 /tmp/hx-x-cat.txt | grep -q "Spec:"

echo "→ spec init"
"$BIN" spec init --name oss-readiness "MFA scaffold" | grep -q "wrote spec:"
test -s .harness/artifacts/specs/oss-readiness.md

echo "→ skill list (empty)"
"$BIN" skill list | grep -q "no skill versions"

echo "→ completion bash + zsh"
"$BIN" completion bash > /tmp/hx-x-comp-b.sh
head -1 /tmp/hx-x-comp-b.sh | grep -q "bash completion for harness"
"$BIN" completion zsh > /tmp/hx-x-comp-z.sh
head -1 /tmp/hx-x-comp-z.sh | grep -q "compdef harness"

if command -v sqlite3 >/dev/null; then
  COUNT=$(sqlite3 .harness/db/harness.sqlite 'select count(*) from sessions;')
  test "$COUNT" -ge 3
fi

echo "✓ e2e-extras passed"
