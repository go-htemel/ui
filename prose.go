package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

// Prose returns a <div class="ui-prose"> that styles the long-form content inside it: rendered
// Markdown (markdown.Render wraps its output in Prose), CMS bodies, notes and journal entries.
// That HTML carries no classes of its own, so the typography lives in ui.css as descendant
// rules of .ui-prose rather than as utilities: a heading scale with balanced wrapping,
// paragraphs at text-base/7, underlined primary links, nested bullet and number lists with
// muted markers, blockquotes with a control-border bar, tables in the Table colors, rules, images
// with the surface radius, inline code as a chip, <kbd>, GFM task-list checkboxes drawn like
// Checkbox (read-only), footnotes and <mark>. Code blocks from the code package keep their own
// look. Light and dark colors are --ui-prose-* variables on .ui-prose, so a scope can retheme
// them; the defaults meet WCAG AA on a white or neutral-900 card and on the page background.
//
// Prose sets no width: add max-w-prose (65ch) with Class for a readable measure. ProseSm and
// ProseLg are the same at text-sm and text-lg; every margin is in em, so they scale as a whole.
//
//	ui.Prose(html.H2().Text("Release notes"), html.P().Text("..."))
func Prose(children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).Class("ui-prose")
}

// ProseSm is Prose at a smaller size (14px body text), for side panels, comments and previews.
func ProseSm(children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).Class("ui-prose ui-prose-sm")
}

// ProseLg is Prose at a larger size (18px body text), for articles and reading views.
func ProseLg(children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).Class("ui-prose ui-prose-lg")
}
