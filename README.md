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
| Icons | `Icon` (Font Awesome), `IconHome`, `IconDocument`, `IconFile`, `IconFolder`, `IconUsers`, `IconSettings`, `IconSidebar`, `IconArrowLeft`, `IconUpload`, `IconMenu`, `IconChevronLeft`/`Right`, `IconEllipsis`, `IconCheck`, `IconX`, `IconChevronDown`, `IconSearch`, `IconPlus`, `IconTrash`, `IconPencil`, `IconCopy`, `IconGripVertical`, `IconSpinner`; for editors `IconBold`, `IconItalic`, `IconStrikethrough`, `IconHeading`, `IconQuote`, `IconCode`, `IconLink`, `IconImage`, `IconListBullet`, `IconListNumbered`, `IconListTask`, `IconTable` | ready |
| Layout | `Stack`, `Inline`, `Container`, `Card` | ready |
| Feedback | `Alert`, `AlertDismissible`, `ToastRegion`, `Toast` | ready |
| Forms | `Field` (`HelpText`, `ErrorText`, `ErrorSignal`), `Input` (text-like, number, search, date/time, file, color, range), `Textarea`, `Select`/`Choice` (customizable picker where supported), `Checkbox`, `Radio`/`RadioGroup`, `Switch` (all drawn with tokens) | ready |
| Overlays | `Dialog`/`DialogActions` (native `<dialog>`), `Drawer`/`DrawerActions` (slide-over `<dialog>`: left, right, bottom), `Dropdown` with `MenuItem`/`MenuLink`/`MenuDivider` (Popover API) | ready |
| Navigation | `Tabs`/`Tab`, `Disclosure`, `Accordion`/`AccordionItem` | ready; no arrow keys yet |
| Drag and drop | `Draggable`, `DropZone`, `DropInto`, `DragData` | ready; mouse only (HTML drag events) |
| Data | `Table` (`TableHead`, `TableBody`, `Th`, `Td`; `Striped`, `Dense`, `StickyHeader`), `SortableTh`, `EmptyState`/`EmptyRow`, `Pagination`, `PaginationSignal`, `Stat`/`StatGrid`, `DescriptionList` | ready |
| Navigation | `Navbar`/`NavbarSignal` with `NavLink` (collapses by container width), `Sidebar`/`SidebarGroup` (collapsible), `Breadcrumbs`/`BreadcrumbsLabel`/`Crumb`, `Steps`/`StepsLabel` | ready; no arrow keys yet |
| Progress | `Progress`, `ProgressSignal`, `Skeleton`, `Spinner` | ready |
| Code | `code.Block`, `code.Inline` (subpackage `ui/code`, server-side highlighting; see Code blocks) | ready |
| Content | `MarkdownEditor`/`MarkdownPreview` (textarea, toolbar, server-rendered preview), `Prose`/`ProseSm`/`ProseLg`, `markdown.Render` (subpackage `ui/markdown`); see Markdown and Prose | ready; toolbar needs `ui.js` |
| Misc | `Tooltip` (CSS only, no Escape; see below), `InputGroup` (labelled by `Field` through its inner control) | ready |
| Layouts | `AppShell` (`SidebarFixed`, `SidebarCollapsible`, `SidebarDark`, `TopbarOnly`), `SplitLayout` (`SplitSelect`, `SplitDetailHeader`), `SettingsLayout`, `StackedLayout`/`StackedTabs`, `PageHeader`, `SectionHeader`, `SkipLink`; `NavLink(...).Icon(...).Badge(...).Ancestor()`; all switch on their container's width (see Layouts) | ready; rail persistence uses localStorage |
| Content inputs | `TagInput`, `FileUpload`, `DateInput`/`DateTimeInput`/`TimeInput`, `SlugField`, `FilterBar`/`FilterSearch`/`FilterReset` (see Content inputs) | ready; tag and file editing need `ui.js` |
| Planned | Command palette | — |

## Layouts

Page frames for admin pages, a CMS, notes and journals. Every layout adapts to its **container**,
not the viewport: its root is a Tailwind v4 CSS container (`@container`) and its breakpoints are
container queries (`@3xl:`, `@2xl:`, `@xl:`). A layout in a narrow column, a preview frame or
another layout's pane takes its narrow form on any screen, and layouts nest (a `SplitLayout`
inside an `AppShell`'s content). They never set an outer max-width: use `Container` for that.

| Layout | Narrow form | Wide form from | Why that width |
| --- | --- | --- | --- |
| `AppShell` | top bar with a menu button; the sidebar opens in a left `Drawer` | `@3xl` (48rem) | a 16rem sidebar plus a 32rem content column, the narrowest that still fits a table or a form beside it |
| `SplitLayout` | one pane, chosen by `$split` (`"list"` / `"detail"`) | `@2xl` (42rem) | a 20rem list and a detail at least as wide |
| `SettingsLayout` | the nav is a labelled `Select` of the sections above the content (navigates on change) | `@xl` (36rem) | a 13rem vertical nav plus a usable form column |
| `PageHeader` | actions wrap under the title | `@xl` (36rem) | room for a title and two buttons on one line |
| `StackedLayout` | side padding 1rem | `@xl` 1.5rem, `@5xl` 2rem | padding only; no switch |

The showcase draws the wide previews of this section as `framed()` containers (container queries
work in any box) but the phone-width ones as 390x760 `<iframe>`s of bare pages
(`/components/layouts/frame/{appshell,split,settings}`). A narrow `AppShell` opens its sidebar in a
native modal `<dialog>`, and a modal dialog lives in the top layer of its *document*: inside a
fixed-size div it would cover the whole viewport, not the preview. An iframe is a real viewport, so
the drawer, its backdrop and the focus trap behave there as they do on a phone. (Your own previews
of a layout that opens a `Drawer` or `Dialog` want the same.)

```go
func adminPage(content ...htemel.FlowContent) htemel.Node {
	return html.Body(ui.AppShell(
		ui.WithBrand(logo()),
		ui.WithSidebar(ui.Sidebar("nav",
			ui.SidebarGroup("content", "Content",
				ui.NavLink("Posts", "/posts", true).Icon(ui.IconDocument()).Badge("12"),
				ui.NavLink("Pages", "/pages", false).Icon(ui.IconFolder())))),
		ui.WithSidebarFooter(ui.ColorSchemeToggle()),
		ui.WithTopbar(search(), html.Div(userMenu()).Class("ml-auto")),
		ui.WithContent(html.Div(ui.PageHeader("Posts", ui.WithActions(newPost())), table()).Class("p-6")),
		ui.SidebarCollapsible(),
	)).Class(ui.BodyClasses() + " h-dvh")
}
```

**`AppShell(opts...)`** fills its parent's height (`h-dvh` on `<body>`); the sidebar and the
content scroll on their own. Slots: `WithBrand`, `WithSidebar`, `WithSidebarFooter`, `WithTopbar`,
`WithContent`. Variants (combinable where it makes sense): `SidebarFixed` (default),
`SidebarCollapsible` (a button collapses the sidebar to a 4rem icon rail; the state is the
`<id>Rail` signal, remembered in `localStorage["ui-shell-rail:<id>"]`), `SidebarDark` (the sidebar
carries the `dark` class, so it renders its dark variant on a `neutral-900` surface in both
schemes) and `TopbarOnly`. `WithShellID` names the ids and signals when a page has several;
`WithEmbedded(label)` renders a preview without `<main>`/`<aside>` for pages that have their own.
`MainID()`, `MenuSignal()` and `RailSignal()` return the names it used.

