package main

import (
	"strconv"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

// demoShellNav is the sidebar of the shell demos: sections of icon links with badges, the
// current page marked active and its parent an ancestor. key keeps each demo's signals apart.
func demoShellNav(key, label string) htemel.FlowContent {
	// Each demo names its nav landmark, since the page holds several shells.
	return ui.Sidebar(key+"Nav",
		ui.SidebarGroup("content", "Content",
			ui.NavLink("Dashboard", "#dashboard", false).Icon(ui.IconHome()),
			ui.NavLink("Posts", "#posts", true).Icon(ui.IconDocument()).Badge("12"),
			ui.NavLink("Pages", "#pages", false).Icon(ui.IconFolder()),
		),
		ui.SidebarGroup("admin", "Admin",
			ui.NavLink("Team", "#team", false).Icon(ui.IconUsers()),
			ui.NavLink("Settings", "#settings", false).Icon(ui.IconSettings()),
		),
	).Aria("label", label)
}

func demoBrand() htemel.FlowContent {
	return html.A(
		html.Span().Class("inline-flex size-8 shrink-0 items-center justify-center rounded-control bg-primary-600 text-sm font-bold text-white").Text("A"),
		html.Span().Class("truncate").Text("Acme CMS"),
	).Href("#").Class("flex min-w-0 items-center gap-x-2 text-sm font-semibold text-neutral-900 dark:text-white")
}

func demoUser() htemel.FlowContent {
	return html.Div(
		ui.AvatarInitials("AL", "Ada Lovelace", ui.WithSize(ui.Sm)),
		html.Div(
			html.P().Class("truncate text-sm font-medium text-neutral-900 dark:text-white").Text("Ada Lovelace"),
			html.P().Class("truncate text-xs text-neutral-600 dark:text-neutral-400").Text("ada@example.com"),
		).Class("min-w-0"),
	).Class("flex items-center gap-x-3 px-2")
}

func demoPostsPage() htemel.FlowContent {
	posts := html.Ul().Role("list").Class("divide-y divide-neutral-200 rounded-surface bg-white ring-1 ring-neutral-950/5 dark:divide-white/10 dark:bg-neutral-900 dark:ring-white/10")
	for i, title := range []string{"Container queries in practice", "Writing a CMS in Go", "Notes on Datastar signals"} {
		state := ui.Badge(ui.Success, "Published")
		if i == 1 {
			state = ui.Badge(ui.Warning, "Draft")
		}
		posts.Children(html.Li(
			html.P().Class("min-w-0 truncate text-sm font-medium text-neutral-900 dark:text-white").Text(title),
			state,
		).Class("flex items-center justify-between gap-4 px-4 py-3"))
	}
	return html.Div(
		ui.PageHeader("Posts", ui.WithHeadingLevel(3), ui.WithDescription("Everything you have written."),
			ui.WithActions(ui.Button(ui.Primary, "New post", ui.WithIcon(ui.IconPlus())))),
		html.Div(posts).Class("mt-6"),
	).Class("p-4 @xl:p-6")
}

func demoSearch(id string) htemel.FlowContent {
	return html.Div(ui.Input(html.InputTypeSearch, "q").Placeholder("Search…").Aria("label", "Search").Id(id)).Class("w-full max-w-xs")
}

func appShellDefault() htemel.FlowContent {
	// SidebarFixed (the default): the sidebar sits beside the content from a 48rem container;
	// below that the menu button opens it in a left drawer. WithEmbedded keeps this preview out of
	// the page's own <main>; an application leaves it out.
	return framed("1100px", "520px", ui.AppShell(
		ui.WithShellID("demoShell"),
		ui.WithEmbedded("Default shell preview"),
		ui.WithBrand(demoBrand()),
		ui.WithSidebar(demoShellNav("demoShell", "Default shell navigation")),
		ui.WithSidebarFooter(demoUser()),
		ui.WithTopbar(demoSearch("demoShellSearch")),
		ui.WithContent(demoPostsPage()),
	))
}

func appShellCollapsible() htemel.FlowContent {
	// The button at the bottom of the sidebar collapses it to an icon rail ($demoRailRail),
	// remembered in localStorage. Labels stay the links' accessible names; icons get a title.
	return framed("1100px", "520px", ui.AppShell(
		ui.WithShellID("demoRail"),
		ui.WithEmbedded("Collapsible shell preview"),
		ui.WithBrand(demoBrand()),
		ui.WithSidebar(demoShellNav("demoRail", "Rail shell navigation")),
		ui.WithSidebarFooter(demoUser()),
		ui.WithTopbar(demoSearch("demoRailSearch")),
		ui.WithContent(demoPostsPage()),
		ui.SidebarCollapsible(),
	))
}

func appShellDark() htemel.FlowContent {
	// SidebarDark puts the dark class on the sidebar, so it renders its dark variant on light
	// content in both color schemes.
	return framed("1100px", "520px", ui.AppShell(
		ui.WithShellID("demoDark"),
		ui.WithEmbedded("Dark sidebar preview"),
		ui.WithBrand(demoBrand()),
		ui.WithSidebar(demoShellNav("demoDark", "Dark shell navigation")),
		ui.WithSidebarFooter(demoUser()),
		ui.WithContent(demoPostsPage()),
		ui.SidebarDark(), ui.SidebarCollapsible(),
	))
}

func appShellTopbar() htemel.FlowContent {
	// TopbarOnly: no persistent sidebar, the content takes the full width.
	links := html.Nav(
		html.A().Href("#").Aria("current", "page").Class("rounded-control px-3 py-2 text-sm font-medium text-primary-700 bg-neutral-100 dark:bg-white/10 dark:text-white").Text("Journal"),
		html.A().Href("#").Class("rounded-control px-3 py-2 text-sm font-medium text-neutral-600 hover:bg-neutral-100 dark:text-neutral-400 dark:hover:bg-white/5").Text("Archive"),
	).Aria("label", "Journal sections").Class("hidden items-center gap-x-1 @xl:flex")
	return framed("1100px", "360px", ui.AppShell(
		ui.WithShellID("demoTop"),
		ui.WithEmbedded("Top bar shell preview"),
		ui.WithBrand(demoBrand()),
		ui.WithTopbar(links, html.Div(ui.AvatarInitials("AL", "Ada Lovelace", ui.WithSize(ui.Sm))).Class("ml-auto")),
		ui.WithContent(html.Div(ui.PageHeader("Today", ui.WithHeadingLevel(3), ui.WithMeta(htemel.Text("Saturday, 31 October")))).Class("p-6")),
		ui.TopbarOnly(),
	))
}

func appShellPhone() htemel.FlowContent {
	// A native modal <dialog> (the Drawer the menu opens) lives in the top layer of its document,
	// so inside a framed() div it would cover the whole page, not the frame. The narrow previews
	// are therefore real viewports: an <iframe> of a bare page that renders only the shell (see
	// phoneShellPage). A 390px viewport is below the 48rem breakpoint, so the sidebar is hidden
	// and the menu button opens it in the drawer, inside the iframe.
	return phoneFrame("appshell", "Phone-width app shell with a menu drawer")
}

// phoneFrame embeds the bare layout page /components/layouts/frame/<name> in a 390x760 iframe
// drawn like a phone: a real viewport, so container queries, dialogs and the top layer behave as
// on a phone.
func phoneFrame(name, title string) htemel.FlowContent {
	return html.Div(
		html.Iframe().Src("/components/layouts/frame/" + name).Title(title).Loading(html.LoadingLazy).
			Class("block h-[760px] w-[390px] max-w-full border-0 bg-white dark:bg-neutral-950"),
	).Class("w-fit max-w-full overflow-hidden rounded-surface ring-1 ring-neutral-950/10 dark:ring-white/10")
}

// phoneShellPage is the page of the phone-width app shell iframe.
func phoneShellPage() htemel.Node {
	return bareDocument("Phone-width app shell", "h-dvh", ui.AppShell(
		ui.WithShellID("phone"),
		ui.WithBrand(demoBrand()),
		ui.WithSidebar(demoShellNav("phone", "Phone shell navigation")),
		ui.WithSidebarFooter(demoUser()),
		ui.WithContent(demoPostsPage()),
	))
}

var demoNotes = []struct{ title, date, body string }{
	{"Groceries", "Today", "Oat milk, lemons, basil, a loaf of sourdough."},
	{"Talk outline", "Yesterday", "Container queries: why the viewport is the wrong question, nesting layouts, framed previews."},
	{"Book notes", "Monday", "Chapter 3 argues that most layout bugs are breakpoint bugs in disguise."},
}

// notesPanes builds the panes of the notes demo for one split signal: the list sets
// $<signal>Note (declared by the example) and shows the detail, which renders every note and
// shows the selected one.
func notesPanes(signal string) (list *html.UlElement, detail, detailHeader *html.DivElement) {
	noteSig := signal + "Note"
	list = html.Ul().Role("list").Class("divide-y divide-neutral-200 dark:divide-white/10")
	detail, detailHeader = html.Div(), html.Div()
	for i, n := range demoNotes {
		idx := strconv.Itoa(i)
		list.Children(html.Li(ui.SplitListItem(signal, "$"+noteSig+" === "+idx, "$"+noteSig+" = "+idx,
			html.Span().Class("block truncate text-sm font-medium text-neutral-900 dark:text-white").Text(n.title),
			html.Span().Class("block text-xs text-neutral-600 dark:text-neutral-400").Text(n.date),
		)))
		detailHeader.Children(html.Div(
			ui.SplitDetailHeaderSignal(signal, n.title, ui.IconButton(ui.Ghost, ui.IconTrash(), "Delete note", ui.WithSize(ui.Sm))),
		).Attrs(datastar.Show("$" + noteSig + " === " + idx)))
		detail.Children(html.P().Class("px-4 py-6 text-sm text-neutral-700 @xl:px-6 dark:text-neutral-300").
			Attrs(datastar.Show("$" + noteSig + " === " + idx)).Text(n.body))
	}
	return list, detail, detailHeader
}

// notesSplit is the notes demo's SplitLayout: a titled list header and the selected note's header
// share one grid row, so their rules align.
func notesSplit(signal string, opts ...ui.SplitOption) *html.DivElement {
	list, detail, detailHeader := notesPanes(signal)
	opts = append(opts, ui.WithPaneLabels("Notes", "Note"),
		ui.WithListHeader(ui.SplitListHeader("Notes", ui.IconButton(ui.Ghost, ui.IconPlus(), "New note", ui.WithSize(ui.Sm)))),
		ui.WithDetailHeader(detailHeader))
	return ui.SplitLayout(list, detail, opts...)
}

func splitLayoutNotes() htemel.FlowContent {
	// Wide: the list and the note side by side; drag the list's corner to resize it. Narrow: one
	// pane at a time, a note opens the detail, Back returns to the list. The narrow form is the
	// second preview, a real 390px viewport (see appShellPhone and phoneSplitPage).
	return html.Div(
		framed("100%", "360px", notesSplit("split", ui.WithResizable())),
		phoneFrame("split", "Phone-width notes split"),
	).Class("flex flex-col gap-6").Attrs(datastar.Signals(map[string]any{"splitNote": 0}))
}

// phoneSplitPage is the page of the narrow split iframe: the same panes, with the layout
// filling the viewport. Its signals live in the frame's own document.
func phoneSplitPage() htemel.Node {
	return bareDocument("Phone-width notes split", "h-dvh", html.Main(
		notesSplit("split"),
	).Class("h-full").Attrs(datastar.Signals(map[string]any{"splitNote": 0})))
}

// settingsNav is the sections of the settings demos.
func settingsNav() []ui.NavItem {
	return []ui.NavItem{
		ui.NavLink("Profile", "#profile", true),
		ui.NavLink("Notifications", "#notifications", false),
		ui.NavLink("Billing", "#billing", false),
		ui.NavLink("Danger zone", "#danger", false),
	}
}

func settingsForm(id string) htemel.FlowContent {
	return ui.Stack(ui.Md,
		ui.SectionHeader("Profile"),
		ui.Field(id+"Name", "Display name", ui.Input(html.InputTypeText, "name").Value("Ada Lovelace")),
		ui.Field(id+"Bio", "Bio", ui.Textarea("bio", 3)),
	)
}

func settingsLayout() htemel.FlowContent {
	// A vertical nav beside the form from a 36rem container; a section picker above it below
	// that, shown in a real 390px viewport (see appShellPhone and phoneSettingsPage).
	return html.Div(
		framed("100%", "auto", html.Div(ui.SettingsLayout(settingsNav(), settingsForm("settingsWide"))).Class("p-6")),
		phoneFrame("settings", "Phone-width settings"),
	).Class("flex flex-col gap-6")
}

// phoneSettingsPage is the page of the narrow settings iframe.
func phoneSettingsPage() htemel.Node {
	return bareDocument("Phone-width settings", "", html.Main(
		ui.SettingsLayout(settingsNav(), settingsForm("settingsNarrow")),
	).Class("p-4"))
}

func stackedLayout() htemel.FlowContent {
	// A record page: header band with breadcrumbs, meta and actions, link tabs, then the content.
	header := ui.PageHeader("Writing a CMS in Go",
		ui.WithHeadingLevel(3),
		ui.WithBreadcrumbs(ui.BreadcrumbsLabel("Post breadcrumb", ui.Crumb{Label: "Posts", Href: "#"}, ui.Crumb{Label: "Writing a CMS in Go"})),
		ui.WithMeta(ui.Badge(ui.Warning, "Draft"), htemel.Text("Edited 2 minutes ago"), htemel.Text("1,240 words")),
		ui.WithActions(ui.Button(ui.Secondary, "Preview"), ui.Button(ui.Primary, "Publish")),
	)
	return framed("100%", "auto", ui.StackedLayout(header,
		ui.StackedTabs("Post sections",
			ui.NavLink("Content", "#content", true),
			ui.NavLink("Comments", "#comments", false).Badge("4"),
			ui.NavLink("Revisions", "#revisions", false),
		),
		html.P().Class("text-sm text-neutral-700 dark:text-neutral-300").Text("The editor goes here."),
	))
}

func pageHeaders() htemel.FlowContent {
	// PageHeader puts the actions beside the title from 36rem; SectionHeader heads a section.
	return ui.Stack(ui.Xl,
		ui.PageHeader("Edit post",
			ui.WithHeadingLevel(3),
			ui.WithBackLink("#posts", "Posts"),
			ui.WithDescription("Changes are saved as a draft until you publish."),
			ui.WithActions(ui.Button(ui.Secondary, "Discard"), ui.Button(ui.Primary, "Save")),
		),
		ui.SectionHeader("Comments", ui.Button(ui.Secondary, "Moderate", ui.WithSize(ui.Sm))),
	)
}
