# Changelog

## Unreleased

### Motion

One light, consistent motion scale, compositor-only (opacity, translate, scale, rotate, grid
rows; never width, height or position) and switched off by `prefers-reduced-motion`
(`motion-reduce:transition-none` on every transitioned element; a golden test enforces it and
the 200ms ceiling):

- Hover and focus colour changes: `transition-colors duration-150` (buttons, menu items, nav
  links, pagination items, tabs, close buttons, the switch track). Colour-only transitions keep
  running under reduced motion: nothing moves.
- Appearing: `duration-200 ease-out`, reversed on the way out. `Dialog` fades and scales from
  98%, `Drawer` slides from its edge, both with a fading backdrop and both animating close too
  (`@starting-style` via `starting:open:`, `transition-discrete` over `display` and `overlay`);
  `Dropdown` fades with a 1px drop the same way; `Tooltip` fades with a 1px rise; a `Toast`
  slides in from the right and fades out when dismissed; `Tabs` panels fade in.
- `Disclosure` and `Accordion` panels open and close by animating a grid row from `0fr` to `1fr`
  with no script: the panel follows its button's `aria-expanded` (`peer-aria-expanded:`) and is
  `visibility: hidden` once collapsed. They no longer use `data-show` / `display: none`.
- `Progress` slides its fill (`translate`) instead of animating `width`; the inline style is now
  `translate: -N%`. `Skeleton` stops pulsing under reduced motion.

### Tabs: no vertical scrollbar

The tab list scrolls horizontally (`overflow-x-auto`), and each tab's `-mb-px` made it 1px
taller than the list, so browsers showed a vertical scrollbar. The baseline is now the border of
a wrapper that the list overlaps by 1px; the tabs keep their 2px underline inside the list. The
axe run asserts `scrollHeight === clientHeight` for every tab list.

### Drawer

- **Added:** `Drawer(signal, side, title, children...)` and `DrawerActions(...)`, slide-over
  panels on the same native modal `<dialog>` as `Dialog` (signal synced both ways, Escape,
  backdrop click, focus return). `DrawerRight` and `DrawerLeft` are full height, `DrawerBottom`
  is a sheet up to 85% of the viewport; the body scrolls and `DrawerActions` stays pinned below
  it. `Dialog` and `Drawer` share their `<dialog>`, header and close handling.
- **Changed:** `Navbar`/`NavbarSignal`'s mobile menu is a right `Drawer` (same signatures and
  signal). The menu button is labelled "Open menu" (was "Toggle menu"), opens the drawer with
  `ui.Opens`, carries `aria-haspopup="dialog"` and controls `<signal>-dialog`; the inline menu
  `<div id="<signal>-menu">`, its click-outside and window Escape handlers are gone (the native
  dialog handles both).

### Landmark names, Field labels InputGroup, disabled pagination ends

- **Added:** an optional landmark name for the navigation components, so a page can hold two of
  them: `Pagination(page, pages, hrefFn, label ...string)` and
  `PaginationSignal(page, pages, signal, action, label ...string)` take a trailing label;
  `BreadcrumbsLabel(label, items...)` and `StepsLabel(label, current, labels...)` are
  `Breadcrumbs` and `Steps` with a name. The defaults stay "Pagination", "Breadcrumb" and
  "Progress"; two of the same on one page need distinct labels.
- **Changed (breaking):** `InputGroup` returns `*ui.InputGroupElement` (the `<div>`, embedded,
  plus `Control()`), not `*html.DivElement`. Chained element methods still work. `Field` labels
  the control inside it: the `id` its `<label for>` points at, `aria-describedby` and
  `aria-invalid` go on the input instead of the wrapper. The new `ui.CompositeControl` interface
  lets your own wrappers do the same. Remove any `aria-label` you added to the inner input as a
  workaround.
- **Changed:** a disabled previous/next item of `Pagination` is a `<span aria-disabled="true">`
  with visually hidden text instead of an `<a>` without `href` carrying `aria-label` (which axe
  reports as a prohibited attribute).

### WCAG AA contrast with the default tokens

The default primary stays sky; the components now use the steps of each family that reach WCAG
2.1 AA (4.5:1 text, 3:1 non-text) in light and dark, measured by the new contrast test and an
axe-core scan of the showcase. Rendered class names change:

