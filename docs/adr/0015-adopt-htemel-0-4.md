# 0015 — Adopt htemel v0.4.0

**Status:** Accepted

## Decision
When htemel v0.4.0 lands (htemel ADRs 0008, 0009, 0011, 0013): migrate the handful of
`.If`/`IfElse` call sites, rely on additive `Class` (delete `withClass` options where they
only existed to avoid replacement), move the showcase handlers to `htemel.Write`/`Handler`,
and rebuild composites on the delegating helper if it shipped.

## Consequences
- Fewer local workarounds; breaking only for apps that depended on `Class` replacing.

## Outcome
Adopted against htemel v0.4.0. No call site used the removed conditionals or the 21 moved global
setters, so the migration was about `Class`: every `AddClass` became `Class`; the `withClass`
option went (callers chain `.Class(...)`); the `code.Block` "use AddClass" warning went. The one
semantic trap was icons: `IconX().Class("size-4")` used to replace `size-5 shrink-0` and would now
render both sizes, so those 22 call sites use `SetClass`. A rendered diff of all 37 showcase pages
against the v0.3 build showed two changes: the bottom drawer's duplicate `max-w-none` is gone, and
the Kanban columns now keep the `DropZone` drag-over styling they used to discard. Showcase pages
render with `htemel.Write`/`Handler`. The delegating helper (htemel ADR 0013) was rejected, so
composites keep their three forwarding methods.
