#!/usr/bin/env sh
# Managed by HarnessX. Re-run `harness install-git-hooks` to refresh.
# Bypass: HARNESS_SKIP_PRECOMMIT=1 git commit
set -eu

if [ "${HARNESS_SKIP_PRECOMMIT:-0}" = "1" ]; then
  exit 0
fi

if ! command -v harness >/dev/null 2>&1; then
  echo "✗ pre-commit: 'harness' binary not on PATH" >&2
  exit 1
fi

exec harness lint
