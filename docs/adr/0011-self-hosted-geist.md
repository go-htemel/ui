# 0011 — Self-hosted Geist, opt-in

**Status:** Accepted

## Context
`system-ui` resolves to Segoe UI on Windows, whose metrics sit text about a pixel high in
buttons. Google Fonts was a third-party request.

## Decision
Ship Geist and Geist Mono variable fonts (OFL, ~70 KB each) in `assets/`, served by
`Assets()`; `FontCSS(prefix)` and `FontPreload(prefix)` opt an app in. The library default
stays the system stack. Body weight stays 400.

## Consequences
- No third-party requests; Geist Mono flows into code blocks via `--font-mono`.
