#!/usr/bin/env bash
# Phase 7 end-to-end: design-to-product against folder + zip fixtures.
set -euo pipefail

unset GOROOT || true

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/bin/harness"

if [[ ! -x "$BIN" ]]; then
  (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/harness)
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cp -R "$ROOT/testdata/projects/sample-go/." "$WORK/"
cp -R "$ROOT/testdata/designs/sample-design" "$WORK/sample-design"
cd "$WORK"
git init -q
git add -A
git -c user.email=t@t -c user.name=t commit -q -m init
"$BIN" init >/dev/null

echo "→ design-to-product (folder)"
"$BIN" design-to-product "convert this design" --source ./sample-design > /tmp/hx-dtp.txt
cat /tmp/hx-dtp.txt
grep -q "Detected Design-to-Product mode" /tmp/hx-dtp.txt
for f in design-manifest feature-map toggle-map roadmap api-contracts flow-map; do
  test -s ".harness/product/${f}.json" || { echo "missing ${f}.json"; exit 1; }
done
grep -q '"pages"' .harness/product/design-manifest.json
grep -q '"feature.signup"' .harness/product/feature-map.json
grep -q '"MVP 0"' .harness/product/roadmap.json

echo "→ design-to-product (zip)"
(cd sample-design && zip -qr ../sample-design.zip .)
"$BIN" design-to-product "use this Claude Design zip" --source ./sample-design.zip > /tmp/hx-dtp-zip.txt
grep -q "Detected Design-to-Product mode" /tmp/hx-dtp-zip.txt
test -s .harness/product/design-manifest.json

echo "→ resolveFromPrompt (path in prompt)"
"$BIN" design-to-product "convert ./sample-design into React parity" > /tmp/hx-dtp-nat.txt
grep -q "source=" /tmp/hx-dtp-nat.txt

if command -v sqlite3 >/dev/null; then
  COUNT="$(sqlite3 .harness/db/harness.sqlite "select count(*) from sessions where mode='design_to_product';")"
  test "$COUNT" -ge 1
fi

echo "✓ e2e-phase7 passed"
