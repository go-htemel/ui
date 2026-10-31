# 0004 — Native `<dialog>`/Popover for overlays; light-DOM custom elements; no shadow DOM

**Status:** Accepted

## Context
The first Dialog was a `div` with `data-show`, a hand-rolled focus trap and scroll lock.
Shadow DOM was considered for encapsulation.

## Decision
Dialog and Drawer are native modal `<dialog>` elements opened with `showModal()` (inert
page, Escape, focus containment and return, top layer, `closedby` with a click-outside
fallback). Dropdown uses the Popover API with anchor positioning and a JS-free positioning
fallback. Custom elements use the light DOM only: shadow roots would hide markup from
Datastar and from Tailwind utilities.

## Consequences
- Less code and better accessibility than the hand-rolled versions.
- Overlays escape framed previews (top layer) — the showcase uses iframes for narrow
  previews.
- Initial focus goes to the first form control, else the dialog itself, so the close
  button never shows a focus ring on open.
