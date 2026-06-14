#!/usr/bin/env bash
# Phase 8 end-to-end: dashboard server up, REST endpoints respond, shut down.
set -euo pipefail

unset GOROOT || true

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/bin/harness"

if [[ ! -x "$BIN" ]]; then
  (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/harness)
fi

WORK="$(mktemp -d)"
PORT=$(awk 'BEGIN { srand(); print 17000 + int(rand()*5000) }')
ADDR="127.0.0.1:${PORT}"
cleanup() {
  if [[ -n "${PID:-}" ]]; then kill "$PID" 2>/dev/null || true; fi
  rm -rf "$WORK"
}
trap cleanup EXIT

cp -R "$ROOT/testdata/projects/sample-go/." "$WORK/"
cp -R "$ROOT/testdata/designs/sample-design" "$WORK/sample-design"
cd "$WORK"
git init -q
git add -A
git -c user.email=t@t -c user.name=t commit -q -m init
"$BIN" init >/dev/null
"$BIN" project index >/dev/null
"$BIN" design-to-product "convert" --source ./sample-design >/dev/null
"$BIN" check >/dev/null || true

"$BIN" dashboard --addr "$ADDR" > /tmp/hx-dashboard.log 2>&1 &
PID=$!

# wait for server to bind
for i in $(seq 1 40); do
  sleep 0.25
  if curl -sf "http://${ADDR}/api/health" >/dev/null; then break; fi
done

echo "→ /api/health"
curl -sf "http://${ADDR}/api/health" | tee /tmp/hx-health.txt | grep -q '"ok":true'

echo "→ /api/sessions"
curl -sf "http://${ADDR}/api/sessions" | tee /tmp/hx-sessions.txt
grep -q '"Mode"' /tmp/hx-sessions.txt

echo "→ /api/design"
curl -sf "http://${ADDR}/api/design" | grep -q '"pages"'

echo "→ /api/roadmap"
curl -sf "http://${ADDR}/api/roadmap" | grep -q '"MVP 0"'

echo "→ /api/sensors"
curl -sf "http://${ADDR}/api/sensors" >/dev/null

echo "→ / (static fallback)"
curl -sf "http://${ADDR}/" | grep -q "HarnessX"

echo "→ /api/logs?tail=5"
curl -sf "http://${ADDR}/api/logs?tail=5" | grep -q '"lines"'

echo "✓ e2e-phase8 passed"
