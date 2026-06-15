#!/usr/bin/env bash
# event: post-tool-use
# description: Append a one-line audit entry per run to .harness/audit/hooks.log.
set -euo pipefail
mkdir -p .harness/audit
printf '%s run=%s agent=%s status=%s\n' \
  "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  "${HARNESS_RUN_ID:-unknown}" \
  "${HARNESS_AGENT:-unknown}" \
  "${HARNESS_RUN_STATUS:-unknown}" \
  >> .harness/audit/hooks.log
exit 0
