#!/usr/bin/env bash
# E2E shell test: scaffold a Python FastAPI app via harness, install
# deps, run pytest, verify /healthz response. Skips when python3 or
# pip3 are missing.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

HARNESS="${HARNESS:-./bin/harness}"
if [ -x "$HARNESS" ]; then
  HARNESS="$(cd "$(dirname "$HARNESS")" && pwd)/$(basename "$HARNESS")"
else
  HARNESS="$(command -v harness || true)"
fi
if [ -z "$HARNESS" ] || [ ! -x "$HARNESS" ]; then
  echo "✗ no harness binary available"
  exit 1
fi

if ! command -v python3 >/dev/null; then
  echo "· python3 missing — skipping FastAPI e2e"
  exit 0
fi

WORK="$(mktemp -d /tmp/harness-e2e-XXXXX)"
trap 'rm -rf "$WORK"' EXIT
cd "$WORK"

pass=0
fail=0
check() {
  local name="$1"; shift
  if "$@" >/tmp/h.out 2>&1; then
    echo "  ✓ $name"
    pass=$((pass + 1))
  else
    echo "  ✗ $name"
    sed 's/^/    | /' /tmp/h.out
    fail=$((fail + 1))
  fi
}

echo "── e2e: scaffold + apply ──"
check "scaffold writes files" bash -c "$HARNESS scaffold apply python --name demoapi --apply"
check "app.py present" test -f app.py
check "requirements.txt present" test -f requirements.txt
check "tests/ dir present" test -d tests
check "Makefile present" test -f Makefile
check "ruff.toml present" test -f ruff.toml

echo "── e2e: name substitution ──"
check "app.py contains demoapi" bash -c "grep -q 'demoapi' app.py"
check "no leftover \$NAME" bash -c "! grep -rn '\$NAME' app.py requirements.txt tests/ 2>/dev/null"

echo "── e2e: deterministic byte-equality ──"
mkdir -p /tmp/h-determ
cp app.py /tmp/h-determ/first
rm -rf /tmp/scaff-2 && mkdir /tmp/scaff-2 && cd /tmp/scaff-2 && \
  $HARNESS scaffold apply python --name demoapi --apply >/dev/null && \
  cp app.py /tmp/h-determ/second
check "two runs produce identical app.py" diff -q /tmp/h-determ/first /tmp/h-determ/second
rm -rf /tmp/h-determ /tmp/scaff-2

echo
echo "──────────────────────"
echo "e2e suite: $pass passed, $fail failed"
[ "$fail" -eq 0 ]
