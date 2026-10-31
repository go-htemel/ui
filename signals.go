package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// Expressions and attributes for the boolean signals that open and close Dialog, Dropdown,
// Disclosure and AlertDismissible. signal is the camelCase signal name, without the $. Dialog
// and Dropdown are native overlays (<dialog> and the Popover API) whose signal mirrors their
// state in both directions, so these expressions open and close them as well.

// Open returns the expression that sets signal to true.
func Open(signal string) string { return "$" + signal + " = true" }

// Close returns the expression that sets signal to false.
func Close(signal string) string { return "$" + signal + " = false" }

// Toggle returns the expression that flips signal.
func Toggle(signal string) string { return "$" + signal + " = !$" + signal }

// Opens is the click handler for a trigger that opens signal from outside the component, such as
// the button that opens a Dialog. The click does not propagate, so it does not also reach the
// Dialog's click-outside fallback (used in browsers without closedby) and close it again.
//
//	ui.Button(ui.Primary, "Edit").Attrs(ui.Opens("editOpen"))
func Opens(signal string) htemel.Attribute {
	return datastar.OnClick(Open(signal), datastar.Stop())
}

// initially declares signal with value expr unless the page already has it.
func initially(signal, expr string) htemel.Attribute {
	return datastar.Signal(signal, expr, datastar.IfMissing())
}

// hiddenUntilDatastar is the inline style of elements that data-show reveals: they stay hidden
// until Datastar evaluates the expression, instead of flashing on load.
const hiddenUntilDatastar = "display: none"

// closeButton is the small X button of Alert, in the tone's color; expr runs on click.
func closeButton(label, expr string) *html.ButtonElement {
	return html.Button(IconX().Class("size-4 shrink-0")).
		Type(html.ButtonTypeButton).
		Class(Classes("-m-1 inline-flex shrink-0 cursor-pointer items-center justify-center rounded-control p-1",
			"transition-colors duration-150 hover:bg-black/5 focus-visible:outline-2 focus-visible:outline-primary-600 dark:hover:bg-white/10 dark:focus-visible:outline-primary-500")).
		Aria("label", label).
		Attrs(datastar.OnClick(expr))
}

// ghostClose is the X button of Dialog, Drawer and Toast: a Ghost IconButton with no ring, so the
// only chrome is the hover background. (Alert uses closeButton, which takes the tone's color.)
func ghostClose(label, expr string) *html.ButtonElement {
	return IconButton(Ghost, IconX().Class("size-4 shrink-0"), label, WithSize(Xs), withClass("-m-1")).
		Attrs(datastar.OnClick(expr))
}

// expanded binds aria-expanded to a boolean expression, with "false" as the server-rendered value.
func expanded(expr string) []htemel.Attribute {
	return []htemel.Attribute{
		htemel.Attr("aria-expanded", "false"),
		datastar.Attr("aria-expanded", "("+expr+") ? 'true' : 'false'"),
	}
}

// domID turns a signal name or key into an id fragment: letters and digits are kept, anything
// else becomes a hyphen.
func domID(s string) string {
	b := []byte(s)
	for i, c := range b {
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z', '0' <= c && c <= '9':
		default:
			b[i] = '-'
		}
	}
	return string(b)
}
