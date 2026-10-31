package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// Disclosure returns a button that shows and hides children, open while signal is true (the
// signal starts false unless the page defines it).
//
// The button is a real <button> with aria-expanded bound to the signal and aria-controls naming
// the panel, which is a role="region" labelled by the button. The panel follows the button's
// aria-expanded through CSS alone (peer-aria-expanded:), so its state cannot drift from what
// assistive technology is told: it opens by animating its grid row from 0fr to 1fr (200ms, no
// height measurement, no script) and, while closed, is visibility: hidden, which removes its
// content from the tab order and the accessibility tree once it has collapsed. With
// prefers-reduced-motion it opens and closes at once.
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
// the open item, or "" when all are closed (the default unless the page defines the signal). Each
// item is a Disclosure; ArrowDown / ArrowUp move focus between the item buttons (wrapping) and
// Home / End jump to the first / last.
func Accordion(signal string, items ...AccordionPanel) *html.DivElement {
	sig := "$" + signal
	root := html.Div().
		Class(Classes("divide-y divide-neutral-200 border-y border-neutral-200 dark:divide-white/10 dark:border-white/10")).
		Attrs(initially(signal, "''"), datastar.OnKeyDown(accordionKeys))
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
	buttonID := id + "-button"
	button := html.Button(
		html.Span().Text(title),
		IconChevronDown().Class("size-5 shrink-0 text-neutral-400 transition-transform duration-200 ease-out motion-reduce:transition-none").
			Attrs(datastar.Class("rotate-180", open)),
	).Type(html.ButtonTypeButton).
		Id(buttonID).
		Class(Classes("peer flex w-full cursor-pointer items-center justify-between gap-x-4 py-3 text-left text-sm font-semibold",
			"text-neutral-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 dark:focus-visible:outline-primary-500 dark:text-white")).
		Aria("controls", panelID).
		Attrs(expanded(open)...).
		Attrs(datastar.OnClick(toggle))
	// The region is a one-row grid whose row grows from 0fr to 1fr when the button (its peer) is
	// aria-expanded; the inner wrapper clips the content while the row is shorter than it.
	panel := html.Div(
		html.Div(
			html.Div(children...).Class(Classes("pb-4 text-sm text-neutral-600 dark:text-neutral-300")),
		).Class("min-h-0 overflow-hidden"),
	).
		Id(panelID).
		Role("region").
		Aria("labelledby", buttonID).
		Class(Classes("invisible grid grid-rows-[0fr] transition-[grid-template-rows,visibility] duration-200 ease-out motion-reduce:transition-none",
			"peer-aria-expanded:visible peer-aria-expanded:grid-rows-[1fr]"))
	return []htemel.FlowContent{button, panel}
}

// accordionKeys moves focus among the item buttons, which are the first child of each item.
const accordionKeys = "const b = [...el.children].map(d => d.firstElementChild), i = b.indexOf(document.activeElement), k = evt.key; " +
	"if (i < 0) return; " +
	"const n = k === 'ArrowDown' ? (i + 1) % b.length : k === 'ArrowUp' ? (i + b.length - 1) % b.length " +
	": k === 'Home' ? 0 : k === 'End' ? b.length - 1 : -1; " +
	"if (n < 0) return; " +
	"evt.preventDefault(); b[n].focus()"
