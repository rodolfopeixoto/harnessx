<!-- mode: design_to_product -->
<!-- description: Tokens, parity to design specs, accessibility, no inline magic colors. -->

## Design rule

- Tokens, not literals. Colors from `tokens.color`, spacing from `tokens.space`, type from `tokens.type`. No `#FFFFFF`, no `padding: 12px`, no `font-family: sans-serif` inline.
- Layout primitives (Stack, Cluster, Grid) before custom flex. Build complex from compositions.
- Accessibility default: keyboard focusable, visible focus ring, ARIA labels on interactive elements, contrast ≥ 4.5:1 for body text.
- Responsive baseline: mobile-first, content-driven breakpoints, no fixed pixel widths above 640px.
- Empty / loading / error states for every async surface. Skeletons over spinners when the layout is known.
- Animations ≤ 200ms, respect `prefers-reduced-motion`.
- Forms: label every field, inline validation, primary action right-aligned (desktop) / full-width (mobile).
- Tests: one snapshot per state (loaded/empty/error), one a11y assertion (axe or jest-axe).
