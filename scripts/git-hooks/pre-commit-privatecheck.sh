#!/usr/bin/env bash
# Pre-commit guardrail — blocks personal / marketing / YouTube-channel
# material from ever reaching the open-source repo.
#
# Rationale: harnessx is public. Rodolfo's YouTube channel material
# (roteiros, thumb briefs, SEO plans, engagement growth, playlists,
# shorts factory, brand-thumb-title matrices) is personal marketing IP
# and must live outside the repo (backup at ~/harnessx-notes/).
#
# Bypass (audit when used): HARNESS_ALLOW_PRIVATE=1 git commit ...
set -euo pipefail

if [[ "${HARNESS_ALLOW_PRIVATE:-}" == "1" ]]; then
  echo "⚠ HARNESS_ALLOW_PRIVATE=1 — private-material check bypassed."
  exit 0
fi

cd "$(git rev-parse --show-toplevel)"

staged="$(git diff --cached --name-only --diff-filter=ACMR || true)"
if [[ -z "$staged" ]]; then
  exit 0
fi

# Path patterns — directories & filename markers that indicate personal
# marketing material.
path_patterns=(
  '^docs/youtube/'
  '^docs/marketing/'
  '^docs/canal/'
  '^docs/roteiros/'
  'THUMB'
  'SHORTS'
  'ENGAGEMENT-GROWTH'
  'BRAND-.*MATRIX'
  'SEO-RETENCAO'
  'PLAYLISTS-GROWTH'
  'SERIE-DENTRO-DO-HARNESSX'
  'SERIE-PAPERS-NA-PRATICA'
)

offenders=()
for f in $staged; do
  for p in "${path_patterns[@]}"; do
    if [[ "$f" =~ $p ]]; then
      offenders+=("$f  [path match: $p]")
      break
    fi
  done
done

# Content patterns — marketing keywords that shouldn't appear in
# open-source docs even if the filename looks innocent.
content_patterns=(
  'CTR ≥'
  'AVD ≥'
  'Subs 30d'
  'thumb brief'
  'Shorts factory'
  'canal do YouTube'
)

for f in $staged; do
  # Skip binary / deleted / self-referential files. A file may opt out of
  # the content scan by including the marker `privatecheck: allow` in a
  # comment (used by docs that legitimately describe the rule).
  [[ ! -f "$f" ]] && continue
  [[ "$f" == "scripts/git-hooks/pre-commit-privatecheck.sh" ]] && continue
  [[ "$f" == *".gitignore" ]] && continue
  if file --mime "$f" 2>/dev/null | grep -q 'charset=binary'; then
    continue
  fi
  if grep -qF 'privatecheck: allow' "$f" 2>/dev/null; then
    continue
  fi
  for p in "${content_patterns[@]}"; do
    if grep -qF -- "$p" "$f" 2>/dev/null; then
      offenders+=("$f  [content match: '$p']")
      break
    fi
  done
done

if [[ ${#offenders[@]} -gt 0 ]]; then
  echo
  echo "✗ Personal marketing material blocked from open-source repo."
  echo
  echo "Offending staged paths / content:"
  for o in "${offenders[@]}"; do
    echo "  - $o"
  done
  echo
  echo "Move these files to ~/harnessx-notes/ (or another private location)"
  echo "and unstage them before committing:"
  echo "  git restore --staged <file>"
  echo
  echo "Emergency bypass (audit trail required):"
  echo "  HARNESS_ALLOW_PRIVATE=1 git commit ..."
  echo
  exit 1
fi

exit 0
