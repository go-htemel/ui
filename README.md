# go-htemel/ui

[![Go Reference](https://pkg.go.dev/badge/github.com/go-htemel/ui.svg)](https://pkg.go.dev/github.com/go-htemel/ui)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Components for [htemel](https://github.com/go-htemel/htemel): Go functions that render
Tailwind-styled HTML, with [Datastar](https://data-star.dev) attributes for every interaction and
no custom JavaScript. Components are typed by the HTML content model, so the compiler rejects a
`Card` inside a paragraph just as htemel rejects a `<div>` there.

## Install

```shell
go get github.com/go-htemel/ui
```

Requires Go 1.24, htemel v0.3.0, Tailwind CSS v4 and Datastar v1.0.4.

## Example

```go
func profileCard() htemel.FlowContent {
	return ui.Card(
		ui.CardHeader("Profile", ui.Badge(ui.Success, "Verified")),
		ui.CardBody(
			ui.Field("name", "Name", ui.Input(html.InputTypeText, "name").Attrs(datastar.Bind("name")),
				ui.ErrorSignal("errors.name")),
		),
		ui.CardFooter(
			ui.Button(ui.Primary, "Save", ui.WithLoading("saving")).
				Attrs(datastar.OnClick(datastar.Post("/profile"))),
		),
	)
}
```

Components return the htemel element (`*html.ButtonElement`, `*html.DivElement`, ...), so you keep
chaining element methods and `datastar` helpers onto them. Sections that only belong inside one
component have their own types: `CardHeader`/`CardBody`/`CardFooter` return a `CardSlot` that only
`Card` accepts, `Tab` returns a `TabItem` for `Tabs`, `AccordionItem` an `AccordionPanel`.

## Components

| Group | Components | Status |
| --- | --- | --- |
| Buttons | `Button`, `IconButton`, `ButtonGroup`; variants, sizes, icon slots, `WithLoading` (data-indicator) | ready |
| Display | `Badge`, `Avatar`, `AvatarInitials`, `Kbd`, `Divider`, `DividerLabel` | ready |
| Icons | `Icon` (Font Awesome), `IconCheck`, `IconX`, `IconChevronDown`, `IconSearch`, `IconPlus`, `IconTrash`, `IconPencil`, `IconGripVertical`, `IconSpinner` | ready |
| Layout | `Stack`, `Inline`, `Container`, `Card` | ready |
| Feedback | `Alert`, `AlertDismissible`, `ToastRegion`, `Toast` | ready |
| Forms | `Field` (`HelpText`, `ErrorText`, `ErrorSignal`), `Input`, `Textarea`, `Select`/`Choice`, `Checkbox`, `Radio`/`RadioGroup`, `Switch` | ready |
| Overlays | `Dialog`/`DialogActions` (native `<dialog>`), `Dropdown` with `MenuItem`/`MenuLink`/`MenuDivider` (Popover API) | ready |
| Navigation | `Tabs`/`Tab`, `Disclosure`, `Accordion`/`AccordionItem` | ready; no arrow keys yet |
| Drag and drop | `Draggable`, `DropZone`, `DropInto`, `DragData` | ready; mouse only (HTML drag events) |
| Data | `Table` (`TableHead`, `TableBody`, `Th`, `Td`; `Striped`, `Dense`, `StickyHeader`), `SortableTh`, `EmptyState`/`EmptyRow`, `Pagination`, `PaginationSignal`, `Stat`/`StatGrid`, `DescriptionList` | ready |
| Navigation | `Navbar`/`NavbarSignal` with `NavLink`, `Sidebar`/`SidebarGroup` (collapsible), `Breadcrumbs`/`Crumb`, `Steps` | ready; no arrow keys yet |
| Progress | `Progress`, `ProgressSignal`, `Skeleton`, `Spinner` | ready |
| Misc | `Tooltip` (CSS only, no Escape; see below), `InputGroup` | ready |
| Planned | Command palette | — |

## Interactivity

State lives in Datastar signals that you name. `Dialog`, `Dropdown`, `Disclosure`, `Tabs`,
`Accordion` and `AlertDismissible` take the signal as their first argument and declare it with
`__ifmissing`, so a server can preset it. `ui.Open`, `ui.Close` and `ui.Toggle` return the
expressions, and `ui.Opens(signal)` is the click handler for a trigger outside the component.
`Dialog` and `Dropdown` are native overlays (a modal `<dialog>` and a `popover`): the browser
opens, closes and focuses them, and their signal mirrors the state in both directions, so it
stays correct after Escape or a click outside and setting it opens or closes them:

```go
ui.Button(ui.Danger, "Delete").Attrs(ui.Opens("confirm"))
ui.Dialog("confirm", "Delete project?", html.P().Text("This cannot be undone."))
```

Toasts are pushed by the server into `ui.ToastRegion()` with datastar-go:

```go
sse := datastar.NewSSE(w, r)
s, _ := htemel.RenderString(ui.Toast(ui.NextToastID(), ui.Success, "Saved", "", 4*time.Second))
sse.PatchElements(s, datastar.WithSelectorID(ui.ToastRegionID), datastar.WithModeAppend())
```

## Tailwind

Tailwind v4 is required (the components use `shadow-xs`, `has-checked:`, `*:` variants and CSS
theme variables). The components are styled with design tokens (see [Theming](#theming)), so your
build needs two files from the module: the token definitions and the list of classes the
components use. Tailwind only generates classes it finds in your sources, and the ui sources sit
in the Go module cache, so the class list stands in for them.

```shell
go run github.com/go-htemel/ui/cmd/themecss > assets/ui.theme.css
go run github.com/go-htemel/ui/cmd/safelist > assets/tailwind.safelist.txt
```

```css
/* assets/app.css */
@import "tailwindcss";
@import "./ui.theme.css";              /* default tokens: one @theme block */
@source "./tailwind.safelist.txt";     /* the classes the components use */

@theme {
  /* your overrides, see Theming */
}
```

```shell
npx @tailwindcss/cli -i assets/app.css -o static/app.css
```

Rerun both `go run` commands after upgrading the module (they read the version your `go.mod`
requires) and commit the two files. The safelist is one static list for every theme: class names
never change, only the values behind the tokens. `ui.ThemeCSS()` returns the same tokens as a
string, and `ui.Assets()` serves them at `/ui.css`, but that file is Tailwind source (an `@theme`
block), not a stylesheet browsers apply on their own. The showcase compiles its CSS in the browser
with the Tailwind browser build and puts `ui.ThemeCSS()` in its `<style type="text/tailwindcss">`
block; real applications should build their CSS with the Tailwind CLI.

## Theming

Components never name a palette color. They use semantic tokens, defined as CSS variables in
`ui.theme.css`, and you retheme by overriding those variables in CSS:

| Token | Utilities | Default | Used for |
| --- | --- | --- | --- |
| `--color-primary-{50..950}` | `bg-primary-700`, `text-primary-700`, `outline-primary-700`, ... | `sky` | primary buttons, links, focus rings, selected and active states, drop targets |
| `--color-neutral-{50..950}` | `bg-neutral-50`, `text-neutral-500`, `ring-neutral-950/5`, ... | `zinc` | surfaces, text, borders, dividers, the page background |
| `--color-success-{50..950}` | `bg-success-50`, `text-success-700`, ... | `green` | Success tone (Badge, Alert, Toast, Progress), positive Stat deltas |
| `--color-warning-{50..950}` | `bg-warning-50`, `text-warning-800`, ... | `amber` | Warning tone |
| `--color-danger-{50..950}` | `bg-danger-600`, `outline-danger-500`, ... | `red` | Danger tone and button, invalid fields, error text, negative Stat deltas |
| `--color-info-{50..950}` | `bg-info-50`, `text-info-700`, ... | `blue` | Info tone |
| `--radius-control` | `rounded-control`, `rounded-l-control`, ... | `--radius-md` | buttons, button and input groups, inputs, menu items, nav links, pagination, tooltips, `Kbd` |
| `--radius-surface` | `rounded-surface` | `--radius-md` | Card, Dialog, Dropdown panel, Alert, Toast, Stat, Table frame |
| `--font-sans` (Tailwind's own) | `font-sans` | Tailwind's sans stack | `ui.BodyClasses()` puts `font-sans` on `<body>` |

Pills and avatars stay `rounded-full`. Each family defines the full 50–950 scale, so your own
markup can use any step (`bg-primary-100`), and `ui.Classes(...)` joins class lists for your own
components.

### Contrast

The shipped defaults meet WCAG 2.1 AA (4.5:1 for text, 3:1 for borders, focus rings and fills) in
light and dark, and the Go contrast test (`go test -run Contrast .`) measures every pair the
components render. The steps follow one convention:

- Solid surfaces with white text (Primary button, current page, completed step) use
  `bg-primary-700`, hover `bg-primary-800`, in both schemes; Danger uses `bg-danger-600`, hover
  `bg-danger-700`. Links and active labels are `text-primary-700` (hover `800`) in light and
  `dark:text-primary-400` in dark.
- Focus rings are `outline-primary-700` (`outline-neutral-500` on neutral controls) in light and
  `primary-500` in dark; input borders are `outline-neutral-500` / `dark:outline-white/35`;
  placeholders `neutral-500` / `dark:neutral-400`.

Light hues such as sky (the default), cyan, amber, lime and yellow only reach 4.5:1 against
white text at the 700 step, which is why the solids use it; a darker hue (indigo, blue, violet)
has room to spare. The test guards the shipped defaults only (the aliases in `assets/ui.css`).
After retheming, run it against your palette: change those aliases in a checkout of this module
and rerun it (`CONTRAST_REPORT=contrast.md go test -run 'Contrast$' .` writes the full table
with a suggested step for every failing pair).

### Overriding tokens

Override in your `@theme` block, after the import, to change the default for the whole app:

```css
@theme {
  --color-primary-50:  var(--color-indigo-50);
  /* ... 100 to 900 ... */
  --color-primary-950: var(--color-indigo-950);
  --color-neutral-500: var(--color-slate-500);   /* and the other neutral steps */
  --radius-control: var(--radius-lg);
  --radius-surface: var(--radius-xl);            /* square: 0, pill: calc(infinity * 1px) */
  --font-sans: "Inter", ui-sans-serif, system-ui, sans-serif;
}
```

Alias the whole scale of a family you change: components use several steps (for example
`bg-primary-700 hover:bg-primary-800 dark:text-primary-400`). Any CSS color works as a value
(`oklch(...)`, hex), not only a Tailwind palette variable. The token families share their names
with nothing in Tailwind's palette except `neutral`: `ui.theme.css` redefines Tailwind's `neutral`
scale as the zinc alias, so `bg-neutral-*` in your own markup follows the token too.

> **Font:** if `--font-sans` falls back to `system-ui`, Windows renders Segoe UI, whose metrics
> place text about a pixel above centre in buttons and badges. Point it at a metrically balanced
> face such as Inter, as above (the showcase does).

### Scoped themes, live

`ui.theme.css` uses plain `@theme`, not `@theme inline`, so the generated utilities read the token
itself (`.bg-primary-600 { background-color: var(--color-primary-600) }`). Redefining a token in
any selector therefore rethemes everything inside it, in the browser, with no Go change, no new
safelist and no new markup:

```css
[data-theme="emerald"] {
  --color-primary-500: var(--color-emerald-500);
  --color-primary-600: var(--color-emerald-600);   /* ... the whole scale ... */
  --radius-control: calc(infinity * 1px);
}
```

```go
html.Div(content...).Attr("data-theme", "emerald") // or set it on <html> to switch the page
```

Switching `data-theme` (the showcase's Theming page does it with a `<select>`, a signal and
`data-effect`) needs no CSS build either. Override the semantic token (`--color-primary-600`), not
the palette color it aliases (`--color-sky-600`): the alias resolves where it is declared, on
`:root`. Tailwind emits the palette variables your CSS references (`var(--color-emerald-600)`
above), but CSS it never sees, such as a per-tenant `style="--color-primary-600: ..."` added at
runtime, should use literal color values.

### Dark mode

Components carry `dark:` variants that pick other steps of the same tokens (`text-neutral-900
dark:text-neutral-100`), so one override covers both modes. Tailwind's `dark:` follows
`prefers-color-scheme` by default; for a class toggle add
`@custom-variant dark (&:where(.dark, .dark *));` to `app.css`. To use different palettes per
mode, override the tokens inside that scope (`.dark { --color-primary-500: ...; }`).

## Assets and progressive enhancement

Everything works without JavaScript files of your own. The one exception is optional: `ui.js`, a
small embedded script (vanilla ES2020, no dependencies, about 150 lines) that defines the
`<ui-sortable>` custom element for touch dragging. Serve it with `ui.Assets()` and load it with
`ui.ScriptTag()`:

```go
mux.Handle("/assets/", http.StripPrefix("/assets", ui.Assets()))   // GET /assets/ui.js (and /assets/ui.css)

html.Head(ui.ScriptTag("/assets")) // <script type="module" src="/assets/ui.js?v=<hash>">
```

The script is served as `text/javascript` with `Cache-Control: public, max-age=31536000,
immutable` and an `ETag`; the `?v=` hash in the tag changes whenever the script does, so a
library upgrade is never stale. It is a module script (deferred), loaded from your own origin and
needing no inline script, so a Content-Security-Policy of `script-src 'self'` allows it. The
components still use Datastar's `data-on` attributes, which Datastar evaluates itself. Skip
`ScriptTag` and nothing breaks: each enhancement has a no-script fallback.

### Touch drag

HTML drag events never come from a touch screen. Wrap your board in `ui.Sortable` and load
`ui.js`, and a long press (about 150 ms; a swipe before that still scrolls) on a `Draggable` lifts
a copy that follows the finger. Releasing it over a `DropZone` runs the same action a mouse drop
runs, so `DropInto` needs no change:

```go
ui.Sortable(board) // board contains DropZone(DropInto("dragged", act), ...) and Draggable(...)
```

`Draggable` and `DropZone` carry `data-ui-drag`, `data-ui-payload` and `data-ui-dropzone` for the
script. It dispatches bubbling `ui:dragstart` (on the dragged element), `ui:drop` (on the zone;
`detail.payload`) and `ui:dragend` (`detail.payload`, `detail.dropped`) events, and sets
`data-over` and `data-dragging` like the native path. Mouse pointers are ignored by the script
and keep the browser's native drag and drop. Without `ui.js`, the `PickUp` and `MoveHere` buttons
remain the touch and keyboard fallback.

## Showcase

```shell
task run   # http://localhost:8080
```

To use another port, pass `ADDR` to the task or the `-addr` flag to the program:

```shell
task run ADDR=:8081
go run ./cmd/showcase -addr :8081
```

`cmd/showcase` renders one page per component group with each live component next to the Go
source that produced it, plus datastar-go demos: server-pushed toasts, live form validation, a
kanban board whose drops post to the server and a table that sorts and pages on the server. The
Theming page switches token sets live through `<html data-theme>`. Example functions live in
`cmd/showcase/examples_*.go`; those files are embedded and parsed at startup, so the displayed
source is always the code that ran.

## Datastar version

The library targets Datastar v1.0.4 and emits its colon key syntax (`data-on:click`,
`data-signals:open`) through htemel's `datastar` package; no attribute is written by hand, and the
showcase tests fail on any `data-*` attribute missing from the Datastar v1.0.4 reference. Load
Datastar with:

```html
<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.4/bundles/datastar.js"></script>
```

## Development

- `task test` runs `go test -race ./...`: golden HTML per component (`go test -run TestGolden
  -update` rewrites `testdata/`), compile-time content assertions, safelist checks and showcase
  page checks.
- `task lint` runs golangci-lint v2.
- `task tailwind-safelist` regenerates `tailwind.safelist.txt` after a class change; a test fails
  until you do. Tests also fail on a raw palette or radius class (`bg-sky-600`, `rounded-md`) in
  a component, and on a token class whose variable `assets/ui.css` does not define.

### Accessibility testing

`task a11y` builds the showcase, serves it on a free port and scans every page with
[axe-core](https://github.com/dequelabs/axe-core) in Chromium (Playwright), in light and dark
(`prefers-color-scheme`), at WCAG 2.1 A/AA. It also opens the interactive states first: the
dialog, the dropdown, a switched tab, an expanded disclosure and accordion, and the mobile
navbar. The report groups violations by rule with impact, count, selectors and pages, and the
task exits non-zero on any WCAG violation. Rules tagged only `best-practice` are listed
separately and do not fail the run. The scan fails loudly, rather than passing, when Tailwind
did not apply (the showcase loads Tailwind and Datastar from CDNs).

```shell
(cd tools/a11y && npm ci && npx playwright install chromium)   # once
task a11y
```

Options, as environment variables: `CHROMIUM_PATH` launches an existing Chromium instead of
Playwright's, and `A11Y_CDN_DIR` serves `tailwind.js` and `datastar.js` from a directory when
the CDNs are unreachable. CI runs it as the separate `a11y` job. Colour contrast between tokens
is also covered without a browser by the Go contrast test.

## License

MIT

## Accessibility

The interactive components follow the matching WAI-ARIA patterns. Overlays use the platform:
`Dialog` is a modal `<dialog>` and `Dropdown` a `popover`, so the top layer, inertness, Escape,
light dismiss and focus return come from the browser. The rest is Datastar attributes and
expressions. Each component's Go documentation lists exactly what it covers.

| Component | Semantics | Keyboard and focus |
| --- | --- | --- |
| `Dialog` | native `<dialog>` opened with `showModal()`, `aria-labelledby` the title | native: page inert behind it, Tab and Shift+Tab stay inside, Escape closes, focus returns to the trigger; initial focus on `autofocus` (`DialogActions` sets it on its first button) |
| `Dropdown` | trigger `popovertarget`, `aria-haspopup`, `aria-controls`, bound `aria-expanded`; `popover="auto"` `role="menu"` and `menuitem` | native: Escape closes and returns focus, click outside closes; ArrowDown/ArrowUp open from the trigger and move between items; Home/End; Tab closes |
| `Tooltip` | `role="tooltip"`, trigger `aria-describedby` | shown on hover and focus (CSS) |
| `Tabs` | `tablist`, `tab`, `tabpanel`, `aria-selected`, `aria-controls`, roving `tabindex` | ArrowLeft/ArrowRight, Home/End |
| `Disclosure`, `Accordion` | `aria-expanded`, `aria-controls`, panel `role="region"` | Accordion: ArrowDown/ArrowUp, Home/End between headers |
| `Draggable`, `DropZone` | `aria-grabbed`, `data-dragging`, `data-over` highlight | `PickUp` and `MoveHere` buttons replace dragging for keyboard and touch |
| `ToastRegion`, `Toast` | polite live region, `role="status"`, labelled Dismiss button | the auto-dismiss countdown pauses on hover and while focus is inside |

### Browser support

The baseline features (`<dialog>` with `showModal()`, the Popover API and its `toggle` and
`beforetoggle` events) are in every current browser: Chrome/Edge 114+, Firefox 125+, Safari 17+.
Newer features are used where available, with a fallback elsewhere:

| Feature | Used for | Supported | Fallback |
| --- | --- | --- | --- |
| `<dialog closedby="any">` | closing a Dialog on a backdrop click | Chrome/Edge 134+, Firefox 141+; not Safari at the time of writing | a Datastar click-outside handler that only acts when `'closedBy' in HTMLDialogElement.prototype` is false |
| CSS anchor positioning (`anchor-name`, `position-anchor`, `position-area`, `position-try-fallbacks`) | placing the Dropdown menu under its trigger and flipping it above near the bottom of the viewport | Chrome/Edge 125+ (`position-try-fallbacks` 128+), Safari 26+; Firefox not in a release at the time of writing | on `beforetoggle`, the menu's `top`/`right` are set from the trigger's bounding box; it does not follow the trigger on scroll or flip |
| `popover="hint"` | not used | Chrome/Edge 133+ only | `Tooltip` stays CSS only: a hint popover needs script to show on hover and focus, plus positioning, and only adds the top layer and Escape |

### Limits

- A modal dialog does not stop the page behind it from scrolling. Add
  `html:has(dialog:modal) { overflow: hidden; }` to your CSS (or the class
  `has-[dialog:modal]:overflow-hidden` on `<html>`) to lock it.
- A dialog or menu that the server patches while it is open loses its open state if the patch
  replaces the element.
- `Tooltip` is clipped by an ancestor with `overflow: hidden` or `auto` and is not dismissed by
  Escape.
- Menus have no type-ahead or submenus; tab lists are horizontal only and not mirrored for
  right-to-left text; and HTML drag events never fire from touch or keyboard, hence the pick-up
  fallback (and `ui.Sortable` for touch). Because of the pick-up fallback `ui.DropInto` clears
  its signal after the action runs.

See [CHANGELOG.md](CHANGELOG.md) for the API changes that came with the native overlays.
