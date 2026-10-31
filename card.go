package ui

import (
	"io"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

// CardSlot is one section of a Card, built by CardHeader, CardBody or CardFooter. It is not flow
// content, so a slot cannot be placed outside a card.
type CardSlot struct {
	el  *html.DivElement
	hdr *cardHeader // set on a CardHeader, so HeadingLevel can rebuild its title
}

// cardHeader is what HeadingLevel needs to rebuild a header with another heading tag.
type cardHeader struct {
	title   string
	actions []htemel.FlowContent
	attrs   []htemel.Attribute
}

// Render implements htemel.Node.
func (s CardSlot) Render(w io.Writer) error { return s.el.Render(w) }

// Attrs adds attributes, such as Datastar attributes, to the section's <div>.
func (s CardSlot) Attrs(attrs ...htemel.Attribute) CardSlot {
	s.el.Attrs(attrs...)
	if s.hdr != nil {
		s.hdr.attrs = append(s.hdr.attrs, attrs...)
	}
	return s
}

// HeadingLevel sets the heading level of a CardHeader's title, 1 to 6 (the default is 3, an
// <h3>): use 2 for a card directly under the page's <h1>, so the outline has no gap. The same
// name and meaning as WithHeadingLevel of PageHeader. It has no effect on CardBody and
// CardFooter, which have no title.
//
//	ui.CardHeader("Publish").HeadingLevel(2)
func (s CardSlot) HeadingLevel(level int) CardSlot {
	if s.hdr == nil {
		return s
	}
	el := cardHeaderElement(s.hdr.title, level, s.hdr.actions)
	el.Attrs(s.hdr.attrs...)
	return CardSlot{el: el, hdr: s.hdr}
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

// CardHeader returns the card's title row: title as an <h3> (HeadingLevel changes the level), and
// actions (buttons, badges) on the right.
func CardHeader(title string, actions ...htemel.FlowContent) CardSlot {
	return CardSlot{el: cardHeaderElement(title, 3, actions), hdr: &cardHeader{title: title, actions: actions}}
}

func cardHeaderElement(title string, level int, actions []htemel.FlowContent) *html.DivElement {
	el := html.Div(
		heading(level, title, "text-base font-semibold text-neutral-900 dark:text-white"),
	).Class(Classes("flex items-center justify-between gap-x-4 px-4 py-4 sm:px-6"))
	if len(actions) > 0 {
		el.Children(html.Div(actions...).Class(Classes("flex shrink-0 items-center gap-x-2")))
	}
	return el
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
