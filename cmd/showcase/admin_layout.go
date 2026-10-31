package main

import (
	"strconv"
	"time"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

// The Admin demo is a set of full-page routes under /admin: each page is a complete document
// holding an AppShell, not an example inside the component pages' frame. adminDocument is the
// document, adminPage the shell around one page's content.

// Admin nav items, the values of adminPage's current argument.
const (
	navDashboard = "dashboard"
	navPosts     = "posts"
	navMedia     = "media"
	navComments  = "comments"
	navNotes     = "notes"
	navSettings  = "settings"
)

// navPostsID is the id of the Posts link; its badge, navPostsID-badge, is patched by the posts list.
const navPostsID = "nav-posts"

const (
	adminThemeKey = "admin-theme"
	adminMainID   = "main"
)

// adminNotices are the toasts a page can show when it loads, chosen by ?notice= (a redirect
// after a create or a delete cannot carry the toast an SSE response would have patched in).
var adminNotices = map[string]struct {
	tone       ui.Tone
	title, msg string
}{
	"created": {ui.Success, "Post created", "Your new post was saved as a draft."},
	"deleted": {ui.Info, "Post deleted", "The post was removed from the list."},
	"reset":   {ui.Info, "Demo data reset", "Posts and notes are back to the seed."},
}

// adminDocument wraps shell in an HTML document. notice, if it names an adminNotices entry, is
// shown as a toast on load. The body applies the stored theme (the Appearance setting) once
// Datastar starts; the color scheme is applied before first paint by ui.ColorSchemeBootstrap.
func adminDocument(title string, shell htemel.FlowContent, notice string) htemel.Node {
	region := ui.ToastRegion()
	if n, ok := adminNotices[notice]; ok {
		region.Children(ui.Toast(ui.NextToastID(), n.tone, n.title, n.msg, 5*time.Second))
	}
	return htemel.Group(
		htemel.Doctype(),
		html.Html(
			documentHead(title),
			html.Body(shell, region).
				Class(ui.BodyClasses()+" h-dvh").
				Attrs(
					datastar.Signal("theme", "''", datastar.IfMissing()),
					datastar.Init("try { $theme = localStorage.getItem('"+adminThemeKey+"') || '' } catch (e) {}"),
					datastar.Effect("$theme ? (document.documentElement.dataset.theme = $theme) : delete document.documentElement.dataset.theme"),
				),
		).Lang("en"),
	)
}

// adminBrand is the logo in the sidebar and, when the sidebar is hidden, the top bar.
func adminBrand() htemel.FlowContent {
	return html.A(
		html.Span().Class("inline-flex size-8 shrink-0 items-center justify-center rounded-control bg-primary-600 text-sm font-bold text-white").Aria("hidden", "true").Text("A"),
		html.Span().Class("truncate").Text("Acme CMS"),
	).Href("/admin").Class("flex min-w-0 items-center gap-x-2 rounded-control text-sm font-semibold text-neutral-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600 dark:text-white dark:focus-visible:outline-primary-500")
}

// adminSidebar is the navigation: current marks the page's own item aria-current="page"; on the
// editor pages Posts is the ancestor.
func adminSidebar(current string, ancestor bool) htemel.FlowContent {
	item := func(key, label, href string, icon htemel.PhrasingContent) ui.NavItem {
		n := ui.NavLink(label, href, current == key).Icon(icon)
		if ancestor && key == current {
			n = ui.NavLink(label, href, false).Icon(icon).Ancestor()
		}
		return n
	}
	posts := item(navPosts, "Posts", "/admin/posts", ui.IconDocument())
	posts = posts.ID(navPostsID).Badge(strconv.Itoa(adminData.counts().Total))
	return ui.Sidebar("adminNav",
		ui.SidebarGroup("manage", "Manage",
			item(navDashboard, "Dashboard", "/admin", ui.IconHome()),
			posts,
			item(navMedia, "Media", "/admin/media", ui.IconImage()),
			item(navComments, "Comments", "/admin/comments", ui.IconQuote()),
			item(navNotes, "Notes", "/admin/notes", ui.IconFile()),
		),
		ui.SidebarGroup("site", "Site",
			item(navSettings, "Settings", "/admin/settings", ui.IconSettings()),
		),
	).Aria("label", "Main")
}

// userMenu is the account Dropdown of the sidebar footer.
func userMenu() htemel.FlowContent {
	trigger := ui.Button(ui.Ghost, "Ada Lovelace",
		ui.WithIcon(html.Span().Aria("hidden", "true").
			Class("inline-flex size-6 shrink-0 items-center justify-center rounded-full bg-neutral-600 text-[0.625rem] font-medium text-white dark:bg-neutral-500").Text("AL")),
		ui.WithTrailingIcon(ui.IconChevronDown()),
		ui.WithRailIcon(),
	).AddClass("w-full min-w-0 justify-start *:last:ml-auto")
	return ui.Dropdown("userMenu", trigger,
		ui.MenuLink("Account settings", "/admin/settings"),
		ui.MenuLink("Writing preferences", "/admin/settings/writing"),
		ui.MenuDivider(),
		ui.MenuLink("Back to showcase", "/"),
	).AddClass("block w-full")
}

// backToShowcase is the top bar's link out of the demo.
func backToShowcase() htemel.FlowContent {
	return html.A(ui.IconArrowLeft().Class("size-4 shrink-0"), htemel.Text("back to showcase")).Href("/").
		Class("inline-flex items-center gap-x-1.5 rounded-control text-sm font-medium text-neutral-600 transition-colors duration-150 hover:text-neutral-900 " +
			"focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600 dark:text-neutral-400 dark:hover:text-white dark:focus-visible:outline-primary-500")
}

// adminPage renders the admin shell around content, as a whole document: current names the nav
// item to mark, ancestor says the page is below it (an editor under Posts), notice an
// adminNotices key to toast on load.
func adminPage(title, current string, ancestor bool, notice string, content ...htemel.FlowContent) htemel.Node {
	shell := ui.AppShell(
		ui.WithBrand(adminBrand()),
		ui.WithSidebar(adminSidebar(current, ancestor)),
		ui.WithSidebarFooter(userMenu(), html.Div(ui.ColorSchemeToggle()).Class("mt-3 flex justify-center @3xl:in-data-ui-rail:mt-0")),
		ui.WithTopbar(backToShowcase()),
		ui.WithContent(content...),
		ui.SidebarCollapsible(),
	)
	return adminDocument(title, shell, notice)
}

// adminBody is the padded column of a page's content inside <main>. The main element is a
// container, so the padding grows with its width, not the viewport's.
func adminBody(children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).Class("mx-auto flex max-w-6xl flex-col gap-y-6 p-4 @xl:p-6 @5xl:p-8")
}

// adminNotFound is the 404 page: an EmptyState in the shell, so the visitor keeps the nav.
func adminNotFound() htemel.Node {
	return adminPage("Page not found", "", false, "",
		adminBody(html.H1().Class("sr-only").Text("Page not found"), ui.EmptyState(ui.IconSearch(), "Page not found",
			"This page does not exist in the demo, or it was moved.",
			html.Div(ui.LinkButton(ui.Primary, "Go to the dashboard", "/admin", ui.WithIcon(ui.IconHome())))).AddClass("py-24")))
}

// adminStub is a section of the nav that the demo does not build out.
func adminStub(title, current, text string) htemel.Node {
	return adminPage(title, current, false, "",
		adminBody(
			ui.PageHeader(title, ui.WithDescription(text)),
			ui.EmptyState(ui.IconFolder(), "Nothing here in the demo", "This section is only a nav entry; the posts, notes and settings pages are fully working.",
				html.Div(ui.LinkButton(ui.Secondary, "Back to the dashboard", "/admin"))),
		))
}
