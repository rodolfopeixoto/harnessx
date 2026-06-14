#!/usr/bin/env bash
set -euo pipefail
unset GOROOT || true
export HARNESS_LANG=en
cd "$(git rev-parse --show-toplevel)"
make build > /dev/null
BIN="$(pwd)/bin/harness"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"; kill $DASH_PID 2>/dev/null || true' EXIT
PROJ="$WORK/p"
mkdir -p "$PROJ"
(cd "$PROJ" && git init -q)
cd "$PROJ"
"$BIN" init > /dev/null

echo "→ autonomy get prints all 5 levels"
"$BIN" autonomy get > "$WORK/aut.txt"
for level in manual plan_and_ask safe_execute full_project_loop scheduled_maintenance; do
  grep -q "$level" "$WORK/aut.txt" || { echo "✗ missing $level"; cat "$WORK/aut.txt"; exit 1; }
done

echo "→ autonomy set validates unknown level"
if "$BIN" autonomy set definitely-not-a-level > "$WORK/bad.txt" 2>&1; then
  echo "✗ unknown level must fail"; cat "$WORK/bad.txt"; exit 1
fi

echo "→ health show prints subsystems"
"$BIN" health show > "$WORK/h.txt"
for sub in tests sensors security perf deps docs design_parity roadmap_readiness memory_freshness configs; do
  grep -q "$sub" "$WORK/h.txt" || { echo "✗ missing subsystem $sub"; cat "$WORK/h.txt"; exit 1; }
done
grep -q "score:" "$WORK/h.txt" || { echo "✗ score line missing"; exit 1; }

echo "→ HTTP /api/autonomy + /api/health"
ADDR="127.0.0.1:17818"
"$BIN" dashboard --addr "$ADDR" > "$WORK/dash.log" 2>&1 &
DASH_PID=$!
sleep 2
code=$(curl -sf -o "$WORK/aut.json" -w "%{http_code}" "http://$ADDR/api/autonomy")
[ "$code" = "200" ] || { echo "✗ /api/autonomy=$code"; cat "$WORK/dash.log"; exit 1; }
grep -q "manual" "$WORK/aut.json" || { echo "✗ payload missing manual"; cat "$WORK/aut.json"; exit 1; }
code=$(curl -sf -o "$WORK/h.json" -w "%{http_code}" "http://$ADDR/api/health/score")
[ "$code" = "200" ] || { echo "✗ /api/health=$code"; cat "$WORK/dash.log"; exit 1; }
grep -q "subsystems" "$WORK/h.json" || { echo "✗ subsystems missing"; cat "$WORK/h.json"; exit 1; }

kill $DASH_PID 2>/dev/null || true
wait 2>/dev/null || true
echo "✓ e2e-phase18 passed"
