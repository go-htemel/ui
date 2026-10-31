package ui

import (
	"io"
	"regexp"
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// Layouts adapt to their container, not the viewport: each root is a CSS container (@container)
// and its breakpoints are container queries (@3xl:, @2xl:, @xl:). A layout in a narrow column, a
// framed preview or another layout's pane takes its narrow form even on a wide screen, and
// layouts nest. They never set an outer max-width; put content in Container for that.

const skipLinkClass = "sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-50 focus:rounded-control focus:bg-white " +
	"focus:px-3 focus:py-2 focus:text-sm focus:font-semibold focus:text-primary-700 focus:shadow-lg focus:ring-1 focus:ring-neutral-950/10 " +
	"focus:outline-2 focus:outline-offset-2 focus:outline-primary-600 " +
	"dark:focus:bg-neutral-800 dark:focus:text-white dark:focus:ring-white/10 dark:focus:outline-primary-500"

// SkipLink returns a "Skip to content" link to href (usually "#main"), visually hidden until it
// receives keyboard focus, when it appears in the top left corner of its positioned ancestor (or
// the page). Put it first in <body> so it is the first Tab stop; AppShell renders one. The target
// should be focusable (tabindex="-1"), as AppShell's <main> is, so focus really moves there.
func SkipLink(href string) *html.AElement {
	return SkipLinkLabel(href, "Skip to content")
}

// SkipLinkLabel is SkipLink with its own text, such as "Skip to editor" or a translation.
func SkipLinkLabel(href, label string) *html.AElement {
	return html.A().Href(href).Class(Classes(skipLinkClass)).Text(label)
}

// AppShellOption configures AppShell: a slot (WithSidebar, WithTopbar, WithBrand,
// WithSidebarFooter, WithContent), a variant (SidebarFixed, SidebarCollapsible, SidebarDark,
// TopbarOnly) or an identity (WithShellID, WithEmbedded).
type AppShellOption func(*shellConfig)

type shellConfig struct {
	sidebar, topbar, footer, content []htemel.FlowContent
	brand                            htemel.FlowContent
	collapsible, dark, topbarOnly    bool
	id, embedded                     string
}

// WithSidebar fills the sidebar, typically with a Sidebar (or a few of them). The same content
// is rendered a second time in the drawer of narrow containers; see AppShell.
func WithSidebar(items ...htemel.FlowContent) AppShellOption {
	return func(c *shellConfig) { c.sidebar = append(c.sidebar, items...) }
}

// WithTopbar fills the top bar, laid out left to right after the menu button and the brand:
// a search field, actions, a user menu. Give the last group ml-auto to push it right.
func WithTopbar(items ...htemel.FlowContent) AppShellOption {
	return func(c *shellConfig) { c.topbar = append(c.topbar, items...) }
}

// WithBrand sets the logo or product name at the top of the sidebar (and in the top bar while the
// sidebar is hidden). In the collapsed rail only its first 2.5rem shows, so lead with a logo.
func WithBrand(node htemel.FlowContent) AppShellOption {
	return func(c *shellConfig) { c.brand = node }
}

// WithSidebarFooter fills the bottom of the sidebar: a user menu, a ColorSchemeToggle. It is
// hidden in the collapsed rail.
func WithSidebarFooter(items ...htemel.FlowContent) AppShellOption {
	return func(c *shellConfig) { c.footer = append(c.footer, items...) }
}

// WithContent fills the content area, the <main> element.
func WithContent(items ...htemel.FlowContent) AppShellOption {
	return func(c *shellConfig) { c.content = append(c.content, items...) }
}

// SidebarFixed is the default variant: the sidebar is shown from the @3xl container width
// (48rem) and opens as a left Drawer from a menu button below it. It undoes SidebarCollapsible
// and TopbarOnly.
func SidebarFixed() AppShellOption {
	return func(c *shellConfig) { c.collapsible, c.topbarOnly = false, false }
}

// SidebarCollapsible adds a button at the bottom of the sidebar that collapses it to a 4rem icon
// rail: link labels, badges, section headings and the footer are hidden (labels stay the links'
// accessible names; icon links carry a title tooltip, see NavItem.Icon) and expanded again. The
// state is the boolean signal RailSignal (true = collapsed), remembered in localStorage under
// AppShellRailKey + ":" + the shell id. The width switches instantly (see AppShell, Motion).
func SidebarCollapsible() AppShellOption {
	return func(c *shellConfig) { c.collapsible, c.topbarOnly = true, false }
}

// SidebarDark draws the sidebar (and its drawer) dark on light content: the sidebar carries the
// dark class, so everything in it renders its dark: variant on a neutral-900 surface, in both
// color schemes. Combine it with SidebarCollapsible.
func SidebarDark() AppShellOption {
	return func(c *shellConfig) { c.dark = true }
}

// TopbarOnly drops the persistent sidebar: the top bar holds the brand and WithTopbar, and the
// content takes the full width. Sidebar content, if any, is reachable from the menu button at
// every width.
func TopbarOnly() AppShellOption {
	return func(c *shellConfig) { c.topbarOnly, c.collapsible = true, false }
}

// WithShellID names the shell when a page has more than one (or nests one): ids and signals are
// derived from it. The default is "shell": signals shellMenu and shellRail, ids shell-sidebar,
// shell-menu-dialog, and <main id="main">; with WithShellID("docs") they are docsMenu, docsRail,
// docs-sidebar and docs-main.
func WithShellID(id string) AppShellOption {
	return func(c *shellConfig) { c.id = id }
}

// WithEmbedded renders the shell as a nested preview rather than the page's frame: the content
// area is a <div role="region" aria-label=label> instead of <main>, and the sidebar a <div>
// rather than an <aside> landmark, so a page that already has its own <main> (a style guide, a
// docs page showing the shell) stays valid. Pair it with WithShellID.
func WithEmbedded(label string) AppShellOption {
	return func(c *shellConfig) { c.embedded = label }
}

// AppShellRailKey prefixes the localStorage key of a collapsible shell's rail state.
const AppShellRailKey = "ui-shell-rail"

// AppShellElement is the shell AppShell returns, a <div> container. MainID, MenuSignal and
// RailSignal name what the shell renders, for links and server patches.
type AppShellElement struct {
	*html.DivElement
	mainID, menuSignal, railSignal string
}

// MainID returns the id of the content area ("main" by default), the skip link's target.
func (e *AppShellElement) MainID() string { return e.mainID }

// MenuSignal returns the signal of the narrow-container drawer (true while it is open).
func (e *AppShellElement) MenuSignal() string { return e.menuSignal }

// RailSignal returns the signal of the collapsed rail (true while collapsed); "" unless
// SidebarCollapsible.
func (e *AppShellElement) RailSignal() string { return e.railSignal }

const (
	shellSidebarClass = "hidden min-h-0 flex-col overflow-hidden border-r border-neutral-200 bg-white text-neutral-900 @3xl:flex " +
		"dark:border-white/10 dark:bg-neutral-900 dark:text-white"
	shellTopbarClass = "flex h-14 shrink-0 items-center gap-x-4 border-b border-neutral-200 bg-white px-4 text-neutral-900 @xl:px-6 " +
		"dark:border-white/10 dark:bg-neutral-900 dark:text-white"
	shellMenuButton = "-ml-2 inline-flex shrink-0 cursor-pointer items-center justify-center rounded-control p-2 text-neutral-600 transition-colors duration-150 " +
		"hover:bg-neutral-100 hover:text-neutral-900 focus-visible:outline-2 focus-visible:outline-primary-600 " +
		"dark:text-neutral-300 dark:hover:bg-white/10 dark:hover:text-white dark:focus-visible:outline-primary-500"
)

// AppShell returns an application frame: a sidebar, a top bar and the content area, for admin
// pages, a CMS, notes. It fills its parent's height (put it in an element with a height: h-dvh
// on <body>, or a framed preview) and the sidebar and the content scroll on their own.
//
//	ui.AppShell(
//		ui.WithBrand(logo),
//		ui.WithSidebar(ui.Sidebar("nav", ui.SidebarGroup("content", "Content",
//			ui.NavLink("Posts", "/posts", true).Icon(ui.IconDocument()).Badge("12")))),
//		ui.WithSidebarFooter(ui.ColorSchemeToggle()),
//		ui.WithTopbar(search, html.Div(userMenu).Class("ml-auto")),
//		ui.WithContent(ui.PageHeader("Posts"), table),
//		ui.SidebarCollapsible(),
//	)
//
// The root is a CSS container. From the @3xl container width (48rem: room for a 16rem sidebar
// and a 32rem content column) the sidebar sits beside the content; below it the sidebar is
// hidden and a menu button in the top bar opens its content in a left Drawer (signal
// MenuSignal). The drawer holds a second rendering of the sidebar and footer slots; ids in that
// copy get the suffix "-drawer" (and aria-controls, aria-labelledby, aria-describedby, for and
// list follow them), so the page has no duplicate ids, while signals are shared: a section opened
// in one is open in the other. The top bar is hidden at @3xl when it would be empty.
//
// # Accessibility
//
// The landmarks are a <header> (banner), an <aside aria-label="Sidebar"> (complementary) and one
// <main id="main" tabindex="-1">, and a SkipLink to it comes first, so Tab from the top of the
// page offers "Skip to content". The drawer is the native modal Drawer (inert page, focus
// trapped, Escape, focus back to the menu button). The rail button names its action ("Collapse
// sidebar" / "Expand sidebar") and binds aria-expanded; collapsed labels are visually hidden, not
// removed, so links keep their names. Links mark the page with aria-current="page" (NavLink).
//
// # Motion
//
// The drawer slides as Drawer does. The rail switches width instantly: animating
// grid-template-columns would lay out the whole content column again on every frame (long
// pages rewrap every line), which the library's motion rules (opacity and transforms only) exist
// to avoid; the icons sit at the same x in both states, so the switch reads as labels appearing
// and disappearing rather than as a jump.
func AppShell(opts ...AppShellOption) *AppShellElement {
	var c shellConfig
	for _, o := range opts {
		if o != nil {
			o(&c)
		}
	}
	key := c.id
	if key == "" {
		key = "shell"
	}
	idp := domID(key)
	sigp := signalID(key)
	menu := sigp + "Menu"
	mainID := "main"
	if c.id != "" {
		mainID = idp + "-main"
	}
	e := &AppShellElement{mainID: mainID, menuSignal: menu}
	hasSidebar := len(c.sidebar) > 0 || len(c.footer) > 0
	persistent := hasSidebar && !c.topbarOnly
	dark := cond(c.dark, "dark scheme-dark")

	root := html.Div(SkipLink("#" + mainID)).Class(Classes("@container relative h-full min-h-0"))
	grid := html.Div().Class(Classes("grid h-full min-h-0 grid-cols-1 grid-rows-[minmax(0,1fr)]",
		cond(persistent, "@3xl:grid-cols-[16rem_minmax(0,1fr)]"),
		cond(persistent && c.collapsible, "@3xl:in-data-ui-rail:grid-cols-[4rem_minmax(0,1fr)]")))
	root.Children(grid)

	if persistent {
		sidebarID := idp + "-sidebar"
		var side *html.DivElement
		body := html.Div(c.sidebar...).Class(Classes("min-h-0 flex-1 overflow-x-hidden overflow-y-auto px-3 py-4"))
		parts := []htemel.FlowContent{}
		if c.brand != nil {
			parts = append(parts, html.Div(c.brand).Class(Classes("flex h-14 shrink-0 items-center overflow-hidden px-4")))
		}
		parts = append(parts, body)
		if len(c.footer) > 0 {
			parts = append(parts, html.Div(c.footer...).Class(Classes("shrink-0 border-t border-neutral-200 px-3 py-3 dark:border-white/10",
				cond(c.collapsible, "@3xl:in-data-ui-rail:hidden"))))
		}
		if c.collapsible {
			rail := sigp + "Rail"
			e.railSignal = rail
			storage := datastar.Quote(AppShellRailKey + ":" + key)
			root.Attrs(
				initially(rail, "false"),
				datastar.Init("try { const v = localStorage.getItem("+storage+"); if (v !== null) $"+rail+" = v === '1' } catch (e) {}"),
				datastar.Attr("data-ui-rail", "$"+rail),
			)
			toggle := IconButton(Ghost, IconSidebar(), "Collapse sidebar", WithSize(Sm)).
				Aria("controls", sidebarID).
				Attrs(
					htemel.Attr("aria-expanded", "true"),
					datastar.Attr("aria-expanded", "$"+rail+" ? 'false' : 'true'"),
					datastar.Attr("aria-label", "$"+rail+" ? 'Expand sidebar' : 'Collapse sidebar'"),
					datastar.Attr("title", "$"+rail+" ? 'Expand sidebar' : 'Collapse sidebar'"),
					datastar.OnClick("$"+rail+" = !$"+rail+"; try { localStorage.setItem("+storage+", $"+rail+" ? '1' : '0') } catch (e) {}"),
				)
			parts = append(parts, html.Div(toggle).Class(Classes("flex shrink-0 border-t border-neutral-200 px-3 py-2 dark:border-white/10")))
		}
		if c.embedded != "" {
			side = html.Div(parts...)
			grid.Children(side.Id(sidebarID).Class(Classes(shellSidebarClass, dark)))
		} else {
			grid.Children(html.Aside(parts...).Id(sidebarID).Aria("label", "Sidebar").Class(Classes(shellSidebarClass, dark)))
		}
	}

	bar := html.Header().Class(Classes(shellTopbarClass, cond(persistent && len(c.topbar) == 0, "@3xl:hidden")))
	if hasSidebar {
		bar.Children(html.Button(IconMenu()).Type(html.ButtonTypeButton).
			Class(Classes(shellMenuButton, cond(persistent, "@3xl:hidden"))).
			Aria("label", "Open menu").
			Aria("haspopup", "dialog").
			Aria("controls", domID(menu)+"-dialog").
			Attrs(expanded("$" + menu)...).
			Attrs(Opens(menu)))
	}
	if c.brand != nil {
		bar.Children(html.Div(c.brand).Class(Classes("flex shrink-0 items-center", cond(persistent, "@3xl:hidden"))))
	}
	if len(c.topbar) > 0 {
		bar.Children(html.Div(c.topbar...).Class(Classes("flex min-w-0 flex-1 items-center gap-x-3")))
	}

	var content htemel.FlowContent
	const contentClass = "@container min-h-0 flex-1 overflow-y-auto outline-none"
	if c.embedded != "" {
		content = html.Div(c.content...).Id(mainID).Role("region").Aria("label", c.embedded).Tabindex(-1).Class(Classes(contentClass))
	} else {
		content = html.Main(c.content...).Id(mainID).Tabindex(-1).Class(Classes(contentClass))
	}
	grid.Children(html.Div(bar, content).Class(Classes("flex min-h-0 min-w-0 flex-col")))

	if hasSidebar {
		drawer := html.Div(suffixIDs(htemel.Group(toNodes(c.sidebar)...), "-drawer"))
		if len(c.footer) > 0 {
			drawer.Children(html.Div(suffixIDs(htemel.Group(toNodes(c.footer)...), "-drawer")).
				Class(Classes("mt-6 border-t border-neutral-200 pt-4 dark:border-white/10")))
		}
		d := DrawerSized(menu, DrawerLeft, Xs, "Menu", drawer)
		if c.dark {
			d.AddClass(dark)
		}
		root.Children(d)
	}
	e.DivElement = root
	return e
}

func toNodes(items []htemel.FlowContent) []htemel.Node {
	out := make([]htemel.Node, len(items))
	for i, it := range items {
		out[i] = it
	}
	return out
}

var (
	startTag = regexp.MustCompile(`<[a-zA-Z][^>]*>`)
	idRefs   = regexp.MustCompile(`(\s(?:id|for|list|aria-controls|aria-labelledby|aria-describedby|aria-owns|aria-activedescendant|popovertarget)=")([^"]*)(")`)
)

// suffixIDs renders n and appends suffix to every id in its start tags, and to the id references
// (for, list, aria-controls, ...) that point at them, so a second rendering of the same nodes
// adds no duplicate ids. Text is escaped by the renderer, so a "<" always starts a tag.
func suffixIDs(n htemel.Node, suffix string) htemel.NodeFunc {
	return func(w io.Writer) error {
		s, err := htemel.RenderString(n)
		if err != nil {
			return err
		}
		s = startTag.ReplaceAllStringFunc(s, func(tag string) string {
			return idRefs.ReplaceAllStringFunc(tag, func(attr string) string {
				m := idRefs.FindStringSubmatch(attr)
				ids := strings.Fields(m[2])
				for i := range ids {
					ids[i] += suffix
				}
				return m[1] + strings.Join(ids, " ") + m[3]
			})
		})
		_, err = io.WriteString(w, s)
		return err
	}
}
