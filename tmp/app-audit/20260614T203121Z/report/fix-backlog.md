# Fix backlog

## P0 — Sessions list

- id: p01-public-landing-desktop
- route: /
- role: operator
- viewport: desktop
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p01-public-landing AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P0 — Sessions list

- id: p01-public-landing-mobile
- route: /
- role: operator
- viewport: mobile
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p01-public-landing AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P0 — Sessions list

- id: p01-public-landing-tablet
- route: /
- role: operator
- viewport: tablet
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p01-public-landing AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P0 — Workspace projects endpoint

- id: p08-workspace-api-desktop
- route: /api/workspace/projects
- role: operator
- viewport: desktop
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p08-workspace-api AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P0 — Catalog kinds endpoint

- id: p09-catalog-api-desktop
- route: /api/catalog/kinds
- role: operator
- viewport: desktop
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p09-catalog-api AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P1 — Sensors page

- id: p02-sensors-desktop
- route: /sensors
- role: operator
- viewport: desktop
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p02-sensors AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P1 — Sensors page

- id: p02-sensors-mobile
- route: /sensors
- role: operator
- viewport: mobile
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p02-sensors AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P1 — Agents page

- id: p03-agents-desktop
- route: /agents
- role: operator
- viewport: desktop
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p03-agents AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P1 — Memory page

- id: p04-memory-desktop
- route: /memory
- role: operator
- viewport: desktop
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p04-memory AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P1 — Settings

- id: p07-settings-desktop
- route: /settings
- role: admin
- viewport: desktop
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p07-settings AUDIT_ROLE=admin bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P1 — Autonomy matrix endpoint

- id: p10-autonomy-api-desktop
- route: /api/autonomy
- role: operator
- viewport: desktop
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p10-autonomy-api AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P2 — Design ingestion view

- id: p05-design-desktop
- route: /design
- role: operator
- viewport: desktop
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p05-design AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

## P2 — Roadmap view

- id: p06-roadmap-desktop
- route: /roadmap
- role: operator
- viewport: desktop
- failure: not_implemented
- reproduce: `AUDIT_FEATURE=p06-roadmap AUDIT_ROLE=operator bin/stack audit`
- expected: status=passed, http=200
- actual: AUDIT_PLAYWRIGHT_SKIP=1
- suggestion: re-run the audit with AUDIT_HEADED=1 to inspect interactively
- accept criteria: audit re-run reports status=passed and visual_diff_pct <= 5

