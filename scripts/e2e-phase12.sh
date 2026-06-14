#!/usr/bin/env bash
# e2e-phase12: capabilities center — discover, plan, install, remove.
set -euo pipefail
unset GOROOT || true
export HARNESS_LANG=en

cd "$(git rev-parse --show-toplevel)"
make build > /dev/null
BIN="$(pwd)/bin/harness"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
PROJ="$WORK/app"
mkdir -p "$PROJ"
(cd "$PROJ" && git init -q)
cp -r templates "$PROJ/"
cd "$PROJ"
"$BIN" init > /dev/null

echo "→ list all kinds"
"$BIN" catalog list > "$WORK/list.txt"
for kind in mcp hook sensor skill context resource plugin; do
  grep -q "^${kind} " "$WORK/list.txt" || { echo "✗ kind $kind missing"; cat "$WORK/list.txt"; exit 1; }
done

echo "→ plan + dry-run"
"$BIN" catalog plan mcp filesystem > "$WORK/plan.txt"
grep -q "create" "$WORK/plan.txt" || { echo "✗ plan missing create"; exit 1; }
"$BIN" catalog install mcp filesystem --dry-run > "$WORK/dry.txt"
grep -q "dry-run" "$WORK/dry.txt" || { echo "✗ dry-run flag"; exit 1; }
[ ! -f .harness/capabilities/mcp/filesystem.yaml ] || { echo "✗ dry-run wrote files"; exit 1; }

echo "→ approval gate without --yes (closed stdin)"
if "$BIN" catalog install mcp filesystem < /dev/null > "$WORK/deny.txt" 2>&1; then
  echo "✗ install must exit non-zero when stdin is closed"; cat "$WORK/deny.txt"; exit 1
fi
[ ! -f .harness/capabilities/mcp/filesystem.yaml ] || { echo "✗ install wrote despite no approval"; exit 1; }

echo "→ install --yes"
"$BIN" catalog install mcp filesystem --yes > "$WORK/install.txt"
grep -q "installed mcp/filesystem" "$WORK/install.txt" || { echo "✗ install message"; cat "$WORK/install.txt"; exit 1; }
[ -f .harness/capabilities/mcp/filesystem.yaml ] || { echo "✗ config not written"; exit 1; }

echo "→ second list reports installed"
"$BIN" catalog list --kind mcp > "$WORK/list2.txt"
grep -q "installed" "$WORK/list2.txt" || { echo "✗ status not updated"; cat "$WORK/list2.txt"; exit 1; }

echo "→ remove"
"$BIN" catalog remove mcp filesystem > "$WORK/rm.txt"
grep -q "removed mcp/filesystem" "$WORK/rm.txt" || { echo "✗ remove message"; exit 1; }
[ ! -f .harness/capabilities/mcp/filesystem.yaml ] || { echo "✗ config still present"; exit 1; }

echo "→ HTTP /api/catalog endpoints"
ADDR="127.0.0.1:17812"
"$BIN" dashboard --addr "$ADDR" > "$WORK/dash.log" 2>&1 &
DASH_PID=$!
trap 'rm -rf "$WORK"; kill $DASH_PID 2>/dev/null || true' EXIT
sleep 2
code=$(curl -sf -o "$WORK/kinds.json" -w "%{http_code}" "http://$ADDR/api/catalog/kinds")
[ "$code" = "200" ] || { echo "✗ /api/catalog/kinds=$code"; exit 1; }
grep -q "mcp" "$WORK/kinds.json" || { echo "✗ kinds missing mcp"; cat "$WORK/kinds.json"; exit 1; }
code=$(curl -sf -o "$WORK/items.json" -w "%{http_code}" "http://$ADDR/api/catalog/items?kind=mcp")
[ "$code" = "200" ] || { echo "✗ /api/catalog/items=$code"; exit 1; }
grep -q "filesystem" "$WORK/items.json" || { echo "✗ items missing filesystem"; cat "$WORK/items.json"; exit 1; }
code=$(curl -sf -o "$WORK/plan.json" -w "%{http_code}" -X POST -H 'Content-Type: application/json' -d '{"kind":"mcp","name":"filesystem"}' "http://$ADDR/api/catalog/plan")
[ "$code" = "200" ] || { echo "✗ /api/catalog/plan=$code"; cat "$WORK/plan.json"; exit 1; }
grep -q "hash" "$WORK/plan.json" || { echo "✗ plan missing hash"; exit 1; }

kill $DASH_PID 2>/dev/null || true
wait 2>/dev/null || true
echo "✓ e2e-phase12 passed"
