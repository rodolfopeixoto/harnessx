#!/usr/bin/env bash
set -euo pipefail
unset GOROOT || true
cd "$(git rev-parse --show-toplevel)"
make build > /dev/null
BIN="$(pwd)/bin/harness"

export AUDIT_PLAYWRIGHT_SKIP=1
export AUDIT_BASE_URL="http://127.0.0.1:7373"

PROJ="$(mktemp -d)"
trap 'rm -rf "$PROJ"' EXIT

cd "$PROJ"
mkdir -p tmp
"$BIN" stack audit > "$PROJ/out.txt" 2>&1 || true

grep -q "Audit finished" "$PROJ/out.txt" || { echo "✗ summary missing"; cat "$PROJ/out.txt"; exit 1; }

ts_dir="$(ls -1 tmp/app-audit | head -1)"
[ -n "$ts_dir" ] || { echo "✗ no run dir"; ls tmp/app-audit; exit 1; }
ROOT="tmp/app-audit/$ts_dir"

for f in json/feature-map.json json/results.json json/summary.json json/visual-diff.json json/layout-metrics.json json/network-errors.json json/console-errors.json json/missing-selectors.json report/audit.html report/fix-backlog.md run.log; do
  [ -f "$ROOT/$f" ] || { echo "✗ missing artifact: $f"; ls -la "$ROOT"; exit 1; }
done

python3 - "$ROOT/json/summary.json" <<'PY'
import json, sys
sum = json.load(open(sys.argv[1]))
for key in ("counts", "visual_counts", "severity_counts", "total_features", "total_results"):
    assert key in sum, f"summary missing {key}"
PY

python3 - "$ROOT/json/results.json" <<'PY'
import json, sys
data = json.load(open(sys.argv[1]))
assert isinstance(data.get("results"), list) and len(data["results"]) > 0
for r in data["results"]:
    for key in ("feature_id", "viewport", "status"):
        assert key in r, f"result missing {key}"
PY

echo "✓ e2e-phase23 passed"
