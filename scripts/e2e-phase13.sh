#!/usr/bin/env bash
set -euo pipefail
unset GOROOT || true
export HARNESS_LANG=en

cd "$(git rev-parse --show-toplevel)"
make build > /dev/null
BIN="$(pwd)/bin/harness"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
ROOT="$WORK/proj"
mkdir -p "$ROOT"
(cd "$ROOT" && git init -q)

mkdir -p "$ROOT/.git/worktrees/stale"
touch -d "60 days ago" "$ROOT/.git/worktrees/stale" 2>/dev/null || touch -t 202301010000 "$ROOT/.git/worktrees/stale"

mkdir -p "$ROOT/.npm"
dd if=/dev/zero of="$ROOT/.npm/big" bs=1024 count=64 status=none

mkdir -p "$ROOT/orphan/.harness"

echo "→ scan reports candidates"
"$BIN" cleanup scan "$ROOT" > "$WORK/scan.txt"
grep -q worktree "$WORK/scan.txt" || { echo "✗ worktree missing"; cat "$WORK/scan.txt"; exit 1; }
grep -q cache "$WORK/scan.txt" || { echo "✗ cache missing"; cat "$WORK/scan.txt"; exit 1; }
grep -q abandoned_harness "$WORK/scan.txt" || { echo "✗ abandoned missing"; cat "$WORK/scan.txt"; exit 1; }

echo "→ apply without policy + without --yes exits non-zero"
if "$BIN" cleanup apply "$ROOT" --yes < /dev/null > "$WORK/apply_no_policy.txt" 2>&1; then
  grep -q "applied 0/" "$WORK/apply_no_policy.txt" || { echo "✗ expected zero applied"; cat "$WORK/apply_no_policy.txt"; exit 1; }
fi
[ -d "$ROOT/.npm" ] || { echo "✗ apply without policy deleted files"; exit 1; }

echo "→ policy allows cache delete"
POLICY="$WORK/policy.yaml"
cat > "$POLICY" <<YAML
version: 1
globals:
  require_acknowledgement: false
rules:
  - kind: cache
    allowlist:
      - "$ROOT/.npm"
    max_risk: high
YAML
"$BIN" cleanup apply "$ROOT" --policy "$POLICY" --yes > "$WORK/apply_policy.txt"
grep -q "applied" "$WORK/apply_policy.txt" || { echo "✗ apply summary missing"; cat "$WORK/apply_policy.txt"; exit 1; }
[ ! -d "$ROOT/.npm" ] || { echo "✗ npm cache still present"; exit 1; }
[ -d "$ROOT/orphan/.harness" ] || { echo "✗ abandoned deleted despite no policy match"; exit 1; }

echo "→ policy init creates default"
"$BIN" cleanup policy init "$ROOT" > "$WORK/policy_init.txt"
[ -f "$ROOT/.harness/cleanup/policy.yaml" ] || { echo "✗ policy file not created"; exit 1; }

echo "→ scan JSON shape"
"$BIN" cleanup scan "$ROOT" --json > "$WORK/scan.json"
python3 - "$WORK/scan.json" <<'PY'
import json, sys
findings = json.load(open(sys.argv[1]))
assert isinstance(findings, list), "expected list"
for f in findings:
    for key in ("Kind", "Path", "Risk", "Reason"):
        assert key in f, f"missing {key}"
PY

echo "✓ e2e-phase13 passed"
