package ui

import (
	"hash/fnv"
	"strconv"
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

// Tooltip returns trigger with a small label that appears above it on hover and keyboard focus.
// It is CSS only (group-hover and group-focus-within), so it needs no signal. When trigger is an
// element (a Button, a link), it gets aria-describedby pointing at the tooltip. Make the trigger
// focusable, as buttons and links are.
//
//	ui.Tooltip(ui.IconButton(ui.Ghost, ui.IconTrash(), "Delete"), "Delete this row")
//
// Tooltip deliberately does not use popover="hint". A hint popover is only shown by script (the
// declarative interestfor attribute is not available across browsers), so it would need
// Datastar handlers for pointer and focus enter and leave, plus the anchor positioning and
// fallback that Dropdown carries; popover="hint" itself is Chromium only (133+), and elsewhere
// the value falls back to a manual popover. What it would add is the top layer (not clipped by
// an overflow container) and Escape to dismiss. Both are limits of this tooltip: it is clipped
// by an ancestor with overflow hidden or auto, and Escape does not hide it (move focus or the
// pointer away instead).
func Tooltip(trigger htemel.PhrasingContent, text string) *html.SpanElement {
	h := fnv.New32a()
	_, _ = h.Write([]byte(text))
	id := "tooltip-" + strconv.FormatUint(uint64(h.Sum32()), 36)
	if el, ok := trigger.(htemel.Element); ok {
		el.SetAttr("aria-describedby", id)
	}
	return html.Span(
		trigger,
		html.Span().Id(id).Role("tooltip").Text(text).
			Class(Classes("pointer-events-none absolute bottom-full left-1/2 z-30 mb-2 -translate-x-1/2 rounded-control bg-neutral-900 px-2 py-1",
				"text-xs font-medium whitespace-nowrap text-white opacity-0 shadow-lg transition-opacity",
				"group-hover:opacity-100 group-focus-within:opacity-100 dark:bg-neutral-700")),
	).Class("group relative inline-flex")
}

// Stat returns a metric card: label, a large value and an optional change. The first delta
// string is shown beside the value, green when it starts with "+" and red when it starts with
// "-" or "−", muted otherwise.
//
//	ui.Stat("Revenue", "$45,231", "+12.5%")
func Stat(label, value string, delta ...string) *html.DivElement {
	figures := html.Div(
		html.P().Class(Classes("text-2xl/8 font-semibold text-neutral-900 dark:text-white")).Text(value),
	).Class("mt-1 flex flex-wrap items-baseline gap-x-2")
	if len(delta) > 0 && delta[0] != "" {
		tone := "text-neutral-500 dark:text-neutral-400"
		switch {
		case strings.HasPrefix(delta[0], "+"):
			tone = "text-success-600 dark:text-success-400"
		case strings.HasPrefix(delta[0], "-"), strings.HasPrefix(delta[0], "−"):
			tone = "text-danger-600 dark:text-danger-400"
		}
		figures.Children(html.P().Class(Classes("shrink-0 text-sm font-medium whitespace-nowrap", tone)).Text(delta[0]))
	}
	return html.Div(
		html.P().Class(Classes("text-sm text-neutral-500 dark:text-neutral-400")).Text(label),
		figures,
	).Class(Classes("min-w-0 rounded-surface bg-white px-4 py-5 ring-1 ring-neutral-950/5 sm:px-6 dark:bg-neutral-900 dark:ring-white/10"))
}

// StatGrid lays Stat cards out in a responsive grid: as many columns as fit at a minimum of
// 11rem each, so cards wrap instead of squeezing their figures when the container is narrow.
func StatGrid(stats ...htemel.FlowContent) *html.DivElement {
	return html.Div(stats...).Class(Classes("grid grid-cols-[repeat(auto-fit,minmax(11rem,1fr))] gap-4"))
}

// DLItem is one term and its details in a DescriptionList. Detail is flow content, so it can hold
// links and badges; DLText builds the common text-only item.
type DLItem struct {
	Term   string
	Detail htemel.FlowContent
}

// DLText returns a DLItem whose details are plain text.
func DLText(term, detail string) DLItem { return DLItem{Term: term, Detail: htemel.Text(detail)} }

// DescriptionList returns a <dl> of terms and details, one pair per row with a rule between rows:
// side by side from the sm breakpoint, stacked below it.
func DescriptionList(pairs ...DLItem) *html.DlElement {
	dl := html.Dl().Class(Classes("grid grid-cols-1 border-b border-neutral-200 sm:grid-cols-3 dark:border-white/10"))
	for _, p := range pairs {
		dd := html.Dd().Class(Classes("pb-4 text-sm text-neutral-900 sm:col-span-2 sm:border-t sm:border-neutral-200 sm:py-4",
			"dark:text-white dark:sm:border-white/10"))
		if p.Detail != nil {
			dd.Children(p.Detail)
		}
		dl.Children(
			html.Dt().Class(Classes("border-t border-neutral-200 pt-4 text-sm font-medium text-neutral-500 sm:col-span-1 sm:py-4",
				"dark:border-white/10 dark:text-neutral-400")).Text(p.Term),
			dd,
		)
	}
	return dl
}

const affixClass = "inline-flex shrink-0 items-center bg-neutral-50 px-3 text-sm text-neutral-500 outline-1 -outline-offset-1 " +
	"outline-neutral-300 dark:bg-white/5 dark:text-neutral-400 dark:outline-white/10"

// InputGroup joins a control with an affix before it, after it or both, such as a currency sign
// or a domain suffix. prefix and suffix may be nil. The control is an Input, Select or any
// FormControl; put the group inside a Field to label it.
//
//	ui.InputGroup(htemel.Text("https://"), ui.Input(html.InputTypeText, "site"), htemel.Text(".com"))
func InputGroup(prefix htemel.PhrasingContent, control FormControl, suffix htemel.PhrasingContent) *html.DivElement {
	g := html.Div().Class(Classes("flex -space-x-px shadow-xs *:relative *:min-w-0 *:rounded-none *:focus-within:z-10",
		"*:first:rounded-l-control *:last:rounded-r-control"))
	if prefix != nil {
		g.Children(html.Span(prefix).Class(Classes(affixClass)))
	}
	g.Children(control)
	if suffix != nil {
		g.Children(html.Span(suffix).Class(Classes(affixClass)))
	}
	return g
}
