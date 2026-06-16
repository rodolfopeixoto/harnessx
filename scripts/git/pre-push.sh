#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

branch="$(git rev-parse --abbrev-ref HEAD)"
case "$branch" in
  feature/*|hotfix/*|release/*) ;;
  *)
    if [ "${HARNESS_RELEASE:-0}" != "1" ]; then
      echo "✗ pre-push: direct push to $branch forbidden; use feature/* (HARNESS_RELEASE=1 to bypass)" >&2
      exit 1
    fi
    ;;
esac

if [ "${HARNESS_SKIP_CI:-0}" = "1" ]; then
  echo "→ HARNESS_SKIP_CI=1 — skipping local pre-push gate"
  exit 0
fi

echo "→ pre-push: make lint"
make lint || exit 1

echo "→ pre-push: go test ./..."
go test ./... > /tmp/pre-push.tests.log 2>&1 \
  || { tail -30 /tmp/pre-push.tests.log; exit 1; }

if [ -x scripts/coverage-gate.sh ]; then
  echo "→ pre-push: coverage gate"
  bash scripts/coverage-gate.sh > /tmp/pre-push.cov.log 2>&1 \
    || { tail -20 /tmp/pre-push.cov.log; exit 1; }
fi

echo "✓ pre-push gate green"
