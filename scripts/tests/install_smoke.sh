#!/usr/bin/env bash
# install.sh smoke test: runs the public installer against a clean
# prefix in a temp dir, verifies the resulting binary boots and reports
# a version on the expected channel.

set -euo pipefail

repo="${HARNESS_REPO:-rodolfopeixoto/harnessx}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "→ smoke prefix: $tmp"

cd "$tmp"
curl -fsSL -o install.sh "https://raw.githubusercontent.com/${repo}/main/scripts/install.sh"
chmod +x install.sh

HARNESS_PREFIX="$tmp/bin" ./install.sh

bin="$tmp/bin/harness"
if [[ ! -x "$bin" ]]; then
  echo "✗ installer did not place harness at $bin"
  exit 1
fi

version="$("$bin" version)"
echo "→ installed: $version"

if ! echo "$version" | grep -q '^harness v0\.[0-9]\+\.[0-9]\+'; then
  echo "✗ unexpected version string: $version"
  exit 1
fi

"$bin" update status --channel stable >/dev/null 2>&1 || {
  echo "✗ harness update status failed"
  exit 1
}

"$bin" --help >/dev/null 2>&1 || {
  echo "✗ harness --help failed"
  exit 1
}

echo "✓ install smoke green"
