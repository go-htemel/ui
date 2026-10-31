# Architecture decision records

Decisions behind go-htemel/ui, in the order they were made. **Accepted** records describe
what was built and the consequences we know; **Proposed** records are agreed direction
not yet implemented. Sharp edges are listed under consequences so that living with one is
a decision, not an oversight.

| # | Title | Status |
|---|---|---|
| [0001](0001-components-return-elements.md) | Components are functions returning concrete elements | Accepted |
| [0002](0002-design-tokens-not-class-rewriting.md) | Theme with Tailwind v4 design tokens, not class rewriting | Accepted |
| [0003](0003-datastar-only-interactivity.md) | Datastar for all interactivity; one small vanilla `ui.js` | Accepted |
| [0004](0004-native-overlays-no-shadow-dom.md) | Native `<dialog>`/Popover for overlays; light-DOM custom elements; no shadow DOM | Accepted |
| [0005](0005-contrast-policy.md) | Measured contrast policy, including the text-field border interpretation | Accepted |
| [0006](0006-motion-rules.md) | Motion: 150/200 ms, compositor-only, reduced-motion everywhere | Accepted |
| [0007](0007-container-queries.md) | Components adapt to their container, not the viewport | Accepted |
| [0008](0008-primary-indigo.md) | Default primary is stock indigo | Accepted |
| [0009](0009-markdown-over-contenteditable.md) | Markdown editor instead of a rich-text editor | Accepted |
| [0010](0010-browser-support-policy.md) | Platform-first with fallbacks; support matrix published, Safari unverified | Accepted (risk) |
| [0011](0011-self-hosted-geist.md) | Self-hosted Geist, opt-in | Accepted |
| [0012](0012-variadic-children-and-trailing-options.md) | Variadic children win over trailing options | Accepted (known wart) |
| [0013](0013-tailwind-cli-consumer-test.md) | CI builds a real consumer app with the Tailwind CLI | Proposed |
| [0014](0014-cross-browser-verification.md) | Human Safari/Firefox pass before v0.2 | Proposed |
| [0015](0015-adopt-htemel-0-4.md) | Adopt htemel v0.4.0 (additive `Class`, no per-element conditionals, composite helper) | Proposed |
