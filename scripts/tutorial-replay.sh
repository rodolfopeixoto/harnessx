#!/usr/bin/env bash
set -euo pipefail

HARNESS_BIN="${HARNESS_BIN:-$(pwd)/bin/harness}"
if [ ! -x "$HARNESS_BIN" ]; then
  echo "building harness binary..."
  go build -o "$HARNESS_BIN" ./cmd/harness
fi

tmp="$(mktemp -d -t harness-tutorial-XXXXXX)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"
echo "==> replay dir: $tmp"

git init -q

"$HARNESS_BIN" init >/dev/null
"$HARNESS_BIN" scaffold apply python --apply >/dev/null
"$HARNESS_BIN" install-git-hooks --hooks pre-push >/dev/null

echo "==> introspection block"
"$HARNESS_BIN" scaffold list   >/dev/null
"$HARNESS_BIN" sensor list     >/dev/null
"$HARNESS_BIN" flow list       >/dev/null
"$HARNESS_BIN" memory list     >/dev/null
"$HARNESS_BIN" routes          >/dev/null

echo "==> resolution block"
"$HARNESS_BIN" doctor          >/dev/null
"$HARNESS_BIN" check           >/dev/null
"$HARNESS_BIN" ci              >/dev/null || echo "(ci red — tool-availability; non-fatal in replay)"

echo "==> ship dry-run (no LLM call)"
"$HARNESS_BIN" ship "add /readiness endpoint" --dry-run --skip-commit >/dev/null

echo "✓ tutorial replay green"
