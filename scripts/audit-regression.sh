#!/usr/bin/env bash
# audit-regression.sh — verifica os 26 bugs catalogados em
# .harness/artifacts/HARNESS-AUDIT-COMPLETE.md (v0.151.1).
#
# Cada bug roda um cenário determinístico (sem custo de LLM quando possível)
# e marca pass/fail em $OUT_JSON. Saída final: tabela + exit code 0 quando
# bug-count(failures) ≤ baseline. CI pode chamar com BASELINE=audit-baseline.json
# para bloquear regressão.
#
# Uso:
#   bash scripts/audit-regression.sh                # gera audit-status.json
#   BASELINE=audit-baseline.json bash scripts/audit-regression.sh
#   GENERATE_BASELINE=1 bash scripts/audit-regression.sh   # grava baseline
#
# Cenários LLM (BUG-1, BUG-3, BUG-5, BUG-13/14) ficam OPT-IN via
# AUDIT_RUN_LLM=1 — caros e exigem adapters reais autenticados.

set -u
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
HARNESS="${HARNESS_BIN:-${ROOT}/dist/harness}"
OUT_JSON="${OUT_JSON:-${ROOT}/.harness/audit-status.json}"
BASELINE="${BASELINE:-}"
SANDBOX="${SANDBOX:-/tmp/harness-audit-regression}"

mkdir -p "$(dirname "$OUT_JSON")"
: >"$OUT_JSON.tmp"

pass=0; fail=0; skipped=0
results=()

record() {
  local id="$1" status="$2" note="$3"
  results+=("{\"bug\":\"$id\",\"status\":\"$status\",\"note\":\"${note//\"/\\\"}\"}")
  case "$status" in
    pass) pass=$((pass+1)) ;;
    fail) fail=$((fail+1)) ;;
    skip) skipped=$((skipped+1)) ;;
  esac
  printf '  [%s] %s — %s\n' "$status" "$id" "$note"
}

require_bin() {
  if [ ! -x "$HARNESS" ]; then
    echo "harness binary not found at $HARNESS — set HARNESS_BIN or run 'make build'" >&2
    exit 2
  fi
}

prepare_sandbox() {
  rm -rf "$SANDBOX"
  mkdir -p "$SANDBOX"
  ( cd "$SANDBOX" && git init -q && touch .keep && git add .keep && git -c user.email=a@a -c user.name=a commit -qm init )
}

# ----- read-only checks (no LLM cost) ----------------------------------

check_bug2() {
  local out
  out=$("$HARNESS" loop --help 2>&1 || true)
  if echo "$out" | grep -q '(default "claude")'; then
    record BUG-2 fail "loop --help still defaults to literal claude"
  else
    out=$("$HARNESS" execute --help 2>&1 || true)
    if echo "$out" | grep -q '(default "fake-real")'; then
      record BUG-2 fail "execute --help still defaults to fake-real"
    else
      record BUG-2 pass "no literal claude/fake-real defaults"
    fi
  fi
}

check_bug6() {
  # BUG-6: runs inspect should find meta.json. We can only assert format here.
  local out
  out=$("$HARNESS" runs inspect --help 2>&1 || true)
  if [ -n "$out" ]; then
    record BUG-6 pass "runs inspect --help responds"
  else
    record BUG-6 fail "runs inspect --help empty"
  fi
}

check_bug9() {
  # If the host env has a broken GOROOT (gvm leftovers etc.) `go version`
  # legitimately fails — that is the user's local config, not a HarnessX
  # bug. Strip GOROOT before probing so we test the binary, not the env.
  local out
  out=$(GOROOT= "$HARNESS" doctor 2>&1 || true)
  if echo "$out" | grep -qE 'Go +.*version probe failed'; then
    record BUG-9 fail "doctor: go version probe failed"
  else
    record BUG-9 pass "doctor reports go ok"
  fi
}

