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
	interCSS       = "https://fonts.googleapis.com/css2?family=Inter:wght@400..700&display=swap"
)

// fontTheme points Tailwind's font-sans at Inter. The system-ui fallback is Segoe UI on Windows,
// whose vertical metrics sit text about a pixel above centre in buttons and badges; Inter's are
// balanced, and it is the face the Tailwind design language assumes.
const fontTheme = `@theme { --font-sans: "Inter", ui-sans-serif, system-ui, sans-serif; }`

// pageCSS is the Tailwind source of every page, the browser-build counterpart of an app.css: the
// default ui tokens, the Inter font, and the alternative token sets of the theming page.
var pageCSS = ui.ThemeCSS() + "\n" + fontTheme + "\n\n" + themeDemoCSS

// page wraps content in the showcase document: CDN assets, the sidebar and the toast region.
func page(title, current string, content ...htemel.FlowContent) htemel.Node {
	return htemel.Group(
		htemel.Doctype(),
		html.Html(
			html.Head(
				html.Meta().Charset(html.CharsetUtf8),
				html.Meta().Name("viewport").Content("width=device-width, initial-scale=1"),
				html.Title().Text(title+" · go-htemel/ui"),
				html.Link().Rel("preconnect").Href("https://fonts.googleapis.com"),
				html.Link().Rel("stylesheet").Href(interCSS),
				html.Script().Src(tailwindJS),
				html.Style().Attr("type", "text/tailwindcss").Text(pageCSS),
				html.Script().Type(html.ScriptTypeModule).Src(datastarJS),
				html.Link().Rel("stylesheet").Href(fontAwesomeCSS),
				ui.ScriptTag("/assets"),
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
		cls := "block rounded-control px-3 py-1.5 text-sm text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900 dark:text-neutral-400 dark:hover:bg-white/5 dark:hover:text-white"
		a := html.A().Href(href).Text(label)
		if active {
			cls = "block rounded-control px-3 py-1.5 text-sm font-semibold bg-neutral-100 text-neutral-900 dark:bg-white/10 dark:text-white"
			a.Aria("current", "page")
		}
		return html.Li(a.Class(ui.Classes(cls)))
	}
	list := html.Ul(link("/", "Overview", current == "")).Class("mt-4 space-y-0.5")
	for _, g := range groups {
		list.Children(link("/components/"+g.Slug, g.Title, current == g.Slug))
	}
	return html.Nav(
		html.A().Href("/").Class("text-base font-bold text-neutral-900 dark:text-white").Text("go-htemel/ui"),
		html.P().Class("mt-1 text-xs text-neutral-500 dark:text-neutral-400").Text("htemel · Tailwind · Datastar"),
		list,
	).Class("shrink-0 border-b border-neutral-200 px-4 py-6 lg:sticky lg:top-0 lg:h-dvh lg:w-60 lg:border-r lg:border-b-0 dark:border-white/10").
		Aria("label", "Components")
}

// groupPage renders one component group: each example's live output at full width, with its
// source in a collapsible block underneath.
func groupPage(g Group) htemel.Node {
	body := html.Div(
		html.H1().Class("text-2xl font-bold tracking-tight text-neutral-900 dark:text-white").Text(g.Title),
		html.P().Class("mt-2 max-w-2xl text-sm text-neutral-600 dark:text-neutral-400").Text(g.Intro),
	)
	for _, e := range g.Examples {
		body.Children(exampleView(e))
	}
	return page(g.Title, g.Slug, body)
}

func exampleView(e Example) *html.SectionElement {
	return html.Section(
		html.H2().Class("text-sm font-semibold text-neutral-900 dark:text-white").Text(e.Title),
		html.Div(e.Node()).Class(ui.Classes("mt-3 min-w-0 rounded-surface bg-white p-6 ring-1 ring-neutral-950/5 dark:bg-neutral-900 dark:ring-white/10")),
		html.Details(
			// list-none and the -webkit- rule hide the browser's disclosure triangle; the chevron icon
			// rotates via group-open: instead.
			html.Summary(
				ui.IconChevronRight().Class("size-3.5 shrink-0 transition-transform duration-150 group-open:rotate-90"),
				html.Span().Text("Show source"),
			).Class("inline-flex cursor-pointer list-none items-center gap-1 text-xs font-medium text-neutral-500 select-none hover:text-neutral-900 [&::-webkit-details-marker]:hidden dark:text-neutral-400 dark:hover:text-white"),
			html.Pre(html.Code().Text(e.Source)).
				Class(ui.Classes("mt-2 max-h-[32rem] min-w-0 overflow-auto rounded-surface bg-neutral-900 p-4 text-xs/5 text-neutral-100 ring-1 ring-white/10 dark:bg-black/40")),
		).Class("group mt-3"),
	).Class("mt-10")
}

func indexPage() htemel.Node {
	cards := html.Div().Class("mt-8 grid gap-4 sm:grid-cols-2 xl:grid-cols-3")
	for _, g := range groups {
		cards.Children(html.A(
			html.H2().Class("text-sm font-semibold text-neutral-900 dark:text-white").Text(g.Title),
			html.P().Class("mt-1 text-sm text-neutral-600 dark:text-neutral-400").Textf("%d examples · %s", len(g.Examples), g.Intro),
		).Href("/components/" + g.Slug).
			Class(ui.Classes("block rounded-surface bg-white p-5 ring-1 ring-neutral-950/5 hover:ring-primary-500 dark:bg-neutral-900 dark:ring-white/10 dark:hover:ring-primary-400")))
	}
	return page("Overview", "", html.Div(
		html.H1().Class("text-2xl font-bold tracking-tight text-neutral-900 dark:text-white").Text("go-htemel/ui"),
		html.P().Class("mt-2 max-w-2xl text-sm text-neutral-600 dark:text-neutral-400").
			Text("Components for htemel: typed Go functions that render Tailwind-styled HTML, with Datastar for every interaction. Each page shows the live component with the Go code that produced it underneath."),
		cards,
	))
}
