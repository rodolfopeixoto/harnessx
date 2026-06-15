<!-- mode: * -->
<!-- description: Never log secrets; validate at boundaries; least privilege. -->

## Security rule

- Never log secrets or their resolved values.
- Validate input at the system boundary (HTTP handler, CLI parser, external file read), not inside helpers.
- Least privilege file modes: 0o600 for secret material, 0o644 for data files, 0o755 for binaries.
- Refuse `..` and absolute paths when extracting archives.
- HTTP responses must not expose stack traces or internal paths.
- `exec.CommandContext` always carries a timeout; never spawn a child without context.
