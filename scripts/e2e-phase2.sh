#!/usr/bin/env bash
# Phase 2 end-to-end smoke test for `harness project index|inspect`.
set -euo pipefail

unset GOROOT || true

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/bin/harness"

if [[ ! -x "$BIN" ]]; then
  echo "building harness…"
  (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/harness)
fi

run_one() {
  local fixture="$1"
  local label="$2"
  local work
  work="$(mktemp -d)"
  trap 'rm -rf "$work"' RETURN
  cp -R "$ROOT/testdata/projects/$fixture/." "$work/"
  cd "$work"
  git init -q
  "$BIN" init >/dev/null

  echo "→ [$label] harness project index"
  "$BIN" project index | tee /tmp/hx-index-out.txt
  for map in profile commands dependencies architecture test-map api-map design-system performance-budget; do
    test -s ".harness/project/${map}.json" || { echo "missing ${map}.json"; exit 1; }
  done

  echo "→ [$label] inspect profile"
  "$BIN" project inspect profile | grep -q '"stacks"'

  echo "→ [$label] incremental run skips everything"
  out2="$("$BIN" project index)"
  echo "$out2" | grep -q "skipped"
  echo "$out2" | grep -qv "updated:"

  echo "→ [$label] --force rebuilds every map"
  out3="$("$BIN" project index --force)"
  echo "$out3" | grep -q "updated:"

  cd - >/dev/null
}

run_one sample-go    "go"
run_one sample-rails "rails"
run_one sample-react "react"

echo "✓ e2e-phase2 passed"