- Landmarks: `<header>`, `<aside aria-label="Sidebar">`, one `<main id="main" tabindex="-1">`,
  and a `SkipLink` to it first (visible on focus).
- Narrow containers get the sidebar in a native modal `Drawer`. The drawer holds a second
  rendering of the sidebar and footer; ids in it get a `-drawer` suffix (references follow), so
  there are no duplicate ids, and the signals are shared.
- The rail hides labels, badges, section headings and the footer *visually*; links keep their
  names, icon links (`NavLink(...).Icon(...)`) carry a `title` tooltip, and the rail button says
  "Collapse sidebar" / "Expand sidebar" with a bound `aria-expanded`.
- Motion: the drawer slides like `Drawer`; the rail switches width **instantly**. Animating
  `grid-template-columns` would lay the content column out again on every frame (a long post
  rewraps every line), which is what the motion rules (opacity and transforms only, checked by
  `TestMotionClasses`) exist to avoid; the icons sit at the same x in both states, so nothing
  jumps.
- Datastar Pro's `data-persist` would also work for the rail; the shell uses `localStorage`
  directly so it works with the free bundle.

**`SplitLayout(list, detail, opts...)`** is master/detail. Options: `WithListWidth` (Sm 16rem, Md
20rem, Lg 24rem, Xl 28rem), `WithResizable` (CSS `resize-x` on the list, 14rem to 60% of the
layout), `WithPaneLabels`, `SplitSignal`. Put `ui.SplitSelect(signal, action)` on each list item
(it shows the detail, runs `action` such as `datastar.Get("/notes/3")` and, when narrow, focuses
the detail pane) and `ui.SplitDetailHeader(title, actions...)` at the top of the detail (its Back
button, shown only when narrow, returns to the list and focuses it).

**`SettingsLayout(nav, content...)`** (`SettingsLayoutLabel` to rename the nav),
**`StackedLayout(header, content...)`** with **`StackedTabs(label, items...)`** (link tabs for the
sections of a record, moved into the header band), **`PageHeader(title, opts...)`** (`WithBackLink`,
`WithBreadcrumbs`, `WithDescription`, `WithMeta`, `WithActions`, `WithHeadingLevel`) and
**`SectionHeader(title, actions...)`** complete the family.

`NavItem` gained three chainable states, all optional and backwards compatible (a plain
`NavLink` renders exactly as before): `.Icon(svg)`, `.Badge("12")` and `.Ancestor()` (the parent of
the current page: full-contrast label, no `aria-current`). The current page is still the
`active` argument, `aria-current="page"`.

## Content inputs

Inputs for writing and publishing. All are `FormControl`s or `CompositeControl`s, so `Field`
labels them and adds help and error text.

| Input | Form value | Script |
| --- | --- | --- |
| `TagInput(name, tags, opts...)` | one hidden `<input name=name>` per chip: `tags=go&tags=css`, read with `r.Form["tags"]`; text left in the field is submitted as one more tag | `<ui-taginput>` adds and removes; without it the server-rendered tags still submit |
| `FileUpload(name, opts...)` | the native `<input type="file">` (visually hidden, inside the browse `<label>`), so a multipart post sends the files | `<ui-fileupload>` adds drop, the file list, previews and removal |
| `DateInput`, `DateTimeInput`, `TimeInput` | `YYYY-MM-DD`, `YYYY-MM-DDTHH:MM`, `HH:MM` | none |
| `SlugField(name, fromSignal)` | the slug, bound to the signal `name` | none: a `data-effect` and an input handler |
| `FilterBar`, `FilterSearch(signal, action)`, `FilterReset(signals...)` | signals sent with the request | none |

- **TagInput**: Enter or a comma adds (a pasted "a, b, c" adds three), Backspace in the empty field
  removes the last tag, duplicates (ignoring case) and tags over `TagMax(n)` are refused and
  announced. `TagSuggestions(...)` fills a `<datalist>` (a picked suggestion is added at once),
  `TagPlaceholder`, and `TagSignal("tags")` mirrors the list into an array signal (one way).
  Each remove button is "Remove go"; a visually hidden hint describes the field; a polite live
  region announces changes; focus returns to the field after a removal.
- **FileUpload**: `FileMultiple`, `FileAccept("image/*")`, `FileHint("PNG or JPG, up to 5 MB")`,
  `FileBrowseLabel`. `FileUploadAction(expr)` runs once per newly chosen file with
  `evt.detail.file`: uploading and progress are the app's. Keep the progress in a signal and
  show it with `ProgressSignal`:

  ```go
  ui.FileUpload("files", ui.FileMultiple(), ui.FileUploadAction(
  	"const x = new XMLHttpRequest(); x.open('POST', '/upload'); "+
  		"x.upload.onprogress = (e) => { $uploadPct = Math.round(100 * e.loaded / e.total) }; x.send(evt.detail.file)"))
  ui.ProgressSignal("uploadPct", 100, ui.ProgressLabel("Upload"))
  ```
- **Date and time**: `DateMin`, `DateMax`, `DateStep`, `DateValue`; they return the
  `*html.InputElement`, styled by the native-control rules of `ui.css`.
- **SlugField**: follows `fromSignal` through a slugify expression (lowercase, accents removed,
  hyphens) until the reader types in it (`$<name>Touched`); clearing it resumes. When editing a
  saved post, preset `{"slug": p.Slug, "slugTouched": true}`.
- **FilterBar** is a wrapping `role="search"` row; `FilterSearch` is a search input bound to a
  signal that runs its action 300ms after the last keystroke (`FilterDebounce`); `FilterReset`
  returns `"$q = ''; $team = ''"` to chain before the reload.

### Custom element contracts (`ui.js`)

| Element | Attributes it reads | Events it dispatches |
| --- | --- | --- |
| `<ui-taginput>` | `data-ui-name`, `data-ui-max`; children `[data-ui-tag]` (with `[data-ui-tag-remove]` and a hidden input), `[data-ui-tag-input]`, `template[data-ui-tag-template]`, `[data-ui-tag-status]` | `ui:tags` (bubbles) `{tags}` after every change; methods `add(text)`, `remove(chip)`, getter `tags` |
| `<ui-fileupload>` | `data-ui-upload`; children `input[type=file]`, `[data-ui-dropzone]` (gets `data-ui-over`), `[data-ui-files]`, `template[data-ui-file-template]`, `[data-ui-file-status]` | `ui:upload` (bubbles) `{file, files}` per newly selected file (with `data-ui-upload`); a `change` on the input whenever the selection changes |

## Interactivity

