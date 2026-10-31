# Changelog

## Unreleased

- **Primary focus rings everywhere.** Secondary and Ghost buttons (and so `IconButton`s, the Dialog / Drawer / Toast X, `ColorSchemeToggle`, the code copy button and the MarkdownEditor tools) ring in `outline-primary-600` / dark `primary-500` instead of `neutral-500`; `MenuItem` focus adds a 2px inset primary ring (dark `primary-400`: `primary-500` is 2.8:1 on the tinted neutral-800 menu) instead of only a tint; the Alert dismiss X uses primary instead of `currentColor`. New contrast rows: Secondary and Ghost rings on cards and the neutral-800 panel, the MenuItem ring, the Alert X ring on each tone tint.
- **Dialog and Drawer initial focus.** With no `autofocus` element inside, the open effect moves focus right after `showModal()` (same task, before paint) to the first enabled `input`/`select`/`textarea` of the body, or else to the `<dialog>` itself (now `tabindex="-1"` and `outline-none`), so the X button no longer shows a focus ring on open; Tab still reaches the X first. `task a11y` checks it from the keyboard: the profile drawer focuses Name, the Delete dialog Cancel, the share sheet the dialog with no outline. `A11Y_PAGES=forms,overlays` limits a harness run to some pages.
- **Soft text-field outlines; 3:1 kept where the box is the control.** `--color-field-border` is now `neutral-300` (dark white at 15%) and `--color-field-border-hover` `neutral-400` (dark white at 30%): the Tailwind UI weight, on Input, Textarea, Select, InputGroup affixes and the MarkdownEditor frame. New `--color-control-border` (`oklch(65% 0.015 286)`, dark white at 35%; 3.24:1 on white) and `--color-control-border-hover` (`neutral-500`, dark white at 55%) carry the previous values for Checkbox, Radio, the off Switch track ring and knob, and the Prose task-list box, blockquote bar and table header rule, so those stay at 3:1. The contrast matrix reports the text-field border rows (rest and hover, 18 rows) as *informational* (never failing, like the disabled rows), citing WCAG Understanding SC 1.4.11 ("a text input identified by its label, fill and focus indicator"); checkbox, radio and switch borders and every focus outline stay enforced. `TestFieldBorderToken` became `TestControlBorderToken`. README "Contrast" documents the strict opt-in (`--color-field-border: var(--color-neutral-500)`). **Upgrading:** regenerate `ui.theme.css` and the safelist; apps that override `--color-field-border` for checkboxes should override `--color-control-border` instead.
- **Markdown: `github.com/go-htemel/ui/markdown`.** `markdown.Render(src, opts...)` converts CommonMark + GFM (tables with alignment, strikethrough, task lists, linkify, footnotes, heading ids) with goldmark, highlights fenced and indented code with `code.Block` (no copy button), sanitises with `markdown.Policy()` (bluemonday `UGCPolicy` plus exactly the markup the renderer emits) and returns the HTML in `ui.Prose` (`*html.DivElement`). Options: `WithHeadingOffset`, `WithIDPrefix`, `WithoutHeadingIDs`, `WithLinkTarget`, `WithSize`, `WithPolicy`, `WithUnsafeHTML` (trusted authors only). Also `HTML` (the string), `Plain` (text for search and meta descriptions) and `Excerpt(src, n)`. Task-list checkboxes get `aria-label="Done"`. New module dependencies `github.com/yuin/goldmark` v1.8.7 and `github.com/microcosm-cc/bluemonday` v1.0.27 (with `aymerick/douceur` v0.2.0 and `gorilla/css` v1.0.1), all compatible with `go 1.24`; the root package imports none of them. Tests cover GFM output, an XSS corpus in every sanitising mode, unsafe mode, heading offsets, ids and targets, code highlighting and deterministic, race-free output.
- **`Prose`, `ProseSm`, `ProseLg`.** Typography for long-form HTML without classes: `.ui-prose` descendant rules in `assets/ui.css` for headings, paragraphs, links, nested lists, blockquotes, tables, rules, images, inline code chips, `kbd`, `mark`, definition lists, figures, GFM task lists and footnotes, light and dark, through `--ui-prose-*` variables; sizes in `em`. 40 new contrast rows (`Prose/...`), all passing. **Upgrading:** regenerate `ui.theme.css` (`cmd/themecss`) and the safelist.
- **`MarkdownEditor`.** A `CompositeControl` (labelled by `Field`, which keeps its hint and counter in `aria-describedby`): formatting toolbar, a textarea bound to a signal, Write / Preview tabs or `WithSplit` (side by side from an `@xl` container width), and a footer with a "Markdown supported" hint and a live word / character count (`WithMaxLength` turns it warning at 90% and danger past the limit). The preview is fetched from the app with a Datastar action (`WithPreview`) and patched by id: `MarkdownPreview(id, content...)`, `PreviewID()`, `MarkdownPreviewID(editorID)`; the editor's live region wraps the patch target (`WithPreviewLabel`). Also `WithSignal`, `WithEditorID`, `WithValue`, `WithRows`, `WithPlaceholder`, `WithAutosize` (`field-sizing: content`), `WithMonospace`, `WithToolbar` (`MarkdownTool` constants, `DefaultMarkdownToolbar`), `WithUploadAction` and `WithHint`. 30 contrast rows. New icons `IconBold`, `IconItalic`, `IconStrikethrough`, `IconHeading`, `IconQuote`, `IconCode`, `IconLink`, `IconImage`, `IconListBullet`, `IconListNumbered`, `IconListTask`, `IconTable`.
- **`<ui-mdeditor>` in `ui.js`.** The toolbar markup only carries `data-ui-md-action`; the new custom element (light DOM, no globals, like `<ui-sortable>`) wraps or prefixes the selection through `insertText` (so Ctrl+Z undoes it and `input` reaches `data-bind`; `setRangeText` plus a dispatched `input` as fallback), handles Ctrl/Cmd+B/I/E/K and Shift+X, continues lists on Enter, runs the toolbar's roving tabindex (arrows, Home, End) and, with `data-ui-upload`, turns pasted or dropped files into `ui:md-upload` events (`detail.file`, `detail.insert(text)`). Until it is defined the toolbar is hidden (`ui-mdeditor:not(:defined)` in `ui.css`).
- Showcase: "Markdown editor" (tabs with a limit, split view, compact note with image paste; `POST /demo/markdown/preview`) and "Prose" (a rendered CMS article, sizes, hand-written HTML) pages. `task a11y` scans both in light and dark plus the rendered preview, and checks a Bold click (value, selection, `input`, the bound signal through the counter, Ctrl+Z), the roving tabindex, list continuation and Ctrl+B, the split preview, and highlighted fenced code and task lists in Prose. The safelist scanner also reads `markdown/`.
- **Self-hosted Geist (opt-in).** The module embeds the Geist and Geist Mono variable fonts (`assets/fonts/Geist-Variable.woff2`, 69,576 bytes, and `GeistMono-Variable.woff2`, 71,272 bytes, both weights 100-900; SIL Open Font License 1.1, copy in `assets/fonts/LICENSE-Geist.txt`), taken from the npm package `geist` 1.9.0. `Assets()` serves them as `font/woff2` at `/fonts/Geist-Variable.woff2` and `/fonts/GeistMono-Variable.woff2` with the same `immutable` year-long caching and `ETag` as `/ui.js`. New `FontCSS(prefix) string` returns the two `@font-face` rules (`font-display: swap`, `src` with a `?v=<hash>` cache buster) plus an `@theme` that sets `--font-sans` to Geist and `--font-mono` to Geist Mono; new `FontPreload(prefix) []*html.LinkElement` returns the matching `<link rel="preload" as="font" type="font/woff2" crossorigin>` elements (same URLs, so each file is fetched once). Apps that do not include `FontCSS` keep the system stacks. See the README, "Fonts". The showcase drops Inter from Google Fonts and uses Geist, so it makes no third-party font request; the a11y harness no longer stubs the Google Fonts hosts.
- **Softer field outlines.** New tokens `--color-field-border` (light `oklch(65% 0.015 286)` = `#8e8e98`, the lightest neutral-hued shade that is 3.05:1 on white and on `neutral-50`: 3.24:1 and 3.11:1; dark white at 35%, 3.1:1 or more over the page, card and field) and `--color-field-border-hover` (`neutral-500`; dark white at 55%), replacing `neutral-500` / `dark:white/35` on Input, Textarea, Select, InputGroup affixes, unchecked Checkbox and Radio, and the off Switch track ring (the knob ring is `field-border-hover`: field-border would be 2.5:1 on the neutral-200 track). Fields darken their outline on hover (`enabled:hover:outline-field-border-hover`); focus (2px primary) and invalid outlines still win. The tokens are redefined under `.dark`, so the classes carry no `dark:` variant. **Upgrading:** regenerate `ui.theme.css` (`cmd/themecss`) and the safelist. For the pale Tailwind UI look set `--color-field-border: var(--color-neutral-300)`; that is below WCAG 1.4.11 (see the README, "Contrast").
- Dialog, Drawer and Toast close buttons are Ghost `IconButton`s (no ring, hover background only; icon `neutral-700` / dark `neutral-300`, 10:1) instead of the bespoke `closeButton`, and now carry a `title` tooltip like every `IconButton`. Alert keeps `closeButton`, which follows the tone colour.
- Contrast test: the dark scheme is measured against the tokens redefined under `.dark` (`wcag.NewDarkTheme`; `wcag` also parses `rgb(r g b / a)`), rows for the field border (on the page, on a card, against the field fill, hover) in both schemes, and `TestFieldBorderToken` pins the light value as the lightest that passes.
- **Code blocks: `github.com/go-htemel/ui/code`.** Server-side syntax highlighting with chroma, in a subpackage so the root package and applications that do not import it stay free of the dependency (new module dependency `github.com/alecthomas/chroma/v2` v2.24.1, which keeps `go 1.24`). `code.Block(lang, src, opts...)` returns a `*BlockElement` (`<figure class="ui-code">` with optional toolbar, a focusable `<pre><code>` of `<span class="line">`s and a `<template>` with the raw source); `code.Inline` returns a highlighted `<code>`. Options: `WithLineNumbers` (CSS counters, never selected or announced), `WithStart`, `WithHighlight`/`WithHighlightRange` (`data-hl`), `WithTitle`, `WithWrap`, `WithoutCopyButton`, `WithID`. Unknown languages and lexer failures fall back to plain text. 14 `tok-*` classes coloured by `--color-code-*` variables in `assets/ui.css` (light on `:root`, dark under `.dark`); 102 new contrast rows, all at least 4.5:1. The copy button is an `IconButton` with a local `_copied_<hash>` signal, a 150ms icon cross-fade and an `aria-live` "Copied". New `IconCopy`. **Upgrading:** regenerate `ui.theme.css` (`cmd/themecss`) and, if you use the package, the safelist (`cmd/safelist` now also scans `code/`).
- Showcase: example sources are highlighted `code.Block`s with line numbers and a copy button, the Theming page's CSS is a titled `css` block, and a new Code page demonstrates the options. The a11y run covers it, including a copied state, selection without line numbers, highlighted lines, scroll and wrap.
- **Custom checkbox, radio and switch.** `Checkbox` and `Radio` are drawn with tokens on an `appearance-none` input (border `neutral-500`, `checked:bg-primary-600`, white check / indeterminate dash / dot via `peer-checked:` siblings, `focus-visible:` outline, disabled states, dark variants, `forced-colors:appearance-auto`), replacing `accent-color` and the native hover halo. The markup gains a `<span class="grid ...">` around the input. `Switch` gets a 3:1 off-state (neutral-500 track and knob rings), hover colours, and a muted disabled state; labels of disabled controls are muted (`has-disabled:`).
- **Native control polish** (`ui-control` class on `Input`, `Textarea`, `Select`; rules in `ui.css`): autofill keeps the field colours (opaque `neutral-800` cover in dark mode), number inputs lose their spin buttons, the search clear button is a themed cross, the date/time picker icon follows `color-scheme`. `Input` styles by type: `file` gets a Secondary-style button, `color` a field-height swatch, `range` a bare primary slider. **Upgrading:** regenerate `ui.theme.css` (`cmd/themecss`) and the safelist.
- **Customizable select picker.** `Select` adds the `ui-select` class; where `appearance: base-select` is supported (Chromium 135+) the open picker is styled like the Dropdown menu, with a primary highlight, a check mark on the selected option and a 200ms fade. Other browsers keep the native popup.
- **`Brand` tone.** `ui.Brand` (`"brand"`): primary tints for `Badge` and `Alert` (`bg-primary-50 text-primary-700`), `border-primary-600` toast accent, `bg-primary-600` progress fill.
- **Info is sky.** `--color-info-*` aliases `sky` instead of `blue`, so Info no longer reads as a second primary next to indigo. All Info pairs still pass (text 5.5:1 light, 10.6:1 dark; toast accent 4.0:1; progress fill 3.2:1).
- **Navbar uses container queries.** The nav is `@container` and switches at its own `@2xl` width (was viewport `md`), so it collapses into the drawer inside narrow columns too. `NavItem.Attrs(...)` adds attributes to a link, and the active highlight now follows `aria-current="page"` (`aria-[current=page]:` utilities; label in `primary-700`), so a client can move it by binding the attribute.
- **Drag and drop: no flash on drop.** `PickUp` also sets `$<signal>Via = 'pickup'` (`ViaSuffix`); `MoveHere` declares that signal, shows only while it is `'pickup'` and clears both. `DropInto` never touches it, so the Move here buttons no longer flash for a frame on every mouse or touch drop.
- Hover colour changes that snapped now ease over 150ms (breadcrumb links, sidebar headings, sortable headers, `MoveHere`); ring changes may use `transition-shadow` (with `motion-reduce:transition-none`).
- Showcase: no more `Refused to apply style from '/components/tailwindcss'` console error (the page CSS no longer contains `@import`; the browser build adds it), "Native inputs" example, two live navbar frames (380px and full width) with client-side active links and toasts, a framed sidebar, scheme-aware source blocks that animate open (`::details-content`), and `Brand` in the people table and theming page. The a11y run fails on console errors and covers the open select picker, the container-query navbar and drops.

