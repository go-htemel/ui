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
| Overlays | `Dialog`/`DialogActions`, `Dropdown` with `MenuItem`/`MenuLink`/`MenuDivider` | ready; no focus trap or arrow keys yet |
| Navigation | `Tabs`/`Tab`, `Disclosure`, `Accordion`/`AccordionItem` | ready; no arrow keys yet |
| Drag and drop | `Draggable`, `DropZone`, `DropInto`, `DragData` | ready; mouse only (HTML drag events) |
| Data | `Table` (`TableHead`, `TableBody`, `Th`, `Td`; `Striped`, `Dense`, `StickyHeader`), `SortableTh`, `EmptyState`/`EmptyRow`, `Pagination`, `PaginationSignal`, `Stat`/`StatGrid`, `DescriptionList` | ready |
| Navigation | `Navbar`/`NavbarSignal` with `NavLink`, `Sidebar`/`SidebarGroup` (collapsible), `Breadcrumbs`/`Crumb`, `Steps` | ready; no arrow keys yet |
| Progress | `Progress`, `ProgressSignal`, `Skeleton`, `Spinner` | ready |
| Misc | `Tooltip` (CSS only, no Escape), `InputGroup` | ready |
| Planned | Command palette | — |

## Interactivity

State lives in Datastar signals that you name. `Dialog`, `Dropdown`, `Disclosure`, `Tabs`,
`Accordion` and `AlertDismissible` take the signal as their first argument and declare it with
`__ifmissing`, so a server can preset it. `ui.Open`, `ui.Close` and `ui.Toggle` return the
expressions, and `ui.Opens(signal)` is the click handler for a trigger outside the component:

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

## Theming

Classes use Tailwind's zinc palette with `dark:` variants and are written for `ui.DefaultTheme`
(primary `sky`, radius `md`, font `font-sans`). Change it once at startup:

```go
ui.SetTheme(ui.Theme{Primary: "indigo", Radius: "lg", Font: "font-serif"})
```

Every `-sky-` class becomes `-indigo-`, every `rounded-md` (and `rounded-l-md`, ...) becomes
`rounded-lg`, and `ui.BodyClasses()` carries the font. `ui.Classes(...)` applies the theme to
your own class strings.

## Tailwind

Tailwind only generates classes it finds in your sources, and the ui sources sit in the Go module
cache. Add the class list to your build instead. For the default theme, use the committed
`tailwind.safelist.txt`; for another theme, generate one:

```shell
go run github.com/go-htemel/ui/cmd/safelist -primary indigo -radius lg > tailwind.safelist.txt
```

```css
/* app.css (Tailwind v4) */
@import "tailwindcss";
@source "./tailwind.safelist.txt";
```

Tailwind v4 is required (the components use `shadow-xs`, `has-checked:` and `*:` variants). The
showcase uses the Tailwind browser build from a CDN; real applications should build their CSS
with the Tailwind CLI.

## Showcase

```shell
make run   # http://localhost:8080
```

`cmd/showcase` renders one page per component group with each live component next to the Go
source that produced it, plus datastar-go demos: server-pushed toasts, live form validation, a
kanban board whose drops post to the server and a table that sorts and pages on the server. Example functions live in
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

- `make test` runs `go test -race ./...`: golden HTML per component (`go test -run TestGolden
  -update` rewrites `testdata/`), compile-time content assertions, safelist checks and showcase
  page checks.
- `make lint` runs golangci-lint v2.
- `make tailwind-safelist` regenerates `tailwind.safelist.txt` after a class change; a test fails
  until you do.

## License

MIT

## Accessibility

The interactive components follow the matching WAI-ARIA patterns using Datastar attributes and
expressions only, with no JavaScript files. Each component's Go documentation lists exactly what
it covers.

| Component | Semantics | Keyboard and focus |
| --- | --- | --- |
| `Dialog` | `role="dialog"`, `aria-modal`, `aria-labelledby` the title | focus moves in on open and back to the trigger on close; Tab and Shift+Tab stay inside; Escape closes; page scroll is locked |
| `Dropdown` | trigger `aria-haspopup`, `aria-controls`, bound `aria-expanded`; `role="menu"` and `menuitem` | ArrowDown/ArrowUp open from the trigger and move between items; Home/End; Escape returns focus to the trigger; Tab closes |
| `Tabs` | `tablist`, `tab`, `tabpanel`, `aria-selected`, `aria-controls`, roving `tabindex` | ArrowLeft/ArrowRight, Home/End |
| `Disclosure`, `Accordion` | `aria-expanded`, `aria-controls`, panel `role="region"` | Accordion: ArrowDown/ArrowUp, Home/End between headers |
| `Draggable`, `DropZone` | `aria-grabbed`, `data-dragging`, `data-over` highlight | `PickUp` and `MoveHere` buttons replace dragging for keyboard and touch |
| `ToastRegion`, `Toast` | polite live region, `role="status"`, labelled Dismiss button | the auto-dismiss countdown pauses on hover and while focus is inside |

Limits that need JavaScript or browser support Datastar does not have: a Dialog does not make the
rest of the page inert (a screen reader's virtual cursor can leave it) and its scroll lock
overwrites an inline `overflow` style on `<body>`; menus have no type-ahead or submenus; tab lists
are horizontal only and not mirrored for right-to-left text; and HTML drag events never fire from
touch or keyboard, hence the pick-up fallback. Because of the pick-up fallback `ui.DropInto` now
clears its signal after the action runs.
