<!-- mode: * -->
<!-- description: React/TS conventions, deterministic state, no useEffect dumps, no any. -->

## Frontend rule

- TypeScript strict. No `any`. Use `unknown` and narrow.
- Components: function components only. Single responsibility per file.
- State: local (`useState`) first, lifted second, store last. Avoid global state for transient UI.
- Side effects: `useEffect` is a smell when used for derived state or syncing with external systems beyond mount/unmount. Prefer event handlers and `useSyncExternalStore`.
- Memo when measured, not by default. Inline objects/arrays are fine until profiler says otherwise.
- Data fetching: React Query / TanStack Query or framework router loader; no naked `useEffect(fetch)`.
- Lists need stable keys. `index` only if items truly cannot reorder.
- A11y: semantic HTML before ARIA. `<button>` for actions, `<a href>` for navigation. Never `onClick` on `<div>`.
- Forms: controlled inputs, schema validation (zod/yup), submit handler returns promise.
- Tests: Testing Library queries by role/label; never by class. One assertion of user-visible outcome per test.
