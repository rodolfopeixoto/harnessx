# P106 — Evidence-bundle verifiers (paper § 5.2.2)

## Context

Paper open challenge: "semantic verification beyond executable
feedback". Sensors today emit pass/fail/skip + a confidence field
(v0.34) that is only sometimes populated. Operators get green from
"build passed" with no idea what scope was actually verified.

This release upgrades `sensors.Result` to carry the **evidence
bundle**: what was checked (Scope), what was confirmed
(Verified), what was not (Unverified), what could still go wrong
(Risks). `harness loop` uses the bundle as a gate: a green attempt
with low Confidence AND non-empty Unverified is treated as failed
verification, not pass.

## What ships

- `sensors.Result` gains `Scope`, `Verified`, `Unverified`, `Risks`
  (all `[]string`) alongside existing `Confidence`.
- New helper `sensors.MakeBundle(scope, verified, unverified, risks)`
  callers can use without boilerplate.
- `devloop.checkVerification(att Attempt) (passed bool, reason string)`
  refuses green when **any attempt sensor** has
  `Confidence < 0.5 && len(Unverified) > 0`.
- Renderer (`internal/app/sensorcmd/sensorcmd.go`) shows verified
  count / unverified count when bundle populated.
- Bundled scanners populate their bundle:
  - `secrets_scan`: Scope=`["*.env", "**/*.{key,pem}"]`,
    Verified=count files scanned, Unverified=files skipped (e.g.
    binary), Risks=none unless a near-miss pattern found.
  - `forbidden_files`: Scope=`["./**"]`, Verified=count paths
    checked.
  - `changed_files`: Scope=`["git diff HEAD"]`.
  - `performance_budget`: Scope=`["build size"]`,
    Unverified=`["runtime memory"]` when only static measurement
    available.

## Critical files

| Path | Change |
|---|---|
| `internal/sensors/types.go` | add 4 fields + `MakeBundle` helper |
| `internal/sensors/scanners.go` | populate bundle in every scanner |
| `internal/devloop/loop.go` | add `checkVerification` gate |
| `internal/app/sensorcmd/sensorcmd.go` | render bundle counts when present |

## Reuse, do not reinvent

- `sensors.Result.Confidence` (v0.34) — already exists, only need to
  populate consistently.
- `devloop.checkRegression` pattern — `checkVerification` mirrors it.
- `app/sensorcmd::detail` — extend to include verified/unverified
  counts only when non-zero.

## Verification

- `make lint` 0 issues.
- `go test -cover ./internal/sensors/... ./internal/devloop/...` —
  new tests for `MakeBundle`, populated-bundle scanners, and the
  `checkVerification` gate.
- Smoke:
  ```
  harness sensor run secrets_scan --root .   # shows verified count
  harness loop "intentionally low-confidence task" \
      --agent fake-real --apply --max-attempts 2
  # devloop refuses green when bundle says unverified + low conf
  ```

## Risks + mitigation

| Risk | Mitigation |
|---|---|
| Existing operators see new fields | Fields optional, omitempty JSON tag |
| Bundle gate breaks legitimate passes | Only triggers when conf < 0.5 AND len(Unverified) > 0; both bars must be true |
| Bundle wording inconsistent across scanners | Helper `MakeBundle` provides canonical struct |

## Acceptance

- Every bundled scanner returns at least Scope + Confidence ≥ 0.5
  on success.
- `harness loop` refuses green when any sensor has
  `Confidence < 0.5 AND len(Unverified) > 0`.
- `internal/sensors` coverage holds ≥ 68%; `internal/devloop` rises
  to ≥ 55%.
