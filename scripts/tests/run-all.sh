#!/usr/bin/env bash
# Runs every scripts/tests/test-*.sh. Aggregates pass/fail.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

failed=0
total=0
for t in scripts/tests/test-*.sh; do
  total=$((total + 1))
  echo "── $(basename "$t") ──"
  if ! bash "$t"; then
    failed=$((failed + 1))
  fi
  echo
done

echo "──────────────────────"
if [ "$failed" -eq 0 ]; then
  echo "✓ shell test harness: $total/$total suites passed"
else
  echo "✗ shell test harness: $failed/$total suites failed"
  exit 1
fi