- **Light / dark / system toggle.** `dark:` now follows a `.dark` class on `<html>` (`@custom-variant dark (&:where(.dark, .dark *))` in `assets/ui.css`, plus `color-scheme` rules so native controls and scrollbars match). New `ColorSchemeBootstrap(opts...)` returns the inline script that applies the stored choice (`ColorSchemeKey`, `"color-scheme"`) before first paint and follows the OS in system mode; `WithNonce` supports strict CSP and `ColorSchemeBootstrapSource` exposes the text for hash sources. New `ColorSchemeToggle(opts...)` is a System/Light/Dark segmented control (`WithLabels` shows the words); new icons `IconSun`, `IconMoon`, `IconMonitor`. **Upgrading:** add `ui.ColorSchemeBootstrap()` to `<head>` before your stylesheet, or the page stays light even on dark-mode systems; to keep OS-only behaviour instead, delete the `@custom-variant` line and the `@layer base` color-scheme block from your copy of the tokens. The a11y scan covers the toggle and a no-flash check.
- **Default primary is indigo.** `primary` aliases stock Tailwind `indigo` (previously `sky`). `cmd/palettetune` remains as a helper for deriving custom palettes. Stock sky is a selectable theme in the showcase.
- Components return to the conventional steps: Primary button, current page and completed step are `bg-primary-600` (hover `700`), links and active tabs `text-primary-600` (hover `700`), focus rings `outline-primary-600`. Dark-mode hovers go lighter: Primary starts at `dark:bg-primary-600` and hovers to `500`; Danger hovers from `dark:bg-danger-700` to `600`. Retheming to a light stock hue (sky, cyan, amber, lime, yellow) now needs the 700 mapping described in the README.
- `DrawerSized(signal, side, size, title, ...)` sets the panel width of left/right drawers (`Xs`…`Xl` → `max-w-xs`…`max-w-2xl`); `Drawer` keeps `Md` right / `Xs` left. Chaining a `max-w-*` class is unreliable because two utilities would compete.
- Danger buttons use `danger-700` (hover `600`, lighter) in dark mode; pure `red-600` glowed on dark surfaces. Still 6:1 with white text.
- `Select` returns `*SelectElement`: a grid wrapper with the library chevron over an `appearance-none` native select (the browser arrow was heavy and unthemed, especially in dark mode). `Attrs`/`SetAttr` still target the `<select>`, and `Field` labels it through `Control()`.

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
