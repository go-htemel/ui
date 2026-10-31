# Changelog

## Unreleased

### Native overlays

`Dialog` and `Dropdown` now use the platform instead of hand-rolled Datastar behaviour.

- **`ui.Dialog` returns `*html.DialogElement`** (was `*html.DivElement`). It renders
  `<dialog closedby="any" aria-labelledby="…">` opened with `showModal()`. Call sites that only
  render it, or chain methods common to all elements (`Class`, `Attrs`, `Id`, ...), compile
  unchanged; code that stored the result in a `*html.DivElement` must change the type.
  - Removed: `role="dialog"`, `aria-modal`, the inline `display: none` and `data-show`, the
    window keydown handler (Escape and the Tab trap), the focus and scroll-lock effect.
  - The signal still opens and closes it (`data-effect`), and `data-on:close` sets it to false
    after Escape, a backdrop click or a `<form method="dialog">` submit.
  - Page scroll is no longer locked; add `html:has(dialog:modal) { overflow: hidden; }`.
  - The element ids are `<signal>-dialog` (new), `<signal>-title` and `<signal>-body`.
- **`ui.DialogActions` adds `autofocus`** to its first child when that is a button or link and
  no child already has `autofocus`, so the dialog opens with focus on its first action.
- **`ui.Dropdown`** keeps its signature. The menu is `popover="auto"` and the trigger gets
  `popovertarget`/`popovertargetaction="toggle"`, so it opens without Datastar.
  - Removed: the trigger's toggle click handler, the menu's `display: none`/`data-show`, the
    click-outside and window Escape handlers.
  - The signal mirrors the menu (`data-on:toggle`) and setting it shows or hides the menu
    (`data-effect`).
  - The trigger gets an inline `style="anchor-name: …"` (replacing any style you set on the
    trigger before passing it) and the menu is placed with inline CSS anchor positioning.
- **`ui.Tooltip`** is unchanged (CSS only); its documentation explains why `popover="hint"` was
  not adopted.
- Removed the unexported focusable-element helpers that only the old Dialog used.
