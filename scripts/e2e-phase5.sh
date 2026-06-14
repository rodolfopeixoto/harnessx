#!/usr/bin/env bash
# Phase 5 end-to-end: context build / inspect, cache hit, --force.
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
"$BIN" project index >/dev/null

echo "→ context build (first run)"
"$BIN" context build "explain main entry point" > /tmp/hx-ctx1.txt
cat /tmp/hx-ctx1.txt
grep -q "^context: built" /tmp/hx-ctx1.txt

echo "→ context build (cache hit)"
"$BIN" context build "explain main entry point" > /tmp/hx-ctx2.txt
cat /tmp/hx-ctx2.txt
grep -q "^context: cache HIT" /tmp/hx-ctx2.txt

echo "→ context build --force"
"$BIN" context build "explain main entry point" --force > /tmp/hx-ctx3.txt
grep -q "^context: built" /tmp/hx-ctx3.txt

echo "→ context inspect (newest)"
"$BIN" context inspect > /tmp/hx-ctx-inspect.txt
grep -q '"task": "explain main entry point"' /tmp/hx-ctx-inspect.txt

test -s .harness/cache/context/*.json

echo "✓ e2e-phase5 passed"
