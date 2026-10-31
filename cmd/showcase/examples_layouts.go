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
	// The same shell in a 375px container: the sidebar is hidden and the menu button opens it in
	// a Drawer (a modal dialog, so it covers the viewport, not just the frame).
	return framed("375px", "600px", ui.AppShell(
		ui.WithShellID("demoPhone"),
		ui.WithEmbedded("Phone shell preview"),
		ui.WithBrand(demoBrand()),
		ui.WithSidebar(demoShellNav("demoPhone", "Phone shell navigation")),
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
func notesPanes(signal string) (list *html.UlElement, detail *html.DivElement) {
	noteSig := signal + "Note"
	list = html.Ul().Role("list").Class("divide-y divide-neutral-200 dark:divide-white/10")
	detail = html.Div()
	for i, n := range demoNotes {
		idx := strconv.Itoa(i)
		list.Children(html.Li(html.Button(
			html.Span().Class("block truncate text-sm font-medium text-neutral-900 dark:text-white").Text(n.title),
			html.Span().Class("block text-xs text-neutral-600 dark:text-neutral-400").Text(n.date),
		).Type(html.ButtonTypeButton).
			Class("block w-full cursor-pointer px-4 py-3 text-left transition-colors duration-150 hover:bg-neutral-100 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-600 aria-[current=true]:bg-neutral-100 dark:hover:bg-white/5 dark:focus-visible:outline-primary-500 dark:aria-[current=true]:bg-white/10").
			Attrs(
				ui.SplitSelect(signal, "$"+noteSig+" = "+idx),
				datastar.Attr("aria-current", "$"+noteSig+" === "+idx+" ? 'true' : false"),
			)))
		detail.Children(html.Div(
			ui.SplitDetailHeaderSignal(signal, n.title, ui.IconButton(ui.Ghost, ui.IconTrash(), "Delete note", ui.WithSize(ui.Sm))),
			html.P().Class("px-4 py-6 text-sm text-neutral-700 @xl:px-6 dark:text-neutral-300").Text(n.body),
		).Attrs(datastar.Show("$" + noteSig + " === " + idx)))
	}
	return list, detail
}

func splitLayoutNotes() htemel.FlowContent {
	// Wide: the list and the note side by side; drag the list's corner to resize it. Narrow (the
	// 375px frame): one pane at a time, a note opens the detail, Back returns to the list.
	wideList, wideNote := notesPanes("split")
	phoneList, phoneNote := notesPanes("phoneSplit")
	return html.Div(
		framed("100%", "360px", ui.SplitLayout(wideList, wideNote, ui.WithResizable(), ui.WithPaneLabels("Notes", "Note"))),
		framed("375px", "360px", ui.SplitLayout(phoneList, phoneNote, ui.SplitSignal("phoneSplit"), ui.WithPaneLabels("Notes (phone)", "Note (phone)"))),
	).Class("flex flex-col gap-6").Attrs(datastar.Signals(map[string]any{"splitNote": 0, "phoneSplitNote": 0}))
}

func settingsLayout() htemel.FlowContent {
	// A vertical nav beside the form from a 36rem container; a row of tabs above it below that.
	nav := func() []ui.NavItem {
		return []ui.NavItem{
			ui.NavLink("Profile", "#profile", true),
			ui.NavLink("Notifications", "#notifications", false),
			ui.NavLink("Billing", "#billing", false),
			ui.NavLink("Danger zone", "#danger", false),
		}
	}
	form := func(id string) htemel.FlowContent {
		return ui.Stack(ui.Md,
			ui.SectionHeader("Profile"),
			ui.Field(id+"Name", "Display name", ui.Input(html.InputTypeText, "name").Value("Ada Lovelace")),
			ui.Field(id+"Bio", "Bio", ui.Textarea("bio", 3)),
		)
	}
	return html.Div(
		framed("100%", "auto", html.Div(ui.SettingsLayout(nav(), form("settingsWide"))).Class("p-6")),
		framed("375px", "auto", html.Div(ui.SettingsLayoutLabel("Settings (narrow)", nav(), form("settingsNarrow"))).Class("p-4")),
	).Class("flex flex-col gap-6")
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
