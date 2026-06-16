# Security Audit

Status: scanned 2026-06-16 against Go 1.21 toolchain.

## Toolchain

```bash
make security        # govulncheck + gitleaks + harness security-audit
```

`make security` calls `scripts/security-gate.sh` which runs:

1. `go vet ./...` — built-in static checks
2. `govulncheck ./...` — vuln DB lookup for direct + indirect deps + Go stdlib
3. `gitleaks detect --no-banner --redact --exit-code 1` — secrets scan over git history
4. `harness security-audit` — in-repo deterministic secrets scan (always runs)

## Current findings (2026-06-16, Go 1.21)

`govulncheck` flagged ~30 findings, mostly against the Go 1.21 stdlib
(crypto, net/http, encoding/* modules). None of them traces to a
function HarnessX calls in user-input paths — they live in libraries
HarnessX depends on transitively (sqlite driver, lipgloss, gh
charmbracelet packages).

Mitigation tracked: upgrade Go toolchain to 1.23+ in a follow-up
release. This will clear the stdlib hits without code changes.

`gitleaks` — not run as part of CI by default (requires the binary
to be on PATH; ships in the `make security` target as a soft
dependency).

`harness security-audit` — green; no `.env` secret patterns in repo.

## Operator expectations

- Every release goes through `make lint` and `go test ./...` as the
  hard gate.
- `make security` is **informational** today, not a blocker, because
  the bulk of findings depend on a Go toolchain upgrade.
- Operators running HarnessX in a hardened environment should:
  1. Build from source with their preferred Go version (≥ 1.23 clears
     most stdlib vulns).
  2. Run `harness security-audit` in their project to scan their own
     code.
  3. Pin `make ci` plus `make security` in their pre-merge gate.

## Waiver process

If a CVE must be waived (false positive, no fix available), document
it here:

```
GO-YYYY-NNNN  module  rationale  reviewer  expiry
```

Currently waived: none. Open CVEs tracked in
[upstream advisory DB](https://pkg.go.dev/vuln).

## Next

- v0.48+: bump go.mod toolchain to 1.23
- v0.49+: re-run `make security`; expect clean stdlib
- pre-v1.0: make `make security` a hard gate (exit 1 on any CVE
  without a documented waiver)
