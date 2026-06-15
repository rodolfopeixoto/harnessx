<!-- mode: * -->
<!-- description: SRP, no comment noise, named constants, intent-revealing identifiers. -->

## Clean-code rule

- One public type per file when reasonable; helpers private.
- gocognit ≤ 25, gocyclo ≤ 15 per function; extract instead of suppressing the lint.
- Comments only for WHY-non-obvious (hidden constraint, workaround, surprising invariant). Never narrate what code does.
- Constants in `internal/platform/constants`; no magic numbers in business logic.
- Names reveal intent. `n int` is ok inside a 5-line loop, never in a public signature.
- Errors include context (`fmt.Errorf("backup: read manifest: %w", err)`); never `_ = err`.
- Tests target the seam, not the line. Prefer table-driven tests with named cases.
