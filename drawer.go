package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

// DrawerSide is the edge of the viewport a Drawer slides in from.
type DrawerSide int

// Drawer sides. DrawerRight is the zero value.
const (
	DrawerRight DrawerSide = iota
	DrawerLeft
	DrawerBottom
)

// drawerSides places the dialog against its edge (the user-agent styles center a modal dialog
// with margin: auto and cap its size; m-0 plus one auto margin and max-*-none undo that) and
// gives the closed state its off-screen translate.
var drawerSides = map[DrawerSide]string{
	DrawerRight: "m-0 ml-auto h-dvh max-h-none w-full border-l translate-x-full open:translate-x-0 starting:open:translate-x-full",
	DrawerLeft:  "m-0 mr-auto h-dvh max-h-none w-full border-r -translate-x-full open:translate-x-0 starting:open:-translate-x-full",
	DrawerBottom: "m-0 mt-auto max-h-[85dvh] w-full max-w-none rounded-t-surface border-t translate-y-full open:translate-y-0 " +
		"starting:open:translate-y-full",
}

// DrawerActionsElement is the button row built by DrawerActions. Drawer pins it below the
// scrolling body.
type DrawerActionsElement struct {
	*html.DivElement
}

// DrawerActions returns the button row at the bottom of a Drawer, aligned right. Pass it among
// the Drawer's children (usually last): the Drawer renders it as a footer that stays in view
// while the body scrolls. Unlike DialogActions it sets no autofocus, so a drawer holding a form
// opens with focus on its first focusable element (the X button, or an autofocus field).
func DrawerActions(children ...htemel.FlowContent) *DrawerActionsElement {
	return &DrawerActionsElement{html.Div(children...).Class(Classes(
		"flex shrink-0 justify-end gap-3 border-t border-neutral-200 px-6 py-4 dark:border-white/10"))}
}

// Drawer returns a slide-over panel: a modal native <dialog> pinned to one edge of the viewport,
// shown with showModal() while signal is true (the signal starts false unless the page defines
// it). It works like Dialog: open it with a trigger carrying ui.Opens(signal); the X button,
// Escape and a click on the backdrop close it, and every way it closes sets the signal back to
// false. The header holds title and the X button, the body scrolls on its own, and a
// DrawerActions child becomes a footer that stays in view:
//
//	ui.Button(ui.Secondary, "Filters").Attrs(ui.Opens("filters"))
//	ui.Drawer("filters", ui.DrawerRight, "Filters",
//		ui.Field("q", "Search", ui.Input(html.InputTypeSearch, "q")),
//		ui.DrawerActions(
//			ui.Button(ui.Secondary, "Cancel").Attrs(datastar.OnClick(ui.Close("filters"))),
//			ui.Button(ui.Primary, "Apply").Attrs(datastar.OnClick(datastar.Get("/search"))),
//		),
//	)
//
// DrawerRight and DrawerLeft are full height (max-w-md and max-w-xs wide by default, see DrawerSized; the full width on small
// screens); DrawerBottom is a sheet up to 85% of the viewport height. The drawer slides in from
// its edge and the backdrop fades, in 200ms, and both reverse on close; with
// prefers-reduced-motion it appears without the slide.
//
// # Accessibility
//
// The same as Dialog, since it is the same native modal: the page behind is inert, Tab stays
// inside, initial focus goes to the first autofocus element or else the first focusable one (the
// X button), focus returns to the trigger on close, and <dialog aria-labelledby> names it after
// the title. The scroll lock note of Dialog applies too.
func Drawer(signal string, side DrawerSide, title string, children ...htemel.FlowContent) *html.DialogElement {
	size := Md
	if side == DrawerLeft {
		size = Xs
	}
	return DrawerSized(signal, side, size, title, children...)
}

// drawerWidths is the panel width of a left or right Drawer by Size; the panel is always the
// full width on screens narrower than that. DrawerBottom ignores it.
var drawerWidths = map[Size]string{
	Xs: "max-w-xs",
	Sm: "max-w-sm",
	Md: "max-w-md",
	Lg: "max-w-lg",
	Xl: "max-w-2xl",
}

// DrawerSized is Drawer with an explicit panel width for DrawerLeft and DrawerRight: Xs (20rem)
// to Xl (42rem). Drawer uses Md on the right and Xs on the left. Chaining a max-w-* class onto a
// Drawer instead is unreliable, because two max-w utilities would compete.
//
//	ui.DrawerSized("editor", ui.DrawerRight, ui.Xl, "Edit article", form)
func DrawerSized(signal string, side DrawerSide, size Size, title string, children ...htemel.FlowContent) *html.DialogElement {
	titleID := domID(signal) + "-title"
	body := html.Div().Id(domID(signal) + "-body").
		Class(Classes("min-h-0 flex-1 overflow-y-auto px-6 py-4 text-sm text-neutral-600 dark:text-neutral-300"))
	var footer []htemel.FlowContent
	for _, c := range children {
		if a, ok := c.(*DrawerActionsElement); ok {
			footer = append(footer, a)
			continue
		}
		body.Children(c)
	}
	panel := html.Div(
		dialogHeader(signal, titleID, title,
			"flex shrink-0 items-center justify-between gap-x-4 border-b border-neutral-200 px-6 py-4 dark:border-white/10"),
		body,
	).Class(Classes("flex h-full max-h-[inherit] flex-col"))
	panel.Children(footer...)
	return modal(signal, titleID, panel,
		Classes("bg-white p-0 shadow-xl border-neutral-200 dark:border-white/10 dark:bg-neutral-900",
			pick(drawerSides, side, DrawerRight), drawerWidth(side, size), overlayTransition, backdropMotion))
}

func drawerWidth(side DrawerSide, size Size) string {
	if side == DrawerBottom {
		return "max-w-none"
	}
	return pick(drawerWidths, size, Md)
}
