# P21 — Per-role page tests + interaction grid

## Acceptance

- `web/dashboard/src/auth/roles.ts` defines `Role` (anonymous, operator, admin) + `RoleContext` provider.
- Every existing page renders under every role; pages that need elevated access call `useRole()` and disable mutating controls accordingly.
- `web/dashboard/src/__tests__/role-grid.test.tsx` walks every (role × page) combination and asserts render + at least one role-visible element.
- Interaction tests cover Tabs switch, InspectorPanel close, CommandPalette open/close/select, DataExplorer search + paginate + onInspect click.
- `scripts/test-grid.sh` runs vitest with coverage and prints a per-file table.

## Risks

- Test grid drift when new pages land. Mitigation: PAGES const drives both the grid + future docs; new page added without an entry fails the grid test.

## Verification

- Vitest passes; coverage of `src/auth/` ≥ 90%.
- `scripts/test-grid.sh` exits 0.
