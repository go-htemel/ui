package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// SplitLayoutSignal is the signal SplitLayout, SplitSelect and SplitDetailHeader use unless
// SplitSignal names another: "list" or "detail", the pane a narrow container shows.
const SplitLayoutSignal = "split"

// SplitOption configures SplitLayout.
type SplitOption func(*splitConfig)

type splitConfig struct {
	signal    string
	width     Size
	resizable bool
	list      string
	detail    string

	listHeader   htemel.FlowContent
	detailHeader htemel.FlowContent
}

// SplitSignal names the pane signal (default SplitLayoutSignal), for a page with two split
// layouts. Pass the same name to SplitSelect and SplitDetailHeaderSignal.
func SplitSignal(name string) SplitOption { return func(c *splitConfig) { c.signal = name } }

// WithListWidth sets the list pane's width beside the detail: Sm 16rem, Md 20rem (the default),
// Lg 24rem, Xl 28rem.
func WithListWidth(s Size) SplitOption { return func(c *splitConfig) { c.width = s } }

// WithResizable lets the reader drag the list pane wider or narrower (CSS resize: the handle is
// the pane's bottom right corner), between 14rem and 60% of the layout.
func WithResizable() SplitOption { return func(c *splitConfig) { c.resizable = true } }

// WithPaneLabels makes the two panes labelled regions (role="region"), such as "Notes" and
// "Note", so screen-reader users can jump between them.
func WithPaneLabels(list, detail string) SplitOption {
	return func(c *splitConfig) { c.list, c.detail = list, detail }
}

// WithListHeader sets the header of the list pane, usually SplitListHeader. The two headers
// (this and WithDetailHeader) are the cells of one grid row, so they are always the same height
// and their bottom rules line up; a list body and a detail body start on the same line below.
// Pass a header for the other pane too, or the layout draws an empty cell with the same rule so
// the one you pass still aligns.
func WithListHeader(node htemel.FlowContent) SplitOption {
	return func(c *splitConfig) { c.listHeader = node }
}

// WithDetailHeader sets the header of the detail pane, usually SplitDetailHeader (or a wrapper
// of several, shown one at a time). See WithListHeader. Content you patch later (a new title,
// other actions) should carry an id so the patch reaches the header, which sits outside the
// detail pane's own element.
func WithDetailHeader(node htemel.FlowContent) SplitOption {
	return func(c *splitConfig) { c.detailHeader = node }
}

var splitWidths = map[Size]string{
	Xs: "@2xl:w-64",
	Sm: "@2xl:w-64",
	Md: "@2xl:w-80",
	Lg: "@2xl:w-96",
	Xl: "@2xl:w-112",
}

