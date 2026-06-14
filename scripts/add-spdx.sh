#!/usr/bin/env bash
# Bulk-add `// SPDX-License-Identifier: MIT` to every Go file under
# internal/ and cmd/ that doesn't already have one. Idempotent.
set -euo pipefail
cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

header='// SPDX-License-Identifier: MIT'

while IFS= read -r f; do
  if head -1 "$f" | grep -q 'SPDX-License-Identifier'; then continue; fi
  tmp="${f}.spdx.tmp"
  { echo "$header"; echo; cat "$f"; } > "$tmp"
  mv "$tmp" "$f"
  echo "  + $f"
done < <(find internal cmd -type f -name '*.go' -not -path '*/testdata/*' -not -name '*_test.go')
echo "✓ SPDX headers up to date"
