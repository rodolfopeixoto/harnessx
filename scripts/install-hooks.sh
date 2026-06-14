#!/usr/bin/env bash
# Convenience wrapper for `make install-hooks`. Lets contributors wire
# the local CI gate without a full Make environment.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
mkdir -p .git/hooks
for h in scripts/git-hooks/*; do
  name="$(basename "$h")"
  cp "$h" ".git/hooks/$name"
  chmod +x ".git/hooks/$name"
  echo "installed .git/hooks/$name"
done
echo
echo "next push will run 'make ci' before reaching the remote."
echo "bypass once for emergencies: HARNESS_SKIP_CI=1 git push"