// SplitLayout returns a master/detail layout: a list pane of fixed width (WithListWidth) beside a
// flexible detail pane, for notes, an inbox, a CMS list and its editor. It fills its parent's
// height and each pane scrolls on its own.
//
//	ui.SplitLayout(
//		noteList, // items carry ui.SplitSelect(ui.SplitLayoutSignal, datastar.Get("/notes/3"))
//		noteBody,
//		ui.WithListHeader(ui.SplitListHeader("Notes", search)),
//		ui.WithDetailHeader(ui.SplitDetailHeader("Groceries", deleteButton)),
//		ui.WithResizable(),
//	)
//
// # Shared baseline
//
// The wide form is one two-column grid with two rows, "auto" and "minmax(0,1fr)": row 1 holds
// the list header and the detail header, row 2 the two scrolling bodies. The header cells carry
// the bottom rule and a 3.5rem minimum height, and share their row, so the rules meet in one
// line (and grow together if one header, say a title with a search field, needs more room),
// and both bodies start at the same y. Without headers row 1 is empty and takes no height. A
// header given to only one pane gets an empty cell opposite it (wide form only). The headers
// are not part of the panes, so they stay in place while the bodies scroll.
//
// The root is a CSS container. From the @2xl container width (42rem: a 20rem list and a detail
// at least as wide) both panes show side by side. Below it the layout stacks and shows one pane,
// chosen by the signal (SplitLayoutSignal, "list" or "detail"; it starts at "list" unless the
// page defines it): SplitSelect on a list item shows the detail, and the Back button of
// SplitDetailHeader the list again. The signal only matters below @2xl, so the server can set it
// freely.
//
// # Accessibility
//
// Both panes have tabindex="-1". When a narrow layout switches to the detail, SplitSelect moves
// focus to the detail pane, and Back moves it to the list pane, so keyboard and screen-reader
// users land where the content changed. WithPaneLabels names the panes as regions.
func SplitLayout(list, detail htemel.FlowContent, opts ...SplitOption) *html.DivElement {
	c := splitConfig{signal: SplitLayoutSignal, width: Md}
	for _, o := range opts {
		if o != nil {
			o(&c)
		}
	}
	sig := "$" + c.signal
	id := domID(c.signal)
	listPane := html.Div(list).Id(id + "-list").Tabindex(-1).
		Class(Classes("col-start-1 row-start-2 min-h-0 min-w-0 overflow-y-auto border-neutral-200 outline-none @2xl:border-r dark:border-white/10",
			pick(splitWidths, c.width, Md),
			cond(c.resizable, "@2xl:max-w-[60cqw] @2xl:min-w-56 @2xl:resize-x"))).
		Attrs(datastar.Classes("{'@max-2xl:hidden': " + sig + " === 'detail'}"))
	detailPane := html.Div(detail).Id(id + "-detail").Tabindex(-1).
		Class(Classes("col-start-1 row-start-2 min-h-0 min-w-0 overflow-y-auto outline-none @2xl:col-start-2 @max-2xl:hidden")).
		Attrs(datastar.Classes("{'@max-2xl:hidden': " + sig + " !== 'detail'}"))
	if c.list != "" {
		listPane.Role("region").Aria("label", c.list)
	}
	if c.detail != "" {
		detailPane.Role("region").Aria("label", c.detail)
	}
	grid := html.Div().
		Class(Classes("grid h-full min-h-0 grid-cols-1 grid-rows-[auto_minmax(0,1fr)] @2xl:grid-cols-[auto_minmax(0,1fr)]"))
	if c.listHeader != nil || c.detailHeader != nil {
		// The list header cell adds no width of its own ([contain:inline-size]): the column is as
		// wide as the list body, including a width the reader dragged it to.
		listCell := html.Div().
			Class(Classes(splitHeaderCell, "col-start-1 @2xl:border-r @2xl:[contain:inline-size]")).
			Attrs(datastar.Classes("{'@max-2xl:hidden': " + sig + " === 'detail'}"))
		detailCell := html.Div().
			Class(Classes(splitHeaderCell, "col-start-1 @2xl:col-start-2 @max-2xl:hidden")).
			Attrs(datastar.Classes("{'@max-2xl:hidden': " + sig + " !== 'detail'}"))
		if c.listHeader != nil {
			listCell.Children(c.listHeader)
		} else {
			listCell.Class("@max-2xl:hidden").Aria("hidden", "true")
		}
		if c.detailHeader != nil {
			detailCell.Children(c.detailHeader)
		} else {
			detailCell.Class("@max-2xl:hidden").Aria("hidden", "true")
		}
		grid.Children(listCell, detailCell)
	}
	grid.Children(listPane, detailPane)
	return html.Div(grid).Class(Classes("@container h-full min-h-0")).Attrs(initially(c.signal, "'list'"))
}

// splitHeaderCell is a header cell of the shared grid row: the rule, the background and the
// minimum height live here, and the header inside is centred in whatever height the row has.
const splitHeaderCell = "row-start-1 flex min-h-14 min-w-0 flex-col justify-center border-b border-neutral-200 bg-white dark:border-white/10 dark:bg-neutral-900"

