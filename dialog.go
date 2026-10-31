package ui

import (
	"regexp"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// Dialog returns a modal native <dialog>, shown with showModal() while signal is true (the
// signal starts false unless the page defines it). It closes on the X button, on Escape and on a
// click on the backdrop, and every way it closes sets the signal back to false. Open it with a
// trigger carrying ui.Opens(signal), and end children with DialogActions for buttons:
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
// A <form method="dialog"> inside the dialog also closes it, and the signal follows.
//
// # Accessibility
//
// The root is a <dialog aria-labelledby> pointing at the title, opened as a modal, so the
// browser provides what the WAI-ARIA dialog pattern asks for:
//
//   - The rest of the page is inert while the dialog is open: it cannot be clicked or focused
//     and is hidden from assistive technology, so a screen reader's virtual cursor stays inside.
//   - Tab and Shift+Tab therefore stay inside the dialog (and the browser's own UI).
//   - Initial focus goes to the first element with autofocus, else the first focusable element
//     (the X button). DialogActions puts autofocus on its first button unless one of its children
//     already has it; an autofocus element earlier in the body wins over it.
//   - On close, focus returns to the element that was focused when the dialog opened, normally
//     the trigger. (Safari does not focus a button on click, so there it may return to the page.)
//   - Escape closes it (a close request); so does a click on the backdrop: natively through
//     closedby="any" where supported (Chrome and Edge 134+, Firefox 141+), and through a
//     Datastar click-outside handler that only acts in browsers without closedby.
//   - The dialog is in the top layer, above every z-index, and its backdrop is the ::backdrop
//     pseudo-element.
//
// Not covered: a modal dialog does not stop the page behind it from scrolling (wheel and touch
// over the backdrop scroll the document). Add `html:has(dialog:modal) { overflow: hidden }` to
// your CSS, or the class has-[dialog:modal]:overflow-hidden on <html>, to lock it. A dialog
// patched by the server while open loses its open state if the patch replaces the element.
func Dialog(signal, title string, children ...htemel.FlowContent) *html.DialogElement {
	titleID := domID(signal) + "-title"
	panel := html.Div(
		html.Div(
			html.H2().Id(titleID).Class(Classes("text-base font-semibold text-neutral-900 dark:text-white")).Text(title),
			closeButton("Close", Close(signal)).AddClass(Classes("text-neutral-500 dark:text-neutral-400")),
		).Class(Classes("flex items-start justify-between gap-x-4")),
		html.Div(children...).Id(domID(signal)+"-body").Class(Classes("mt-3 text-sm text-neutral-600 dark:text-neutral-300")),
	).Class("p-6").
		// The panel fills the dialog, so a click outside it is a click on the backdrop.
		Attrs(datastar.OnClick(dialogBackdropClick(signal), datastar.Outside()))
	return html.Dialog(panel).
		Id(domID(signal)+"-dialog").
		Aria("labelledby", titleID).
		Closedby(html.ClosedbyAny).
		Class(Classes("m-auto w-[calc(100%-2rem)] max-w-lg rounded-surface bg-white p-0 shadow-xl ring-1 ring-neutral-950/5",
			"backdrop:bg-neutral-950/50 dark:bg-neutral-900 dark:ring-white/10 dark:backdrop:bg-neutral-950/70")).
		Attrs(
			initially(signal, "false"),
			datastar.Effect(dialogEffect(signal)),
			datastar.On("close", Close(signal)),
		)
}

// dialogEffect opens the dialog as a modal while the signal is true and closes it when the
// signal turns false, without calling showModal or close twice.
func dialogEffect(signal string) string {
	return "$" + signal + " ? (el.open || el.showModal()) : (el.open && el.close())"
}

// dialogBackdropClick closes the dialog on a click outside its panel, only in browsers without
// closedby (which handle light dismiss themselves).
func dialogBackdropClick(signal string) string {
	return "if ('closedBy' in HTMLDialogElement.prototype) return; " + Close(signal)
}

// autofocusAttr matches a start tag carrying the autofocus attribute in rendered HTML.
var autofocusAttr = regexp.MustCompile(`<[^>]*\sautofocus[\s=>]`)

// DialogActions returns the button row at the bottom of a Dialog, aligned right (stacked on
// small screens with the primary action on top). Unless one of children already has autofocus,
// the first child gets it when it is a button or link, so a Dialog opens with focus on its first
// action (put the safe choice, such as Cancel, first).
func DialogActions(children ...htemel.FlowContent) *html.DivElement {
	if len(children) > 0 && !anyAutofocus(children) {
		switch first := children[0].(type) {
		case *html.ButtonElement:
			first.Autofocus(true)
		case *html.AElement:
			first.Autofocus(true)
		}
	}
	return html.Div(children...).Class(Classes("mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end"))
}

func anyAutofocus(nodes []htemel.FlowContent) bool {
	for _, n := range nodes {
		if s, err := htemel.RenderString(n); err == nil && autofocusAttr.MatchString(s) {
			return true
		}
	}
	return false
}
