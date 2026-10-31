package main

import (
	"regexp"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
	"github.com/go-htemel/ui/code"
)

// CDN assets for the showcase only. A real application builds its CSS with the Tailwind CLI
// (see the README) instead of compiling it in the browser.
const (
	datastarJS     = "https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.4/bundles/datastar.js"
	tailwindJS     = "https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"
	fontAwesomeCSS = "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/css/all.min.css"
)

// pageCSS is the Tailwind source of every page, the browser-build counterpart of an app.css: the
// default ui tokens, the self-hosted Geist fonts (ui.FontCSS), and the alternative token sets of the theming page.
//
// It deliberately contains no @import, not even in a comment (hence the comments are stripped
// from ui.ThemeCSS(), whose header shows an @import example). The browser build prepends
// `@import "tailwindcss";` itself when the text of its <style type="text/tailwindcss"> blocks
// has no "@import" anywhere, and resolves that import internally. Writing the import ourselves
// also works for Tailwind, but Chromium's preload scanner then sees `@import "tailwindcss"` at
// the top of a <style> element, fetches /components/tailwindcss as a stylesheet and logs
// "Refused to apply style ... MIME type" on every page.
var pageCSS = cssComment.ReplaceAllString(ui.ThemeCSS(), "") + "\n" + ui.FontCSS("/assets") + "\n" + detailsCSS + "\n\n" + themeDemoCSS

// detailsCSS animates the "Show source" <details> open and closed with no script: the
// ::details-content box goes from block-size 0 to auto (interpolate-size: allow-keywords makes
// auto animatable) and content-visibility flips discretely at the end of the close. Chromium
// 131+ animates; other browsers open and close instantly.
const detailsCSS = `details.ui-details {
  interpolate-size: allow-keywords;
}
details.ui-details::details-content {
  block-size: 0;
  overflow: clip;
  transition: block-size 200ms ease-out, content-visibility 200ms allow-discrete;
}
details.ui-details[open]::details-content {
  block-size: auto;
}
@media (prefers-reduced-motion: reduce) {
  details.ui-details::details-content {
    transition: none;
  }
}`

// cssComment matches a CSS comment.
var cssComment = regexp.MustCompile(`(?s)/\*.*?\*/`)

// documentHead is the <head> every showcase document shares: CDN assets, the color scheme
// bootstrap, the self-hosted fonts and ui.js.
func documentHead(title string) *html.HeadElement {
	head := html.Head(
		html.Meta().Charset(html.CharsetUtf8),
		html.Meta().Name("viewport").Content("width=device-width, initial-scale=1"),
		html.Title().Text(title+" · go-htemel/ui"),
		// Before the stylesheet and the Tailwind script, so the stored color scheme is applied
		// before the first paint.
		ui.ColorSchemeBootstrap(),
	)
	// The Geist files are served by ui.Assets; preload them so the text does not swap fonts.
	for _, l := range ui.FontPreload("/assets") {
		head.Children(l)
	}
	head.Children(
		html.Script().Src(tailwindJS),
		html.Style().Attr("type", "text/tailwindcss").Text(pageCSS),
		html.Script().Type(html.ScriptTypeModule).Src(datastarJS),
		html.Link().Rel("stylesheet").Href(fontAwesomeCSS),
		ui.ScriptTag("/assets"),
	)
	return head
}

// bareDocument is a showcase document with nothing but body in it: the same head as every page
// (color scheme bootstrap, fonts, Tailwind block, Datastar, ui.js), no sidebar and no toast
// region. bodyClass is added to the base body classes. The framed phone previews of the Layouts
// page and the admin demo embed such documents in an <iframe>, which is a real viewport.
func bareDocument(title string, bodyClass string, body ...htemel.FlowContent) htemel.Node {
	return htemel.Group(
		htemel.Doctype(),
		html.Html(
			documentHead(title),
			html.Body(body...).Class(ui.Classes(ui.BodyClasses(), bodyClass)),
		).Lang("en"),
	)
}

