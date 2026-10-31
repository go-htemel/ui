# 0009 — Markdown editor instead of a rich-text editor

**Status:** Accepted

## Context
Blog/CMS/notes/journal use cases. Options: Markdown + server preview, a hand-rolled
contenteditable editor, or wrapping ProseMirror/TipTap.

## Decision
`MarkdownEditor` (textarea + toolbar driven by `<ui-mdeditor>`, debounced server preview
over SSE) and `markdown.Render` (goldmark GFM + bluemonday, code via the `code` package,
wrapped in `Prose`). No contenteditable. A real editor wrapper, if ever needed, is a
separate module.

## Consequences
- Robust, hypermedia-native, zero client sanitisation. Not Word-like; accepted.
- goldmark/bluemonday/chroma stay out of the root package via subpackages.