// SplitSelect is the click handler of a list item in a SplitLayout: it shows the detail pane
// (signal = 'detail'), runs action (a Datastar expression that loads the item, such as
// datastar.Get("/notes/3"); may be empty) and, when the layout is narrow, moves focus to the
// detail pane. On an <a href> add your own datastar.Prevent() handler if the link should not
// navigate.
func SplitSelect(signal, action string) htemel.Attribute {
	id := domID(signal)
	expr := "$" + signal + " = 'detail'; "
	if action != "" {
		expr += action + "; "
	}
	expr += "requestAnimationFrame(() => { const l = document.getElementById('" + id + "-list'), d = document.getElementById('" + id + "-detail'); " +
		"if (d && l && !l.checkVisibility()) d.focus({preventScroll: true}) })"
	return datastar.OnClick(expr)
}

// SplitDetailHeader returns the header row of a SplitLayout detail pane, for WithDetailHeader: a
// Back button (shown only while the layout is narrow; it returns to the list and focuses it),
// the title as an <h2> and actions on the right. It is the content of the header cell, which
// brings the rule, the background and the height shared with the list header.
func SplitDetailHeader(title string, actions ...htemel.FlowContent) *html.DivElement {
	return SplitDetailHeaderSignal(SplitLayoutSignal, title, actions...)
}

// SplitDetailHeaderSignal is SplitDetailHeader for a SplitLayout with its own SplitSignal.
func SplitDetailHeaderSignal(signal, title string, actions ...htemel.FlowContent) *html.DivElement {
	id := domID(signal)
	back := Button(Ghost, "Back", WithIcon(IconArrowLeft().SetClass("size-4 shrink-0")), WithSize(Sm)).
		Class("-ml-2 @2xl:hidden").
		Attrs(datastar.OnClick("$" + signal + " = 'list'; requestAnimationFrame(() => document.getElementById('" + id + "-list')?.focus({preventScroll: true}))"))
	row := html.Div(
		back,
		html.H2().Class(Classes("min-w-0 flex-1 truncate text-base font-semibold text-neutral-900 dark:text-white")).Text(title),
	).Class(Classes("flex items-center gap-3 px-4 @xl:px-6"))
	if len(actions) > 0 {
		row.Children(html.Div(actions...).Class(Classes("flex shrink-0 items-center gap-2")))
	}
	return row
}

// SplitListHeader returns the header row of a SplitLayout list pane, for WithListHeader: the
// title as an <h2> and, beside it, controls (a search field, an add button) that fill the rest
// of the row without making it taller than SplitDetailHeader. The cell around it shares its row
// with the detail header, so both rules align.
//
//	ui.SplitListHeader("Notes", ui.FilterSearch("nq", datastar.Get("/notes/list")))
func SplitListHeader(title string, controls ...htemel.FlowContent) *html.DivElement {
	row := html.Div(
		html.H2().Class(Classes("shrink-0 text-base font-semibold text-neutral-900 dark:text-white")).Text(title),
	).Class(Classes("flex items-center gap-3 px-4"))
	if len(controls) > 0 {
		row.Children(html.Div(controls...).Class(Classes("flex min-w-0 flex-1 items-center justify-end gap-2 *:min-w-0")))
	}
	return row
}

// SplitListItem returns a row of a SplitLayout's list: a SelectableItem (primary tint and left
// bar while selectedExpr is true, aria-current="true") whose click is SplitSelect(signal,
// action). The same treatment serves every selectable list, so do not hand-roll it.
//
//	ui.SplitListItem(ui.SplitLayoutSignal, "$noteId === 3", datastar.Get("/notes/3")+"; $noteId = 3", title, date)
func SplitListItem(signal, selectedExpr, action string, children ...htemel.PhrasingContent) *html.ButtonElement {
	return SelectableItem(selectedExpr, children...).Attrs(SplitSelect(signal, action))
}
