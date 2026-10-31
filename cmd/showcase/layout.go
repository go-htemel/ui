package main

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

// CDN assets for the showcase only. A real application builds its CSS with the Tailwind CLI
// (see the README) instead of compiling it in the browser.
const (
	datastarJS     = "https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.4/bundles/datastar.js"
	tailwindJS     = "https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"
	fontAwesomeCSS = "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/css/all.min.css"
)

// page wraps content in the showcase document: CDN assets, the sidebar and the toast region.
func page(title, current string, content ...htemel.FlowContent) htemel.Node {
	return htemel.Group(
		htemel.Doctype(),
		html.Html(
			html.Head(
				html.Meta().Charset(html.CharsetUtf8),
				html.Meta().Name("viewport").Content("width=device-width, initial-scale=1"),
				html.Title().Text(title+" · go-htemel/ui"),
				html.Script().Src(tailwindJS),
				html.Script().Type(html.ScriptTypeModule).Src(datastarJS),
				html.Link().Rel("stylesheet").Href(fontAwesomeCSS),
			),
			html.Body(
				html.Div(
					sidebar(current),
					html.Main(content...).Class("min-w-0 flex-1 px-4 py-8 sm:px-8 lg:py-10"),
				).Class("flex min-h-dvh w-full flex-col lg:flex-row"),
				ui.ToastRegion(),
			).Class(ui.BodyClasses()),
		).Lang("en"),
	)
}

func sidebar(current string) *html.NavElement {
	link := func(href, label string, active bool) *html.LiElement {
		cls := "block rounded-md px-3 py-1.5 text-sm text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-white/5 dark:hover:text-white"
		a := html.A().Href(href).Text(label)
		if active {
			cls = "block rounded-md px-3 py-1.5 text-sm font-semibold bg-zinc-100 text-zinc-900 dark:bg-white/10 dark:text-white"
			a.Aria("current", "page")
		}
		return html.Li(a.Class(ui.Classes(cls)))
	}
	list := html.Ul(link("/", "Overview", current == "")).Class("mt-4 space-y-0.5")
	for _, g := range groups {
		list.Children(link("/components/"+g.Slug, g.Title, current == g.Slug))
	}
	return html.Nav(
		html.A().Href("/").Class("text-base font-bold text-zinc-900 dark:text-white").Text("go-htemel/ui"),
		html.P().Class("mt-1 text-xs text-zinc-500 dark:text-zinc-400").Text("htemel · Tailwind · Datastar"),
		list,
	).Class("shrink-0 border-b border-zinc-200 px-4 py-6 lg:sticky lg:top-0 lg:h-dvh lg:w-60 lg:border-r lg:border-b-0 dark:border-white/10").
		Aria("label", "Components")
}

// groupPage renders one component group: each example's live output at full width, with its
// source in a collapsible block underneath.
func groupPage(g Group) htemel.Node {
	body := html.Div(
		html.H1().Class("text-2xl font-bold tracking-tight text-zinc-900 dark:text-white").Text(g.Title),
		html.P().Class("mt-2 max-w-2xl text-sm text-zinc-600 dark:text-zinc-400").Text(g.Intro),
	)
	for _, e := range g.Examples {
		body.Children(exampleView(e))
	}
	return page(g.Title, g.Slug, body)
}

func exampleView(e Example) *html.SectionElement {
	return html.Section(
		html.H2().Class("text-sm font-semibold text-zinc-900 dark:text-white").Text(e.Title),
		html.Div(e.Node()).Class(ui.Classes("mt-3 min-w-0 rounded-md bg-white p-6 ring-1 ring-zinc-950/5 dark:bg-zinc-900 dark:ring-white/10")),
		html.Details(
			html.Summary().Class("cursor-pointer select-none text-xs font-medium text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-white").Text("Show source"),
			html.Pre(html.Code().Text(e.Source)).
				Class(ui.Classes("mt-2 max-h-[32rem] min-w-0 overflow-auto rounded-md bg-zinc-900 p-4 text-xs/5 text-zinc-100 ring-1 ring-white/10 dark:bg-black/40")),
		).Class("mt-3"),
	).Class("mt-10")
}

func indexPage() htemel.Node {
	cards := html.Div().Class("mt-8 grid gap-4 sm:grid-cols-2 xl:grid-cols-3")
	for _, g := range groups {
		cards.Children(html.A(
			html.H2().Class("text-sm font-semibold text-zinc-900 dark:text-white").Text(g.Title),
			html.P().Class("mt-1 text-sm text-zinc-600 dark:text-zinc-400").Textf("%d examples · %s", len(g.Examples), g.Intro),
		).Href("/components/" + g.Slug).
			Class(ui.Classes("block rounded-md bg-white p-5 ring-1 ring-zinc-950/5 hover:ring-sky-500 dark:bg-zinc-900 dark:ring-white/10 dark:hover:ring-sky-400")))
	}
	return page("Overview", "", html.Div(
		html.H1().Class("text-2xl font-bold tracking-tight text-zinc-900 dark:text-white").Text("go-htemel/ui"),
		html.P().Class("mt-2 max-w-2xl text-sm text-zinc-600 dark:text-zinc-400").
			Text("Components for htemel: typed Go functions that render Tailwind-styled HTML, with Datastar for every interaction. Each page shows the live component with the Go code that produced it underneath."),
		cards,
	))
}
