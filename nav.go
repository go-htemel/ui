package ui

import (
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// NavItem is one link of a Navbar or a Sidebar section, built by NavLink.
type NavItem struct {
	label, href string
	active      bool
	ancestor    bool
	icon        htemel.PhrasingContent
	badge       string
	attrs       []htemel.Attribute
}

// NavLink returns a navigation link. active marks the current page: the link gets
// aria-current="page", which is what highlights it.
func NavLink(label, href string, active bool) NavItem {
	return NavItem{label: label, href: href, active: active}
}

// Attrs returns the link with attributes added to its <a>, such as a Datastar click handler. The
// highlight follows aria-current, so a client-side router can move it by binding that attribute:
//
//	ui.NavLink("Team", "/team", false).Attrs(
//		datastar.OnClick("$page = 'team'", datastar.Prevent()),
//		datastar.Attr("aria-current", "$page === 'team' ? 'page' : false"))
func (n NavItem) Attrs(attrs ...htemel.Attribute) NavItem {
	n.attrs = append(append([]htemel.Attribute(nil), n.attrs...), attrs...)
	return n
}

// Icon returns the link with an icon before its label, such as ui.IconHome(). An icon link
// keeps its label in a <span> and carries title=label: in the icon rail of a collapsed AppShell
// (SidebarCollapsible) the label is visually hidden but stays the link's accessible name, and the
// title is its tooltip.
func (n NavItem) Icon(icon htemel.PhrasingContent) NavItem {
	n.icon = icon
	return n
}

// Badge returns the link with a count or short label after its text, such as "12" unread posts.
// The badge is part of the link's accessible name ("Drafts 12").
func (n NavItem) Badge(text string) NavItem {
	n.badge = text
	return n
}

// Ancestor marks the link as an ancestor of the current page (Posts while an editor for one post
// is open): its label is semibold and full contrast, without the tint and without
// aria-current="page", which belongs to the page itself.
func (n NavItem) Ancestor() NavItem {
	n.ancestor = true
	return n
}

// navLinkClass styles idle links and, through aria-current="page", the active one: a neutral tint
// with the label in the primary color, so the current page stands out on white and on neutral-50.
const navLinkClass = "block rounded-control px-3 py-2 text-sm font-medium text-neutral-600 transition-colors duration-150 " +
	"hover:bg-neutral-100 hover:text-neutral-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600 " +
	"aria-[current=page]:bg-neutral-100 aria-[current=page]:font-semibold aria-[current=page]:text-primary-700 " +
	"dark:text-neutral-400 dark:hover:bg-white/5 dark:hover:text-white dark:focus-visible:outline-primary-500 " +
	"dark:aria-[current=page]:bg-white/10 dark:aria-[current=page]:text-white"

// navAncestorClass is the ancestor state (NavItem.Ancestor): the label at full contrast.
const navAncestorClass = "data-ui-ancestor:font-semibold data-ui-ancestor:text-neutral-900 dark:data-ui-ancestor:text-white"

// navRichClass replaces the block layout of navLinkClass for links with an icon or a badge. In
// the rail of a collapsed AppShell (an ancestor with data-ui-rail, from an @3xl container) the
// label and the badge become visually hidden; they stay the accessible name.
const (
	navRichClass  = "flex items-center gap-x-3"
	navIconClass  = "size-5 shrink-0 text-neutral-500 in-aria-[current=page]:text-primary-700 dark:text-neutral-400 dark:in-aria-[current=page]:text-white"
	navLabelClass = "min-w-0 flex-1 truncate @3xl:in-data-ui-rail:sr-only"
	navBadgeClass = "ml-auto shrink-0 rounded-full bg-neutral-100 px-2 py-0.5 text-xs font-medium text-neutral-700 dark:bg-white/10 dark:text-neutral-300 @3xl:in-data-ui-rail:sr-only"
)

func (n NavItem) link() *html.AElement { return n.linkClass(navLinkClass) }

// linkClass renders the link with base, a class list that starts with "block " (replaced by a
// flex row when the link has an icon or a badge).
func (n NavItem) linkClass(base string) *html.AElement {
	a := html.A().Href(n.href)
	if n.icon == nil && n.badge == "" {
		a.Text(n.label).Class(Classes(base, cond(n.ancestor, navAncestorClass)))
	} else {
		a.Class(Classes(strings.Replace(base, "block ", navRichClass+" ", 1), cond(n.ancestor, navAncestorClass)))
		if n.icon != nil {
			if el, ok := n.icon.(htemel.Element); ok {
				el.SetAttr("class", Classes(navIconClass))
			}
			a.Children(n.icon).Title(n.label)
		}
		label := html.Span().Text(n.label)
		if n.icon != nil {
			label.Class(Classes(navLabelClass))
		} else {
			label.Class("min-w-0 flex-1 truncate")
		}
		a.Children(label)
		if n.badge != "" {
			a.Children(html.Span().Class(Classes(navBadgeClass)).Text(n.badge))
		}
	}
	if n.active {
		a.Aria("current", "page")
	}
	if n.ancestor && !n.active {
		a.Attrs(htemel.Attr("data-ui-ancestor", ""))
	}
	return a.Attrs(n.attrs...)
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

// Navbar returns a top bar with brand on the left and the links beside it. The bar is a CSS
// container (@container) and switches on its own width, not the viewport's: from the @2xl
// container size (42rem) the links sit inline; below it they move into a Drawer that slides in
// from the right, opened by a menu button and held in the signal NavbarMenuSignal. The drawer's
// X button, Escape and a click on the backdrop close it, and focus returns to the menu button. A
// navbar in a narrow column therefore collapses even on a wide screen.
//
//	ui.Navbar(html.A().Href("/").Text("Acme"),
//		ui.NavLink("Dashboard", "/", true), ui.NavLink("Team", "/team", false))
//
// Use NavbarSignal when a page has more than one navbar (and give each nav its own aria-label;
// the default is "Main").
func Navbar(brand htemel.FlowContent, items ...NavItem) *html.NavElement {
	return NavbarSignal(NavbarMenuSignal, brand, items...)
}

// NavbarSignal is Navbar with the mobile menu state in the signal named signal.
func NavbarSignal(signal string, brand htemel.FlowContent, items ...NavItem) *html.NavElement {
	menu := Drawer(signal, DrawerRight, "Menu", navList("space-y-1", items))
	toggle := html.Button(IconMenu()).Type(html.ButtonTypeButton).
		Class(Classes("-mr-2 inline-flex cursor-pointer items-center justify-center rounded-control p-2 text-neutral-600 transition-colors duration-150 hover:bg-neutral-100",
			"focus-visible:outline-2 focus-visible:outline-primary-600 dark:focus-visible:outline-primary-500 @2xl:hidden dark:text-neutral-300 dark:hover:bg-white/10")).
		Aria("label", "Open menu").
		Aria("haspopup", "dialog").
		Aria("controls", domID(signal)+"-dialog").
		Attrs(expanded("$" + signal)...).
		Attrs(Opens(signal))
	bar := html.Div(
		html.Div(
			brand,
			navList("hidden items-center gap-x-1 @2xl:flex", items),
		).Class(Classes("flex items-center gap-x-8")),
		toggle,
	).Class(Classes("mx-auto flex h-14 max-w-7xl items-center justify-between px-4 @xl:px-6 @5xl:px-8"))
	return html.Nav(bar, menu).
		Aria("label", "Main").
		Class(Classes("@container border-b border-neutral-200 bg-white dark:border-white/10 dark:bg-neutral-900"))
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
			Class(Classes("flex w-full cursor-pointer items-center justify-between rounded-control px-3 py-1.5 text-left text-xs font-semibold transition-colors duration-150",
				"tracking-wide text-neutral-500 uppercase hover:text-neutral-900 focus-visible:outline-2 focus-visible:outline-primary-600 dark:focus-visible:outline-primary-500",
				"dark:text-neutral-400 dark:hover:text-white @3xl:in-data-ui-rail:hidden")).
			Aria("controls", id).
			Attrs(expanded(open)...).
			Attrs(datastar.OnClick(Toggle(path)))
		// In the icon rail of a collapsed AppShell the heading is hidden, so every section is shown
		// (!block beats the inline display: none of data-show).
		list := navList("mt-1 space-y-0.5 @3xl:in-data-ui-rail:!block @3xl:in-data-ui-rail:mt-0", s.items).Id(id).Style(hiddenUntilDatastar).Attrs(datastar.Show(open))
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
				Class(Classes("text-neutral-500 transition-colors duration-150 hover:text-neutral-900 focus-visible:outline-2 focus-visible:outline-primary-600 dark:focus-visible:outline-primary-500",
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
