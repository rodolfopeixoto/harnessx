#!/usr/bin/env bash
set -euo pipefail
WEB_MIN="${WEB_MIN:-75}"
cd "$(git rev-parse --show-toplevel)"
cd web/dashboard

if ! npm ls @vitest/coverage-v8 >/dev/null 2>&1; then
  echo "→ installing @vitest/coverage-v8 (dev-only)"
  npm install --no-save --no-fund --no-audit --silent @vitest/coverage-v8 >/dev/null
fi

echo "→ web coverage (vitest --coverage)  threshold: ${WEB_MIN}%"
npx vitest run --coverage --coverage.reporter=text-summary --coverage.reporter=json-summary > /tmp/web-cov.log 2>&1 \
  || { tail -40 /tmp/web-cov.log; exit 1; }

if [ ! -f coverage/coverage-summary.json ]; then
  echo "✗ coverage-summary.json missing"; tail -40 /tmp/web-cov.log; exit 1
fi

total=$(node -e 'process.stdout.write(String(JSON.parse(require("fs").readFileSync("coverage/coverage-summary.json")).total.lines.pct))')
echo "→ web lines coverage: ${total}%"
awk -v t="$total" -v g="$WEB_MIN" 'BEGIN { exit !(t+0 >= g+0) }' \
  || { echo "✗ web ${total}% < ${WEB_MIN}%"; exit 1; }
echo "✓ web coverage gate passed"
