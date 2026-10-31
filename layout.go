package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

var gaps = map[Size]string{
	Xs: "gap-1",
	Sm: "gap-2",
	Md: "gap-4",
	Lg: "gap-6",
	Xl: "gap-8",
}

// Stack lays children out in a column (flex-col) with a gap: Xs is 0.25rem, Sm 0.5rem, Md 1rem,
// Lg 1.5rem, Xl 2rem.
func Stack(gap Size, children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).Class(Classes("flex flex-col", pick(gaps, gap, Md)))
}

// Inline lays children out in a wrapping row, vertically centred, with a gap (see Stack).
func Inline(gap Size, children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).Class(Classes("flex flex-wrap items-center", pick(gaps, gap, Md)))
}

// Container centres children in a column at most 80rem wide, with responsive side padding.
func Container(children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).Class(Classes("mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8"))
}

// Divider returns a horizontal rule, <hr>.
func Divider() *html.HrElement {
	return html.Hr().Class(Classes("border-zinc-200 dark:border-white/10"))
}

// DividerLabel returns a horizontal rule with label centred on it, such as "or".
func DividerLabel(label string) *html.DivElement {
	line := func() *html.SpanElement {
		return html.Span().Class(Classes("h-px flex-1 bg-zinc-200 dark:bg-white/10")).Aria("hidden", "true")
	}
	return html.Div(
		line(),
		html.Span().Class(Classes("text-xs font-medium text-zinc-500 dark:text-zinc-400")).Text(label),
		line(),
	).Class(Classes("flex items-center gap-x-3"))
}
