# 0015 — Adopt htemel v0.4.0

**Status:** Proposed

## Decision (proposed)
When htemel v0.4.0 lands (htemel ADRs 0008, 0009, 0011, 0013): migrate the handful of
`.If`/`IfElse` call sites, rely on additive `Class` (delete `withClass` options where they
only existed to avoid replacement), move the showcase handlers to `htemel.Write`/`Handler`,
and rebuild composites on the delegating helper if it shipped.

## Consequences
- Fewer local workarounds; breaking only for apps that depended on `Class` replacing.
