package ui_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

var (
	_ htemel.FlowContent = ui.AppShell()
	_ htemel.FlowContent = ui.SkipLink("#main")
	_ htemel.FlowContent = ui.SplitLayout(nil, nil)
	_ htemel.FlowContent = ui.SplitDetailHeader("x")
	_ htemel.FlowContent = ui.SettingsLayout(nil)
	_ htemel.FlowContent = ui.SettingsLayoutLabel("x", nil)
	_ htemel.FlowContent = ui.StackedLayout(nil)
	_ htemel.FlowContent = ui.StackedTabs("x")
	_ htemel.FlowContent = ui.PageHeader("x")
	_ htemel.FlowContent = ui.SectionHeader("x")
	_ htemel.Attribute   = ui.SplitSelect("split", "")
	// SkipLink is phrasing content, like the <a> it is, so it fits in a paragraph too.
	_ htemel.PhrasingContent = ui.SkipLink("#main")
)

func shellNav() htemel.FlowContent {
	return ui.Sidebar("nav", ui.SidebarGroup("content", "Content",
		ui.NavLink("Posts", "/posts", true).Icon(ui.IconDocument()).Badge("12"),
		ui.NavLink("Pages", "/pages", false).Icon(ui.IconFolder()),
	))
}

func init() {
	brand := html.A().Href("/").Text("Acme")
	addGolden(
		golden{"skip-link", ui.SkipLink("#main")},
		golden{"sidebar-rich", ui.Sidebar("s", ui.SidebarGroup("g", "Content",
			ui.NavLink("Posts", "/posts", false).Icon(ui.IconDocument()).Ancestor(),
			ui.NavLink("Drafts", "/drafts", true).Badge("3")))},
		golden{"app-shell", ui.AppShell(ui.WithBrand(brand), ui.WithSidebar(shellNav()),
			ui.WithSidebarFooter(ui.ColorSchemeToggle()), ui.WithContent(html.P().Text("Body")))},
		golden{"app-shell-collapsible", ui.AppShell(ui.WithBrand(brand), ui.WithSidebar(shellNav()),
			ui.WithTopbar(html.Span().Text("Search")), ui.WithContent(html.P().Text("Body")), ui.SidebarCollapsible())},
		golden{"app-shell-dark", ui.AppShell(ui.WithBrand(brand), ui.WithSidebar(shellNav()), ui.SidebarDark(), ui.WithShellID("admin"))},
		golden{"app-shell-topbar", ui.AppShell(ui.WithBrand(brand), ui.WithTopbar(html.Span().Text("Search")), ui.TopbarOnly(),
			ui.WithContent(html.P().Text("Body")))},
		golden{"app-shell-embedded", ui.AppShell(ui.WithSidebar(shellNav()), ui.WithEmbedded("Preview"), ui.WithShellID("demo"))},
		golden{"split-layout", ui.SplitLayout(html.P().Text("List"), html.Div(ui.SplitDetailHeader("Note", ui.Button(ui.Secondary, "Edit")), html.P().Text("Body")))},
		golden{"split-layout-options", ui.SplitLayout(html.P().Text("List"), html.P().Text("Detail"),
			ui.SplitSignal("pane"), ui.WithListWidth(ui.Lg), ui.WithResizable(), ui.WithPaneLabels("Notes", "Note"))},
		golden{"settings-layout", ui.SettingsLayout([]ui.NavItem{
			ui.NavLink("Profile", "/p", true), ui.NavLink("Billing", "/b", false).Icon(ui.IconSettings())}, html.P().Text("Form"))},
		golden{"stacked-layout", ui.StackedLayout(ui.PageHeader("Post"), ui.StackedTabs("Post sections",
			ui.NavLink("Overview", "/o", true), ui.NavLink("Comments", "/c", false).Badge("4")), html.P().Text("Body"))},
		golden{"stacked-layout-plain", ui.StackedLayout(ui.PageHeader("Post"), html.P().Text("Body"))},
		golden{"page-header", ui.PageHeader("Edit post",
			ui.WithBackLink("/posts", "Posts"),
			ui.WithBreadcrumbs(ui.Breadcrumbs(ui.Crumb{Label: "Posts", Href: "/posts"}, ui.Crumb{Label: "Edit"})),
			ui.WithDescription("Changes are saved as a draft."),
			ui.WithMeta(ui.Badge(ui.Warning, "Draft"), htemel.Text("Saved 2 minutes ago")),
			ui.WithActions(ui.Button(ui.Primary, "Publish")),
			ui.WithHeadingLevel(2))},
		golden{"section-header", ui.SectionHeader("Comments", ui.Button(ui.Secondary, "Add"))},
	)
}