State lives in Datastar signals that you name. `Dialog`, `Drawer`, `Dropdown`, `Disclosure`,
`Tabs`, `Accordion` and `AlertDismissible` take the signal as their first argument and declare it with
`__ifmissing`, so a server can preset it. `ui.Open`, `ui.Close` and `ui.Toggle` return the
expressions, and `ui.Opens(signal)` is the click handler for a trigger outside the component.
`Dialog`, `Drawer` and `Dropdown` are native overlays (a modal `<dialog>` and a `popover`): the browser
opens, closes and focuses them, and their signal mirrors the state in both directions, so it
stays correct after Escape or a click outside and setting it opens or closes them:

```go
ui.Button(ui.Danger, "Delete").Attrs(ui.Opens("confirm"))
ui.Dialog("confirm", "Delete project?", html.P().Text("This cannot be undone."))

ui.Button(ui.Secondary, "Filters").Attrs(ui.Opens("filters"))
ui.Drawer("filters", ui.DrawerRight, "Filters", form, ui.DrawerActions(cancel, apply))
```

`Navbar` is a CSS container: it shows its links inline from the `@2xl` container size (42rem)
and the menu button with a right `Drawer` below it, whatever the viewport. Tailwind's container
`@md` (28rem) is too narrow for a brand and four links, hence `@2xl`, close to the old
viewport `md` (48rem). `NavLink(...).Attrs(...)` adds attributes to a link; the highlight
follows `aria-current="page"`, so binding it moves the active link client-side:

```go
ui.NavLink("Team", "/team", false).Attrs(
	datastar.OnClick("$page = 'team'", datastar.Prevent()),
	datastar.Attr("aria-current", "$page === 'team' ? 'page' : false"))
```

Toasts are pushed by the server into `ui.ToastRegion()` with datastar-go:

```go
sse := datastar.NewSSE(w, r)
s, _ := htemel.RenderString(ui.Toast(ui.NextToastID(), ui.Success, "Saved", "", 4*time.Second))
sse.PatchElements(s, datastar.WithSelectorID(ui.ToastRegionID), datastar.WithModeAppend())
```

## Tailwind

