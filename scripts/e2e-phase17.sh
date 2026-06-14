#!/usr/bin/env bash
set -euo pipefail
unset GOROOT || true
export HARNESS_LANG=en
cd "$(git rev-parse --show-toplevel)"
make build > /dev/null
BIN="$(pwd)/bin/harness"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"; kill $DASH_PID 2>/dev/null || true' EXIT
export HARNESS_HOME="$WORK/home"
mkdir -p "$HARNESS_HOME"

PROJ="$WORK/p"
mkdir -p "$PROJ"
(cd "$PROJ" && git init -q)
cp -r templates "$PROJ/"
cd "$PROJ"
"$BIN" init > /dev/null
"$BIN" project add "$PROJ" --slug demo > /dev/null

echo "→ CLI palette search hits commands + capabilities"
"$BIN" palette search filesystem > "$WORK/cli.txt"
grep -q "filesystem" "$WORK/cli.txt" || { echo "✗ missing filesystem"; cat "$WORK/cli.txt"; exit 1; }
"$BIN" palette search settings > "$WORK/cli2.txt"
grep -q "Open settings" "$WORK/cli2.txt" || { echo "✗ missing builtin command"; cat "$WORK/cli2.txt"; exit 1; }

echo "→ HTTP /api/palette returns same shape"
ADDR="127.0.0.1:17817"
"$BIN" dashboard --addr "$ADDR" > "$WORK/dash.log" 2>&1 &
DASH_PID=$!
sleep 2
code=$(curl -sf -o "$WORK/pal.json" -w "%{http_code}" "http://$ADDR/api/palette?q=settings")
[ "$code" = "200" ] || { echo "✗ /api/palette=$code"; cat "$WORK/dash.log"; exit 1; }
grep -q "router_path" "$WORK/pal.json" || { echo "✗ shape"; cat "$WORK/pal.json"; exit 1; }

kill $DASH_PID 2>/dev/null || true
wait 2>/dev/null || true
echo "✓ e2e-phase17 passed"
