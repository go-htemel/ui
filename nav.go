package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// NavItem is one link of a Navbar or a Sidebar section, built by NavLink.
type NavItem struct {
	label, href string
	active      bool
}

// NavLink returns a navigation link. active marks the current page: the link is highlighted and
// gets aria-current="page".
func NavLink(label, href string, active bool) NavItem {
	return NavItem{label: label, href: href, active: active}
}

const (
	navLinkIdle   = "font-medium text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900 dark:text-neutral-400 dark:hover:bg-white/5 dark:hover:text-white"
	navLinkActive = "bg-neutral-100 font-semibold text-neutral-900 dark:bg-white/10 dark:text-white"
	navLinkBase   = "block rounded-control px-3 py-2 text-sm transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-700 dark:focus-visible:outline-primary-500"
)

func (n NavItem) link() *html.AElement {
	a := html.A().Href(n.href).Text(n.label)
	if n.active {
		return a.Class(Classes(navLinkBase, navLinkActive)).Aria("current", "page")
	}
	return a.Class(Classes(navLinkBase, navLinkIdle))
}

func navList(class string, items []NavItem) *html.UlElement {
	ul := html.Ul().Class(Classes(class))
	for _, it := range items {
		ul.Children(html.Li(it.link()))
	}
	return ul
}

// NavbarMenuSignal is the signal Navbar uses for its mobile menu.
const NavbarMenuSignal = "navOpen"

// Navbar returns a top bar with brand on the left and the links beside it. Below the md
// breakpoint the links move into a Drawer that slides in from the right, opened by a menu button
// and held in the signal NavbarMenuSignal; the drawer's X button, Escape and a click on the
// backdrop close it, and focus returns to the menu button.
//
//	ui.Navbar(html.A().Href("/").Text("Acme"),
//		ui.NavLink("Dashboard", "/", true), ui.NavLink("Team", "/team", false))
//
// Use NavbarSignal when a page has more than one navbar.
func Navbar(brand htemel.FlowContent, items ...NavItem) *html.NavElement {
	return NavbarSignal(NavbarMenuSignal, brand, items...)
}

// NavbarSignal is Navbar with the mobile menu state in the signal named signal.
func NavbarSignal(signal string, brand htemel.FlowContent, items ...NavItem) *html.NavElement {
	menu := Drawer(signal, DrawerRight, "Menu", navList("space-y-1", items))
	toggle := html.Button(IconMenu()).Type(html.ButtonTypeButton).
		Class(Classes("-mr-2 inline-flex cursor-pointer items-center justify-center rounded-control p-2 text-neutral-600 transition-colors duration-150 hover:bg-neutral-100",
			"focus-visible:outline-2 focus-visible:outline-primary-700 dark:focus-visible:outline-primary-500 md:hidden dark:text-neutral-300 dark:hover:bg-white/10")).
		Aria("label", "Open menu").
		Aria("haspopup", "dialog").
		Aria("controls", domID(signal)+"-dialog").
		Attrs(expanded("$" + signal)...).
		Attrs(Opens(signal))
	bar := html.Div(
		html.Div(
			brand,
			navList("hidden items-center gap-x-1 md:flex", items),
		).Class(Classes("flex items-center gap-x-8")),
		toggle,
	).Class(Classes("mx-auto flex h-14 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8"))
	return html.Nav(bar, menu).
		Aria("label", "Main").
		Class(Classes("border-b border-neutral-200 bg-white dark:border-white/10 dark:bg-neutral-900"))
}

// SidebarSection is a titled, collapsible group of links in a Sidebar, built by SidebarGroup.
type SidebarSection struct {
	key, title string
	items      []NavItem
	closed     bool
}

// SidebarGroup returns a section of a Sidebar. key identifies it in the sidebar's signal and
// title is its heading. A section starts open; see Closed.
func SidebarGroup(key, title string, items ...NavItem) SidebarSection {
	return SidebarSection{key: key, title: title, items: items}
}

// Closed returns the section starting collapsed. A section that holds the active link starts
// open regardless.
func (s SidebarSection) Closed() SidebarSection {
	s.closed = true
	return s
}

