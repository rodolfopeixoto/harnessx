#!/usr/bin/env bash
# Bypass guardrail — blocks commits that set HARNESS_SKIP_CI=1 or
# HARNESS_SKIP_SPEC_GATE=1 without an audited HARNESS_BYPASS_REASON.
# When a reason is supplied the commit proceeds but the bypass is
# logged to .harness/logs/gate-bypass.jsonl.
#
# See docs/bypass-guardrail.md and .harness/artifacts/specs/p49-bypass-guardrail.md.
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
  echo "✗ bypass guardrail: ${skipped[*]} set without HARNESS_BYPASS_REASON"
  echo "  → provide justification, e.g.:"
  echo "      HARNESS_BYPASS_REASON=\"hotfix #123 — prod outage\" git commit ..."
  exit 1
fi

log_dir="$(git rev-parse --show-toplevel)/.harness/logs"
mkdir -p "$log_dir"
ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
user="$(git config user.email 2>/dev/null || echo unknown)"
hooks="$(IFS=,; echo "${skipped[*]}")"
# minimal JSON — quote reason via python for safety when jq is absent
escaped_reason="$(printf '%s' "$reason" | sed 's/\\/\\\\/g; s/"/\\"/g')"
printf '{"ts":"%s","user":"%s","hooks_skipped":"%s","reason":"%s","hook":"pre-commit"}\n' \
  "$ts" "$user" "$hooks" "$escaped_reason" >> "$log_dir/gate-bypass.jsonl"

echo "⚠ bypass logged: ${skipped[*]} — reason: $reason"
exit 0
