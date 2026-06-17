#!/usr/bin/env bash
set -euo pipefail
unset GOROOT || true

GLOBAL_MIN="${GLOBAL_MIN:-49}"
CORE_MIN="${CORE_MIN:-58}"
COVER_PROFILE="${COVER_PROFILE:-coverage.out}"
CORE_REGEX='internal/(domain|intent|memory|router|sensors|spec|plan|index|context|design|optimize|skills|platform/|adapters/sqlite|workspace|catalog|cleanup|runtime|importwiz|stale|palette|autonomy|autopilot|audit|health)'

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