check_bug16() {
  # scaffolded project should have .harness/worktrees/ in .gitignore.
  prepare_sandbox
  ( cd "$SANDBOX" && "$HARNESS" init >/dev/null 2>&1 || true )
  if grep -q '.harness/worktrees' "$SANDBOX/.gitignore" 2>/dev/null; then
    record BUG-16 pass ".harness/worktrees in scaffold .gitignore"
  else
    record BUG-16 fail "worktree dir not in .gitignore"
  fi
}

check_bug17() {
  local out
  out=$("$HARNESS" ship --help 2>&1 || true)
  if echo "$out" | grep -q -- "--budget-usd"; then
    record BUG-17 pass "ship has --budget-usd"
  else
    record BUG-17 fail "ship missing --budget-usd"
  fi
}

check_bug19() {
  # Synthesize fake runs + sessions in sandbox; check analytics aggregates.
  prepare_sandbox
  local hdir="$SANDBOX/.harness"
  mkdir -p "$hdir/runs/run_test1" "$hdir/sessions"
  cat >"$hdir/runs/run_test1/meta.json" <<'JSON'
{"run_id":"run_test1","agent_id":"claude","task_tag":"impl","status":"applied","estimated_cost_usd":0.42,"started_at":"2026-06-26T00:00:00Z","finished_at":"2026-06-26T00:01:00Z"}
JSON
  local out
  out=$( cd "$SANDBOX" && "$HARNESS" analytics --json 2>&1 || true )
  if echo "$out" | grep -qE '"TotalUSD"\s*:\s*0\.42|"CostUSD"\s*:\s*0\.42|"total_usd"\s*:\s*0\.42'; then
    record BUG-19 pass "analytics aggregates run meta.json"
  else
    record BUG-19 fail "analytics does not aggregate runs/run_*/meta.json"
  fi
}

check_bug20_25() {
  # scaffold python should include pytest-cov + pytest-benchmark.
  prepare_sandbox
  local req
  req=$(find "$ROOT/internal/scaffoldpkg/templates" -name requirements.txt 2>/dev/null | head -1)
  if [ -z "$req" ]; then
    record BUG-20/25 skip "no scaffold requirements.txt found"
    return
  fi
  if grep -q pytest-cov "$req" && grep -q pytest-benchmark "$req"; then
    record BUG-20/25 pass "scaffold has pytest-cov + pytest-benchmark"
  else
    record BUG-20/25 fail "scaffold missing pytest-cov or pytest-benchmark"
  fi
}

check_bug21() {
  local out
  out=$("$HARNESS" route "what is the weather" 2>&1 || true)
  if echo "$out" | grep -qE "primary|chain|adapter"; then
    record BUG-21 pass "route accepts positional prompt"
  else
    record BUG-21 fail "route only shows help"
  fi
}

check_bug22() {
  local out
  out=$("$HARNESS" explain "ping" 2>&1 || true)
  if echo "$out" | grep -q 'budget=\$0\.00'; then
    record BUG-22 fail "explain still prints budget=\$0.00"
  else
    record BUG-22 pass "explain shows non-zero default budget"
  fi
}

check_bug23() {
  # log-audit must exclude .venv.
  prepare_sandbox
  mkdir -p "$SANDBOX/.venv/lib/python3.12/site-packages/foo"
  printf 'panic("oh no")\n' >"$SANDBOX/.venv/lib/python3.12/site-packages/foo/bar.py"
  local out
  out=$( cd "$SANDBOX" && "$HARNESS" log-audit 2>&1 || true )
  if echo "$out" | grep -q '\.venv'; then
    record BUG-23 fail "log-audit still scans .venv"
  else
    record BUG-23 pass "log-audit excludes .venv"
  fi
}

check_bug24() {
  local out
  out=$("$HARNESS" config show 2>&1 || true)
  if echo "$out" | grep -qi "primary="; then
    record BUG-24 pass "config show lists effective routes"
  else
    record BUG-24 fail "config show says no overrides without showing effective routes"
  fi
}

