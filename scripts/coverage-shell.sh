#!/usr/bin/env bash
set -euo pipefail
SHELL_MIN="${SHELL_MIN:-80}"
cd "$(git rev-parse --show-toplevel)"

if ! command -v bashcov >/dev/null 2>&1; then
  echo "  (bashcov not installed; gem install bashcov simplecov)"
  echo "  (advisory: skipping shell coverage gate this run)"
  exit 0
fi

echo "→ shell coverage (bashcov)  threshold: ${SHELL_MIN}%"
rm -rf coverage/shell
bashcov --root scripts/lib --command "bash scripts/tests/run-all.sh" > /tmp/shellcov.log 2>&1 \
  || { tail -40 /tmp/shellcov.log; exit 1; }

if [ ! -f coverage/.last_run.json ]; then
  echo "✗ bashcov did not produce coverage/.last_run.json"; exit 1
fi
total=$(node -e 'process.stdout.write(String(JSON.parse(require("fs").readFileSync("coverage/.last_run.json")).result.line))')
echo "→ shell lines coverage: ${total}%"
awk -v t="$total" -v g="$SHELL_MIN" 'BEGIN { exit !(t+0 >= g+0) }' \
  || { echo "✗ shell ${total}% < ${SHELL_MIN}%"; exit 1; }
echo "✓ shell coverage gate passed"
