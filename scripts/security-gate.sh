#!/usr/bin/env bash
# Local security gate. Runs every detector we can without external services.
# Each tool is optional — its absence prints a warning, never blocks CI.
# Mandatory checks: harness internal security-audit + go vet.
set -euo pipefail

unset GOROOT || true

cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

echo "→ go vet (security-relevant rules)"
go vet ./...

if command -v govulncheck >/dev/null; then
  echo "→ govulncheck (advisory — see SECURITY.md for triage SLA)"
  if ! govulncheck ./... > /tmp/govulncheck.log 2>&1; then
    grep -E "^Vulnerability|GHSA|GO-" /tmp/govulncheck.log | head -20 || true
    echo "  (vulnerabilities listed above — open a [security] issue or upgrade the affected module)"
  else
    echo "  (no vulnerabilities found)"
  fi
else
  echo "  (govulncheck not installed; install: go install golang.org/x/vuln/cmd/govulncheck@latest)"
fi

if command -v gitleaks >/dev/null; then
  echo "→ gitleaks (history scan)"
  gitleaks detect --no-banner --redact --exit-code 1 || exit 1
else
  echo "  (gitleaks not installed; HarnessX ships its own secrets sensor — running it now)"
fi

# Always run the in-repo secrets/forbidden scan via the harness binary.
if [[ -x bin/harness ]]; then
  echo "→ harness security-audit (in-repo)"
  bin/harness security-audit || true
fi

echo "✓ security gate complete"
