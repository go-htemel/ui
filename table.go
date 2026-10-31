package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// TableOption configures a Table.
type TableOption func(*tableConfig)

type tableConfig struct {
	striped, dense, sticky bool
	caption                string
}

// Striped shades every other body row.
func Striped() TableOption { return func(c *tableConfig) { c.striped = true } }

// Dense uses tighter cell padding.
func Dense() TableOption { return func(c *tableConfig) { c.dense = true } }

// StickyHeader keeps the header row at the top while the body scrolls: the table is capped at
// 24rem and scrolls inside its frame.
func StickyHeader() TableOption { return func(c *tableConfig) { c.sticky = true } }

// TableCaption adds a caption that screen readers announce and sighted users do not see.
func TableCaption(text string) TableOption { return func(c *tableConfig) { c.caption = text } }

// Table returns a data table inside a frame that scrolls sideways on narrow screens. Build the
// parts with TableHead and TableBody; only <tr> elements fit in a body:
//
//	ui.Table(
//		ui.TableHead(ui.Th(htemel.Text("Name")), ui.SortableTh("Role", "role", "sort", datastar.Get("/people"))),
//		ui.TableBody(html.Tr(ui.Td(htemel.Text("Ada")), ui.Td(htemel.Text("Engineer")))),
//		ui.Striped(),
//	)
//
// Striped, Dense and StickyHeader add their classes to the head and body you pass, so they only
// style rows that were built for this Table call. To update a table from the server, patch the
// whole Table (give the frame an id with .Id) rather than just a body built without the options.
func Table(head *html.TheadElement, body *html.TbodyElement, opts ...TableOption) *html.DivElement {
	var c tableConfig
	for _, opt := range opts {
		if opt != nil {
			opt(&c)
		}
	}
	table := html.Table().Class(Classes("min-w-full border-separate border-spacing-0 text-left"))
	if c.caption != "" {
		table.Children(html.Caption().Class("sr-only").Text(c.caption))
	}
	table.Children(head, body)
	if c.striped {
		body.AddClass(Classes("*:even:bg-neutral-50 dark:*:even:bg-white/5"))
	}
	if c.dense {
		head.AddClass(Classes("*:*:py-2"))
		body.AddClass(Classes("*:*:py-2"))
	}
	frame := "overflow-x-auto"
	if c.sticky {
		head.AddClass(Classes("*:*:sticky *:*:top-0 *:*:z-10"))
		frame = "max-h-96 overflow-auto"
	}
	return html.Div(table).Class(Classes("rounded-surface bg-white ring-1 ring-neutral-950/5 dark:bg-neutral-900 dark:ring-white/10", frame))
}

// TableHead returns the <thead> with one row of header cells (Th and SortableTh).
func TableHead(cells ...*html.ThElement) *html.TheadElement {
	row := html.Tr()
	for _, c := range cells {
		row.Children(c)
	}
	return html.Thead(row)
}

// TableBody returns a <tbody> of rows. Build rows with html.Tr and Td.
func TableBody(rows ...*html.TrElement) *html.TbodyElement {
	b := html.Tbody().Class(Classes("*:last:*:border-b-0"))
	for _, r := range rows {
		b.Children(r)
	}
	return b
}

// Th returns a column header cell (scope="col").
func Th(children ...htemel.FlowContent) *html.ThElement {
	return html.Th(children...).Scope(html.ScopeCol).Class(Classes(thClass))
}

// Td returns a data cell.
func Td(children ...htemel.FlowContent) *html.TdElement {
	return html.Td(children...).Class(Classes("border-b border-neutral-200 px-3 py-3.5 text-sm text-neutral-700",
		"dark:border-white/10 dark:text-neutral-300"))
}

const thClass = "border-b border-neutral-200 bg-neutral-50 px-3 py-3.5 text-left text-xs font-semibold tracking-wide " +
	"text-neutral-600 uppercase dark:border-white/10 dark:bg-neutral-900 dark:text-neutral-300"

// SortDirSignal returns the name of the signal that holds the sort direction ("asc" or "desc")
// for the sort-key signal sortSignal: sortSignal + "Dir".
func SortDirSignal(sortSignal string) string { return sortSignal + "Dir" }

// SortableTh returns a column header whose button sorts by key. A click sets the signal
// sortSignal to key and flips the direction signal SortDirSignal(sortSignal) between "asc" and
// "desc" (a new key starts "asc"), then runs action, usually a request that re-renders the rows:
//
//	ui.SortableTh("Name", "name", "sort", datastar.Get("/people"))
//
// sets $sort and $sortDir and sends them to the server with the request; the handler reads them
// with datastar.ReadSignals and patches the table body. The signals start as "" and "asc" unless
// the page defines them. The header carries aria-sort and an arrow that follows the signals. An
// empty action only updates the signals.
func SortableTh(label, key, sortSignal, action string) *html.ThElement {
	dir := SortDirSignal(sortSignal)
	sortKey, sortDir := "$"+sortSignal, "$"+dir
	k := datastar.Quote(key)
	active := sortKey + " === " + k
	click := "(" + active + ") ? (" + sortDir + " = " + sortDir + " === 'asc' ? 'desc' : 'asc') : (" +
		sortKey + " = " + k + ", " + sortDir + " = 'asc')"
	if action != "" {
		click += "; " + action
	}
	arrow := IconChevronDown().Class("size-4 shrink-0 transition-transform").
		Attrs(
			datastar.Style("opacity", "("+active+") ? 1 : 0.3"),
			datastar.Class("rotate-180", active+" && "+sortDir+" === 'asc'"),
		)
	button := html.Button(html.Span().Text(label), arrow).
		Type(html.ButtonTypeButton).
		Class(Classes("group -mx-1 inline-flex cursor-pointer items-center gap-x-1 rounded-control px-1 py-0.5 font-semibold uppercase",
			"tracking-wide hover:text-neutral-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600",
			"dark:hover:text-white")).
		Attrs(datastar.OnClick(click))
	return html.Th(button).Scope(html.ScopeCol).Class(Classes(thClass)).
		Aria("sort", "none").
		Attrs(
			initially(sortSignal, "''"),
			initially(dir, "'asc'"),
			datastar.Attr("aria-sort", "("+active+") ? ("+sortDir+" === 'asc' ? 'ascending' : 'descending') : 'none'"),
		)
}

// EmptyState returns a centred placeholder for a list or table with nothing in it: an optional
// icon, a title, a description and an optional action such as a Button. Use it on its own, or
// inside a table with EmptyRow. icon, description and action may be nil or empty.
func EmptyState(icon htemel.PhrasingContent, title, description string, action htemel.FlowContent) *html.DivElement {
	d := html.Div().Class(Classes("flex flex-col items-center px-6 py-12 text-center"))
	if icon != nil {
		d.Children(html.Div(icon).Class(Classes("mb-3 text-neutral-400 *:size-8 dark:text-neutral-500")))
	}
	d.Children(html.P().Class(Classes("text-sm font-semibold text-neutral-900 dark:text-white")).Text(title))
	if description != "" {
		d.Children(html.P().Class(Classes("mt-1 max-w-sm text-sm text-neutral-500 dark:text-neutral-400")).Text(description))
	}
	if action != nil {
		d.Children(html.Div(action).Class("mt-4"))
	}
	return d
}

// EmptyRow returns a table row whose single cell spans columns columns and holds state, usually
// an EmptyState. Pass it as the only row of a TableBody that has no data.
func EmptyRow(columns int, state htemel.FlowContent) *html.TrElement {
	return html.Tr(html.Td(state).Colspan(columns).Class(Classes("px-3")))
}
