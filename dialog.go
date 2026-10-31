package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// Dialog returns a modal dialog shown while signal is true (the signal starts false unless the
// page defines it). It closes on the X button, on a click outside the panel and on Escape, traps focus and locks page scroll while open (see Accessibility). Open
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
// # Accessibility
//
// The root is role="dialog" aria-modal="true" aria-labelledby the title (it is not a native
// <dialog>, so Datastar's data-show controls it). What the Datastar expressions achieve:
//
//   - Initial focus: on open, focus moves to the first [autofocus] element, else the first
//     focusable control in the body, else the close button.
//   - Focus return: the element focused when the dialog opened (normally the trigger) is focused
//     again on close. Safari does not focus a button on click, so there it falls back to nothing.
//   - Focus trap: Tab and Shift+Tab wrap inside the dialog, and Tab from outside it (focus lost
//     to the page) is pulled back in. The keydown listener is on the window and only acts while
//     the dialog is open.
//   - Scroll lock: while open, body.style.overflow is "hidden" and restored to "" on close.
//     This overwrites an inline overflow style you set on <body>, and with two dialogs open at
//     once the first to close unlocks the page.
//   - Escape closes.
//
// Not achieved: the rest of the page is not made inert or aria-hidden (that needs the inert
// attribute set on every sibling, which a child cannot do), so a screen reader's virtual cursor
// can still leave the dialog; and a dialog patched out of the DOM while open does not restore
// focus or scrolling.
func Dialog(signal, title string, children ...htemel.FlowContent) *html.DivElement {
	titleID := domID(signal) + "-title"
	bodyID := domID(signal) + "-body"
	panel := html.Div(
		html.Div(
			html.H2().Id(titleID).Class(Classes("text-base font-semibold text-zinc-900 dark:text-white")).Text(title),
			closeButton("Close", Close(signal)),
		).Class(Classes("flex items-start justify-between gap-x-4")),
		html.Div(children...).Id(bodyID).Class(Classes("mt-3 text-sm text-zinc-600 dark:text-zinc-300")),
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
		Attrs(
			initially(signal, "false"),
			datastar.Show("$"+signal),
			datastar.On("keydown", dialogKeys(signal), datastar.Window()),
			datastar.Effect(dialogFocus(signal, bodyID)),
		)
}

// dialogKeys closes on Escape and keeps Tab inside the dialog while it is open.
func dialogKeys(signal string) string {
	return "if (!$" + signal + ") return; " +
		"if (evt.key === 'Escape') { " + Close(signal) + "; return } " +
		"if (evt.key !== 'Tab') return; " +
		"const f = " + jsFocusables("el") + ", a = document.activeElement, inside = el.contains(a); " +
		"if (!f.length) { evt.preventDefault(); return } " +
		"const last = f[f.length - 1]; " +
		"if (evt.shiftKey && (a === f[0] || !inside)) { evt.preventDefault(); last.focus() } " +
		"else if (!evt.shiftKey && (a === last || !inside)) { evt.preventDefault(); f[0].focus() }"
}

// dialogFocus locks page scroll and moves focus in on open, and restores both on close. It only
// acts when the signal actually changes, so the initial run does not touch the page.
func dialogFocus(signal, bodyID string) string {
	return "const open = !!$" + signal + "; " +
		"if (open === !!el._open) return; " +
		"el._open = open; " +
		"document.body.style.overflow = open ? 'hidden' : ''; " +
		"if (open) { el._ret = document.activeElement; " +
		"setTimeout(() => { const body = el.querySelector('#" + bodyID + "'); " +
		"const first = (r) => " + jsFocusables("r") + "[0]; " +
		"(el.querySelector('[autofocus]') || (body && first(body)) || first(el))?.focus() }) } " +
		"else if (el._ret) { el._ret.focus(); el._ret = null }"
}

// DialogActions returns the button row at the bottom of a Dialog, aligned right (stacked on
// small screens with the primary action on top).
func DialogActions(children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).Class(Classes("mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end"))
}