| Where | Before | After |
| --- | --- | --- |
| Primary button, current page, completed step (white text) | `bg-primary-600 hover:bg-primary-700 dark:bg-primary-600 dark:hover:bg-primary-500` | `bg-primary-700 hover:bg-primary-800` in both schemes |
| Danger button | `hover:bg-danger-500 dark:bg-danger-500 dark:hover:bg-danger-400` | `bg-danger-600 hover:bg-danger-700` in both schemes |
| Link button, active tab, current step number | `text-primary-600` (Link hover `primary-500`) | `text-primary-700` (hover `800`); dark stays `primary-400` |
| Focus rings | `outline-primary-600`; Secondary/Ghost `outline-neutral-400` | `outline-primary-700 dark:outline-primary-500`; `outline-neutral-500` |
| Input and InputGroup affix border | `outline-neutral-300 dark:outline-white/10` | `outline-neutral-500 dark:outline-white/35` |
| Placeholder | `placeholder:text-neutral-400 dark:placeholder:text-neutral-500` | `placeholder:text-neutral-500 dark:placeholder:text-neutral-400` |
| Toast accent border | `border-{tone}-500` (Neutral `400`) | `border-{tone}-600 dark:border-{tone}-500` (Neutral `500`) |
| Close button (Alert, Toast, Dialog) | `opacity-60 hover:opacity-100` | full opacity; Toast and Dialog `text-neutral-500 dark:text-neutral-400` |
| Alert body | `opacity-90` | full opacity, the tone color |
| Stat positive delta | `text-success-600` | `text-success-700` |
| Progress fills | `bg-{info,success,warning,danger}-500` | `bg-info-600`, `bg-success-700`, `bg-warning-700`, `bg-danger-600`, each `dark:bg-*-500` |
| DropZone outline, MoveHere border | `primary-500` | `primary-600` |

If you retheme with another light hue (cyan, amber, lime, yellow), the 700 step is what keeps
white text readable; rerun `go test -run Contrast .` against your aliases (see README, Contrast).

### Design tokens replace themes (breaking)

Components are now styled with Tailwind v4 design tokens, CSS variables you override in your own
CSS, instead of classes written for sky/zinc/md that `SetTheme` rewrote at render time. Retheming
needs no Go change, works per scope (`[data-theme="..."]`) and live in the browser, and one
safelist serves every theme.

- **Removed:** `ui.Theme`, `ui.SetTheme`, `ui.CurrentTheme`, `ui.DefaultTheme`, `Theme.Classes`
  and the class cache; the `-primary`, `-radius` and `-font` flags of `cmd/safelist`, and the
  `SAFELIST_FLAGS` variable of `task tailwind-safelist`.
- **Added:** `assets/ui.css`, a plain `@theme` block with the default tokens, exposed as
  `ui.ThemeCSS()`, served by `ui.Assets()` at `/ui.css`, and printed by `cmd/themecss`.
- **Changed:** `ui.Classes(...)` only joins and normalizes whitespace. `ui.BodyClasses()` returns a
  constant (`font-sans`, the neutral page colors, antialiasing, the modal scroll lock).
- Every component class uses tokens; rendered class names change accordingly (see the table).
  With the default tokens the rendered colors and corners are the same as before.

Migration:

1. Generate the tokens and the safelist and import both after Tailwind:

   ```shell
   go run github.com/go-htemel/ui/cmd/themecss > assets/ui.theme.css
   go run github.com/go-htemel/ui/cmd/safelist > assets/tailwind.safelist.txt
   ```

   ```css
   @import "tailwindcss";
   @import "./ui.theme.css";
   @source "./tailwind.safelist.txt";
   ```

2. Replace the `SetTheme` call with CSS overrides in your `@theme` block:

   | Before (Go) | After (CSS) |
   | --- | --- |
   | `ui.SetTheme(ui.Theme{Primary: "indigo"})` | `--color-primary-50: var(--color-indigo-50);` ... `--color-primary-950: var(--color-indigo-950);` |
   | `ui.Theme{Radius: "lg"}` | `--radius-control: var(--radius-lg); --radius-surface: var(--radius-lg);` |
   | `ui.Theme{Font: "font-serif"}` | `--font-sans: ui-serif, Georgia, serif;` (or put `font-serif` on `<body>` yourself) |
   | `go run .../cmd/safelist -primary indigo -radius lg` | `go run .../cmd/safelist` (one list for every theme) |
   | `ui.CurrentTheme().Classes(s)`, `t.Classes(s)` | `ui.Classes(s)` with token classes |

3. Rename raw classes in your own markup that should follow the theme:

   | Before | After |
   | --- | --- |
   | `*-sky-*` (`bg-sky-600`, `focus-visible:outline-sky-600`) | `*-primary-*` |
   | `*-zinc-*` | `*-neutral-*` |
   | `*-green-*` | `*-success-*` |
   | `*-amber-*` | `*-warning-*` |
   | `*-red-*` | `*-danger-*` |
   | `*-blue-*` (Info tone) | `*-info-*` |
   | `rounded-md` on controls (buttons, inputs, menu items, links) | `rounded-control` (`rounded-l-control`, ...) |
   | `rounded-md` on surfaces (cards, dialogs, panels, alerts, toasts) | `rounded-surface` |

   Tests that compare rendered HTML need the new class names.

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
