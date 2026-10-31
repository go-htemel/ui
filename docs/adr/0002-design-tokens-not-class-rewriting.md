# 0002 — Theme with Tailwind v4 design tokens, not class rewriting

**Status:** Accepted (supersedes the `SetTheme` string-rewriting approach)

## Context
The first theme implementation rewrote class strings at build time (`-sky-` → `-indigo-`)
with per-theme safelists, three knobs, one theme per process.

## Decision
Components use semantic utilities (`bg-primary-600`, `text-neutral-500`,
`rounded-control`, `rounded-surface`, `success/warning/danger/info` families) defined as
CSS variables in a shipped `@theme` block (`assets/ui.css`, `cmd/themecss`). Apps override
variables globally or under `[data-theme]` scopes; nothing rebuilds. Plain `@theme` (not
`inline`) so scoped overrides apply live. One static safelist, generated from class-bearing
calls only.

## Consequences
- Per-tenant and per-scope themes with zero Go changes; the showcase switches live.
- Tailwind v4.1+ is required; the library is Tailwind-specific by design.
- A retheme to a light hue can fail contrast; the matrix test guards the defaults and the
  README says to run it after retheming.