func TestAppShellLandmarks(t *testing.T) {
	got := render(t, ui.AppShell(ui.WithBrand(htemel.Text("Acme")), ui.WithSidebar(shellNav()), ui.WithContent(html.P().Text("x"))))
	wantAll(t, got,
		`<a href="#main" class="sr-only focus:not-sr-only`,
		`<aside id="shell-sidebar" aria-label="Sidebar"`,
		`<header class="flex h-14`,
		`<main id="main" tabindex="-1"`,
		`aria-controls="shellMenu-dialog"`, `<dialog id="shellMenu-dialog"`,
	)
	if n := strings.Count(got, "<main"); n != 1 {
		t.Errorf("%d <main> elements, want 1", n)
	}
	// The drawer copy renames ids and the references to them, so no id appears twice.
	ids := regexp.MustCompile(`\sid="([^"]+)"`).FindAllStringSubmatch(got, -1)
	seen := map[string]bool{}
	for _, m := range ids {
		if seen[m[1]] {
			t.Errorf("duplicate id %q", m[1])
		}
		seen[m[1]] = true
	}
	wantAll(t, got, `id="nav-content-drawer"`, `aria-controls="nav-content-drawer"`)
	// The narrow top bar is hidden at @3xl when it holds nothing else.
	wantAll(t, got, `dark:text-white @3xl:hidden"`)
}

func TestAppShellEmbedded(t *testing.T) {
	got := render(t, ui.AppShell(ui.WithSidebar(shellNav()), ui.WithEmbedded("App preview"), ui.WithShellID("demo")))
	for _, gone := range []string{"<main", "<aside"} {
		if strings.Contains(got, gone) {
			t.Errorf("embedded shell renders %s", gone)
		}
	}
	wantAll(t, got, `<a href="#demo-main"`, `id="demo-main" role="region" aria-label="App preview" tabindex="-1"`, `id="demoMenu-dialog"`)
}

func TestAppShellRail(t *testing.T) {
	e := ui.AppShell(ui.WithSidebar(shellNav()), ui.SidebarCollapsible(), ui.WithShellID("cms"))
	if e.RailSignal() != "cmsRail" || e.MenuSignal() != "cmsMenu" || e.MainID() != "cms-main" {
		t.Errorf("signals %q %q, main %q", e.RailSignal(), e.MenuSignal(), e.MainID())
	}
	got := render(t, e)
	wantAll(t, got,
		`data-signals:cms-rail__ifmissing="false"`,
		`data-attr:data-rail="$cmsRail"`,
		`localStorage.getItem('ui-shell-rail:cms')`,
		`@3xl:in-data-rail:grid-cols-[4rem_minmax(0,1fr)]`,
		`aria-controls="cms-sidebar" aria-expanded="true"`,
		`data-attr:aria-label="$cmsRail ? 'Expand sidebar' : 'Collapse sidebar'"`,
		// Labels are visually hidden in the rail, not removed, and icon links carry a title.
		`title="Posts"`, `@3xl:in-data-rail:sr-only`,
	)
	if strings.Count(got, `aria-expanded="true"`) != 1 {
		t.Errorf("the rail button must carry exactly one aria-expanded")
	}
	if ui.AppShell(ui.WithSidebar(shellNav())).RailSignal() != "" {
		t.Error("RailSignal without SidebarCollapsible")
	}
	// The motion rule: the rail switches width, it does not transition the grid columns.
	if strings.Contains(got, "transition-[grid-template-columns]") {
		t.Error("the rail must not animate grid-template-columns")
	}
}

