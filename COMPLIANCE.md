# Compliance posture

HarnessX is designed to be drop-in for private companies, public-sector
projects, and open-source teams. This document maps what HarnessX is and
is not, so legal / security / procurement reviewers can ship a sign-off
without source-reading the entire repository.

## License

| Layer | License | File |
|---|---|---|
| HarnessX itself | MIT | `LICENSE` |
| Third-party Go modules | various (audited) | `THIRD_PARTY_LICENSES.md`, `dist/third_party_licenses.csv` |
| React dashboard runtime | MIT (React, Vite) | `web/dashboard/package.json` |
| LSP servers, agent CLIs | external — operator installs separately | n/a |

**Blocked licenses (enforced by `make licenses`):** AGPL-3.0, GPL-2.0,
GPL-3.0, LGPL-2.1, LGPL-3.0, SSPL, EUPL. Any PR introducing a transitive
dep with one of these fails the gate. Override only with explicit legal
sign-off.

## Data handling

HarnessX is **local-first by design**:

- All persistent state lives under `<project>/.harness/`. Nothing is
  uploaded to a remote service.
- No outbound network calls are made by the `harness` binary itself.
  Agent CLIs (Claude, Codex, Gemini, Kimi) talk to their respective
  vendors over the network; that connection is owned by the operator,
  not by HarnessX.
- Local sqlite database under `.harness/db/harness.sqlite`. JSONL event
  log under `.harness/logs/events.jsonl`. Both are append-mostly.
- Dashboard binds to `127.0.0.1:7373` by default; rebind only on trusted
  networks.

### Personal data (GDPR-style)

HarnessX does not collect personal data. The project root, the current
user's OS-reported `$USER`, and the project name (folder basename) end
up in sqlite session rows. To erase:

```bash
rm -rf .harness/db .harness/logs .harness/cache .harness/artifacts
```

`harness session show <id>` and `harness memory list` operate purely on
local state; no remote queries occur.

### Secrets

`.env`, `id_rsa`, `*.pem`, `secrets.yml` and similar files are blocked
by the `forbidden_files` sensor. AWS / Slack / GitHub / PEM patterns
are blocked by `secrets_scan`. Memory promotion rejects content
matching the same vocabulary (`internal/memory.Promote`).

## Audit trail

Every operation persists evidence:

- `sessions` + `runs` + `sensor_results` + `agent_certifications` +
  `metrics` + `memories` + `skill_versions` + `artifacts` in
  `.harness/db/harness.sqlite` (see `internal/adapters/sqlite/migrations/0001_init.sql`).
- One JSONL line per event in `.harness/logs/events.jsonl` with size
  rotation at 10 MiB.
- Generated specs / plans / reports / perf snapshots under
  `.harness/artifacts/` with content-hashed filenames.

The dashboard `/api/*` surface is read-only — write paths only go
through the CLI which runs through the same telemetry layer.

## Supply chain

- `go-licenses` (`make licenses`) emits `THIRD_PARTY_LICENSES.md` +
  `NOTICE` + machine-readable CSV before every release.
- `make sbom` produces `dist/sbom.cyclonedx.json` (CycloneDX 1.5).
  Falls back to a stdlib-only Python generator when `syft` is absent.
- `make security` runs `govulncheck` + the in-repo `harness
  security-audit` (forbidden files, forbidden commands, secrets scan).
- Reproducible builds via `go build -trimpath -ldflags '-s -w' -X …`.
  No CGO. Single static binary per OS/arch.
- Release tarballs include SHA-256 sums (`make release`).

## Build provenance

`scripts/install.sh` verifies the SHA-256 of each downloaded tarball
before installing. Pair with sigstore / cosign at the distribution
edge for stronger guarantees (optional; not enforced by HarnessX).

Recommended addition for regulated environments:

```bash
# After `make release`
cosign sign-blob --yes dist/harness-linux-amd64.tar.gz \
  --output-signature dist/harness-linux-amd64.tar.gz.sig
```

## Local-only CI/CD

There is no hosted runner. Every `git push` runs through the local
`pre-push` hook (`make ci` = lint + tests + coverage gate + all phase
e2e). Releases are produced by `make cd` on the maintainer's
workstation. This eliminates dependency on third-party CI providers
and keeps source code, build environment, and signing keys on machines
the operator controls.

## SOC-2 mapping (informational)

| Control area | HarnessX feature |
|---|---|
| CC6.1 Logical access | Loopback-bound dashboard; no remote admin surface. |
| CC6.6 Encryption in transit | Operator-controlled (HTTPS at agent CLI level). |
| CC7.1 Detection | `secrets_scan`, `forbidden_files`, `forbidden_commands`, `go_vuln`, brakeman/bandit/cargo-audit per stack. |
| CC7.2 Monitoring | JSONL event log + sqlite metrics + dashboard panels. |
| CC7.3 Incident response | `harness security-audit` + `harness report` produce evidence packets. |
| CC7.4 Vulnerability mgmt | `make security` + `make licenses` gated on every release. |
| CC8.1 Change mgmt | Spec-driven development; GitFlow; commit-msg + pre-push hooks. |

This is a mapping, not a certification. Auditors should treat it as a
starting checklist, not evidence of compliance.

## Reporting non-compliance

If you believe HarnessX or its outputs violate a license, breach a
regulation, or mishandle data, report via the channel in `SECURITY.md`.
We acknowledge within 72 hours and aim for a fix or mitigation within
14 days.
