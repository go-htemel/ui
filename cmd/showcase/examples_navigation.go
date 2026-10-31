package main

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

func navbar() htemel.FlowContent {
	// Narrow the window: below md the links collapse behind the menu button.
	return ui.NavbarSignal("demoNav",
		html.A().Href("#").Class("text-base font-bold text-neutral-900 dark:text-white").Text("Acme"),
		ui.NavLink("Dashboard", "#dashboard", true),
		ui.NavLink("Team", "#team", false),
		ui.NavLink("Projects", "#projects", false),
		ui.NavLink("Reports", "#reports", false),
	)
}

func sidebarNav() htemel.FlowContent {
	return html.Div(
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
	).Class("max-w-xs")
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
