package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

const buttonBase = "inline-flex items-center justify-center gap-x-1.5 rounded-control font-semibold whitespace-nowrap " +
	"cursor-pointer transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 " +
	"disabled:cursor-not-allowed disabled:opacity-50"

var buttonVariants = map[Variant]string{
	Primary: "bg-primary-600 text-white shadow-xs hover:bg-primary-700 focus-visible:outline-primary-600 " +
		"dark:bg-primary-700 dark:hover:bg-primary-600 dark:focus-visible:outline-primary-500",
	Secondary: "bg-white text-neutral-900 shadow-xs ring-1 ring-inset ring-neutral-300 hover:bg-neutral-50 focus-visible:outline-neutral-500 " +
		"dark:bg-white/10 dark:text-white dark:shadow-none dark:ring-white/10 dark:hover:bg-white/20",
	Outline: "ring-1 ring-inset ring-primary-600 text-primary-600 hover:bg-primary-50 hover:text-primary-700 focus-visible:outline-primary-600 " +
		"dark:ring-primary-500 dark:focus-visible:outline-primary-500 dark:text-primary-400 dark:hover:text-primary-400 dark:hover:bg-primary-500/10",
	Ghost: "text-neutral-700 hover:bg-neutral-100 focus-visible:outline-neutral-500 " +
		"dark:text-neutral-300 dark:hover:bg-white/10",
	Danger: "bg-danger-600 text-white shadow-xs hover:bg-danger-700 focus-visible:outline-danger-600 " +
		"dark:bg-danger-700 dark:hover:bg-danger-600 dark:focus-visible:outline-danger-500",
	Link: "text-primary-600 underline-offset-4 hover:text-primary-700 hover:underline focus-visible:outline-primary-600 " +
		"dark:text-primary-400 dark:hover:text-primary-300 dark:focus-visible:outline-primary-500",
}

var buttonSizes = map[Size]string{
	Xs: "px-2 py-1 text-xs",
	Sm: "px-2.5 py-1.5 text-sm",
	Md: "px-3 py-2 text-sm",
	Lg: "px-3.5 py-2.5 text-sm",
	Xl: "px-4 py-3 text-base",
}

// linkSizes keep the text size of buttonSizes without the padding, so a Link button lines up
// with surrounding text.
var linkSizes = map[Size]string{
	Xs: "text-xs",
	Sm: "text-sm",
	Md: "text-sm",
	Lg: "text-sm",
	Xl: "text-base",
}

var iconButtonSizes = map[Size]string{
	Xs: "p-1",
	Sm: "p-1.5",
	Md: "p-2",
	Lg: "p-2.5",
	Xl: "p-3",
}

// Button returns a <button type="button"> styled for variant v, with label as its text. Chain
// element methods and Datastar attributes onto the result:
//
//	ui.Button(ui.Primary, "Save", ui.WithLoading("saving")).Attrs(datastar.OnClick(datastar.Post("/save")))
//
// Options: WithSize, WithIcon, WithTrailingIcon, WithLoading. Use .Type(html.ButtonTypeSubmit)
// for a submit button.
func Button(v Variant, label string, opts ...Option) *html.ButtonElement {
	o := collect(opts)
	size := pick(buttonSizes, o.size, Md)
	if v == Link {
		size = pick(linkSizes, o.size, Md)
	}
	b := html.Button().
		Type(html.ButtonTypeButton).
		Class(Classes(buttonBase, pick(buttonVariants, v, Primary), size))
	b.Children(leadingIcon(o.icon, o.loading))
	b.Text(label)
	if o.trailing != nil {
		b.Children(o.trailing)
	}
	return b.Attrs(loadingAttrs(o.loading)...)
}

// IconButton returns a square button holding only icon. label is required: it becomes the
// aria-label and the tooltip (title), because the button has no visible text.
//
// Options: WithSize, WithLoading.
func IconButton(v Variant, icon htemel.PhrasingContent, label string, opts ...Option) *html.ButtonElement {
	o := collect(opts)
	b := html.Button().
		Type(html.ButtonTypeButton).
		Class(Classes(buttonBase, pick(buttonVariants, v, Primary), pick(iconButtonSizes, o.size, Md))).
		Aria("label", label).
		Title(label)
	b.Children(leadingIcon(icon, o.loading))
	return b.Attrs(loadingAttrs(o.loading)...)
}

// ButtonGroup joins buttons into one segmented control (role="group"). Pass Secondary or
// Outline buttons; the group squares their inner corners.
func ButtonGroup(buttons ...*html.ButtonElement) *html.SpanElement {
	g := html.Span().
		Class(Classes("isolate inline-flex -space-x-px rounded-control shadow-xs",
			"*:rounded-none *:shadow-none *:focus-visible:z-10 *:first:rounded-l-control *:last:rounded-r-control")).
		Role("group")
	for _, b := range buttons {
		g.Children(b)
	}
	return g
}

// leadingIcon renders icon, and with a loading signal a spinner that replaces it while the
// signal is true.
func leadingIcon(icon htemel.PhrasingContent, loading string) htemel.PhrasingContent {
	if loading == "" {
		if icon == nil {
			return htemel.Group()
		}
		return icon
	}
	spinner := IconSpinner().Attrs(datastar.Show("$" + loading)).Style(hiddenUntilDatastar)
	if icon == nil {
		return spinner
	}
	return htemel.Group(
		html.Span(icon).Class("contents").Attrs(datastar.Show("!$"+loading)),
		spinner,
	)
}

func loadingAttrs(signal string) []htemel.Attribute {
	if signal == "" {
		return nil
	}
	return []htemel.Attribute{
		datastar.Indicator(signal),
		datastar.Attr("disabled", "$"+signal),
		datastar.Attr("aria-busy", "$"+signal+" ? 'true' : false"),
	}
}
