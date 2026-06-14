#!/usr/bin/env bash
# Phase 3 end-to-end: agent list / add / discover / certify.
set -euo pipefail

unset GOROOT || true

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/bin/harness"

if [[ ! -x "$BIN" ]]; then
  (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/harness)
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cd "$WORK"
git init -q
"$BIN" init >/dev/null

echo "→ agent list (bundled only)"
"$BIN" agent list | tee /tmp/hx-agent-list.txt
grep -q "^fake " /tmp/hx-agent-list.txt
grep -q "^claude " /tmp/hx-agent-list.txt
grep -q "bundled:" /tmp/hx-agent-list.txt

echo "→ agent add fake"
"$BIN" agent add fake
test -s ".harness/config/agents/fake.yaml"

echo "→ agent list (project override)"
"$BIN" agent list | grep -E "fake\s+.*\.harness/config/agents/fake\.yaml"

echo "→ agent discover fake-cli"
"$BIN" agent discover fake-cli | grep -q "^id: fake$"

echo "→ agent certify fake --skip-run"
"$BIN" agent certify fake --skip-run | tee /tmp/hx-cert.txt
grep -qE "status: (passed|partial|failed)" /tmp/hx-cert.txt

if command -v sqlite3 >/dev/null; then
  COUNT="$(sqlite3 .harness/db/harness.sqlite 'select count(*) from agent_certifications;')"
  test "$COUNT" -ge 1
fi

echo "✓ e2e-phase3 passed"
