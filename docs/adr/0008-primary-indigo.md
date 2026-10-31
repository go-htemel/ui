# 0008 — Default primary is stock indigo

**Status:** Accepted (sky tuned and "Emerald Sky" were tried and rejected)

## Context
Sky-600 carries white text at only 4.0:1, so sky needs the 700 step for solids and has no
lighter hover step in dark mode. A custom green→blue drifting palette passed every test but
failed the eyeball test in context: tints leaned green, surfaces blue.

## Decision
`primary` aliases Tailwind indigo. Conventional steps hold with margin: 600 solids
(6.5:1), 700 light hover, 600→500 dark hover (4.6:1), 600/400 text. `info` aliases sky so it
does not read as a second primary; a `Brand` tone gives primary-tinted badges for labels.

## Consequences
- No special rules; anyone retheming to a light hue is told by the contrast test.
- `cmd/palettetune` remains for custom palettes.
