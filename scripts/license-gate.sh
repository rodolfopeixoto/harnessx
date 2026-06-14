#!/usr/bin/env bash
# license-gate: produce THIRD_PARTY_LICENSES.md + NOTICE + sbom.cyclonedx.json
# and block on copyleft licenses that conflict with HarnessX's MIT posture.
#
# Tools:
#   go-licenses  (mandatory; install: go install github.com/google/go-licenses@latest)
#   syft         (optional; richer SBOM in CycloneDX format)
set -euo pipefail

unset GOROOT || true

cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

BLOCKED='AGPL-3.0|GPL-3.0|GPL-2.0|LGPL-3.0|LGPL-2.1|SSPL|EUPL'

mkdir -p dist

if ! command -v go-licenses >/dev/null; then
  echo "✗ go-licenses missing — install via:"
  echo "    go install github.com/google/go-licenses@latest"
  exit 1
fi

echo "→ collecting third-party licenses"
csv="dist/third_party_licenses.csv"
go-licenses report ./cmd/harness --template "$(dirname "$0")/license-report.tpl" > THIRD_PARTY_LICENSES.md 2> /tmp/license.log || {
  echo "  (template render failed; falling back to CSV report)"
  go-licenses csv ./cmd/harness > "$csv"
}

# Also emit raw CSV unconditionally (machine-readable).
go-licenses csv ./cmd/harness > "$csv"

echo "→ NOTICE file"
{
  echo "HarnessX"
  echo "Copyright (c) 2026 Rodolfo Peixoto and contributors."
  echo
  echo "This product bundles the following third-party Go modules under their"
  echo "respective licenses. Full text of each license is reachable from the"
  echo "URLs listed in dist/third_party_licenses.csv."
  echo
  awk -F, '{ printf "  %s\t%s\t%s\n", $1, $2, $3 }' "$csv"
} > NOTICE

echo "→ checking for blocked licenses ($BLOCKED)"
if grep -E ",($BLOCKED)," "$csv" > /tmp/blocked.txt; then
  echo "✗ blocked license detected:"
  cat /tmp/blocked.txt
  exit 1
fi

echo "→ SBOM (CycloneDX)"
if command -v syft >/dev/null; then
  syft scan dir:. -o cyclonedx-json > dist/sbom.cyclonedx.json
else
  # Fallback: minimal hand-rolled CycloneDX from go list.
  python3 "$(dirname "$0")/sbom-fallback.py" > dist/sbom.cyclonedx.json
fi

echo "✓ license + SBOM gate passed"
echo "  THIRD_PARTY_LICENSES.md  — human-readable"
echo "  NOTICE                   — distribution-ready"
echo "  dist/third_party_licenses.csv"
echo "  dist/sbom.cyclonedx.json"
