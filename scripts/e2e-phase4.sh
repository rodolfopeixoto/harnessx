#!/usr/bin/env bash
# Phase 4 end-to-end: sensor list/run, check, ci, forbidden-file detection.
set -euo pipefail

unset GOROOT || true

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/bin/harness"

if [[ ! -x "$BIN" ]]; then
  (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/harness)
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cp -R "$ROOT/testdata/projects/sample-go/." "$WORK/"
cd "$WORK"
git init -q
git add -A
git -c user.email=t@t -c user.name=t commit -q -m init
"$BIN" init >/dev/null

echo "→ sensor list"
"$BIN" sensor list | tee /tmp/hx-sensor-list.txt
grep -q "^forbidden_files" /tmp/hx-sensor-list.txt
grep -q "^go_vet" /tmp/hx-sensor-list.txt

echo "→ harness check (clean project)"
"$BIN" check | tee /tmp/hx-check.txt
grep -q "summary:" /tmp/hx-check.txt
# Forbidden + go_vet must pass on a clean go fixture.
grep -q '\[✓\] forbidden_files' /tmp/hx-check.txt
grep -q '\[✓\] go_vet' /tmp/hx-check.txt

if command -v sqlite3 >/dev/null; then
  COUNT="$(sqlite3 .harness/db/harness.sqlite 'select count(*) from sensor_results;')"
  test "$COUNT" -ge 4
fi

echo "→ inject .env and ensure forbidden_files fails"
printf 'SECRET=abc\n' > .env
set +e
"$BIN" ci > /tmp/hx-ci.txt
code=$?
set -e
cat /tmp/hx-ci.txt
test "$code" -ne 0
grep -q '\[✗\] forbidden_files' /tmp/hx-ci.txt

echo "→ targeted sensor run"
rm .env
"$BIN" sensor run forbidden_files > /tmp/hx-targeted.txt
grep -q '\[✓\] forbidden_files' /tmp/hx-targeted.txt

echo "✓ e2e-phase4 passed"
