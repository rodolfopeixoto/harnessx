#!/usr/bin/env bash
# e2e-phase11: workspace registry CRUD end-to-end.
set -euo pipefail
unset GOROOT || true
export HARNESS_LANG=en

cd "$(git rev-parse --show-toplevel)"
make build > /dev/null
BIN="$(pwd)/bin/harness"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
export HARNESS_HOME="$WORK/home"
mkdir -p "$HARNESS_HOME"

APP_A="$WORK/app-a"
APP_B="$WORK/app-b"
mkdir -p "$APP_A" "$APP_B"
(cd "$APP_A" && git init -q)
(cd "$APP_B" && git init -q)

echo "→ add two projects"
"$BIN" project add "$APP_A" --name "App A" --slug app-a > "$WORK/add-a.txt"
"$BIN" project add "$APP_B" --name "App B" --slug app-b > "$WORK/add-b.txt"
grep -q "registered App A" "$WORK/add-a.txt" || { echo "✗ add A"; exit 1; }
grep -q "registered App B" "$WORK/add-b.txt" || { echo "✗ add B"; exit 1; }

echo "→ list shows both"
"$BIN" project list > "$WORK/list.txt"
grep -q "app-a" "$WORK/list.txt" || { echo "✗ list missing app-a"; cat "$WORK/list.txt"; exit 1; }
grep -q "app-b" "$WORK/list.txt" || { echo "✗ list missing app-b"; cat "$WORK/list.txt"; exit 1; }

echo "→ idempotent add"
"$BIN" project add "$APP_A" --name "App A v2" --slug app-a > "$WORK/add-a2.txt"
"$BIN" project list > "$WORK/list2.txt"
[ "$(grep -c "^[* ] app-" "$WORK/list2.txt")" -eq 2 ] || { echo "✗ duplicated"; cat "$WORK/list2.txt"; exit 1; }

echo "→ switch + current"
"$BIN" project switch app-a > "$WORK/switch.txt"
grep -q "active: app-a" "$WORK/switch.txt" || { echo "✗ switch"; exit 1; }
(cd "$WORK" && "$BIN" project current > "$WORK/cur.txt")
grep -q "source: active" "$WORK/cur.txt" || { echo "✗ current source=active"; cat "$WORK/cur.txt"; exit 1; }

echo "→ flag precedence wins over active"
(cd "$WORK" && "$BIN" project current --project app-b > "$WORK/curb.txt")
grep -q "slug:   app-b" "$WORK/curb.txt" || { echo "✗ flag precedence"; cat "$WORK/curb.txt"; exit 1; }
grep -q "source: flag" "$WORK/curb.txt" || { echo "✗ source flag"; cat "$WORK/curb.txt"; exit 1; }

echo "→ archive hides from default list"
"$BIN" project archive app-a > /dev/null
"$BIN" project list > "$WORK/list3.txt"
grep -q "app-a" "$WORK/list3.txt" && { echo "✗ archived shown"; exit 1; }
"$BIN" project list --archived > "$WORK/list4.txt"
grep -q "archived" "$WORK/list4.txt" || { echo "✗ --archived missing flag"; cat "$WORK/list4.txt"; exit 1; }
"$BIN" project unarchive app-a > /dev/null

echo "→ dashboard /api/workspace endpoints"
ADDR="127.0.0.1:17811"
(cd "$APP_A" && HARNESS_HOME="$HARNESS_HOME" "$BIN" init > /dev/null)
cd "$APP_A"
"$BIN" dashboard --addr "$ADDR" > "$WORK/dash.log" 2>&1 &
DASH_PID=$!
trap 'rm -rf "$WORK"; kill $DASH_PID 2>/dev/null || true' EXIT
sleep 2
code=$(curl -sf -o "$WORK/projects.json" -w "%{http_code}" "http://$ADDR/api/workspace/projects")
[ "$code" = "200" ] || { echo "✗ /api/workspace/projects=$code"; cat "$WORK/dash.log"; exit 1; }
grep -q "app-a" "$WORK/projects.json" || { echo "✗ projects payload missing app-a"; cat "$WORK/projects.json"; exit 1; }

code=$(curl -sf -o "$WORK/cur.json" -w "%{http_code}" "http://$ADDR/api/workspace/current")
[ "$code" = "200" ] || { echo "✗ /api/workspace/current=$code"; exit 1; }

kill $DASH_PID 2>/dev/null || true
wait 2>/dev/null || true
echo "✓ e2e-phase11 passed"
