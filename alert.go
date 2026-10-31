package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// Alert returns an inline message box, <div>, in tone t, with a bold title and optional body
// content. Warning and Danger alerts get role="alert" (announced immediately); the other tones
// get role="status".
func Alert(t Tone, title string, children ...htemel.FlowContent) *html.DivElement {
	return alertBox(t, title, children)
}

// AlertDismissible is an Alert with a close button. It is shown while signal is true; the signal
// starts true unless the page already defines it, and the close button sets it to false.
func AlertDismissible(signal string, t Tone, title string, children ...htemel.FlowContent) *html.DivElement {
	return alertBox(t, title, children).
		Children(closeButton("Dismiss", Close(signal)).AddClass("ml-auto self-start")).
		Attrs(initially(signal, "true"), datastar.Show("$"+signal))
}

func alertBox(t Tone, title string, children []htemel.FlowContent) *html.DivElement {
	role := "status"
	if t == Warning || t == Danger {
		role = "alert"
	}
	content := html.Div(
		html.P().Class("font-semibold").Text(title),
	).Class(Classes("min-w-0 flex-1 space-y-1"))
	if len(children) > 0 {
		content.Children(html.Div(children...).Class("opacity-90"))
	}
	return html.Div(content).
		Class(Classes("flex gap-x-3 rounded-md p-4 text-sm ring-1 ring-inset", pick(toneSoft, t, Neutral))).
		Role(role)
}
