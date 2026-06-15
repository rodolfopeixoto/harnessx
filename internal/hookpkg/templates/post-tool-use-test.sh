#!/usr/bin/env bash
# event: post-tool-use
# description: Run go test ./... (or npm test) after the agent run; non-zero blocks apply.
set -euo pipefail
echo "[post-tool-use-test] run=$HARNESS_RUN_ID agent=$HARNESS_AGENT" >&2

if [[ -f go.mod ]]; then
  go test ./... >&2 || { echo "[post-tool-use-test] go test failed" >&2; exit 1; }
elif [[ -f package.json ]]; then
  npm test >&2 || { echo "[post-tool-use-test] npm test failed" >&2; exit 1; }
fi

exit 0
