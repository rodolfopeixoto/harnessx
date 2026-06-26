#!/usr/bin/env bash
# Phase 6 end-to-end: ask, plan, run, feature, bugfix, report.
set -euo pipefail

unset GOROOT || true

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/bin/harness"

if [[ ! -x "$BIN" ]]; then
  (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/harness)
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cp -R "$ROOT/testdata/projects/sample-go/." "$WORK/"
cd "$WORK"
git init -q
git add -A
git -c user.email=t@t -c user.name=t commit -q -m init
"$BIN" init >/dev/null
"$BIN" project index >/dev/null

echo "→ ask"
"$BIN" ask "What does main do?" > /tmp/hx-ask.txt
cat /tmp/hx-ask.txt
grep -q "Question Mode" /tmp/hx-ask.txt
test -s .harness/runs/*/report.md

echo "→ feature"
"$BIN" feature "add greet function with tests" --yes --budget 0.5 > /tmp/hx-feat.txt
cat /tmp/hx-feat.txt
grep -q "Spec written:" /tmp/hx-feat.txt
grep -q "Plan written:" /tmp/hx-feat.txt
grep -q "Report written:" /tmp/hx-feat.txt
test -s .harness/artifacts/specs/*.md
test -s .harness/artifacts/plans/*.md

echo "→ bugfix"
"$BIN" bugfix "fix crash on empty input" --yes > /tmp/hx-bug.txt
grep -q "Detected intent: bugfix" /tmp/hx-bug.txt

echo "→ run (natural form via classifier)"
"$BIN" run "optimize Docker image size" --yes > /tmp/hx-run.txt
grep -q "Detected intent: optimization" /tmp/hx-run.txt

echo "→ natural prompt"
"$BIN" "review the latest diff" > /tmp/hx-nat.txt
grep -q "Detected intent: review" /tmp/hx-nat.txt

echo "→ plan (no execute)"
"$BIN" plan "create product search with filters" > /tmp/hx-plan.txt
grep -q "Plan written:" /tmp/hx-plan.txt

echo "→ report --last"
"$BIN" report > /tmp/hx-rpt.txt
head -1 /tmp/hx-rpt.txt | grep -q ".md"
grep -q "# Summary" /tmp/hx-rpt.txt

if command -v sqlite3 >/dev/null; then
  COUNT="$(sqlite3 .harness/db/harness.sqlite 'select count(*) from sessions;')"
  test "$COUNT" -ge 5
fi

echo "✓ e2e-phase6 passed"
