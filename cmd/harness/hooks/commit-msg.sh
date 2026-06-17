#!/usr/bin/env sh
# Managed by HarnessX. Re-run `harness install-git-hooks` to refresh.
# Enforces Conventional Commits subject line.
# Bypass: HARNESS_SKIP_COMMITMSG=1 git commit
set -eu

if [ "${HARNESS_SKIP_COMMITMSG:-0}" = "1" ]; then
  exit 0
fi

msg_file="$1"
subject=$(head -n 1 "$msg_file")

pattern='^(feat|fix|chore|docs|refactor|test|perf|build|ci|style|revert)(\([a-z0-9-]+\))?!?: .{1,}'

if ! printf '%s' "$subject" | grep -Eq "$pattern"; then
  echo "✗ commit-msg: subject must follow Conventional Commits (type(scope)?: description)" >&2
  echo "  got: $subject" >&2
  exit 1
fi
