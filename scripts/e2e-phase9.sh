#!/usr/bin/env bash
# Phase 9 end-to-end: perf-snapshot, perf-compare, image/dep/log/security audits, optimize meta cmd.
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
# inject a Dockerfile + a noisy log line
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
"$BIN" init >/dev/null
"$BIN" project index >/dev/null

echo "→ perf-snapshot baseline"
"$BIN" perf-snapshot --label baseline --report > /tmp/hx-snap1.txt
cat /tmp/hx-snap1.txt
grep -q "^snapshot:" /tmp/hx-snap1.txt
ls .harness/artifacts/perf/*.json >/dev/null

echo "→ image-audit"
"$BIN" image-audit > /tmp/hx-img.txt
cat /tmp/hx-img.txt
grep -q "docker.latest_tag" /tmp/hx-img.txt
grep -q "docker.no_user" /tmp/hx-img.txt

echo "→ dependency-audit"
"$BIN" dependency-audit > /tmp/hx-dep.txt
grep -q "dependencies: total=" /tmp/hx-dep.txt

echo "→ log-audit"
"$BIN" log-audit > /tmp/hx-log.txt
grep -q "noisy.go" /tmp/hx-log.txt

echo "→ security-audit"
"$BIN" security-audit > /tmp/hx-sec.txt
grep -q "summary:" /tmp/hx-sec.txt

echo "→ perf-snapshot after"
"$BIN" perf-snapshot --label after >/dev/null

echo "→ perf-compare"
"$BIN" perf-compare > /tmp/hx-cmp.txt
cat /tmp/hx-cmp.txt
grep -q "^compare:" /tmp/hx-cmp.txt
ls .harness/artifacts/reports/perf-compare-*.md >/dev/null

echo "→ optimize meta"
"$BIN" optimize resources > /tmp/hx-opt.txt
grep -q "Cycle A" /tmp/hx-opt.txt
grep -q "Cycle B" /tmp/hx-opt.txt
grep -q "Cycle D" /tmp/hx-opt.txt

echo "✓ e2e-phase9 passed"
