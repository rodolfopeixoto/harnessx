#!/usr/bin/env bash
# event: pre-tool-use
# description: Run golangci-lint + go vet before letting the agent modify files; block on first failure.
set -euo pipefail
echo "[pre-tool-use-lint] run=$HARNESS_RUN_ID agent=$HARNESS_AGENT" >&2

if command -v go >/dev/null 2>&1 && [[ -f go.mod ]]; then
  go vet ./... >&2 || { echo "[pre-tool-use-lint] go vet failed" >&2; exit 1; }
fi

if command -v golangci-lint >/dev/null 2>&1 && [[ -f go.mod ]]; then
  golangci-lint run ./... >&2 || { echo "[pre-tool-use-lint] golangci-lint failed" >&2; exit 1; }
fi

exit 0
