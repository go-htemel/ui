# 0001 — Components are functions returning concrete elements

**Status:** Accepted

## Context
A component library could return an opaque `Node`, a struct with options, or the element
it builds. htemel's elements are typed by content model and chainable.

## Decision
Components are plain functions returning the concrete htemel element (`*html.ButtonElement`)
so callers keep chaining (`.Attrs(datastar.OnClick(...))`). Variants and sizes are typed
enums; rarely used knobs are functional options. Parts that only belong inside one component
(`CardSlot`, `TabItem`, `AccordionPanel`) are their own types and not content. Composites
that wrap a real control (`Select`, `InputGroup`, `TagInput`, `FileUpload`,
`MarkdownEditor`) implement `CompositeControl` so `Field` labels the inner control and
`Attrs`/`SetAttr` reach it.

## Consequences
- Pages read as a description of the page; the Admin demo is mostly composition.
- Sharp edge: every composite re-plumbs four forwarding methods (htemel ADR 0013).
- Sharp edge: `Class()` on a component replaced its base classes (htemel ADR 0009);
  until that lands, use `AddClass`/`WithRailIcon`-style options.
