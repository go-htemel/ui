package ui_test

import (
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

var (
	_ htemel.FlowContent = ui.Navbar(htemel.Text("x"))
	_ htemel.FlowContent = ui.NavbarSignal("m", htemel.Text("x"))
	_ htemel.FlowContent = ui.Sidebar("s")
	_ htemel.FlowContent = ui.Breadcrumbs()
	_ htemel.FlowContent = ui.Steps(1)
	_ htemel.FlowContent = ui.IconChevronLeft()
	_ htemel.FlowContent = ui.IconChevronRight()
	_ htemel.FlowContent = ui.IconMenu()
)

func init() {
	addGolden(
		golden{"navbar", ui.Navbar(html.A().Href("/").Text("Acme"),
			ui.NavLink("Dashboard", "/", true), ui.NavLink("Team", "/team", false))},
		golden{"sidebar", ui.Sidebar("sidebar",
			ui.SidebarGroup("general", "General", ui.NavLink("Overview", "/", true), ui.NavLink("Activity", "/activity", false)),
			ui.SidebarGroup("billing", "Billing", ui.NavLink("Invoices", "/invoices", false)).Closed(),
		)},
		golden{"breadcrumbs", ui.Breadcrumbs(
			ui.Crumb{Label: "Home", Href: "/"}, ui.Crumb{Label: "Projects", Href: "/projects"}, ui.Crumb{Label: "Apollo"})},
		golden{"breadcrumbs-single", ui.Breadcrumbs(ui.Crumb{Label: "Home", Href: "/"})},
		golden{"steps", ui.Steps(2, "Account", "Billing", "Confirm")},
	)
}

func TestSidebarSections(t *testing.T) {
	// A closed section that holds the active link still starts open.
	s, err := htemel.RenderString(ui.Sidebar("nav",
		ui.SidebarGroup("a", "A", ui.NavLink("x", "/x", true)).Closed(),
		ui.SidebarGroup("b-2", "B", ui.NavLink("y", "/y", false)).Closed(),
	))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, `data-signals:nav.a__ifmissing="true"`) || !strings.Contains(s, `data-signals:nav.b2__ifmissing="false"`) {
		t.Errorf("unexpected section signals: %s", s)
	}
}

func TestBreadcrumbCurrent(t *testing.T) {
	s, _ := htemel.RenderString(ui.Breadcrumbs(ui.Crumb{Label: "Home", Href: "/"}, ui.Crumb{Label: "Here", Href: "/here"}))
	if strings.Count(s, `aria-current="page"`) != 1 || strings.Contains(s, `href="/here"`) {
		t.Errorf("the last crumb must be the current page and not a link: %s", s)
	}
}
