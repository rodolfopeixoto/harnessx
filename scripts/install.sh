#!/usr/bin/env bash
# HarnessX one-line installer.
#
#   curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh | bash
#
# Detects OS + arch, downloads the matching tarball from the latest GitHub
# release, verifies SHA-256, installs into ${HARNESS_PREFIX:-/usr/local/bin}.
#
# Flags:
#   --dry-run        plan + verify checksum without installing
#   --prefix <dir>   install dir (overrides HARNESS_PREFIX env)
set -euo pipefail

REPO="${HARNESS_REPO:-rodolfopeixoto/harnessx}"
PREFIX="${HARNESS_PREFIX:-/usr/local/bin}"
DRY_RUN=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1; shift ;;
    --prefix) PREFIX="$2"; shift 2 ;;
    --help|-h) grep -E "^# " "$0" | sed 's/^# //'; exit 0 ;;
    *) echo "unknown flag: $1" >&2; exit 2 ;;
  esac
done

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac
case "$os" in
  darwin|linux) ;;
  *) echo "unsupported OS: $os" >&2; exit 1 ;;
esac

target="harness-${os}-${arch}"
tag="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | \
       grep -m1 '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"
if [[ -z "${tag}" ]]; then
  echo "could not resolve latest tag from GitHub API" >&2
  exit 1
fi

url="https://github.com/${REPO}/releases/download/${tag}/${target}.tar.gz"
sha_url="${url}.sha256"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
echo "→ resolved tag: ${tag}"
echo "→ downloading ${url}"
curl -fsSL "${url}"     -o "${tmp}/${target}.tar.gz"
curl -fsSL "${sha_url}" -o "${tmp}/${target}.tar.gz.sha256"

echo "→ verifying SHA-256"
(cd "${tmp}" && shasum -a 256 -c "${target}.tar.gz.sha256")

echo "→ extracting"
tar -xzf "${tmp}/${target}.tar.gz" -C "${tmp}"

if [[ "${DRY_RUN}" -eq 1 ]]; then
  echo "→ dry-run: would install ${tmp}/${target} to ${PREFIX}/harness"
  "${tmp}/${target}" version
  exit 0
fi

dest="${PREFIX}/harness"
echo "→ installing ${dest}"
if [[ -w "${PREFIX}" ]]; then
  install -m 0755 "${tmp}/${target}" "${dest}"
else
  sudo install -m 0755 "${tmp}/${target}" "${dest}"
fi

echo
"${dest}" version
echo
echo "next steps:"
echo "  cd your-project"
echo "  harness init"
echo "  harness doctor"
echo "  harness stack tour --keep            # walk every feature deterministically"
echo "  harness completion bash > /etc/bash_completion.d/harness   # optional"
