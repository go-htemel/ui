# 0005 — Measured contrast policy

**Status:** Accepted

## Context
Nothing measured colour until a Go WCAG matrix (exact Tailwind OKLCH values) and an
axe-core run (both schemes, overlays open) were added; 32 pairs failed, then the fixes
made field borders stark at 4.8:1.

## Decision
- Text ≥ 4.5:1, large text and non-text UI ≥ 3:1, enforced by `go test` over a curated
  matrix (~570 pairs) and by axe in CI; disabled controls reported, not enforced.
- Text-field borders follow WCAG 1.4.11's Understanding text: a field identified by label,
  fill and a ≥3:1 focus ring does not need a 3:1 boundary. Default `--color-field-border`
  is neutral-300 / white-15 (Tailwind UI weight); rows are informational. Checkbox, radio
  and switch boundaries stay enforced at 3:1 (`--color-control-border`). Strict opt-in
  documented.
- Focus rings are primary everywhere; solids keep ≥4.5:1 with white in both schemes, dark
  mode hovers lighter.

## Consequences
- Colour changes are safe to make: the tests say immediately what broke.
- Accepted risk: informational rows can drift; review them when retheming.
