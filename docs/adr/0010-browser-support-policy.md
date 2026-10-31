# 0010 — Platform-first with fallbacks; support matrix published, Safari unverified

**Status:** Accepted (risk)

## Decision
Use modern platform features with tested fallbacks: `closedby`, anchor positioning,
customizable `<select>` (`appearance: base-select`), `@starting-style`,
`interpolate-size`, container queries. The README carries a support table.

## Consequences
- Chromium users get the full experience; Firefox gets plainer popovers and native select
  lists; nothing breaks.
- **Risk accepted for v0.1:** no human has looked at the library in Safari or Firefox.
  ADR 0014 schedules that before v0.2.
