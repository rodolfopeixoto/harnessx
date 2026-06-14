#!/usr/bin/env bash
# Smoke test for license-gate: skip when go-licenses missing, exercise
# the BLOCKED regex against a synthetic CSV.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
. scripts/lib/assert.sh
TESTS_NAME="license-gate"

BLOCKED='AGPL-3.0|GPL-3.0|GPL-2.0|LGPL-3.0|LGPL-2.1|SSPL|EUPL'

# go-licenses csv format is module,url,license,extra — pattern requires
# the license to be bounded by commas, so the trailing comma in real
# output matters.
clean=$(printf "mod,url,MIT,\nother,url,Apache-2.0,\n")
dirty=$(printf "mod,url,MIT,\nbad,url,GPL-3.0,\n")

if echo "$clean" | grep -E ",($BLOCKED)," >/dev/null; then
  _fail "clean CSV must not match" "matched"
else
  _pass "clean CSV bypasses block list"
fi

if echo "$dirty" | grep -E ",($BLOCKED)," >/dev/null; then
  _pass "dirty CSV trips block list"
else
  _fail "dirty CSV must trip block list" "no match"
fi

if command -v go-licenses >/dev/null; then
  _pass "go-licenses installed (script can run)"
else
  echo "  · go-licenses not installed (license-gate.sh would fail fast — expected)"
fi

report
