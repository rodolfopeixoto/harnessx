#!/usr/bin/env bash
# Verify install-hooks.sh copies every hook into .git/hooks under a temp
# repo, and marks them executable.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
. scripts/lib/assert.sh
TESTS_NAME="install-hooks"

src="$(pwd)/scripts/git-hooks"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

(cd "$tmp" && git init -q)
mkdir -p "$tmp/scripts/git-hooks"
cp "$src"/* "$tmp/scripts/git-hooks/"
cp scripts/install-hooks.sh "$tmp/scripts/install-hooks.sh"

(cd "$tmp" && bash scripts/install-hooks.sh >/dev/null)

for h in pre-commit commit-msg pre-push; do
  if [ -x "$tmp/.git/hooks/$h" ]; then
    _pass "installed $h"
  else
    _fail "installed $h" "missing or not executable"
  fi
done

report
