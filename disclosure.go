package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// Disclosure returns a button that shows and hides children, open while signal is true (the
// signal starts false unless the page defines it).
func Disclosure(signal, title string, children ...htemel.FlowContent) *html.DivElement {
	return html.Div(
		disclosureParts(domID(signal), title, "$"+signal, Toggle(signal), children)...,
	).Attrs(initially(signal, "false"))
}

// AccordionPanel is one section of an Accordion, built by AccordionItem.
type AccordionPanel struct {
	key, title string
	children   []htemel.FlowContent
}

// AccordionItem returns a section of an Accordion: key identifies it in the signal, title is its
// button text and children are shown while it is open.
func AccordionItem(key, title string, children ...htemel.FlowContent) AccordionPanel {
	return AccordionPanel{key: key, title: title, children: children}
}

// Accordion returns stacked disclosures of which at most one is open: signal holds the key of
// the open item, or "" when all are closed (the default unless the page defines the signal).
func Accordion(signal string, items ...AccordionPanel) *html.DivElement {
	sig := "$" + signal
	root := html.Div().
		Class(Classes("divide-y divide-zinc-200 border-y border-zinc-200 dark:divide-white/10 dark:border-white/10")).
		Attrs(initially(signal, "''"))
	for _, it := range items {
		key := datastar.Quote(it.key)
		open := sig + " === " + key
		toggle := sig + " = " + open + " ? '' : " + key
		root.Children(html.Div(
			disclosureParts(domID(signal)+"-"+domID(it.key), it.title, open, toggle, it.children)...,
		))
	}
	return root
}

// disclosureParts builds the toggle button and the panel shown while open is true.
func disclosureParts(id, title, open, toggle string, children []htemel.FlowContent) []htemel.FlowContent {
	panelID := id + "-panel"
	button := html.Button(
		html.Span().Text(title),
		IconChevronDown().Class("size-5 shrink-0 text-zinc-400 transition-transform duration-200").
			Attrs(datastar.Class("rotate-180", open)),
	).Type(html.ButtonTypeButton).
		Class(Classes("flex w-full cursor-pointer items-center justify-between gap-x-4 py-3 text-left text-sm font-semibold",
			"text-zinc-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-600 dark:text-white")).
		Aria("controls", panelID).
		Attrs(expanded(open)...).
		Attrs(datastar.OnClick(toggle))
	panel := html.Div(children...).
		Id(panelID).
		Class(Classes("pb-4 text-sm text-zinc-600 dark:text-zinc-300")).
		Style(hiddenUntilDatastar).
		Attrs(datastar.Show(open))
	return []htemel.FlowContent{button, panel}
}
