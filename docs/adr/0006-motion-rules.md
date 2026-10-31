# 0006 — Motion: 150/200 ms, compositor-only, reduced-motion everywhere

**Status:** Accepted

## Decision
Colour/opacity on hover and focus: 150 ms. Things appearing or leaving: 200 ms ease-out,
both directions (`@starting-style` + `transition-discrete` for dialogs, drawers,
popovers, toasts). Only opacity, transform and grid rows animate; never width, height or
position. Every non-colour transition carries `motion-reduce:transition-none`. A golden
based test enforces all of this.

## Consequences
- Consistent feel without per-component judgement.
- The collapsible rail switches width instantly by design (animating grid columns would
  reflow long content every frame).
