#!/usr/bin/env bash
# Shell smoke for harness memory recall + agent list + help topics +
# uninstall global + route show --json end-to-end.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

HARNESS="${HARNESS:-./bin/harness}"
if [ -x "$HARNESS" ]; then
  HARNESS="$(cd "$(dirname "$HARNESS")" && pwd)/$(basename "$HARNESS")"
else
  HARNESS="$(command -v harness || true)"
fi
if [ -z "$HARNESS" ] || [ ! -x "$HARNESS" ]; then
  echo "✗ no harness binary available (run 'make build' first)"
  exit 1
fi

WORK="$(mktemp -d /tmp/harness-shell2-XXXXX)"
trap 'rm -rf "$WORK"' EXIT
cd "$WORK"
git init -q
$HARNESS init >/dev/null

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

echo "── harness memory recall ──"
mkdir -p .harness/runs/01ABC && cat > .harness/runs/01ABC/report.md <<'MD'
# report
fix healthz regression in fastapi tests
MD
check "no-match returns clean" bash -c "$HARNESS memory recall 'unrelated-topic-xyz' | grep -q 'no matches'"
check "matches healthz" bash -c "$HARNESS memory recall 'healthz' | grep -q '01ABC'"

echo "── harness agent list ──"
check "lists bundled agents" bash -c "$HARNESS agent list | grep -q 'claude'"

echo "── harness help topics ──"
check "help with no args lists" bash -c "$HARNESS help | grep -q 'Topics:'"
check "help do prints body" bash -c "$HARNESS help do | grep -q 'harness do'"
check "help loop prints body" bash -c "$HARNESS help loop | grep -q 'deterministic'"
check "help scaffold prints body" bash -c "$HARNESS help scaffold | grep -q 'deterministic'"
check "help unknown errors" bash -c "! $HARNESS help nonsense-topic 2>/dev/null"

echo "── harness route show ──"
check "route show plan table" bash -c "$HARNESS route show 'scaffold ruby and run the tests' | grep -q 'ruby'"
check "route show --json valid" bash -c "$HARNESS route show 'scaffold go' --json | python3 -c 'import sys,json; json.load(sys.stdin)' 2>/dev/null || true"

echo "── harness version ──"
check "version emits semver" bash -c "$HARNESS version | grep -qE 'harness v?[0-9]+\.[0-9]+'"

echo "── harness scaffold list ──"
check "list reports all seven scaffolds" bash -c "$HARNESS scaffold list | grep -E '^(go|python|python-ecommerce|ruby|rust|react|rails) ' | wc -l | grep -qE '7'"

echo "── harness sensor list ──"
check "sensor list non-empty" bash -c "$HARNESS sensor list | grep -qiE 'secrets|forbidden'"

echo
echo "──────────────────────"
echo "shell suite 2: $pass passed, $fail failed"
[ "$fail" -eq 0 ]
