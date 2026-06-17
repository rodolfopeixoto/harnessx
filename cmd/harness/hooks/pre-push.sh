#!/usr/bin/env sh
# Managed by HarnessX. Re-run `harness install-git-hooks` to refresh.
# Bypass: HARNESS_SKIP_CI=1 git push
set -eu

if [ "${HARNESS_SKIP_CI:-0}" = "1" ]; then
  echo "→ HARNESS_SKIP_CI=1 — skipping harness ci"
  exit 0
fi

if ! command -v harness >/dev/null 2>&1; then
  echo "✗ pre-push: 'harness' binary not on PATH" >&2
  exit 1
fi

exec harness ci
