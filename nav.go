package ui

import (
	"strconv"
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
	id          string
	signal      string // Signal: the link sets this signal to value instead of navigating
	value       string
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
// The badge is part of the link's accessible name ("Drafts 12"). With ID set the badge gets the
// id NavBadgeID(id), so the server can patch the count live (see NavBadge).
func (n NavItem) Badge(text string) NavItem {
	n.badge = text
	return n
}

// ID returns the link with id as the id of its <a>, and, if it has a Badge, "<id>-badge" as the
// id of the badge (NavBadgeID). The ids are what a server patch targets to update a count.
// Inside an AppShell the narrow-container drawer holds a second copy of the sidebar whose ids
// end in "-drawer" (see AppShell), so patch both: NavBadge(id, text) and
// NavBadge(id, text).Id(NavBadgeID(id) + "-drawer").
func (n NavItem) ID(id string) NavItem {
	n.id = id
	return n
}

// Signal makes the link drive a signal instead of navigating: a click sets $signal to value (the
// default is prevented, so href is the no-script fallback and what a middle click opens) and
// aria-current="page" follows the signal ($signal === value), which is what highlights the link.
// Use it for sections of one page that the server renders into panels, or patches in, like
// Tabs but with link semantics:
//
//	ui.NavLink("Profile", "#profile", true).Signal("section", "profile")
//
// The active link declares the signal with its value unless the page already has it. In a
// StackedTabs or SettingsLayout the narrow-container Select sets and follows the same signal; use
// one signal per nav.
func (n NavItem) Signal(signal, value string) NavItem {
	n.signal, n.value = signal, value
	return n
}

// NavBadgeID returns the id of the badge of the NavItem whose ID is id: id + "-badge".
func NavBadgeID(id string) string { return id + "-badge" }

// NavBadge returns the badge of a nav link with text, carrying the id NavBadgeID(id), so a
// server patch (a morph by id) replaces a count rendered by NavItem.Badge on a link with
// NavItem.ID(id):
//
//	ui.NavLink("Posts", "/posts", true).ID("nav-posts").Badge("12")
//	patch(sse, ui.NavBadge("nav-posts", "13"))
func NavBadge(id, text string) *html.SpanElement {
	return html.Span().Id(NavBadgeID(id)).Class(Classes(navBadgeClass)).Text(text)
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
	if n.id != "" {
		a.Id(n.id)
	}
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
			badge := html.Span().Class(Classes(navBadgeClass)).Text(n.badge)
			if n.id != "" {
				badge.Id(NavBadgeID(n.id))
			}
			a.Children(badge)
		}
	}
	if n.active {
		a.Aria("current", "page")
	}
	if n.ancestor && !n.active {
		a.Attrs(htemel.Attr("data-ui-ancestor", ""))
	}
	if n.signal != "" {
		q := datastar.Quote(n.value)
		a.Attrs(
			datastar.OnClick("$"+n.signal+" = "+q, datastar.Prevent()),
			datastar.Attr("aria-current", "$"+n.signal+" === "+q+" ? 'page' : false"),
		)
		if n.active {
			a.Attrs(initially(n.signal, q))
		}
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
// A deep trail can be shortened with BreadcrumbsCollapse.
func Breadcrumbs(items ...Crumb) *html.NavElement {
	return BreadcrumbsLabel("Breadcrumb", items...)
}

// BreadcrumbsLabel is Breadcrumbs with label as the name of the <nav> landmark instead of
// "Breadcrumb". Landmarks of the same role need different names for screen-reader users to tell
// them apart, so give each trail on a page its own ("Breadcrumb", "Folder path").
func BreadcrumbsLabel(label string, items ...Crumb) *html.NavElement {
	return breadcrumbs(navLabel([]string{label}, "Breadcrumb"), 0, items)
}

// BreadcrumbsCollapse is Breadcrumbs for a trail that may be deep: when there are more than limit
// crumbs it keeps the first, the last limit-2 and puts an ellipsis button (IconEllipsis, named
// "Show N more pages") between them. The button opens a small list of the hidden crumbs that
// are links (a native <details>, no script; it opens and closes on click or Enter / Space).
// limit below 3 is 3. Chain .Aria("label", ...) on the result to name the landmark.
//
//	ui.BreadcrumbsCollapse(4, crumbs...) // Home › … › Billing › Invoice 42
func BreadcrumbsCollapse(limit int, items ...Crumb) *html.NavElement {
	return breadcrumbs("Breadcrumb", limit, items)
}

const (
	crumbLinkClass = "text-neutral-500 transition-colors duration-150 hover:text-neutral-900 focus-visible:outline-2 focus-visible:outline-primary-600 dark:focus-visible:outline-primary-500 " +
		"dark:text-neutral-400 dark:hover:text-white"
	crumbMoreClass = "flex cursor-pointer list-none items-center rounded-control p-0.5 text-neutral-500 transition-colors duration-150 hover:bg-neutral-100 hover:text-neutral-900 " +
		"focus-visible:outline-2 focus-visible:outline-primary-600 dark:text-neutral-400 dark:hover:bg-white/10 dark:hover:text-white dark:focus-visible:outline-primary-500"
	crumbPanelClass = "absolute top-full left-0 z-20 mt-1 w-56 rounded-surface bg-white p-1 shadow-lg ring-1 ring-neutral-950/5 dark:bg-neutral-800 dark:ring-white/10"
)

func breadcrumbs(name string, limit int, items []Crumb) *html.NavElement {
	var hidden []Crumb
	hideAt := -1
	if limit > 0 {
		limit = max(limit, 3)
		if len(items) > limit {
			keep := limit - 2 // trailing crumbs
			hideAt, hidden = 1, items[1:len(items)-keep]
			items = append(append([]Crumb{items[0]}, Crumb{}), items[len(items)-keep:]...)
		}
	}
	list := html.Ol().Class(Classes("flex flex-wrap items-center gap-x-1.5 text-sm"))
	for i, c := range items {
		last := i == len(items)-1
		var entry htemel.FlowContent
		switch {
		case i == hideAt:
			entry = crumbMore(hidden)
		case last || c.Href == "":
			s := html.Span().Text(c.Label)
			if last {
				s.Aria("current", "page").Class(Classes("font-medium text-neutral-900 dark:text-white"))
			} else {
				s.Class(Classes("text-neutral-500 dark:text-neutral-400"))
			}
			entry = s
		default:
			entry = html.A().Href(c.Href).Text(c.Label).Class(Classes(crumbLinkClass))
		}
		li := html.Li(entry).Class("flex items-center gap-x-1.5")
		if !last {
			li.Children(IconChevronRight().Class("size-4 shrink-0 text-neutral-400"))
		}
		list.Children(li)
	}
	return html.Nav(list).Aria("label", name)
}

// crumbMore is the ellipsis disclosure of a collapsed trail: the hidden crumbs, as links where
// they have an href.
func crumbMore(hidden []Crumb) htemel.FlowContent {
	panel := html.Ul().Class(Classes(crumbPanelClass))
	for _, c := range hidden {
		if c.Href == "" {
			panel.Children(html.Li(html.Span().Class(Classes(menuItemClass, "cursor-default")).Text(c.Label)))
			continue
		}
		panel.Children(html.Li(html.A().Href(c.Href).Class(Classes(menuItemClass)).Text(c.Label)))
	}
	label := "Show " + strconv.Itoa(len(hidden)) + " more " + map[bool]string{true: "page", false: "pages"}[len(hidden) == 1]
	return html.Details(
		html.Summary(IconEllipsis().Class("size-5")).Aria("label", label).Title(label).Class(Classes(crumbMoreClass)),
		panel,
	).Class("relative")
}

// navSelect returns the section picker that replaces a link nav in a narrow container: a
// labelled Select (label name, id derived from it) with one option per item, the active one
// selected. Choosing an option navigates to the item's href, or, for an item with Signal, sets
// its signal; the select also follows the signal, so a click on a wide-form link moves it. The
// wrapper carries class. SettingsLayout and StackedTabs share it.
func navSelect(name string, items []NavItem, class string) *html.DivElement {
	options := make([]html.SelectChild, 0, len(items))
	signal := ""
	for _, it := range items {
		if it.signal != "" && signal == "" {
			signal = it.signal
		}
	}
	for _, it := range items {
		value := it.href
		if signal != "" && it.signal == signal {
			value = it.value
		}
		c := Choice(value, it.label).Selected(it.active)
		if signal != "" && it.signal != signal {
			c.SetAttr("data-ui-href", it.href)
		}
		options = append(options, c)
	}
	id := strings.ToLower(domID(name)) + "-section"
	picker := Select("section", options...)
	if signal == "" {
		picker.Attrs(datastar.On("change", "window.location.href = evt.target.value"))
	} else {
		sig := "$" + signal
		picker.Attrs(
			datastar.On("change", "const o = evt.target.selectedOptions[0]; o.dataset.uiHref ? (window.location.href = o.dataset.uiHref) : ("+sig+" = o.value)"),
			datastar.Effect("const o = [...el.options].find((o) => o.value === "+sig+"); if (o) el.value = o.value"),
		)
	}
	picker.SetAttr("id", id)
	return html.Div(
		html.Label().For(id).Class("sr-only").Text("Section"),
		picker,
	).Class(class)
}
