#!/usr/bin/env bash
set -euo pipefail
unset GOROOT || true
cd "$(git rev-parse --show-toplevel)"
make build > /dev/null
BIN="$(pwd)/bin/harness"

export AUDIT_PLAYWRIGHT_SKIP=1
PROJ="$(mktemp -d)"
trap 'rm -rf "$PROJ"' EXIT
cd "$PROJ"
mkdir -p tmp
"$BIN" stack audit > "$PROJ/out.txt" 2>&1 || true

ts_dir="$(ls -1 tmp/app-audit/ 2>/dev/null | head -1)"
[ -n "$ts_dir" ] || { echo "✗ no project dir"; ls tmp/app-audit; exit 1; }
run_dir="$(ls -1 tmp/app-audit/$ts_dir | head -1)"
ROOT="tmp/app-audit/$ts_dir/$run_dir"

bundle="$ROOT/report/audit-bundle.zip"
[ -f "$bundle" ] || { echo "✗ bundle missing"; ls "$ROOT/report/"; exit 1; }

UNPACK="$PROJ/unpacked"
mkdir -p "$UNPACK"
unzip -q "$bundle" -d "$UNPACK"

for need in BUNDLE_INDEX.md json/feature-map.json json/results.json json/summary.json json/cli-flows.json json/inventory.json report/audit.html report/fix-backlog.md run.log; do
  [ -f "$UNPACK/$need" ] || { echo "✗ bundle missing $need"; ls -R "$UNPACK"; exit 1; }
done

python3 - "$UNPACK/json/cli-flows.json" <<'PY'
import json, sys
data = json.load(open(sys.argv[1]))
assert "flows" in data and len(data["flows"]) > 0
for flow in data["flows"]:
    for key in ("name", "args", "exit_code", "duration_ms", "stdout", "stderr"):
        assert key in flow, f"missing {key} in {flow}"
PY

python3 - "$UNPACK/json/inventory.json" <<'PY'
import json, sys
inv = json.load(open(sys.argv[1]))
for key in ("go_files", "go_test_files", "tsx_files", "shell_scripts", "spec_files"):
    assert key in inv, f"missing {key}"
PY

echo "✓ e2e-phase24 passed"
