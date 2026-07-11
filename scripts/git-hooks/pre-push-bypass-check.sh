#!/usr/bin/env bash
# Pre-push bypass detector. Emits warning + audit log entry when a push
# is being performed with gates disabled (HARNESS_SKIP_CI=1 or
# HARNESS_SKIP_SPEC_GATE=1). Requires HARNESS_BYPASS_REASON to allow.
#
# See docs/bypass-guardrail.md.
set -euo pipefail

skipped=()
if [[ "${HARNESS_SKIP_CI:-}" == "1" ]]; then
  skipped+=("HARNESS_SKIP_CI")
fi
if [[ "${HARNESS_SKIP_SPEC_GATE:-}" == "1" ]]; then
  skipped+=("HARNESS_SKIP_SPEC_GATE")
fi

if [[ ${#skipped[@]} -eq 0 ]]; then
  exit 0
fi

reason="${HARNESS_BYPASS_REASON:-}"
if [[ -z "$reason" ]]; then
  echo "✗ pre-push bypass guardrail: ${skipped[*]} set without HARNESS_BYPASS_REASON"
  echo "  Set HARNESS_BYPASS_REASON with a documented justification (issue/ticket)."
  exit 1
fi

log_dir="$(git rev-parse --show-toplevel)/.harness/logs"
mkdir -p "$log_dir"
ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
user="$(git config user.email 2>/dev/null || echo unknown)"
hooks="$(IFS=,; echo "${skipped[*]}")"
escaped_reason="$(printf '%s' "$reason" | sed 's/\\/\\\\/g; s/"/\\"/g')"
printf '{"ts":"%s","user":"%s","hooks_skipped":"%s","reason":"%s","hook":"pre-push"}\n' \
  "$ts" "$user" "$hooks" "$escaped_reason" >> "$log_dir/gate-bypass.jsonl"

echo "⚠ gates skipped: ${skipped[*]} — reason: $reason (audit logged)"
exit 0
