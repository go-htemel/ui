package main

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

func sortableTable() htemel.FlowContent {
	// Clicking a header sets $sort and $sortDir, then GETs /demo/people with every signal; the
	// server patches the table body and the pager (ids people-body and people-pager).
	get := datastar.Get("/demo/people")
	return html.Div(
		ui.Table(
			ui.TableHead(
				ui.SortableTh("Name", "name", "sort", get),
				ui.SortableTh("Role", "role", "sort", get),
				ui.SortableTh("Team", "team", "sort", get),
			),
			peopleBody(peopleQuery{SortDir: "asc", Page: 1}),
			ui.TableCaption("People, sortable by every column"),
		),
		html.Div(peoplePager(peopleQuery{Page: 1})).Class("mt-4"),
	).Attrs(datastar.Signals(map[string]any{"sort": "", "sortDir": "asc", "page": 1}))
}

func tableVariants() htemel.FlowContent {
	rows := func() *html.TbodyElement {
		body := ui.TableBody()
		for _, p := range demoPeople[:6] {
			body.Children(html.Tr(ui.Td(htemel.Text(p.Name)), ui.Td(htemel.Text(p.Role)), ui.Td(htemel.Text(p.Team))))
		}
		return body
	}
	head := func() *html.TheadElement {
		return ui.TableHead(ui.Th(htemel.Text("Name")), ui.Th(htemel.Text("Role")), ui.Th(htemel.Text("Team")))
	}
	return ui.Stack(ui.Lg,
		ui.Table(head(), rows(), ui.Striped(), ui.Dense()),
		// The header stays put while the body scrolls inside the frame.
		ui.Table(head(), rows(), ui.StickyHeader(), ui.TableCaption("Scrolls under a sticky header")),
	)
}

func emptyStates() htemel.FlowContent {
	return ui.Stack(ui.Lg,
		ui.Table(
			ui.TableHead(ui.Th(htemel.Text("Name")), ui.Th(htemel.Text("Role"))),
			ui.TableBody(ui.EmptyRow(2, ui.EmptyState(ui.IconSearch(),
				"No people found", "Try a different search or clear the filters.",
				ui.Button(ui.Secondary, "Clear filters", ui.WithSize(ui.Sm))))),
		),
		ui.EmptyState(ui.IconPlus(), "No projects yet", "Create your first project to get started.",
			ui.Button(ui.Primary, "New project", ui.WithIcon(ui.IconPlus()))),
	)
}

func paginationLinks() htemel.FlowContent {
	href := func(page int) string { return "#page-" + string(rune('0'+page%10)) }
	// Several on one page: each landmark gets its own name (the default is "Pagination").
	return ui.Stack(ui.Md,
		ui.Pagination(1, 4, href, "Pagination, first page"),
		ui.Pagination(5, 24, href, "Pagination, middle page"),
		ui.Pagination(24, 24, href, "Pagination, last page"),
	)
}

func paginationSignal() htemel.FlowContent {
	// The buttons set $demoPage; there is no request, so the pages beyond the first window only
	// show after the control is re-rendered. See the sortable table for the server round-trip.
	return html.Div(
		ui.PaginationSignal(1, 5, "demoPage", "", "Demo pages"),
		html.P().Class("mt-3 text-sm text-neutral-600 dark:text-neutral-400").
			Text("Page ").Children(html.Span().Attrs(datastar.Text("$demoPage"))),
	)
}

func statsAndLists() htemel.FlowContent {
	return ui.Stack(ui.Lg,
		ui.StatGrid(
			ui.Stat("Revenue", "$45,231", "+12.5%"),
			ui.Stat("Bounce rate", "38.2%", "-2.1%"),
			ui.Stat("Open tickets", "17"),
			ui.Stat("Uptime", "99.98%", "30 days"),
		),
		ui.DescriptionList(
			ui.DLText("Full name", "Ada Lovelace"),
			ui.DLText("Email", "ada@example.com"),
			ui.DLItem{Term: "Status", Detail: ui.Badge(ui.Success, "Active")},
			ui.DLText("About", "Wrote the first published program for Babbage's Analytical Engine."),
		),
	)
}
