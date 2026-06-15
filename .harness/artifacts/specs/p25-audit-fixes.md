# P25 — audit findings fixed

## Acceptance

- Audit pass_rate ≥ 95% on a fresh checkout (zero data, no project init).
- /api/profile, /api/design, /api/roadmap, /api/toggles, /api/features return 200 empty envelopes when their backing files are absent.
- Shell + DataExplorer respect mobile viewport (no horizontal overflow at 390 px).
- Audit Playwright spec ignores favicon noise.

## Risks

- Empty envelopes mask real configuration mistakes. Mitigation: dashboard still shows "no X yet" empty state — readable signal, no silent failure.

## Verification

- bin/stack audit reports pass_rate = 1.0 on develop tip.
