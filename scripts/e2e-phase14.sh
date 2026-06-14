#!/usr/bin/env bash
set -euo pipefail
unset GOROOT || true
cd "$(git rev-parse --show-toplevel)"

echo "→ Go coverage gate honours floor"
GLOBAL_MIN=10 CORE_MIN=10 bash scripts/coverage-gate.sh > /tmp/cv-low.log 2>&1
grep -q "coverage gate passed" /tmp/cv-low.log || { echo "✗ low-floor pass"; tail -10 /tmp/cv-low.log; exit 1; }

echo "→ Go coverage gate rejects when floor too high"
if GLOBAL_MIN=99 CORE_MIN=10 bash scripts/coverage-gate.sh > /tmp/cv-high.log 2>&1; then
  echo "✗ high-floor must fail"; tail -10 /tmp/cv-high.log; exit 1
fi
grep -q "global coverage" /tmp/cv-high.log || { echo "✗ output shape"; tail -10 /tmp/cv-high.log; exit 1; }

echo "→ shell coverage gate present + executable"
[ -x scripts/coverage-shell.sh ] || { echo "✗ coverage-shell.sh missing"; exit 1; }
bash scripts/coverage-shell.sh > /tmp/sh-cov.log 2>&1 || true
grep -q "shell coverage gate passed\|advisory: skipping shell coverage gate" /tmp/sh-cov.log \
  || { echo "✗ shell gate output shape"; tail -10 /tmp/sh-cov.log; exit 1; }

echo "→ web coverage gate script present + executable"
[ -x scripts/coverage-web.sh ] || { echo "✗ coverage-web.sh missing"; exit 1; }

echo "✓ e2e-phase14 passed"