// Sidebar returns a vertical navigation of collapsible sections. signal names an object signal
// with one boolean per section, keyed by the SidebarGroup key ($sidebar.reports for signal
// "sidebar" and key "reports"); the server can preset any of them. Keys are reduced to lowercase
// letters and digits.
func Sidebar(signal string, sections ...SidebarSection) *html.NavElement {
	nav := html.Nav().Aria("label", "Sidebar").Class(Classes("flex flex-col gap-y-1"))
	for _, s := range sections {
		key := toastKey(s.key)
		path := signal + "." + key
		open := "$" + path
		startOpen := !s.closed
		for _, it := range s.items {
			startOpen = startOpen || it.active
		}
		id := domID(signal) + "-" + key
		button := html.Button(
			html.Span().Text(s.title),
			IconChevronDown().Class("size-4 shrink-0 text-neutral-400 transition-transform duration-200 ease-out motion-reduce:transition-none").
				Attrs(datastar.Class("-rotate-90", "!"+open)),
		).Type(html.ButtonTypeButton).
			Class(Classes("flex w-full cursor-pointer items-center justify-between rounded-control px-3 py-1.5 text-left text-xs font-semibold",
				"tracking-wide text-neutral-500 uppercase hover:text-neutral-900 focus-visible:outline-2 focus-visible:outline-primary-700 dark:focus-visible:outline-primary-500",
				"dark:text-neutral-400 dark:hover:text-white")).
			Aria("controls", id).
			Attrs(expanded(open)...).
			Attrs(datastar.OnClick(Toggle(path)))
		list := navList("mt-1 space-y-0.5", s.items).Id(id).Style(hiddenUntilDatastar).Attrs(datastar.Show(open))
		nav.Children(html.Div(button, list).Class("py-1").
			Attrs(datastar.Signal(path, boolExpr(startOpen), datastar.IfMissing())))
	}
	return nav
}

func boolExpr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Crumb is one step of Breadcrumbs. A crumb without an Href, and the last one, is the current
// page and is not a link.
type Crumb struct {
	Label, Href string
}

// Breadcrumbs returns the path to the current page, a <nav aria-label="Breadcrumb"> with an
// ordered list of links separated by chevrons. The last crumb is the current page
// (aria-current="page").
//
//	ui.Breadcrumbs(ui.Crumb{"Home", "/"}, ui.Crumb{"Projects", "/projects"}, ui.Crumb{Label: "Apollo"})
//
// A page with two breadcrumb trails needs a distinct name for each landmark: use BreadcrumbsLabel.
func Breadcrumbs(items ...Crumb) *html.NavElement {
	return BreadcrumbsLabel("Breadcrumb", items...)
}

// BreadcrumbsLabel is Breadcrumbs with label as the name of the <nav> landmark instead of
// "Breadcrumb". Landmarks of the same role need different names for screen-reader users to tell
// them apart, so give each trail on a page its own ("Breadcrumb", "Folder path").
func BreadcrumbsLabel(label string, items ...Crumb) *html.NavElement {
	list := html.Ol().Class(Classes("flex flex-wrap items-center gap-x-1.5 text-sm"))
	for i, c := range items {
		last := i == len(items)-1
		var entry htemel.FlowContent
		switch {
		case last || c.Href == "":
			s := html.Span().Text(c.Label)
			if last {
				s.Aria("current", "page").Class(Classes("font-medium text-neutral-900 dark:text-white"))
			} else {
				s.Class(Classes("text-neutral-500 dark:text-neutral-400"))
			}
			entry = s
		default:
			entry = html.A().Href(c.Href).Text(c.Label).
				Class(Classes("text-neutral-500 hover:text-neutral-900 focus-visible:outline-2 focus-visible:outline-primary-700 dark:focus-visible:outline-primary-500",
					"dark:text-neutral-400 dark:hover:text-white"))
		}
		li := html.Li(entry).Class("flex items-center gap-x-1.5")
		if !last {
			li.Children(IconChevronRight().Class("size-4 shrink-0 text-neutral-400"))
		}
		list.Children(li)
	}
	return html.Nav(list).Aria("label", navLabel([]string{label}, "Breadcrumb"))
}
