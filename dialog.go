package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// Dialog returns a modal dialog shown while signal is true (the signal starts false unless the
// page defines it). It closes on the X button, on a click outside the panel and on Escape. Open
// it with a trigger carrying ui.Opens(signal), and end children with DialogActions for buttons:
//
//	ui.Button(ui.Danger, "Delete").Attrs(ui.Opens("confirmDelete"))
//	ui.Dialog("confirmDelete", "Delete project?",
//		html.P().Text("This cannot be undone."),
//		ui.DialogActions(
//			ui.Button(ui.Secondary, "Cancel").Attrs(datastar.OnClick(ui.Close("confirmDelete"))),
//			ui.Button(ui.Danger, "Delete").Attrs(datastar.OnClick(datastar.Delete("/projects/1"))),
//		),
//	)
//
// The panel is a role="dialog" aria-modal element, not a native <dialog>, so Datastar's data-show
// controls it. Focus is not trapped inside the panel and the page behind it still scrolls; move
// focus into it yourself (for example with autofocus on its first control) where that matters.
func Dialog(signal, title string, children ...htemel.FlowContent) *html.DivElement {
	titleID := domID(signal) + "-title"
	panel := html.Div(
		html.Div(
			html.H2().Id(titleID).Class(Classes("text-base font-semibold text-zinc-900 dark:text-white")).Text(title),
			closeButton("Close", Close(signal)),
		).Class(Classes("flex items-start justify-between gap-x-4")),
		html.Div(children...).Class(Classes("mt-3 text-sm text-zinc-600 dark:text-zinc-300")),
	).Class(Classes("relative w-full max-w-lg rounded-md bg-white p-6 shadow-xl ring-1 ring-zinc-950/5",
		"dark:bg-zinc-900 dark:ring-white/10")).
		Attrs(datastar.OnClick(Close(signal), datastar.Outside()))
	return html.Div(
		html.Div().Class(Classes("fixed inset-0 bg-zinc-950/50 dark:bg-zinc-950/70")).Aria("hidden", "true"),
		panel,
	).Class(Classes("fixed inset-0 z-50 flex items-center justify-center p-4")).
		Role("dialog").
		Aria("modal", "true").
		Aria("labelledby", titleID).
		Style(hiddenUntilDatastar).
		Attrs(initially(signal, "false"), datastar.Show("$"+signal), closeOnEscape(signal))
}

// DialogActions returns the button row at the bottom of a Dialog, aligned right (stacked on
// small screens with the primary action on top).
func DialogActions(children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).Class(Classes("mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end"))
}
