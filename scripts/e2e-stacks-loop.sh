#!/usr/bin/env bash
set -u
BIN="${HARNESS:-harness}"
WORK="$(mktemp -d)"
STACKS=(go python python-ecommerce ruby rails react rust java kotlin swift elixir php dotnet)

pass=0
fail=0
skip=0
FAILED_LIST=()

cleanup() {
  rm -rf "$WORK" >/dev/null 2>&1 || true
}
trap cleanup EXIT

for s in "${STACKS[@]}"; do
  echo "── stack: $s ──"
  dir="$WORK/$s"
  mkdir -p "$dir"
  (cd "$dir" && "$BIN" scaffold apply "$s" --apply --name "audit-$s") > "$dir/scaffold.log" 2>&1
  if [ $? -ne 0 ]; then
    echo "  ✗ scaffold failed"
    fail=$((fail+1))
    FAILED_LIST+=("$s:scaffold")
    continue
  fi
  (cd "$dir" && "$BIN" sensor list) > "$dir/sensors.log" 2>&1
  sensors=$(grep -c "" "$dir/sensors.log" || echo 0)
  echo "  scaffold ok ($(ls "$dir" | wc -l | tr -d ' ') top-level entries, $sensors sensors registered)"
  (cd "$dir" && "$BIN" ci) > "$dir/ci.log" 2>&1
  rc=$?
  status_line=$(grep -E "^summary:" "$dir/ci.log" || echo "(no summary)")
  if [ "$rc" -eq 0 ]; then
    echo "  ci pass — $status_line"
    pass=$((pass+1))
  elif grep -q "summary: " "$dir/ci.log"; then
    failures=$(grep "summary:" "$dir/ci.log" | head -1)
    echo "  ci ran (rc=$rc) — $failures"
    if echo "$failures" | grep -qE "0 failed"; then
      pass=$((pass+1))
    else
      fail=$((fail+1))
      FAILED_LIST+=("$s:ci")
      grep -E "\[✗\]" "$dir/ci.log" | head -3 | sed 's/^/    /'
    fi
  else
    echo "  ci skipped (no detectable tools, rc=$rc)"
    skip=$((skip+1))
  fi
  rm -rf "$dir/.harness" "$dir/node_modules" "$dir/vendor" "$dir/.venv" "$dir/venv" "$dir/build" "$dir/target" "$dir/_build" "$dir/deps" "$dir/bin" "$dir/obj" 2>/dev/null || true
done

echo
echo "══════════════════════"
echo "stacks: $((pass+fail+skip))  pass: $pass  fail: $fail  skip: $skip"
if [ $fail -gt 0 ]; then
  echo "failed: ${FAILED_LIST[*]}"
  exit 1
fi
exit 0
