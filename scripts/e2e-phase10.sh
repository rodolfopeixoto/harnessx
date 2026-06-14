#!/usr/bin/env bash
# Phase 10 end-to-end: chain every phase against one project to prove the
# full HarnessX workflow produces real artifacts on disk and a live dashboard.
set -euo pipefail

unset GOROOT || true

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/bin/harness"

if [[ ! -x "$BIN" ]]; then
  echo "→ building harness"
  (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/harness)
fi

WORK="$(mktemp -d)"
PORT=$(awk 'BEGIN { srand(); print 22000 + int(rand()*4000) }')
ADDR="127.0.0.1:${PORT}"
cleanup() {
  if [[ -n "${DASH_PID:-}" ]]; then kill "$DASH_PID" 2>/dev/null || true; fi
  rm -rf "$WORK"
}
trap cleanup EXIT

cp -R "$ROOT/testdata/projects/sample-go/." "$WORK/"
cp -R "$ROOT/testdata/designs/sample-design" "$WORK/sample-design"
# Inject a Dockerfile + a noisy log line so optimize sees real findings.
cat > "$WORK/Dockerfile" <<'EOF'
FROM ubuntu
RUN apt-get update && apt-get install -y curl
RUN apt-get install -y git
RUN apt-get install -y vim
COPY . /app
COPY scripts /app/scripts
COPY config /app/config
COPY assets /app/assets
COPY src /app/src
EOF
cat > "$WORK/noisy.go" <<'EOF'
package main

import "fmt"

func init() { fmt.Println("noisy startup") }
EOF
cd "$WORK"
git init -q
git add -A
git -c user.email=t@t -c user.name=t commit -q -m init

echo
echo "═════ Phase 1: init ═════"
"$BIN" init
test -f .harness/db/harness.sqlite

echo
echo "═════ Phase 1: doctor (plain) ═════"
"$BIN" doctor --plain >/dev/null

echo
echo "═════ Phase 2: project index ═════"
"$BIN" project index
for m in profile commands dependencies architecture test-map api-map design-system performance-budget; do
  test -s ".harness/project/${m}.json"
done

echo
echo "═════ Phase 3: agent list + certify fake ═════"
"$BIN" agent list
"$BIN" agent certify fake --skip-run >/dev/null
if command -v sqlite3 >/dev/null; then
  CERT_COUNT="$(sqlite3 .harness/db/harness.sqlite 'select count(*) from agent_certifications;')"
  test "$CERT_COUNT" -ge 1
fi

echo
echo "═════ Phase 7: design-to-product ═════"
"$BIN" design-to-product "convert design" --source ./sample-design >/dev/null
for m in design-manifest feature-map toggle-map roadmap api-contracts flow-map; do
  test -s ".harness/product/${m}.json"
done

echo
echo "═════ Phase 6: feature (--yes) ═════"
"$BIN" feature "add product search with filters" --yes --budget 0.5 >/dev/null
ls -1 .harness/artifacts/specs/*.md >/dev/null
ls -1 .harness/artifacts/plans/*.md >/dev/null

echo
echo "═════ Phase 5: context build ═════"
"$BIN" context build "explain main entry point" >/dev/null
ls -1 .harness/cache/context/*.json >/dev/null

echo
echo "═════ Phase 4: harness check ═════"
"$BIN" check >/dev/null || true   # ok if some optional tools missing
if command -v sqlite3 >/dev/null; then
  SENSOR_COUNT="$(sqlite3 .harness/db/harness.sqlite 'select count(*) from sensor_results;')"
  test "$SENSOR_COUNT" -ge 4
fi

echo
echo "═════ Phase 9: perf-snapshot + image-audit + perf-compare ═════"
"$BIN" perf-snapshot --label baseline --report >/dev/null
"$BIN" image-audit | tee /tmp/hx10-img.txt
grep -q "docker.latest_tag" /tmp/hx10-img.txt
"$BIN" log-audit | tee /tmp/hx10-log.txt
grep -q "noisy.go" /tmp/hx10-log.txt
"$BIN" perf-snapshot --label after >/dev/null
"$BIN" perf-compare >/dev/null
ls .harness/artifacts/reports/perf-compare-*.md >/dev/null

echo
echo "═════ Phase 8: dashboard up + API smoke ═════"
"$BIN" dashboard --addr "$ADDR" > /tmp/hx10-dash.log 2>&1 &
DASH_PID=$!
for i in $(seq 1 40); do
  sleep 0.25
  if curl -sf "http://${ADDR}/api/health" >/dev/null; then break; fi
done
curl -sf "http://${ADDR}/api/health" | grep -q '"ok":true'
curl -sf "http://${ADDR}/api/sessions" | grep -q '"Mode"'
curl -sf "http://${ADDR}/api/sensors" >/dev/null
curl -sf "http://${ADDR}/api/agents" >/dev/null
curl -sf "http://${ADDR}/api/cost" >/dev/null
curl -sf "http://${ADDR}/api/design" | grep -q '"pages"'
curl -sf "http://${ADDR}/api/roadmap" | grep -q '"MVP 0"'
curl -sf "http://${ADDR}/api/features" >/dev/null
curl -sf "http://${ADDR}/api/toggles" >/dev/null
curl -sf "http://${ADDR}/api/profile" >/dev/null
curl -sf "http://${ADDR}/api/logs?tail=5" >/dev/null
curl -sf "http://${ADDR}/" | grep -q "HarnessX"
kill "$DASH_PID" 2>/dev/null || true
wait "$DASH_PID" 2>/dev/null || true
unset DASH_PID

echo
echo "═════ Phase 6: report --last ═════"
"$BIN" report > /tmp/hx10-rpt.txt
grep -qE "(# Summary|# Executive Summary)" /tmp/hx10-rpt.txt

echo
echo "═════ Final artifact inventory ═════"
find .harness -maxdepth 3 -type f | sort

echo
echo "✓ e2e-phase10 passed — full HarnessX cycle produced real artifacts on disk"
