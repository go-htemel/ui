package main

import (
	"net/url"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

// demoNavLinks are the navbar demo's links. Clicking one stays on the page: it sets $navActive,
// which moves aria-current (and so the highlight), closes the mobile drawer, and asks the
// server for a toast.
func demoNavLinks(menu string) []ui.NavItem {
	var items []ui.NavItem
	for _, l := range []struct{ key, label string }{
		{"dashboard", "Dashboard"}, {"team", "Team"}, {"projects", "Projects"}, {"reports", "Reports"},
	} {
		toast := datastar.Post("/demo/toast?tone=info&message=&title=" + url.QueryEscape("Navigated to "+l.label))
		items = append(items, ui.NavLink(l.label, "#"+l.key, l.key == "dashboard").Attrs(
			datastar.OnClick("$navActive = '"+l.key+"'; $"+menu+" = false; "+toast, datastar.Prevent()),
			datastar.Attr("aria-current", "$navActive === '"+l.key+"' ? 'page' : false"),
		))
	}
	return items
}

func navbar() htemel.FlowContent {
	// The navbar switches on its own width (a container query), not the viewport's: the 380px
	// frame shows the menu button and drawer, the full-width one the inline links. Both share
	// $navActive.
	brand := func() *html.AElement {
		return html.A().Href("#").Class("text-base font-bold text-neutral-900 dark:text-white").Text("Acme")
	}
	frame := "overflow-hidden rounded-surface ring-1 ring-neutral-950/10 dark:ring-white/10"
	caption := "mb-2 text-xs font-medium text-neutral-500 dark:text-neutral-400"
	return html.Div(
		html.Div(
			html.P().Class(caption).Text("380px container"),
			html.Div(ui.NavbarSignal("demoNav", brand(), demoNavLinks("demoNav")...).Aria("label", "Demo, narrow")).
				Class(frame+" w-full max-w-[380px]"),
		),
		html.Div(
			html.P().Class(caption).Text("Full width"),
			html.Div(ui.NavbarSignal("demoNavWide", brand(), demoNavLinks("demoNavWide")...).Aria("label", "Demo, wide")).
				Class(frame),
		),
	).Class("space-y-6").Attrs(datastar.Signal("navActive", datastar.Quote("dashboard"), datastar.IfMissing()))
}

func sidebarNav() htemel.FlowContent {
	// Framed like an app shell, so the active link (aria-current, a tint with a primary label)
	// stands out against the page beside it.
	return html.Div(
		html.Div(
			ui.Sidebar("demoSidebar",
				ui.SidebarGroup("general", "General",
					ui.NavLink("Overview", "#overview", true),
					ui.NavLink("Activity", "#activity", false),
				),
				ui.SidebarGroup("billing", "Billing",
					ui.NavLink("Invoices", "#invoices", false),
					ui.NavLink("Payment methods", "#payment", false),
				),
				ui.SidebarGroup("danger", "Advanced",
					ui.NavLink("Export data", "#export", false),
				).Closed(),
			),
		).Class("w-56 shrink-0 border-r border-neutral-200 bg-white p-3 dark:border-white/10 dark:bg-neutral-900"),
		html.Div(
			html.P().Class("text-sm font-semibold text-neutral-900 dark:text-white").Text("Overview"),
			html.P().Class("mt-1 text-sm text-neutral-600 dark:text-neutral-400").Text("The page the active link points at."),
		).Class("min-w-0 flex-1 p-6"),
	).Class("flex overflow-hidden rounded-surface bg-neutral-50 ring-1 ring-neutral-950/10 dark:bg-neutral-950 dark:ring-white/10")
}

func breadcrumbs() htemel.FlowContent {
	// Two trails on one page need distinct landmark names: BreadcrumbsLabel names the second.
	return ui.Stack(ui.Md,
		ui.Breadcrumbs(
			ui.Crumb{Label: "Home", Href: "#"},
			ui.Crumb{Label: "Projects", Href: "#projects"},
			ui.Crumb{Label: "Apollo"},
		),
		ui.BreadcrumbsLabel("Settings breadcrumb", ui.Crumb{Label: "Home", Href: "#"}, ui.Crumb{Label: "Settings"}),
	)
}

func steps() htemel.FlowContent {
	// StepsLabel names each landmark; Steps alone would give three navs named "Progress".
	return ui.Stack(ui.Lg,
		ui.StepsLabel("Signup progress, step 1", 1, "Account", "Billing", "Confirm"),
		ui.StepsLabel("Signup progress, step 2", 2, "Account", "Billing", "Confirm"),
		ui.StepsLabel("Signup progress, done", 4, "Account", "Billing", "Confirm"),
	)
}
