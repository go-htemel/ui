package ui

import (
	"strconv"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

const pageItemClass = "inline-flex min-w-9 items-center justify-center gap-x-1 rounded-control px-3 py-1.5 text-sm font-medium transition-colors duration-150 " +
	"text-neutral-700 ring-1 ring-inset ring-neutral-300 hover:bg-neutral-50 focus-visible:outline-2 focus-visible:outline-offset-2 " +
	"focus-visible:outline-primary-600 dark:focus-visible:outline-primary-500 aria-[current=page]:bg-primary-600 aria-[current=page]:text-white aria-[current=page]:ring-primary-600 " +
	"aria-[current=page]:hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-50 " +
	"dark:text-neutral-300 dark:ring-white/10 dark:hover:bg-white/5 dark:aria-[current=page]:bg-primary-700 dark:aria-[current=page]:hover:bg-primary-600 " +
	"dark:aria-[current=page]:text-white dark:aria-[current=page]:ring-primary-700"

const pageGapClass = "inline-flex min-w-9 items-center justify-center px-1 py-1.5 text-sm text-neutral-500 dark:text-neutral-400"

// pageWindow returns the page numbers Pagination shows, with 0 for an ellipsis. Up to seven
// pages are all shown. More are cut to seven slots, so the control keeps its width as you page:
// the first and last page always show, with the current page and its neighbours, and a gap of a
// single page is shown as that page instead of an ellipsis.
func pageWindow(page, pages int) []int {
	if pages <= 7 {
		out := make([]int, pages)
		for i := range out {
			out[i] = i + 1
		}
		return out
	}
	switch {
	case page <= 4:
		return []int{1, 2, 3, 4, 5, 0, pages}
	case page >= pages-3:
		return []int{1, 0, pages - 4, pages - 3, pages - 2, pages - 1, pages}
	}
	return []int{1, 0, page - 1, page, page + 1, 0, pages}
}

func clampPage(page, pages int) (int, int) {
	pages = max(pages, 1)
	return min(max(page, 1), pages), pages
}

// Pagination returns page links for a server-rendered list: previous, numbered pages with
// ellipses and next, in a <nav aria-label="Pagination">. page is the current page (1-based) of
// pages; hrefFn returns the URL of a page (nil links to "?page=N"). The current page is marked
// aria-current="page". At either end, previous or next is not a link but a <span
// aria-disabled="true"> whose name is visually hidden text, so it is announced and skipped by Tab.
//
// label, optional, names the navigation landmark instead of "Pagination". Give each a distinct
// label when a page has two ("Results pages", "Comments pages"): landmarks of the same role
// need different names for screen-reader users to tell them apart.
//
//	ui.Pagination(3, 12, func(p int) string { return "/people?page=" + strconv.Itoa(p) })
func Pagination(page, pages int, hrefFn func(page int) string, label ...string) *html.NavElement {
	page, pages = clampPage(page, pages)
	if hrefFn == nil {
		hrefFn = func(n int) string { return "?page=" + strconv.Itoa(n) }
	}
	link := func(name string, icon htemel.PhrasingContent, target int, disabled bool, rel string) htemel.FlowContent {
		if disabled {
			// An <a> without href has no role, and aria-label is not allowed on it; a span with
			// hidden text is announced as plain text, which is what an inactive control is.
			return html.Span(icon, html.Span().Class("sr-only").Text(name)).
				Class(Classes(pageItemClass, "pointer-events-none opacity-50")).Aria("disabled", "true")
		}
		a := html.A(icon).Class(Classes(pageItemClass)).Aria("label", name)
		a.Href(hrefFn(target))
		if rel != "" {
			a.Rel(rel)
		}
		return a
	}
	return paginationNav(page, pages, func(n int) htemel.FlowContent {
		a := html.A().Href(hrefFn(n)).Class(Classes(pageItemClass)).Aria("label", "Page "+strconv.Itoa(n)).Text(strconv.Itoa(n))
		if n == page {
			a.Aria("current", "page")
		}
		return a
	}, link("Previous page", IconChevronLeft(), page-1, page <= 1, "prev"),
		link("Next page", IconChevronRight(), page+1, page >= pages, "next")).
		Aria("label", navLabel(label, "Pagination"))
}

// PaginationSignal returns the same control driven by a Datastar signal instead of links: signal
// holds the current page (declared as page unless the page already defines it), each button sets
// it and then runs action, usually a request that re-renders the list, which receives the new
// page with the other signals. An empty action only changes the signal.
//
//	ui.PaginationSignal(1, 8, "page", datastar.Get("/people"))
//
// The numbered pages are rendered for page, so have the server re-render the control (give it an
// id: PaginationSignal(...).Id("pager")) when the page changes, as it does with the list. label
// names the landmark, as for Pagination.
func PaginationSignal(page, pages int, signal, action string, label ...string) *html.NavElement {
	page, pages = clampPage(page, pages)
	sig := "$" + signal
	run := func(set string) htemel.Attribute {
		if action != "" {
			set += "; " + action
		}
		return datastar.OnClick(set)
	}
	step := func(label string, icon htemel.PhrasingContent, delta int, disabled bool, bound string) htemel.FlowContent {
		return html.Button(icon).Type(html.ButtonTypeButton).Class(Classes(pageItemClass)).Aria("label", label).
			BoolAttr("disabled", disabled).
			Attrs(datastar.Attr("disabled", bound), run(sig+" = "+sig+" "+signedDelta(delta)))
	}
	return paginationNav(page, pages, func(n int) htemel.FlowContent {
		num := strconv.Itoa(n)
		b := html.Button().Type(html.ButtonTypeButton).Class(Classes(pageItemClass)).Aria("label", "Page "+num).Text(num).
			Attrs(run(sig+" = "+num), datastar.Attr("aria-current", sig+" === "+num+" ? 'page' : false"))
		if n == page {
			b.Aria("current", "page")
		}
		return b
	}, step("Previous page", IconChevronLeft(), -1, page <= 1, sig+" <= 1"),
		step("Next page", IconChevronRight(), 1, page >= pages, sig+" >= "+strconv.Itoa(pages))).
		Aria("label", navLabel(label, "Pagination")).
		Attrs(initially(signal, strconv.Itoa(page)))
}

func signedDelta(d int) string {
	if d < 0 {
		return "- " + strconv.Itoa(-d)
	}
	return "+ " + strconv.Itoa(d)
}

// paginationNav lays out previous, the numbered items with ellipses, and next.
func paginationNav(page, pages int, item func(int) htemel.FlowContent, prev, next htemel.FlowContent) *html.NavElement {
	list := html.Ul(html.Li(prev))
	for _, n := range pageWindow(page, pages) {
		if n == 0 {
			list.Children(html.Li(html.Span(IconEllipsis().Class("size-4")).Class(Classes(pageGapClass)).Aria("hidden", "true")))
			continue
		}
		list.Children(html.Li(item(n)))
	}
	list.Children(html.Li(next))
	return html.Nav(list.Class(Classes("flex flex-wrap items-center gap-1")))
}

// navLabel returns the first non-empty label, or def: the optional landmark name of Pagination,
// Breadcrumbs and Steps.
func navLabel(label []string, def string) string {
	for _, l := range label {
		if l != "" {
			return l
		}
	}
	return def
}
