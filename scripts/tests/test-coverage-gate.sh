#!/usr/bin/env bash
# Unit test for scripts/coverage-aggregate.py — exercises the core/global
# split without invoking go test (which is too slow for a smoke harness).
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
. scripts/lib/assert.sh
TESTS_NAME="coverage-gate"

REGEX='internal/(domain|intent|memory|router|sensors|spec|plan|index|context|design|optimize|skills|platform/)'

input_pass=$(cat <<'EOF'
github.com/x/repo/internal/intent/intent.go:1:	IntentDetect		95.0%
github.com/x/repo/internal/router/router.go:1:	Pick			80.0%
github.com/x/repo/cmd/harness/main.go:1:		main			0.0%
EOF
)
input_fail=$(cat <<'EOF'
github.com/x/repo/internal/intent/intent.go:1:	IntentDetect		20.0%
github.com/x/repo/cmd/harness/main.go:1:		main			0.0%
EOF
)

if echo "$input_pass" | python3 scripts/coverage-aggregate.py "$REGEX" 70 > /tmp/cov.pass.log 2>&1; then
  _pass "passes when all core ≥ threshold"
else
  _fail "passes when all core ≥ threshold" "$(cat /tmp/cov.pass.log)"
fi

if echo "$input_fail" | python3 scripts/coverage-aggregate.py "$REGEX" 70 > /tmp/cov.fail.log 2>&1; then
  _fail "rejects when core below threshold" "exit was 0"
else
  _pass "rejects when core below threshold"
fi

report
