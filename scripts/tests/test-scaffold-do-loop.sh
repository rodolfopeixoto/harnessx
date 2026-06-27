#!/usr/bin/env bash
# Shell smoke test for harness scaffold + harness do + harness route show
# + harness runs prune + harness do --json.
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

WORK="$(mktemp -d /tmp/harness-shell-XXXXX)"
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

echo "── harness scaffold list ──"
check "lists 13 langs" bash -c "$HARNESS scaffold list | tail -n +2 | wc -l | grep -qE '^[[:space:]]*13'"

echo "── harness scaffold show python ──"
check "shows scaffold.yaml" bash -c "$HARNESS scaffold show python | grep -q 'language: python'"

echo "── harness route show ──"
check "rule-based decomposer" bash -c "$HARNESS route show 'scaffold python and add a /healthz endpoint' | grep -q 'scaffold:python'"
check "--json schema_version=1" bash -c "$HARNESS route show 'scaffold python' --json | grep -q '\"schema_version\": 1'"

echo "── harness scaffold apply python ──"
git init -q "$WORK"
check "dry-run lists files" bash -c "$HARNESS scaffold apply python --name demo | grep -q 'app.py'"
check "writes files with --apply" bash -c "$HARNESS scaffold apply python --name demo --apply && test -f app.py"

echo "── harness init ──"
check "init bootstraps .harness" bash -c "$HARNESS init && test -f .harness/config/harness.yaml"
check "hook stub present" bash -c "test -x .harness/hooks/pre-tool-use.sh"

echo "── harness runs prune ──"
mkdir -p .harness/runs/01OLDRUN
echo "{}" > .harness/runs/01OLDRUN/meta.json
touch -t 202401010000 .harness/runs/01OLDRUN
check "dry-run finds old run" bash -c "$HARNESS runs prune --older-than 30d | grep -q '01OLDRUN'"
check "--apply deletes" bash -c "$HARNESS runs prune --older-than 30d --apply && ! test -d .harness/runs/01OLDRUN"

echo "── harness uninstall project ──"
check "wipes .harness" bash -c "$HARNESS uninstall project --yes && ! test -d .harness"

echo
echo "──────────────────────"
echo "shell suite: $pass passed, $fail failed"
[ "$fail" -eq 0 ]
