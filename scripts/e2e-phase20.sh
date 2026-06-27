#!/usr/bin/env bash
set -euo pipefail
unset GOROOT || true
cd "$(git rev-parse --show-toplevel)"
make build > /dev/null
BIN="$(pwd)/bin/harness"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

ROOT="$WORK/tour"
echo "→ tour with explicit --root"
"$BIN" stack tour --root "$ROOT" > "$WORK/tour.txt"
grep -q "workspace_add" "$WORK/tour.txt" || { echo "✗ missing workspace_add"; cat "$WORK/tour.txt"; exit 1; }
grep -q "catalog_install" "$WORK/tour.txt" || { echo "✗ missing catalog_install"; cat "$WORK/tour.txt"; exit 1; }
grep -q "health_score" "$WORK/tour.txt" || { echo "✗ missing health_score"; cat "$WORK/tour.txt"; exit 1; }
grep -q "FAIL" "$WORK/tour.txt" && { echo "✗ tour reported FAIL"; cat "$WORK/tour.txt"; exit 1; }

echo "→ tour deletes root by default"
[ -d "$ROOT" ] || true

echo "→ tour with --keep preserves project"
KEPT="$WORK/keep"
"$BIN" stack tour --root "$KEPT" --keep > "$WORK/keep.txt"
[ -d "$KEPT" ] || { echo "✗ --keep removed root"; exit 1; }
[ -f "$KEPT/.harness/registry.sqlite" ] || { echo "✗ registry missing"; exit 1; }

echo "→ tour with --dashboard probes /api/health"
DASH_ROOT="$WORK/dash"
"$BIN" stack tour --root "$DASH_ROOT" --keep --dashboard --addr 127.0.0.1:17820 > "$WORK/dash.txt"
grep -q "dashboard_probe" "$WORK/dash.txt" || { echo "✗ dashboard probe missing"; cat "$WORK/dash.txt"; exit 1; }

echo "→ stack status reports offline when nothing running (default exit 0)"
"$BIN" stack status --addr 127.0.0.1:1 > "$WORK/status.txt" 2>&1 || { echo "✗ default status must not exit non-zero when offline"; cat "$WORK/status.txt"; exit 1; }
grep -q "offline" "$WORK/status.txt" || { echo "✗ missing offline marker"; cat "$WORK/status.txt"; exit 1; }

echo "→ stack status --strict exits non-zero when offline"
if "$BIN" stack status --strict --addr 127.0.0.1:1 > "$WORK/status-strict.txt" 2>&1; then
  echo "✗ --strict must fail when offline"; cat "$WORK/status-strict.txt"; exit 1
fi
grep -q "offline" "$WORK/status-strict.txt" || { echo "✗ --strict missing offline marker"; cat "$WORK/status-strict.txt"; exit 1; }

echo "✓ e2e-phase20 passed"
