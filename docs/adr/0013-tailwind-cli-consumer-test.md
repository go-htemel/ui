# 0013 — CI builds a real consumer app with the Tailwind CLI

**Status:** Proposed

## Context
The showcase compiles CSS with the Tailwind browser build and CDN Datastar. A production
app's path — Tailwind CLI, `@import "./ui.theme.css"`, `@source` of the safelist — is
documented but not exercised by CI.

## Decision (proposed)
Add a tiny consumer module under `tools/consumer/` built in CI with the Tailwind CLI,
asserting the emitted CSS contains every class in the safelist and that the app renders
one page using `Assets()`, `FontCSS`, `ColorSchemeBootstrap`.

## Consequences
- Catches safelist/token regressions where users would actually hit them.
