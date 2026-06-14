#!/usr/bin/env bash
# spec-gate: Spec-Driven Development enforcement.
#
# If a push adds new Go files under a core package, the same push must add
# a matching spec at .harness/artifacts/specs/p<N>-*.md.
#
# Bypass: HARNESS_SKIP_SPEC_GATE=1 (audited).
set -euo pipefail

remote="${1:-origin}"
core_regex='^internal/(workspace|catalog|cleanup|importwiz|stale|palette|autonomy|autopilot|audit|health)/[^/]+\.go$'
spec_glob='.harness/artifacts/specs/p[0-9]*.md'

if ! git rev-parse --verify "${remote}/HEAD" >/dev/null 2>&1; then
  base="$(git rev-list --max-parents=0 HEAD | tail -1)"
else
  base="$(git merge-base HEAD "${remote}/HEAD" 2>/dev/null || git rev-list --max-parents=0 HEAD | tail -1)"
fi

added_core="$(git diff --name-only --diff-filter=A "$base"..HEAD | grep -E "$core_regex" || true)"
[ -z "$added_core" ] && exit 0

added_specs="$(git diff --name-only --diff-filter=A "$base"..HEAD | grep -E "${spec_glob//\*/.*}" || true)"

if [ -z "$added_specs" ]; then
  echo "✗ spec-gate: new files in core packages without a matching spec"
  echo "  files:"
  echo "$added_core" | sed 's/^/    /'
  echo "  expected at .harness/artifacts/specs/p<N>-<slug>.md"
  echo "  bypass: HARNESS_SKIP_SPEC_GATE=1 (audited)"
  exit 1
fi

echo "✓ spec-gate: $(echo "$added_specs" | wc -l | tr -d ' ') spec(s) cover new core files"
