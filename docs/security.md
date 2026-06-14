# Security

HarnessX is a **safe agent harness**. Every layer assumes adversarial input — from the prompts users type to the YAML adapters they install.

## Block-list

The following actions are blocked at the framework level:

- Reading `.env` (and friends) unless explicitly approved.
- Editing or printing secrets.
- Destructive shell commands: `chmod -R 777`, `rm -rf /`, `sudo rm -rf /`, `curl | bash`, `wget | bash`.
- Force pushes (`git push --force`).
- Disabling tests (`--no-verify`).
- Deleting tests to pass gates (Phase 9 forbidden-files allowlist).
- Removing security tools to reduce image size.
- Changing production config without approval.

Enforcement points:

| Layer | Sensor / guard |
|---|---|
| Project tree | `forbidden_files` (.env, *.pem, id_rsa, secrets.yml, …) |
| Shell snippets | `forbidden_commands` (regex) |
| Source files | `secrets_scan` (AKIA, AWS secret, Slack tokens, PEM blocks, GitHub tokens, `api_key='…'`) |
| ZIP extraction | `internal/design.extractZip` rejects path traversal + 200 MiB cap |
| Agent execution | router stops fallback on `FailureAuth` / `FailurePermanent` |
| Memory promotion | `internal/memory.Promote` rejects sensitive content with same vocabulary as `secrets_scan` |

## Security sensors

```bash
harness security-audit   # runs forbidden_files + forbidden_commands + secrets_scan +
                         # go_vuln + ruby_brakeman + py_bandit + rust_audit
```

Per-stack security sensors are `OptionalTool: true` so missing binaries skip rather than fail. Install the relevant tool to elevate the audit:

- Go: `govulncheck`
- Rails: `brakeman`, `bundle-audit`
- Python: `bandit`, `pip-audit`
- Rust: `cargo audit`
- Docker: `hadolint`

## Memory policy

Memory entries are written through `internal/memory.Promote`. The gate rejects:

- Missing `evidence_run_id` (every memory must trace back to a run).
- Confidence below `0.4`.
- Content matching the sensitive-token regex.
- Empty content.

Forbidden memory categories (per spec §11): unverified assumptions, failed agent claims, secrets/tokens/credentials, temporary guesses, stale results without timestamp.

## Network isolation

HarnessX itself makes zero outbound network calls. Agent CLIs are invoked through adapters; their network access is the user's responsibility.

The dashboard binds to `127.0.0.1:7373` by default. Re-bind to `0.0.0.0` only for trusted networks.

## Audit trail

Every action persists to `.harness/db/harness.sqlite` plus a JSONL line in `.harness/logs/events.jsonl`. Sessions, runs, sensor results, agent certifications, and artifacts all carry timestamps + IDs. The dashboard surfaces the same data read-only.
