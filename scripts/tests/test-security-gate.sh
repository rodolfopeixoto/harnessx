#!/usr/bin/env bash
# Smoke test: security-gate.sh has expected structure and is executable.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
. scripts/lib/assert.sh
TESTS_NAME="security-gate"

[ -x scripts/security-gate.sh ] && _pass "executable" || _fail "executable" "scripts/security-gate.sh is not executable"

src=$(cat scripts/security-gate.sh)
assert_contains "runs go vet" "go vet" "$src"
assert_contains "optional govulncheck" "govulncheck" "$src"
assert_contains "optional gitleaks" "gitleaks" "$src"
assert_contains "runs harness security-audit" "security-audit" "$src"

report
