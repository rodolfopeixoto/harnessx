#!/usr/bin/env bash
# event: pre-tool-use
# description: Refuse runs whose prompt mentions force-push or git reset --hard.
set -euo pipefail

if [[ -n "${HARNESS_RUN_PROMPT:-}" ]]; then
  if echo "$HARNESS_RUN_PROMPT" | grep -Ei '(push --force|reset --hard|--force-with-lease)' >/dev/null 2>&1; then
    echo "[pre-tool-use-noforce] refusing: prompt mentions force-push / reset --hard" >&2
    exit 1
  fi
fi
exit 0
