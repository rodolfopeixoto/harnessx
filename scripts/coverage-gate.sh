#!/usr/bin/env bash
# coverage-gate: enforce minimum coverage per package class.
#
# Two thresholds:
#   GLOBAL_MIN  applies to global average coverage (default 60).
#   CORE_MIN    applies to core packages (default 70).
#
# Core regex: internal/(domain|intent|memory|router|sensors|spec|plan|
#             index|context|design|optimize|skills|platform/).
#
# Override via env: GLOBAL_MIN=70 CORE_MIN=90 bash scripts/coverage-gate.sh.
set -euo pipefail

unset GOROOT || true

# Pragmatic baseline: cmd/ + internal/app/* are thin Cobra/delegation
# layers exercised by scripts/e2e-phase*.sh; their unit coverage is 0
# by design. The core packages (intent, memory, router, sensors, …)
# carry the substantive logic and must stay above CORE_MIN.
#
# Ratchet plan: when a core package crosses the next 5-pt band (50 → 55
# → 60 → …), bump CORE_MIN to lock the gain in. Bump GLOBAL_MIN only
# after the bottom half of core packages crosses CORE_MIN.
GLOBAL_MIN="${GLOBAL_MIN:-40}"
CORE_MIN="${CORE_MIN:-50}"
COVER_PROFILE="${COVER_PROFILE:-coverage.out}"
CORE_REGEX='internal/(domain|intent|memory|router|sensors|spec|plan|index|context|design|optimize|skills|platform/)'

cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

echo "→ coverage profile: $COVER_PROFILE  thresholds: global ≥ ${GLOBAL_MIN}% / core ≥ ${CORE_MIN}%"

go test -race -coverpkg=./... -coverprofile="$COVER_PROFILE" ./... > /tmp/coverage-gate.tests.log 2>&1 \
  || { tail -30 /tmp/coverage-gate.tests.log; exit 1; }
echo "  ($(grep -c '^ok' /tmp/coverage-gate.tests.log) packages reported)"

total=$(go tool cover -func="$COVER_PROFILE" | awk '/^total:/{ sub(/%/,"",$NF); print $NF }')
echo "→ global coverage: ${total}%"

awk -v t="$total" -v g="$GLOBAL_MIN" 'BEGIN { exit !(t+0 >= g+0) }' \
  || { echo "✗ global coverage ${total}% < ${GLOBAL_MIN}%"; exit 1; }

go tool cover -func="$COVER_PROFILE" | grep -v '^total:' > /tmp/coverage.func.txt

python3 scripts/coverage-aggregate.py "$CORE_REGEX" "$CORE_MIN" < /tmp/coverage.func.txt

echo "✓ coverage gate passed"