func TestAppShellDark(t *testing.T) {
	got := render(t, ui.AppShell(ui.WithSidebar(shellNav()), ui.SidebarDark()))
	wantAll(t, got, `dark:text-white dark scheme-dark"`)
	if !regexp.MustCompile(`<dialog id="shellMenu-dialog"[^>]*class="[^"]*\bdark scheme-dark"`).MatchString(got) {
		t.Error("the drawer of a dark sidebar must be dark too")
	}
}

func TestNavItemStates(t *testing.T) {
	plain := render(t, ui.Navbar(htemel.Text("x"), ui.NavLink("A", "/a", false)))
	if strings.Contains(plain, `<a href="/a" title`) || strings.Contains(plain, `/a"><span`) {
		t.Errorf("a plain NavLink must render as before: %s", plain)
	}
	anc := render(t, ui.Sidebar("s", ui.SidebarGroup("g", "G", ui.NavLink("Posts", "/p", false).Ancestor())))
	wantAll(t, anc, `data-ancestor`, `data-ancestor:font-semibold`)
	if strings.Contains(anc, `aria-current`) {
		t.Error("an ancestor is not the current page")
	}
	badge := render(t, ui.Sidebar("s", ui.SidebarGroup("g", "G", ui.NavLink("Drafts", "/d", true).Badge("3"))))
	wantAll(t, badge, `aria-current="page"`, `>Drafts</span><span class="ml-auto`, `>3</span>`)
}

func TestSplitLayout(t *testing.T) {
	got := render(t, ui.SplitLayout(html.P().Text("L"), html.P().Text("D")))
	wantAll(t, got,
		`data-signals:split__ifmissing="'list'"`,
		`id="split-list" tabindex="-1"`, `id="split-detail" tabindex="-1"`,
		`data-class="{'@max-2xl:hidden': $split === 'detail'}"`,
		`data-class="{'@max-2xl:hidden': $split !== 'detail'}"`,
		`@2xl:grid-cols-[auto_minmax(0,1fr)]`, `@2xl:w-80`,
	)
	sel := render(t, html.Button().Attrs(ui.SplitSelect("split", "@get('/n/1')")))
	wantAll(t, sel, `$split = 'detail'; @get('/n/1'); requestAnimationFrame`, `d.focus({preventScroll: true})`)
	head := render(t, ui.SplitDetailHeaderSignal("pane", "Note"))
	wantAll(t, head, `$pane = 'list'`, `getElementById('pane-list')`, `@2xl:hidden`, `<h2 class="min-w-0 flex-1 truncate`)
}

func TestPageHeaderLevels(t *testing.T) {
	if got := render(t, ui.PageHeader("T")); !strings.Contains(got, "<h1 ") {
		t.Errorf("default level: %s", got)
	}
	if got := render(t, ui.PageHeader("T", ui.WithHeadingLevel(3))); !strings.Contains(got, "<h3 ") {
		t.Errorf("level 3: %s", got)
	}
	got := render(t, ui.PageHeader("T", ui.WithBackLink("/posts", "Posts")))
	wantAll(t, got, `<a href="/posts"`, `</svg>Posts</a>`)
}

func TestLayoutNavLabels(t *testing.T) {
	wantAll(t, render(t, ui.SettingsLayout([]ui.NavItem{ui.NavLink("P", "/p", true)})), `<nav aria-label="Settings"`, `aria-current="page"`)
	wantAll(t, render(t, ui.StackedTabs("Post sections", ui.NavLink("O", "/o", true))), `<nav aria-label="Post sections"`)
}
