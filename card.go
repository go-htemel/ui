package ui

import (
	"io"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

// CardSlot is one section of a Card, built by CardHeader, CardBody or CardFooter. It is not flow
// content, so a slot cannot be placed outside a card.
type CardSlot struct {
	el *html.DivElement
}

// Render implements htemel.Node.
func (s CardSlot) Render(w io.Writer) error { return s.el.Render(w) }

// Attrs adds attributes, such as Datastar attributes, to the section's <div>.
func (s CardSlot) Attrs(attrs ...htemel.Attribute) CardSlot {
	s.el.Attrs(attrs...)
	return s
}

// Card returns a bordered surface, <div>, whose sections are divided by rules.
//
//	ui.Card(
//		ui.CardHeader("Team", ui.Button(ui.Secondary, "Invite")),
//		ui.CardBody(html.P().Text("3 members")),
//		ui.CardFooter(ui.Button(ui.Primary, "Save")),
//	)
func Card(slots ...CardSlot) *html.DivElement {
	c := html.Div().
		Class(Classes("divide-y divide-neutral-200 overflow-hidden rounded-surface bg-white shadow-sm ring-1 ring-neutral-950/5",
			"dark:divide-white/10 dark:bg-neutral-900 dark:ring-white/10"))
	for _, s := range slots {
		c.Children(htemel.Unchecked(s))
	}
	return c
}

// CardHeader returns the card's title row: title as an <h3>, and actions (buttons, badges) on
// the right.
func CardHeader(title string, actions ...htemel.FlowContent) CardSlot {
	el := html.Div(
		html.H3().Class(Classes("text-base font-semibold text-neutral-900 dark:text-white")).Text(title),
	).Class(Classes("flex items-center justify-between gap-x-4 px-4 py-4 sm:px-6"))
	if len(actions) > 0 {
		el.Children(html.Div(actions...).Class(Classes("flex shrink-0 items-center gap-x-2")))
	}
	return CardSlot{el: el}
}

// CardBody returns the card's main section.
func CardBody(children ...htemel.FlowContent) CardSlot {
	return CardSlot{el: html.Div(children...).Class(Classes("px-4 py-5 text-sm text-neutral-700 sm:p-6 dark:text-neutral-300"))}
}

// CardFooter returns the card's bottom row, with its children aligned right.
func CardFooter(children ...htemel.FlowContent) CardSlot {
	return CardSlot{el: html.Div(children...).
		Class(Classes("flex items-center justify-end gap-x-3 bg-neutral-50 px-4 py-3 sm:px-6 dark:bg-white/5"))}
}
