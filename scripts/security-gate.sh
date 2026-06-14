#!/usr/bin/env bash
# Local security gate. Runs every detector we can without external services.
# Each tool is optional — its absence prints a warning, never blocks CI.
# Mandatory checks: harness internal security-audit + go vet.
set -euo pipefail

unset GOROOT || true

cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

echo "→ go vet (security-relevant rules)"
go vet ./...

# Add user-local GOBIN to PATH so `go install` targets are reachable.
export PATH="$PATH:$HOME/go/bin"

if ! command -v govulncheck >/dev/null; then
  echo "✗ govulncheck missing — install via:"
  echo "    go install golang.org/x/vuln/cmd/govulncheck@latest"
  exit 1
fi
echo "→ govulncheck"
if ! govulncheck ./... > /tmp/govulncheck.log 2>&1; then
  grep -E "^Vulnerability|GHSA|GO-" /tmp/govulncheck.log | head -20 || true
  echo "✗ vulnerabilities detected — upgrade affected modules or document waiver in SECURITY.md"
  exit 1
fi
echo "  (no vulnerabilities found)"

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
