# 0003 — Datastar for all interactivity; one small vanilla `ui.js`

**Status:** Accepted

## Context
Hypermedia-first apps with Datastar; the owner wanted no JS framework and no bundler.

## Decision
Every interaction is expressed with htemel's `datastar` helpers (never hand-written
`data-*`; a test enforces it). Behaviour that genuinely needs script (touch drag, editor
toolbar, tag input, file upload) lives in light-DOM custom elements in a single embedded
`assets/ui.js` served by `ui.Assets()`, dispatching plain DOM events (`ui:drop`,
`ui:tags`, `ui:upload`) that components listen to with `data-on`. Library-owned attributes
use the `data-ui-*` namespace. The only inline script is `ColorSchemeBootstrap` (no-flash
dark mode), with a nonce/hash option for CSP.

## Consequences
- Progressive enhancement by construction: without `ui.js` the pick-up buttons, plain
  textarea and native inputs still work.
- Sharp edge: handler expressions are JavaScript in Go strings; `datastar.Quote`,
  `ui.Then` and the custom elements keep them short, but there is no type help inside them.
- Datastar Pro features (`data-persist`) are not assumed; rail state uses localStorage.