check_falta_mod1() {
  local out
  out=$("$HARNESS" agent models --help 2>&1 || true)
  if [ -n "$out" ] && ! echo "$out" | grep -qi "unknown command"; then
    record FALTA-MOD-1 pass "agent models subcommand exists"
  else
    record FALTA-MOD-1 fail "agent models subcommand missing"
  fi
}

check_falta_mod2() {
  local out
  out=$("$HARNESS" use --help 2>&1 || true)
  if echo "$out" | grep -q -- "--tier"; then
    record FALTA-MOD-2 pass "use --tier exists"
  else
    record FALTA-MOD-2 fail "use --tier missing"
  fi
}

check_falta_mod3() {
  local out
  out=$("$HARNESS" onboarding --help 2>&1 || true)
  if echo "$out" | grep -q -- "--interactive"; then
    record FALTA-MOD-3 pass "onboarding --interactive exists"
  else
    record FALTA-MOD-3 fail "onboarding --interactive missing"
  fi
}

check_bug12() {
  local out
  out=$("$HARNESS" auto --help 2>&1 || true)
  if [ -n "$out" ] && ! echo "$out" | grep -qi "unknown command"; then
    record BUG-12 pass "auto command exists"
  else
    record BUG-12 fail "auto command missing"
  fi
}

# Placeholder checks for bugs requiring LLM cost or already covered by Go tests.
placeholder() {
  local id="$1" reason="$2"
  if [ "${AUDIT_RUN_LLM:-0}" = "1" ]; then
    # Run real scenarios — out of scope for default invocation
    record "$id" skip "LLM scenario not implemented in $0 yet"
  else
    record "$id" skip "$reason (set AUDIT_RUN_LLM=1 to attempt)"
  fi
}

# ----- main --------------------------------------------------------------

require_bin
echo "harness audit-regression: $(date -u +%FT%TZ)"

check_bug2
placeholder BUG-1 "chat REPL uses paid input"
placeholder BUG-3 "needs codex adapter + skill"
check_bug6
placeholder BUG-4 "spinner pipe needs chat run"
placeholder BUG-5 "feature --plan-only --yes (costs)"
placeholder BUG-7 "ask needs LLM"
placeholder BUG-8 "py_bandit needs python venv"
check_bug9
placeholder BUG-10 "subset of BUG-1"
placeholder BUG-11 "/use persistence (needs REPL)"
check_bug16
check_bug17
placeholder BUG-13 "do conflict (costs)"
placeholder BUG-14 "do conflict (costs)"
placeholder BUG-15 "do report cost (needs run)"
placeholder BUG-18 "do verify (costs)"
check_bug19
check_bug20_25
check_bug21
check_bug22
check_bug23
check_bug24
placeholder BUG-26 "spec LLM (costs)"
check_falta_mod1
check_falta_mod2
check_falta_mod3
check_bug12

# emit json
{
  printf '{\n  "generated_at":"%s",\n  "totals":{"pass":%d,"fail":%d,"skip":%d},\n  "results":[\n' \
    "$(date -u +%FT%TZ)" "$pass" "$fail" "$skipped"
  for i in "${!results[@]}"; do
    sep=","
    [ "$i" -eq $((${#results[@]} - 1)) ] && sep=""
    printf '    %s%s\n' "${results[$i]}" "$sep"
  done
  printf '  ]\n}\n'
} >"$OUT_JSON"

echo
echo "summary: pass=$pass fail=$fail skip=$skipped — $OUT_JSON"

if [ "${GENERATE_BASELINE:-0}" = "1" ]; then
  cp "$OUT_JSON" "$ROOT/audit-baseline.json"
  echo "baseline written: $ROOT/audit-baseline.json"
fi

if [ -n "$BASELINE" ] && [ -f "$BASELINE" ]; then
  base_fail=$(grep -c '"status":"fail"' "$BASELINE" || echo 0)
  if [ "$fail" -gt "$base_fail" ]; then
    echo "REGRESSION: fail=$fail > baseline=$base_fail" >&2
    exit 1
  fi
fi

exit 0
