#!/usr/bin/env bash
set -euo pipefail
unset GOROOT || true
export HARNESS_LANG=en

cd "$(git rev-parse --show-toplevel)"
make build > /dev/null
BIN="$(pwd)/bin/harness"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"; pkill -f hx-p16 2>/dev/null || true' EXIT
export HARNESS_HOME="$WORK/home"
mkdir -p "$HARNESS_HOME"

PROJ="$WORK/demo"
mkdir -p "$PROJ"
(cd "$PROJ" && git init -q)
echo '{"name":"demo"}' > "$PROJ/package.json"

echo "→ wizard imports project + fingerprints"
"$BIN" project import "$PROJ" --name Demo --slug demo --yes > "$WORK/import.txt"
grep -q "imported Demo" "$WORK/import.txt" || { echo "✗ import message"; cat "$WORK/import.txt"; exit 1; }
[ -f "$PROJ/.harness/project/fingerprints.json" ] || { echo "✗ fingerprints missing"; exit 1; }

echo "→ stale empty right after import"
"$BIN" project stale "$PROJ" > "$WORK/stale.txt"
grep -q "no stale files" "$WORK/stale.txt" || { echo "✗ expected no stale"; cat "$WORK/stale.txt"; exit 1; }

echo "→ mutate package.json triggers stale"
echo '{"name":"demo","version":"0.2.0"}' > "$PROJ/package.json"
"$BIN" project stale "$PROJ" > "$WORK/stale2.txt"
grep -q "package.json" "$WORK/stale2.txt" || { echo "✗ expected stale package.json"; cat "$WORK/stale2.txt"; exit 1; }

echo "→ HTTP /api/workspace/stale/<slug>"
ADDR="127.0.0.1:17816"
cd "$PROJ"
"$BIN" init > /dev/null
"$BIN" dashboard --addr "$ADDR" > "$WORK/dash.log" 2>&1 &
DASH_PID=$!
sleep 2
code=$(curl -sf -o "$WORK/stale.json" -w "%{http_code}" "http://$ADDR/api/workspace/stale/demo" || echo FAIL)
[ "$code" = "200" ] || { echo "✗ stale endpoint=$code"; cat "$WORK/dash.log"; kill $DASH_PID 2>/dev/null; exit 1; }
grep -q "package.json" "$WORK/stale.json" || { echo "✗ payload missing package.json"; cat "$WORK/stale.json"; kill $DASH_PID 2>/dev/null; exit 1; }
kill $DASH_PID 2>/dev/null || true
wait 2>/dev/null || true

echo "✓ e2e-phase16 passed"
