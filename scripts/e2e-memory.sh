#!/usr/bin/env bash
# Hardening 3 end-to-end: harness memory list/promote.
set -euo pipefail

unset GOROOT || true
export HARNESS_LANG=en

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/bin/harness"

if [[ ! -x "$BIN" ]]; then
  (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/harness)
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cd "$WORK"
git init -q
"$BIN" init >/dev/null

# pull a run id from the bootstrap session
RUN_ID="$(sqlite3 .harness/db/harness.sqlite 'select id from runs order by started_at desc limit 1;')"
test -n "$RUN_ID"

echo "→ promote with evidence"
"$BIN" memory promote --content "tests use rspec, not minitest" \
  --kind convention --scope project \
  --run-id "$RUN_ID" --confidence 0.85 | tee /tmp/hx-mem.txt
grep -q "promoted memory" /tmp/hx-mem.txt

echo "→ list shows the entry"
"$BIN" memory list | tee /tmp/hx-mem-list.txt
grep -q "rspec, not minitest" /tmp/hx-mem-list.txt

echo "→ low-confidence rejected"
set +e
"$BIN" memory promote --content "low" --run-id "$RUN_ID" --confidence 0.1 >/tmp/hx-mem-low.txt 2>&1
rc=$?
set -e
test "$rc" -ne 0

echo "→ missing evidence rejected"
set +e
"$BIN" memory promote --content "no evidence" --confidence 0.9 >/tmp/hx-mem-noev.txt 2>&1
rc=$?
set -e
test "$rc" -ne 0

echo "→ sensitive content rejected"
set +e
"$BIN" memory promote --content "AKIAIOSFODNN7EXAMPLE was leaked" \
  --run-id "$RUN_ID" --confidence 0.9 >/tmp/hx-mem-sec.txt 2>&1
rc=$?
set -e
test "$rc" -ne 0

if command -v sqlite3 >/dev/null; then
  COUNT="$(sqlite3 .harness/db/harness.sqlite 'select count(*) from memories;')"
  test "$COUNT" -eq 1
fi

echo "✓ e2e-memory passed"