Tailwind v4 is required (the components use `shadow-xs`, `has-checked:`, `*:` variants and CSS
theme variables, container queries). The components are styled with design tokens (see [Theming](#theming)), so your
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
@import "./ui.theme.css";              /* default tokens: @theme, the dark variant, color-scheme */
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
block, the dark variant and a base layer), not a stylesheet browsers apply on their own. The showcase compiles its CSS in the browser
with the Tailwind browser build and puts `ui.ThemeCSS()` in its `<style type="text/tailwindcss">`
block; real applications should build their CSS with the Tailwind CLI and add
`ui.ColorSchemeBootstrap()` to `<head>` (see [Light and dark mode](#light-and-dark-mode)).

## Theming

Components never name a palette color. They use semantic tokens, defined as CSS variables in
`ui.theme.css`, and you retheme by overriding those variables in CSS:

| Token | Utilities | Default | Used for |
| --- | --- | --- | --- |
| `--color-primary-{50..950}` | `bg-primary-600`, `text-primary-600`, `outline-primary-600`, ... | `indigo` | primary buttons, links, focus rings, selected and active states, drop targets |
| `--color-neutral-{50..950}` | `bg-neutral-50`, `text-neutral-500`, `ring-neutral-950/5`, ... | `zinc` | surfaces, text, borders, dividers, the page background |
| `--color-success-{50..950}` | `bg-success-50`, `text-success-700`, ... | `green` | Success tone (Badge, Alert, Toast, Progress), positive Stat deltas |
| `--color-warning-{50..950}` | `bg-warning-50`, `text-warning-800`, ... | `amber` | Warning tone |
| `--color-danger-{50..950}` | `bg-danger-600`, `outline-danger-500`, ... | `red` | Danger tone and button, invalid fields, error text, negative Stat deltas |
| `--color-info-{50..950}` | `bg-info-50`, `text-info-700`, ... | `sky` | Info tone (sky, not blue, so it does not read as a second primary next to indigo) |
| `--radius-control` | `rounded-control`, `rounded-l-control`, ... | `--radius-md` | buttons, button and input groups, inputs, menu items, nav links, pagination, tooltips, `Kbd` |
| `--radius-surface` | `rounded-surface` | `--radius-md` | Card, Dialog, Dropdown panel, Alert, Toast, Stat, Table frame |
| `--color-field-border`, `--color-field-border-hover` | `outline-field-border`, ... | `neutral-300`, `neutral-400`; dark: white at 15% and 30% | outline of Input, Textarea, Select, InputGroup affixes and the MarkdownEditor frame (see [Contrast](#contrast)) |
| `--color-control-border`, `--color-control-border-hover` | `border-control-border`, `ring-control-border`, ... | `oklch(65% 0.015 286)`, `neutral-500`; dark: white at 35% and 55% | 3:1 boundary of Checkbox and Radio, the off Switch track (see [Contrast](#contrast)) |
| `--font-sans` (Tailwind's own) | `font-sans` | Tailwind's sans stack | `ui.BodyClasses()` puts `font-sans` on `<body>`; `ui.FontCSS` sets it to Geist (see [Fonts](#fonts)) |

The `Brand` tone (`ui.Badge(ui.Brand, "Team")`, also Alert, Toast and Progress) uses the
primary family's tints (`bg-primary-50 text-primary-700`, dark `bg-primary-400/10
text-primary-300`), for labels that belong to your product rather than a status.

Pills and avatars stay `rounded-full`. Each family defines the full 50–950 scale, so your own
markup can use any step (`bg-primary-100`), and `ui.Classes(...)` joins class lists for your own
components.

### Contrast

The shipped defaults meet WCAG 2.1 AA (4.5:1 for text, 3:1 for borders, focus rings and fills) in
light and dark, and the Go contrast test (`go test -run Contrast .`) measures every pair the
components render. The steps follow one convention:

- Solid surfaces with white text (Primary button, current page) use `bg-primary-600`, hover
  `bg-primary-700` in light (indigo-600 is 6.5:1 with white). In dark they use
  `dark:bg-primary-600` and hover to the *lighter* `dark:hover:bg-primary-500` (4.6:1 with
  white). The Tailwind UI pairing 500 to 400 does not pass: indigo-400 is only 3.1:1 with
  white. Danger is `bg-danger-600`, hover `700` in light and `dark:bg-danger-700`, hover
  `dark:hover:bg-danger-600` (lighter, 4.8:1) in dark. Progress fills and completed Steps sit
  at 600 (light) and 500 (dark fill).
- Links and active labels are `text-primary-600` (hover `700`) in light and
  `dark:text-primary-400` (hover `300`) in dark.
- Focus rings are `outline-primary-600` (`outline-neutral-500` on neutral controls) in light and
  `primary-500` in dark; placeholders are `neutral-500` / `dark:neutral-400`.
- Text-field outlines (Input, Textarea, Select, InputGroup affixes, the MarkdownEditor frame) use
  `--color-field-border`, the soft Tailwind UI weight: `neutral-300` (white at 15% in dark),
  darkening on hover to `--color-field-border-hover` (`neutral-400`, white at 30%). That resting
  outline is about 1.5:1, and that is a deliberate reading of WCAG 1.4.11: its Understanding
  document says a text input identified by its label, its fill and its focus indicator does not
  need a 3:1 boundary. What does need 3:1 is the focused state, and every focus outline is a 2px
  primary ring (`outline-primary-600`, dark `primary-500`) that the contrast test enforces. The
  test still measures the resting and hover outlines and reports them as *informational* rows
  (`CONTRAST_REPORT=report.md go test -run Contrast .`); they never fail.
- Checkbox, Radio and the off Switch track always keep a 3:1 boundary, because there the box *is*
  the control (the example WCAG itself gives). They use their own token, `--color-control-border`:
  the lightest neutral-hued shade that clears 3:1 on both white and the page background,
  `oklch(65% 0.015 286)`, `#8e8e98` (3.24:1 on white, 3.11:1 on `neutral-50`), hover `neutral-500`;
  in dark white at 35% (hover 55%). Those rows are enforced.
- Both tokens flip with `.dark`, so components carry no `dark:` variant for them. Secondary and
  file-input button rings, the Table frame and `Kbd` are decoration and stay at their pale
  `neutral-300` / `neutral-950/5` ring.

  **Strict opt-in.** A team that wants 3:1 boundaries on every field points the field token at a
  darker step (light in your `@theme`, dark in a `.dark` rule):

  ```css
  @theme { --color-field-border: var(--color-neutral-500); --color-field-border-hover: var(--color-neutral-600); }
  @layer base { .dark { --color-field-border: rgb(255 255 255 / 0.4); --color-field-border-hover: rgb(255 255 255 / 0.55); } }
  ```

  (or reuse the checkbox value: `--color-field-border: var(--color-control-border)`). The values
  follow the zinc neutral; if you retheme `neutral`, rerun the contrast test and adjust
  `--color-control-border`.

#### The default primary: indigo

The default `primary` is stock Tailwind `indigo`: every step is a plain alias
(`--color-primary-600: var(--color-indigo-600)`), like the other families. There is no custom
palette. The showcase's Theming page has "Sky (Tailwind stock)" and other themes to try. To build
your own palette, `go run ./cmd/palettetune` is a helper that solves OKLCH lightness per step
against the WCAG rules; its output is for custom palettes, not the default.

#### Retheming to a stock palette

The test guards the shipped defaults only (the tokens in `assets/ui.css`). Light hues such as
sky, cyan, amber, lime and yellow only reach 4.5:1 against white text at the 700 step, so with
one of those the solids need `bg-primary-700` (hover 800) and links `text-primary-700`: map the
tokens one step darker (`--color-primary-600: var(--color-sky-700)`, `--color-primary-700:
var(--color-sky-800)`) or accept the contrast failures. Indigo (the default), blue and violet have
room to spare. After retheming, run the test against your palette: change the tokens in a
checkout of this module and rerun it (`CONTRAST_REPORT=contrast.md go test -run 'Contrast$' .`
writes the full table with a suggested step for every failing pair).

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
`bg-primary-600 hover:bg-primary-700 dark:text-primary-400`). Any CSS color works as a value
(`oklch(...)`, hex), not only a Tailwind palette variable. The token families share their names
with nothing in Tailwind's palette except `neutral`: `ui.theme.css` redefines Tailwind's `neutral`
scale as the zinc alias, so `bg-neutral-*` in your own markup follows the token too.

> **Font:** if `--font-sans` falls back to `system-ui`, Windows renders Segoe UI, whose metrics
> place text about a pixel above centre in buttons and badges. Point it at a metrically balanced
> face such as Inter, as above, or use the bundled Geist (next section; the showcase does).

### Fonts

The library ships [Geist](https://vercel.com/font) and Geist Mono as opt-in, self-hosted
variable fonts. Geist is a clean neutral sans designed for interfaces, with balanced vertical
metrics (no Segoe UI offset in buttons and badges) and a matching monospace for code blocks. They
are served from your own origin, so no page view makes a request to a third-party font service
(no Google Fonts, no IP address or referrer handed to a CDN, nothing for a privacy policy or a
strict `font-src 'self'` policy to carve out). The two woff2 files (70 KB and 71 KB, weights
100 to 900 each) are embedded in the module and served by `ui.Assets()` as `font/woff2` with the
same year-long `immutable` caching and `ETag` as `ui.js`, under `/fonts/`.

Add the stylesheet next to the tokens and the preload links to `<head>`, with the prefix where
you mounted `Assets`:

```go
// Tailwind source: tokens, then fonts
css := ui.ThemeCSS() + ui.FontCSS("/assets")

head := html.Head( /* ... */ )
for _, l := range ui.FontPreload("/assets") { // <link rel="preload" as="font" type="font/woff2" crossorigin href="/assets/fonts/...?v=<hash>">
	head.Children(l)
}
```

`FontCSS(prefix)` returns two `@font-face` rules (`font-display: swap`, `font-weight: 100 900`,
`src: url(<prefix>/fonts/....woff2?v=<hash>)`) and an `@theme` block with
`--font-sans: "Geist", ui-sans-serif, system-ui, sans-serif` and `--font-mono: "Geist Mono",
ui-monospace, SFMono-Regular, Menlo, monospace`, so `font-sans` (on `<body>` through
`BodyClasses`) and `font-mono` (code blocks, `Kbd`) use them. `FontPreload(prefix)` returns the
two `<link>` elements with exactly the same URLs, content hash included, so the browser fetches
each file once and the first paint already has the face. Without the preload the text shows the
fallback first and swaps; the metrics differ slightly, so preloading avoids a visible reflow.

Fonts are opt-in. An application that does not include `FontCSS` keeps Tailwind's system stacks
and never requests a font file (the files are only served, not linked). To use another face, leave
`FontCSS` out and set `--font-sans` / `--font-mono` in your own `@theme`; to keep Geist for text
and change only the code face, override `--font-mono` after `FontCSS`. The body weight stays at
the default 400: Geist reads well there, and a heavier base such as 450 would shrink the step to
`font-medium` that labels and headings rely on.

Geist and Geist Mono are copyright Vercel, in collaboration with basement.studio, licensed under
the [SIL Open Font License 1.1](assets/fonts/LICENSE-Geist.txt). The OFL allows bundling and
redistributing the fonts with software, as this module does, but not selling them by themselves;
keep the licence file with the font files if you copy them out of the module.

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
the palette color it aliases (for a stock theme, `--color-sky-600`): the alias resolves where it is declared, on
`:root`. Tailwind emits the palette variables your CSS references (`var(--color-emerald-600)`
above), but CSS it never sees, such as a per-tenant `style="--color-primary-600: ..."` added at
runtime, should use literal color values.

### Dark mode

Components carry `dark:` variants that pick other steps of the same tokens (`text-neutral-900
dark:text-neutral-100`), so one override covers both modes. Which mode applies, and how visitors
choose it, is described in [Light and dark mode](#light-and-dark-mode). To use different palettes
per mode, override the tokens inside the dark scope (`.dark { --color-primary-500: ...; }`).

## Light and dark mode

Dark mode follows a `.dark` class on `<html>`, not the operating-system setting alone. `ui.css`
declares `@custom-variant dark (&:where(.dark, .dark *));`, so every `dark:` utility keys off that
class, and a small `@layer base` block sets `color-scheme` from it (`html.dark { color-scheme:
dark }`, `html:not(.dark) { color-scheme: light }`) so native controls, scrollbars and the Select
arrow match the page.

Three pieces make it work:

1. **The bootstrap.** `ui.ColorSchemeBootstrap()` is an inline `<script>` that reads
   `localStorage["color-scheme"]` (`ui.ColorSchemeKey`: `"light"`, `"dark"`, or absent for
   "system") and toggles the class: dark when the choice is `dark`, or when it is not `light` and
   `prefers-color-scheme: dark` matches. It also listens for OS changes, so a page in system mode
   follows the OS live. Put it in `<head>` **before** your stylesheet or the Tailwind script and
   render it on every page. It is the only inline script in the library, deliberately: a deferred
   or external script would run after the first paint and dark-mode visitors would see a flash of
   the light page.

   ```go
   html.Head(
       ui.ColorSchemeBootstrap(),
       html.Link().Rel("stylesheet").Href("/static/app.css"),
       ui.ScriptTag("/assets"),
   )
   ```

2. **Content-Security-Policy, if you have one.** A policy without `'unsafe-inline'` blocks
   the script. Pass a per-response nonce, `ui.ColorSchemeBootstrap(ui.WithNonce(nonce))`, and send
   `script-src 'self' 'nonce-<nonce>'`; or allow its hash (the text is
   `ui.ColorSchemeBootstrapSource()`) as `'sha256-...'`.

3. **The toggle.** `ui.ColorSchemeToggle()` is a System / Light / Dark segmented control
   (`ui.WithLabels()` adds visible words to the icons). Each button sets the class, stores the
   choice, and `aria-pressed` follows the `colorScheme` signal, which is initialised from the
   stored choice. System removes the stored choice and re-derives the class from the OS. Place it
   anywhere (the showcase puts it at the bottom of its sidebar); several toggles on one page stay
   in step. It needs Datastar and the bootstrap.

```go
html.Div(ui.ColorSchemeToggle(ui.WithLabels())).Class("p-4")
```

**OS-only apps.** If you do not want a toggle and are happy with `prefers-color-scheme`, delete
the `@custom-variant dark` line and the `@layer base` color-scheme block from your copy of
`ui.css` (`go run github.com/go-htemel/ui/cmd/themecss > assets/ui.theme.css`, then edit); `dark:`
reverts to Tailwind's media-query default and neither the bootstrap nor the toggle is needed.
The reverse is the trap: with the class variant kept and no bootstrap, the page is always light
unless your own code sets `.dark`.

## Form controls

`Checkbox`, `Radio` and `Switch` are drawn entirely with tokens: the real `<input>` keeps its
semantics, keyboard and form behaviour but has `appearance-none`, and the box, the circle, the
check, the indeterminate dash and the dot are drawn with utilities (`border-neutral-500`,
`checked:bg-primary-600`, `peer-checked:` marks, `focus-visible:` outlines, muted `disabled:`
states, `dark:` variants). There is no `accent-color`, so browsers draw no hover halo. In forced
colors mode (Windows High Contrast) the native controls come back. Borders, fills, marks and the
switch knob meet 3:1 (SC 1.4.11) in both schemes; the contrast test has a row for each.

`Input`, `Textarea` and `Select` carry the class `ui-control` (and `Select` also `ui-select`).
`ui.css` uses it for what utilities cannot reach, so keep `ui.theme.css` imported:

- **Autofill.** Chrome and Safari paint autofilled fields yellow or blue with an `!important`
  rule. An inset `box-shadow` in the field colour covers it and `-webkit-text-fill-color`
  restores the text. The cover has to be opaque, so dark mode uses `neutral-800` (the field
  itself is the translucent `bg-white/5`). Browsers cannot fake autofill for a test, so this is
  verified by inspection and by checking the compiled CSS.
- **Native inputs.** `number` has no spin buttons (arrow keys still step); the `search` clear
  button is a cross in the text colour; the `date`/`time` picker icon keeps its native glyph,
  which follows `color-scheme` (no invert filter needed); `file` gets a small Secondary-style
  button (`file:` utilities); `color` is a field-height swatch; `range` is a bare slider in the
  primary colour.
- **Customizable select.** Where `appearance: base-select` is supported, the open picker is
  styled like the Dropdown menu: tokens, `rounded-surface`, shadow and ring, padded options with
  a primary highlight on hover and keyboard focus, a check mark (`::checkmark`) on the selected
  option, and a 200ms fade with `@starting-style` (none under reduced motion). The library
  chevron stays the icon (`::picker-icon` is hidden), so the closed select looks the same
  everywhere. These rules are unlayered so `appearance: base-select` beats the `appearance-none`
  utility.

| Browser | Select picker |
| --- | --- |
| Chrome / Edge 135+ | customizable picker (`appearance: base-select`) |
| Firefox, Safari (at the time of writing), older Chromium | native popup; the closed select is styled the same |

## Code blocks

Syntax highlighting runs on the server, in its own package so that applications that do not use it
do not link the highlighter ([chroma](https://github.com/alecthomas/chroma) v2; the root `ui`
package has no dependency on it):

```go
import "github.com/go-htemel/ui/code"

code.Block("go", src, code.WithLineNumbers(), code.WithTitle("main.go"), code.WithHighlight(3))
html.P(htemel.Text("Call "), code.Inline("go", `ui.Button(ui.Primary, "Save")`))
```

`Block(lang, src, opts...)` returns a `*code.BlockElement`, a `<figure class="ui-code">` (flow
content; `Attrs`, `SetAttr` and `AddClass` reach the figure) with an optional toolbar, a focusable
`<pre><code>` with one `<span class="line">` per line, and a `<template>` holding the raw source.
`Inline(lang, src)` returns a highlighted `<code>`. `lang` is a chroma lexer name or alias (`go`,
`css`, `html`, `js`, `json`, `bash`, `sql`, `yaml`, `markdown`, ...); an unknown or empty one, or a
lexer that fails, renders plain text, never an error. All text is escaped.

| Option | Effect |
| --- | --- |
| `WithLineNumbers()` | gutter of CSS counters: never selected, copied or announced |
| `WithStart(n)` | number of the first line; highlight numbers follow it |
| `WithHighlight(lines...)`, `WithHighlightRange(from, to)` | `data-ui-hl` on those lines: primary tint and a left bar |
| `WithTitle(s)` | file name in a toolbar above the code; also the figure's `aria-label` |
| `WithWrap()` | soft wrap instead of horizontal scrolling (the default) |
| `WithoutCopyButton()` | the copy button is on by default |
| `WithID(s)` | name of the copy button's signal, see below |

Cap the height with `AddClass("[--ui-code-max-height:24rem]")`; the code scrolls, the toolbar stays.

**Copy button.** An `IconButton` (`IconCopy`, which fades to `IconCheck` for 1.5s in 150ms,
without motion under `prefers-reduced-motion`) writes the template's text to the clipboard, so
the copy is the source, not the rendered lines; it needs Datastar and a secure context (https or
localhost). A visually hidden `aria-live="polite"` region says "Copied". The signal behind it is
local to the browser (`_copied_<id>`, the underscore keeps it out of requests) and `<id>` is the
first 8 hex digits of a SHA-1 of language and source, so the same input renders the same HTML.
Two blocks with identical language and source on one page share that signal and would both show
the check mark; give them distinct `WithID` values.

**Tokens and colours.** Chroma's token types are reduced to 14 classes: `tok-keyword`,
`tok-builtin`, `tok-name`, `tok-function`, `tok-type`, `tok-string`, `tok-number`, `tok-literal`,
`tok-attribute`, `tok-tag`, `tok-comment`, `tok-operator`, `tok-punctuation`, `tok-error`
(plain identifiers and whitespace have no span). The colours are variables in `assets/ui.css`,
light on `:root` and dark under `.dark`: `--color-code-keyword` (primary 700 / 300),
`-builtin` (primary 600 / 400), `-name` (neutral 800 / 200), `-function` (info 700 / 300),
`-type` (info 800 / 200), `-string` (success 800 / 300), `-number` and `-literal` (warning 800 /
300), `-attribute` (info 700 / 300), `-tag` and `-error` (danger 700 / 300), `-comment` and
`-punctuation` (neutral 600 / 400), `-operator` (neutral 700 / 300), plus `-text`, `-gutter`,
`-title`, `-surface`, `-border`, `-highlight` and `-highlight-bar`. Override them in your own
`:root` / `.dark` rules to retheme. Every pair is at least 4.5:1 on the code surface (neutral-50;
black/40 in dark), on the surface over a card and on a highlighted line; the contrast test has
the rows (`Code/...`). Regenerate `ui.theme.css` (`cmd/themecss`) to get the rules, and
`tailwind.safelist.txt` (`cmd/safelist` reads `code/` as well as the root package).

## Markdown

For blogs, CMS pages, admin forms, notes and journal entries: a Markdown editor (a textarea with
a toolbar, not contenteditable) whose preview is rendered on the server, and a renderer that
turns stored Markdown into sanitised HTML.

### Rendering: `github.com/go-htemel/ui/markdown`

A subpackage, so applications that do not import it do not link the parser or the sanitiser
([goldmark](https://github.com/yuin/goldmark) v1.8.7 and
[bluemonday](https://github.com/microcosm-cc/bluemonday) v1.0.27; both keep `go 1.24`, and the
root `ui` package imports neither):

```go
import "github.com/go-htemel/ui/markdown"

markdown.Render(post.Body)                                   // *html.DivElement: ui.Prose around the HTML
markdown.Render(post.Body, markdown.WithHeadingOffset(1))    // # becomes <h2> under the page's <h1>
markdown.Render(note.Text, markdown.WithSize(ui.Sm))         // ui.ProseSm
markdown.HTML(post.Body)                                     // the sanitised HTML as a string
markdown.Excerpt(post.Body, 160)                             // plain-text teaser, cut at a word
```

CommonMark plus GitHub Flavored Markdown: tables (alignment kept as `align` attributes), ~~strike~~,
task lists, bare URLs, footnotes, and heading ids made from the text. Fenced and indented code
blocks are highlighted by the `code` package (`code.Block` without its copy button, which needs a
`<template>` and a signal the sanitiser removes). `Plain(src)` strips the markup for search
indexes and meta descriptions; `Excerpt(src, n)` cuts that to `n` characters plus "…".

| Option | Effect |
| --- | --- |
| `WithHeadingOffset(n)` | shift headings down `n` levels (capped at `<h6>`) |
| `WithIDPrefix(p)` | prefix heading and footnote ids; use it whenever the content shares a page with other ids (an editor preview, two posts) |
| `WithoutHeadingIDs()` | no heading ids |
| `WithLinkTarget("_blank")` | `target` on links (not on `#fragment` links); `_blank` adds `rel="noopener"` |
| `WithSize(ui.Sm)` / `WithSize(ui.Lg)` | `ProseSm` / `ProseLg` instead of `Prose` |
| `WithPolicy(p)` | sanitise with your own bluemonday policy; start from `markdown.Policy()` |
| `WithUnsafeHTML()` | pass raw HTML through and skip sanitising: trusted authors only |

**Sanitising.** By default raw HTML in the source is not rendered (goldmark omits it), dangerous
URL schemes are dropped, and the output then goes through `markdown.Policy()`: bluemonday's
`UGCPolicy` (text, lists, tables, images and links; `http`, `https`, `mailto` and relative URLs
only; `rel="nofollow"` on links; no `style`, `<script>`, `<iframe>`, `<object>`, forms or event
handlers) extended with exactly what the renderer emits: the code block classes (`ui-code`,
`line`, `tok-*`, `language-*`), `tabindex="0"` on `<pre>` and the figure's `aria-label`,
task-list classes and `<input type="checkbox" disabled aria-label="Done">`, `align` on cells,
footnote classes and DPUB roles, and link targets. The tests run an XSS corpus (`<script>`,
`javascript:` / `vbscript:` / `data:` links and images, entity-encoded schemes, `onerror=`,
`<svg onload>`, `<iframe srcdoc>`, nested HTML, `style`, `<base>`, `<meta>`, forms) through every
mode except unsafe and assert none of it survives.

### Editing: `ui.MarkdownEditor`

```go
ed := ui.MarkdownEditor("body",
	ui.WithPreview(datastar.Post("/posts/preview")),
	ui.WithValue(post.Body),
	ui.WithMaxLength(20000),
)
ui.Field("body", "Body", ed, ui.HelpText("Shown on the post page."))
```

The preview comes from your server: the action posts the signals (the textarea is bound to a
signal named after the field, or `WithSignal`), and the handler patches the preview with
datastar-go:

```go
func preview(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Body string `json:"body"`
	}
	if err := datastar.ReadSignals(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body := markdown.Render(in.Body, markdown.WithIDPrefix("body-md-"), markdown.WithSize(ui.Sm))
	html, _ := htemel.RenderString(ui.MarkdownPreview(ui.MarkdownPreviewID("body"), body))
	_ = datastar.NewSSE(w, r).PatchElements(html)
}
```

`MarkdownPreview(id, content...)` is the patch target (`MarkdownEditorElement.PreviewID()`, or
`MarkdownPreviewID(editorID)` without the editor at hand). It sits inside the editor's preview
region (`role="region"`, `aria-live="polite"`), which a patch never replaces.

| Option | Effect |
| --- | --- |
| `WithPreview(action)` | add a preview fetched with a Datastar action; in tabs mode it runs when Preview is chosen |
| `WithSplit()` | preview beside the textarea once the editor is 36rem wide (`@xl` container query), stacked below that; fetched on load and 400ms after the last keystroke |
| `WithPreviewLabel(s)` | name of the preview region (default "Preview"); give each editor on a page its own |
| `WithSignal(name)` | signal of the text (default: the field name) |
| `WithEditorID(id)` | id the parts are named after (default: the name); the preview is `<id>-preview` |
| `WithValue(s)`, `WithRows(n)`, `WithPlaceholder(s)` | initial text (and counts), visible rows (default 8), placeholder |
| `WithMaxLength(n)` | "used / max characters" counter: warning colour from 90%, danger past the limit; no `maxlength` attribute, so pastes are never cut silently |
| `WithAutosize()` | `field-sizing: content` (Chromium 123+), 10rem to 70vh |
| `WithMonospace()` | `font-mono` textarea (the default is the sans font) |
| `WithToolbar(tools...)` | pick and order the buttons (`MarkdownBold`, `MarkdownItalic`, `MarkdownStrikethrough`, `MarkdownHeading`, `MarkdownQuote`, `MarkdownCode`, `MarkdownLink`, `MarkdownImage`, `MarkdownBulletList`, `MarkdownNumberedList`, `MarkdownTaskList`, `MarkdownTable`, `MarkdownSeparator`); none removes the toolbar |
| `WithUploadAction(action)` | handle pasted or dropped files: see below |
| `WithHint(text, href)` | footer hint (default "Markdown supported" linking to the CommonMark help); empty text removes it |

It is a `CompositeControl`: `Field` labels the textarea and its `aria-describedby` keeps the
editor's hint and counter after Field's help and error; `Attrs` reaches the textarea; `Id`,
`Class` and `AddClass` are the frame's.

**The toolbar needs `ui.js`.** Its buttons carry only `data-ui-md-action="bold"` (and so on); the
`<ui-mdeditor>` element in `ui.js` does the editing, so no attribute holds script and the markup
stays small. A button wraps the selection (`**bold**`, `_italic_`, `~~strike~~`, `` `code` `` or
a fenced block for several lines, `[text](https://)` / `![alt](https://)` with the URL selected),
or toggles a prefix on every selected line (`## `, `> `, `- `, `1. `, `- [ ] `); `MarkdownTable`
inserts a small table. Pressing it again undoes the wrap. Edits go through the browser's
`insertText` command (falling back to `setRangeText` plus an `input` event), so Ctrl+Z works and
`data-bind` sees every change. Shortcuts: Ctrl/Cmd+B, I, E, K and Ctrl/Cmd+Shift+X; Enter in a
list item continues the list and Enter on an empty item ends it. Tab is not captured, so keyboard
users can always leave the textarea. Without `ui.js` the toolbar stays hidden
(`ui-mdeditor:not(:defined)`) and the textarea, counter and preview still work.

**Uploads.** With `WithUploadAction`, pasting or dropping files on the textarea inserts nothing;
the element dispatches a bubbling `ui:md-upload` event per file, which runs your action with
`evt.detail.file` (the `File`) and `evt.detail.insert(text)` (insert Markdown at the caret, as if
typed). Uploading is yours:

```go
ui.WithUploadAction("fetch('/uploads', {method: 'POST', body: evt.detail.file})" +
	".then(r => r.text()).then(url => evt.detail.insert('![' + evt.detail.file.name + '](' + url + ')'))")
```

**Accessibility.** The toolbar is `role="toolbar"` (`aria-label="Formatting"`,
`aria-controls` the textarea) with a roving tabindex (one tab stop; ArrowLeft/ArrowRight,
Home/End); each button has an `aria-label`, a `title` naming its shortcut and
`aria-keyshortcuts`. The Write / Preview tabs follow `Tabs` (roles, roving focus, arrow keys,
the 200ms panel fade, none under reduced motion); the toolbar hides on the Preview tab. The
whole frame shows the 2px focus ring while anything inside has focus. Contrast rows cover the
text, placeholder, frame, focus ring, toolbar icons, hint link and the three counter states.

## Prose

`ui.Prose(children...)` styles long-form HTML that carries no classes: rendered Markdown, CMS
bodies, help pages. It is a `<div class="ui-prose">` whose descendant rules live in
`assets/ui.css` (Tailwind cannot put a class on every element of sanitised HTML, and no
typography plugin is needed):

```go
ui.Prose(html.H2().Text("Release notes"), html.P().Text("..."))   // or markdown.Render(src)
ui.ProseSm(...)                                                    // 14px body: comments, side panels
ui.ProseLg(...).AddClass("max-w-prose")                            // 18px body: reading views
```

Headings (balanced wrapping, `scroll-margin-top` for anchors), paragraphs at `text-base/7`,
underlined primary links with a focus outline, bullet and number lists with nested styles and
muted markers, blockquotes with a `control-border` bar, tables in the `Table` colours, rules,
images with `rounded-surface` and `max-width: 100%`, inline code as a chip (`neutral-100`,
`white/10` in dark), `<kbd>`, `<mark>`, definition lists, figure captions, GFM task-list
checkboxes drawn like `Checkbox` (read-only) and footnotes. Code blocks from the `code` package
keep their own look. All spacing is in `em`, so the sizes scale as a whole; Prose sets no width.
Colours are `--ui-prose-*` variables on `.ui-prose` (`body`, `headings`, `link`, `link-hover`,
`muted`, `marker`, `code`, `code-bg`, `th`, `th-bg`, `rule`, `quote-border`, `check`) with dark
values, so a scope can retheme them; the contrast test has rows for each text pair on a card and
on the page (`Prose/...`).

## Assets and progressive enhancement

Everything works without JavaScript files of your own. The one exception is optional: `ui.js`, a
small embedded script (vanilla ES2020, no dependencies, about 590 lines) that defines four custom
elements: `<ui-sortable>` for touch dragging, `<ui-mdeditor>` for the `MarkdownEditor` toolbar
and shortcuts (see Markdown), and `<ui-taginput>` and `<ui-fileupload>` for the content inputs
(see Content inputs). Serve it with `ui.Assets()` and load it with
`ui.ScriptTag()`:

```go
mux.Handle("/assets/", http.StripPrefix("/assets", ui.Assets()))   // GET /assets/ui.js (and /assets/ui.css, /assets/fonts/*.woff2)

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
`data-ui-over` and `data-ui-dragging` like the native path. Mouse pointers are ignored by the script
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
Code page shows the `code` package, the Markdown editor page posts its preview to
`/demo/markdown/preview` (rendered with the `markdown` package), the Prose page renders a CMS
article, and the Theming page switches token sets live through `<html data-theme>`. Example functions live in
`cmd/showcase/examples_*.go`; those files are embedded and parsed at startup, so the displayed
source is always the code that ran.

### Admin demo

"Admin demo" in the showcase sidebar opens `/admin`, a believable blog/CMS admin built only from
this library, as full-page routes outside the component pages' frame (`cmd/showcase/admin_*.go`;
a "back to showcase" link sits in its top bar). It is the integration test of the component
list: an `AppShell` with a collapsible rail (its state persists across routes), nav with icons
and a post-count badge, an account `Dropdown` and `ColorSchemeToggle` in the footer; a dashboard
(`StatGrid`, a `Table` of recent activity, `Progress`); a **posts list** with a `FilterBar`,
`SortableTh` columns, bulk select, row menus, confirm `Dialog`s, `PaginationSignal` and an
`EmptyState`, all served by one `GET /admin/posts/list` SSE handler that reads the signals and
patches only the table body, the select-all cell, the pager and the count line (so the search
field keeps its focus while you type); a **post editor** (`StackedLayout`, `SlugField`,
`MarkdownEditor` with a server-rendered split preview, `TagInput`, `FileUpload`, a Publish card,
`ErrorSignal` validation, a toast and a patched "Saved" line, a preview `Drawer`); **settings**
(`SettingsLayout`, `Prose` help text, the `data-theme` switch of the Theming page, remembered in
`localStorage`); **notes** (`SplitLayout` with server search and an inline Edit toggle) and a 404
`EmptyState`. Posts, notes and settings live in server memory, like the kanban board, and are
restored by `POST /admin/reset` ("Reset demo data" under Settings). The pages hold no
hand-written `data-*` attributes and no script but `ColorSchemeBootstrap`, which tests guard.

Screenshots are not stored in the repository. `task a11y` drives every admin page and state
(and `A11Y_PAGES=admin` limits a run to them); to look at the pages run `task run` and open
<http://localhost:8080/admin>. The review screenshots (dashboard light and dark, the posts list
with its bulk bar, the editor in dark, notes, settings, and the phone-width dashboard with its
drawer) were written to `/tmp/claude-0/-home-user/6cbb6813-54d5-5200-9cd8-e90c002cf8a2/scratchpad/admin/` by a throwaway Playwright script and are not part of
the module.

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
(the OS setting is emulated and picked up through the color scheme bootstrap; each scan also
checks the `.dark` class and the painted background), at WCAG 2.1 A/AA. It also opens the interactive states first: the
dialog, the drawer and the bottom sheet, the dropdown, a switched tab, an expanded disclosure
and accordion, the narrow navbar's drawer, the open customizable select picker (when the
Chromium supports `base-select`), the color scheme toggle switched to the other scheme (then restored via System), and the Markdown editor's server-rendered preview. The report groups violations by rule with impact, count, selectors and pages, and the
task exits non-zero on any WCAG violation. Behaviour checks run alongside and fail it
too: the stored scheme is applied before first paint (no flash: `.dark` is on `<html>` at
`domcontentloaded` when the choice differs from the OS), no tab list overflows vertically (the scrollbar a tab's negative margin used to cause), and
a Dialog's opacity really transitions when it opens and closes, the demo navbars switch
layout by container width, a drop never flashes the Move here buttons, a Markdown toolbar
click wraps the selection, fires `input`, updates the bound signal and undoes with Ctrl+Z (plus
the toolbar's roving tabindex, list continuation, the split preview and highlighted fenced code
in Prose), and no page logs a
console error (failed loads of the blocked external CDNs excepted). Rules tagged only `best-practice` are listed
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
`Dialog` and `Drawer` are modal `<dialog>`s and `Dropdown` a `popover`, so the top layer, inertness, Escape,
light dismiss and focus return come from the browser. The rest is Datastar attributes and
expressions. Each component's Go documentation lists exactly what it covers.

| Component | Semantics | Keyboard and focus |
| --- | --- | --- |
| `Dialog` | native `<dialog>` opened with `showModal()`, `aria-labelledby` the title | native: page inert behind it, Tab and Shift+Tab stay inside, Escape closes, focus returns to the trigger; initial focus on `autofocus` (`DialogActions` sets it on its first button) |
| `Drawer` | the same native modal `<dialog>` as `Dialog`, pinned to an edge | as `Dialog`; initial focus on an `autofocus` element, else the first form control, else the drawer itself (never a ring on the X); `Navbar`'s mobile menu is one |
| `Dropdown` | trigger `popovertarget`, `aria-haspopup`, `aria-controls`, bound `aria-expanded`; `popover="auto"` `role="menu"` and `menuitem` | native: Escape closes and returns focus, click outside closes; ArrowDown/ArrowUp open from the trigger and move between items; Home/End; Tab closes |
| `Tooltip` | `role="tooltip"`, trigger `aria-describedby` | shown on hover and focus (CSS) |
| `Tabs` | `tablist`, `tab`, `tabpanel`, `aria-selected`, `aria-controls`, roving `tabindex` | ArrowLeft/ArrowRight, Home/End |
| `Disclosure`, `Accordion` | `aria-expanded`, `aria-controls`, panel `role="region"` | Accordion: ArrowDown/ArrowUp, Home/End between headers |
| `Draggable`, `DropZone` | `aria-grabbed`, `data-ui-dragging`, `data-ui-over` highlight | `PickUp` and `MoveHere` buttons replace dragging for keyboard and touch |
| `ToastRegion`, `Toast` | polite live region, `role="status"`, labelled Dismiss button | the auto-dismiss countdown pauses on hover and while focus is inside |

### Motion

Transitions are short and only animate what the compositor handles: colour changes on hover and
focus take 150ms (`transition-colors duration-150`); things that appear take 200ms `ease-out`
and reverse when they go (Dialog fade and 2% scale, Drawer slide, Dropdown and Tooltip fade with
a 1px shift, Toast slide-in and fade-out, Tabs panel fade, Disclosure and Accordion height via a
`grid-rows-[0fr]` to `[1fr]` transition). Every transitioned element carries
`motion-reduce:transition-none`, so `prefers-reduced-motion: reduce` shows the end state at once;
colour-only transitions are left on, since nothing moves. The overlay transitions use
`@starting-style` and `transition-behavior: allow-discrete` (Chrome/Edge 117+, Safari 17.4+,
Firefox 129+); older browsers open and close them without animation.

### Browser support

The baseline features (`<dialog>` with `showModal()`, the Popover API and its `toggle` and
`beforetoggle` events) are in every current browser: Chrome/Edge 114+, Firefox 125+, Safari 17+.
Newer features are used where available, with a fallback elsewhere:

| Feature | Used for | Supported | Fallback |
| --- | --- | --- | --- |
| `<dialog closedby="any">` | closing a Dialog on a backdrop click | Chrome/Edge 134+, Firefox 141+; not Safari at the time of writing | a Datastar click-outside handler that only acts when `'closedBy' in HTMLDialogElement.prototype` is false |
| CSS anchor positioning (`anchor-name`, `position-anchor`, `position-area`, `position-try-fallbacks`) | placing the Dropdown menu under its trigger and flipping it above near the bottom of the viewport | Chrome/Edge 125+ (`position-try-fallbacks` 128+), Safari 26+; Firefox not in a release at the time of writing | on `beforetoggle`, the menu's `top`/`right` are set from the trigger's bounding box; it does not follow the trigger on scroll or flip |
| `appearance: base-select` | the customizable `Select` picker | Chrome/Edge 135+ | the native popup |
| Container queries (`@container`, `@2xl:`, `@max-2xl:`, `cqw`) | `Navbar`, `AppShell`, `SplitLayout`, `SettingsLayout`, `PageHeader` and `StackedLayout` switching by their own width | Chrome/Edge 105+, Safari 16+, Firefox 110+ | none needed (baseline) |
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
- `Pagination`, `Breadcrumbs` and `Steps` are named `<nav>` landmarks ("Pagination",
  "Breadcrumb", "Progress"). Two of the same on one page need distinct names: pass a label
  (`Pagination(..., "Results pages")`, `BreadcrumbsLabel`, `StepsLabel`).

See [CHANGELOG.md](CHANGELOG.md) for the API changes that came with the native overlays.
