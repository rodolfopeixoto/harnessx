#!/usr/bin/env bash
# Phase 1 end-to-end smoke test.
set -euo pipefail

unset GOROOT || true

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/bin/harness"

if [[ ! -x "$BIN" ]]; then
  echo "building harness…"
  (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/harness)
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cd "$WORK"

git init -q

echo "→ harness version"
"$BIN" version | grep -q "harness "

echo "→ harness init"
"$BIN" init
test -f .harness/db/harness.sqlite

echo "→ harness doctor"
DOCTOR_OUT="$("$BIN" doctor --plain || true)"
echo "$DOCTOR_OUT" | grep -q "HarnessX Doctor"
echo "$DOCTOR_OUT" | grep -q "Project"

echo "→ harness logs"
LOGS_OUT="$("$BIN" logs --tail 5)"
echo "$LOGS_OUT" | grep -q '"stage":"init"'

if command -v sqlite3 >/dev/null; then
  COUNT="$(sqlite3 .harness/db/harness.sqlite 'select count(*) from sessions;')"
  test "$COUNT" -eq 1
fi

echo "✓ e2e-phase1 passed"