// page wraps content in the showcase document: CDN assets, the sidebar and the toast region.
func page(title, current string, content ...htemel.FlowContent) htemel.Node {
	return htemel.Group(
		htemel.Doctype(),
		html.Html(
			documentHead(title),
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

func sidebar(current string) htemel.FlowContent {
	link := func(href, label string, active bool) *html.LiElement {
		cls := "block rounded-control px-3 py-1.5 text-sm text-neutral-600 transition-colors duration-150 hover:bg-neutral-100 hover:text-neutral-900 dark:text-neutral-400 dark:hover:bg-white/5 dark:hover:text-white"
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
	demo := html.Div(
		html.P().Class("px-3 pb-1 text-xs font-semibold tracking-wide text-neutral-500 uppercase dark:text-neutral-400").Text("Demos"),
		html.Ul(link("/admin", "Admin demo", false)).Class("space-y-0.5"),
	).Class("mt-6")
	return html.Aside(
		html.Nav(
			html.A().Href("/").Class("text-base font-bold text-neutral-900 dark:text-white").Text("go-htemel/ui"),
			html.P().Class("mt-1 text-xs text-neutral-500 dark:text-neutral-400").Text("htemel · Tailwind · Datastar"),
			list,
			demo,
		).Aria("label", "Components"),
		html.Div(ui.ColorSchemeToggle()).Class("mt-6 lg:mt-auto lg:pt-6"),
	).Class("flex shrink-0 flex-col border-b border-neutral-200 px-4 py-6 lg:sticky lg:top-0 lg:h-dvh lg:w-60 lg:border-r lg:border-b-0 dark:border-white/10").
		Aria("label", "Sidebar")
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
			// rotates via group-open: instead, in step with the 200ms height transition of detailsCSS.
			// focus:outline-none drops the focus rectangle a mouse click leaves; keyboard focus
			// still gets the focus-visible outline.
			html.Summary(
				ui.IconChevronRight().Class("size-3.5 shrink-0 transition-transform duration-200 ease-out group-open:rotate-90 motion-reduce:transition-none"),
				html.Span().Text("Show source"),
			).Class("inline-flex cursor-pointer list-none items-center gap-1 rounded-control text-xs font-medium text-neutral-500 transition-colors duration-150 select-none hover:text-neutral-900 focus:outline-none focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-solid focus-visible:outline-primary-600 [&::-webkit-details-marker]:hidden dark:text-neutral-400 dark:hover:text-white dark:focus-visible:outline-primary-500"),
			sourceBlocks(e),
		).Class("ui-details group mt-3"),
	).Class("mt-10")
}

// sourceBlocks is the highlighted source of an example: the Go function, then any snippets. The
// blocks scroll inside a capped height; the copy button is the default of the code package.
func sourceBlocks(e Example) *htemel.GroupElement {
	const capHeight = "mt-2 [--ui-code-max-height:32rem]"
	nodes := []htemel.Node{code.Block("go", e.Source, code.WithLineNumbers()).AddClass(capHeight)}
	for _, s := range e.Snippets {
		nodes = append(nodes, code.Block(s.Lang, s.Src, code.WithTitle(s.Title), code.WithLineNumbers()).AddClass(capHeight))
	}
	return htemel.Group(nodes...)
}

func indexPage() htemel.Node {
	cards := html.Div().Class("mt-8 grid gap-4 sm:grid-cols-2 xl:grid-cols-3")
	for _, g := range groups {
		cards.Children(html.A(
			html.H2().Class("text-sm font-semibold text-neutral-900 dark:text-white").Text(g.Title),
			html.P().Class("mt-1 text-sm text-neutral-600 dark:text-neutral-400").Textf("%d examples · %s", len(g.Examples), g.Intro),
		).Href("/components/" + g.Slug).
			Class(ui.Classes("block rounded-surface bg-white p-5 ring-1 ring-neutral-950/5 transition-shadow duration-150 motion-reduce:transition-none hover:ring-primary-500 dark:bg-neutral-900 dark:ring-white/10 dark:hover:ring-primary-400")))
	}
	return page("Overview", "", html.Div(
		html.H1().Class("text-2xl font-bold tracking-tight text-neutral-900 dark:text-white").Text("go-htemel/ui"),
		html.P().Class("mt-2 max-w-2xl text-sm text-neutral-600 dark:text-neutral-400").
			Text("Components for htemel: typed Go functions that render Tailwind-styled HTML, with Datastar for every interaction. Each page shows the live component with the Go code that produced it underneath."),
		cards,
	))
}

// framed draws node in a fixed-size preview frame: a CSS container (@container) with the page
// background, clipped and ringed like a window, width wide (capped at the example's width) and
// height high ("auto" to fit the content). The layouts switch on their container's width, so a
// 375px frame shows their phone form whatever the viewport is.
func framed(width, height string, node htemel.FlowContent) *html.DivElement {
	style := "width: " + width + "; max-width: 100%"
	if height != "auto" {
		style += "; height: " + height
	}
	return html.Div(node).
		Class(ui.Classes("@container overflow-hidden rounded-surface bg-neutral-50 text-neutral-900 ring-1 ring-neutral-950/10",
			"dark:bg-neutral-950 dark:text-neutral-100 dark:ring-white/10")).
		Style(style)
}
