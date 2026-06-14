# Security policy

## Supported versions

HarnessX is pre-1.0. The `main` branch is supported; tagged releases are
supported for the duration noted in their release notes.

## Reporting a vulnerability

**Do not open a public issue for security bugs.**

Use one of:

1. GitHub Security Advisories — https://github.com/ropeixoto/harnessx/security/advisories/new
2. Email — security@harnessx.dev (PGP optional)

We will acknowledge receipt within 72 hours and aim to publish a fix or
mitigation within 14 days. Coordinated disclosure timelines are
negotiable for complex issues — please include a suggested timeline in
your report.

## Scope

In scope:

- Code execution / sandbox escape in the `harness` binary.
- Path traversal through `harness design-to-product` ZIP ingest.
- SQL injection through any persisted user input.
- Secrets leakage in logs, dashboards, or memory promotion.
- Authentication / authorisation issues in the dashboard HTTP server.
- Race conditions causing data corruption in `.harness/db/harness.sqlite`.

Out of scope:

- Vulnerabilities in third-party agent CLIs (Claude Code, Codex, Gemini,
  Kimi) themselves — report those upstream.
- Issues requiring physical access to the user's machine.

## Hall of fame

We credit reporters in the release notes unless anonymity is requested.
