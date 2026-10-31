package ui_test

import (
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

func init() {
	crumbs := []ui.Crumb{{Label: "Home", Href: "/"}, {Label: "A", Href: "/a"}, {Label: "B", Href: "/b"}, {Label: "C", Href: "/c"}, {Label: "D"}}
	tabs := []ui.NavItem{
		ui.NavLink("Profile", "#profile", true).Signal("section", "profile"),
		ui.NavLink("Billing", "#billing", false).Signal("section", "billing"),
	}
	addGolden(
		golden{"link-button-primary-icon", ui.LinkButton(ui.Primary, "New post", "/posts/new", ui.WithIcon(ui.IconPlus()))},
		golden{"link-button-secondary-trailing", ui.LinkButton(ui.Secondary, "Next", "/next", ui.WithSize(ui.Sm), ui.WithTrailingIcon(ui.IconChevronRight()))},
		golden{"icon-link-button", ui.IconLinkButton(ui.Ghost, ui.IconPencil(), "Edit", "/edit", ui.WithSize(ui.Sm))},
		golden{"button-rail-icon", ui.Button(ui.Ghost, "Ada", ui.WithIcon(ui.IconHome()), ui.WithTrailingIcon(ui.IconChevronDown()), ui.WithRailIcon())},
		golden{"card-header-level-2", ui.Card(ui.CardHeader("Publish", ui.Badge(ui.Info, "Draft")).HeadingLevel(2))},
		golden{"sidebar-id-badge", ui.Sidebar("s", ui.SidebarGroup("g", "Content",
			ui.NavLink("Posts", "/posts", false).Icon(ui.IconDocument()).ID("nav-posts").Badge("12")))},
		golden{"nav-badge", ui.NavBadge("nav-posts", "13")},
		golden{"stacked-tabs", ui.StackedTabs("Sections", ui.NavLink("Overview", "/o", true), ui.NavLink("Comments", "/c", false))},
		golden{"stacked-tabs-signal", ui.StackedTabs("Sections", tabs...)},
		golden{"settings-layout-signal", ui.SettingsLayout(tabs, html.P().Text("Body"))},
		golden{"select-checkboxes", htemel.Group(
			ui.Checkbox("", ui.SelectAll("sel", []string{"p1", "p2"})...),
			ui.Checkbox("One", ui.SelectRow("sel", "p1")))},
		golden{"table-striped-dense-sticky", ui.Table(ui.TableHead(ui.Th(htemel.Text("A"))),
			ui.TableBody(html.Tr(ui.Td(htemel.Text("1")))), ui.Striped(), ui.Dense(), ui.StickyHeader())},
		golden{"breadcrumbs-collapsed", ui.BreadcrumbsCollapse(4, crumbs...)},
		golden{"breadcrumbs-collapse-short", ui.BreadcrumbsCollapse(4, crumbs[:3]...)},
		golden{"selectable-item", ui.SelectableItem("$id === 3", html.Span().Text("Groceries"))},
		golden{"split-list-item", ui.SplitListItem("split", "$id === 3", "$id = 3", html.Span().Text("Groceries"))},
		golden{"app-shell-rail-footer", ui.AppShell(ui.WithSidebar(html.P().Text("nav")),
			ui.WithSidebarFooter(ui.Button(ui.Ghost, "Ada", ui.WithIcon(ui.IconHome()), ui.WithRailIcon()), ui.ColorSchemeToggle()),
			ui.WithContent(html.P().Text("Body")), ui.SidebarCollapsible())},
	)
}

func TestLinkButtonSharesButtonClasses(t *testing.T) {
	class := func(n htemel.Node) string {
		s := render(t, n)
		i := strings.Index(s, `class="`) + len(`class="`)
		return s[i : i+strings.Index(s[i:], `"`)]
	}
	for _, v := range []ui.Variant{ui.Primary, ui.Secondary, ui.Outline, ui.Ghost, ui.Danger, ui.Link} {
		if a, b := class(ui.Button(v, "x")), class(ui.LinkButton(v, "x", "/")); a != b {
			t.Errorf("%s: LinkButton class differs from Button\n%s\n%s", v, a, b)
		}
		if a, b := class(ui.IconButton(v, ui.IconX(), "x")), class(ui.IconLinkButton(v, ui.IconX(), "x", "/")); a != b {
			t.Errorf("%s: IconLinkButton class differs from IconButton", v)
		}
	}
	got := render(t, ui.IconLinkButton(ui.Ghost, ui.IconX(), "Close", "/c"))
	for _, want := range []string{`<a `, `href="/c"`, `aria-label="Close"`, `title="Close"`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
	if strings.Contains(got, "role=") || strings.Contains(render(t, ui.LinkButton(ui.Primary, "x", "/")), "role=") {
		t.Error("link buttons must not carry a role")
	}
}

func TestSelectHelpers(t *testing.T) {
	got := render(t, ui.Checkbox("", ui.SelectAll("sel", []string{"p1", "p2"})...))
	for _, want := range []string{`el.indeterminate = n > 0 && n < ['p1', 'p2'].length`, `data-on:change=`, `$sel[k] = on`} {
		if !strings.Contains(got, want) {
			t.Errorf("SelectAll: missing %q in %s", want, got)
		}
	}
	if got := render(t, ui.Checkbox("", ui.SelectRow("sel", "p1"))); !strings.Contains(got, `data-bind:sel.p1`) {
		t.Errorf("SelectRow: %s", got)
	}
	if ui.SelectedCount("sel") != "Object.values($sel).filter(Boolean).length" || !strings.Contains(ui.SelectClear("sel"), "$sel[k] = false") {
		t.Error("SelectedCount / SelectClear")
	}
}

func TestThen(t *testing.T) {
	if got, want := ui.Then(datastar.Post("/x"), "a()"), datastar.Post("/x")+".then(() => { a() })"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestNavSignalAndSelect(t *testing.T) {
	tabs := []ui.NavItem{ui.NavLink("A", "#a", true).Signal("sec", "a"), ui.NavLink("B", "#b", false).Signal("sec", "b")}
	got := render(t, ui.StackedTabs("Sections", tabs...))
	for _, want := range []string{`data-on:click__prevent="$sec = 'a'"`, `data-attr:aria-current="$sec === 'b' ? 'page' : false"`,
		`data-signals:sec__ifmissing="'a'"`, `<option value="b">B</option>`, `$sec = o.value`, `@xl:hidden`, `class="@container"`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
	if href := render(t, ui.StackedTabs("Sections", ui.NavLink("A", "/a", true))); !strings.Contains(href, `window.location.href = evt.target.value`) {
		t.Errorf("href tabs lack the navigating select: %s", href)
	}
}

func TestNavBadgeID(t *testing.T) {
	got := render(t, ui.Sidebar("s", ui.SidebarGroup("g", "G", ui.NavLink("Posts", "/p", false).ID("nav-posts").Badge("3"))))
	if !strings.Contains(got, `<a href="/p" id="nav-posts"`) && !strings.Contains(got, `id="nav-posts"`) || !strings.Contains(got, `id="nav-posts-badge"`) {
		t.Errorf("ids missing: %s", got)
	}
	if strings.Contains(render(t, ui.Sidebar("s", ui.SidebarGroup("g", "G", ui.NavLink("Posts", "/p", false).Badge("3")))), "-badge") {
		t.Error("a badge without ID must have no id")
	}
}

func TestBadgeNowrapAndCardLevel(t *testing.T) {
	if !strings.Contains(render(t, ui.Badge(ui.Info, "Very long label")), "whitespace-nowrap") {
		t.Error("Badge must not wrap")
	}
	got := render(t, ui.CardHeader("T").HeadingLevel(2).Attrs(htemel.Attr("data-x", "1")))
	if !strings.Contains(got, "<h2") || strings.Contains(got, "<h3") || !strings.Contains(got, `data-x="1"`) {
		t.Errorf("HeadingLevel: %s", got)
	}
	if !strings.Contains(render(t, ui.CardHeader("T")), "<h3") {
		t.Error("default level is 3")
	}
}

func TestTableClassesOnFrame(t *testing.T) {
	plain := render(t, ui.Table(ui.TableHead(ui.Th(htemel.Text("A"))), ui.TableBody(), ui.Striped(), ui.Dense()))
	for _, want := range []string{"ui-table-striped", "ui-table-dense"} {
		if !strings.Contains(plain, want) {
			t.Errorf("missing %s", want)
		}
	}
	// A body built on its own (as a server patch) is the same markup with or without options.
	if a, b := render(t, ui.TableBody()), render(t, ui.TableBody()); a != b || strings.Contains(a, "even:") {
		t.Errorf("TableBody carries option classes: %s", a)
	}
}

func TestBreadcrumbsCollapse(t *testing.T) {
	crumbs := []ui.Crumb{{Label: "Home", Href: "/"}, {Label: "A", Href: "/a"}, {Label: "B", Href: "/b"}, {Label: "C", Href: "/c"}, {Label: "D"}}
	got := render(t, ui.BreadcrumbsCollapse(4, crumbs...))
	for _, want := range []string{`aria-label="Show 2 more pages"`, `href="/a"`, `href="/b"`, `aria-current="page"`, `<details`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Count(got, `<li class="flex items-center`) != 4 {
		t.Errorf("want 4 items (Home, ellipsis, C, D): %s", got)
	}
	if strings.Contains(render(t, ui.BreadcrumbsCollapse(5, crumbs...)), "<details") {
		t.Error("5 crumbs within limit 5 must not collapse")
	}
}

func TestSelectableItem(t *testing.T) {
	got := render(t, ui.SplitListItem("split", "$id === 3", "$id = 3", html.Span().Text("x")))
	for _, want := range []string{`data-attr:aria-current="($id === 3) ? 'true' : false"`, `aria-[current=true]:bg-primary-50`,
		`aria-[current=true]:border-primary-600`, `dark:aria-[current=true]:bg-primary-500/10`, `dark:aria-[current=true]:border-primary-400`, `data-on:click=`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
}

func TestRailFooterStaysVisible(t *testing.T) {
	got := render(t, ui.AppShell(ui.WithSidebar(html.P().Text("n")), ui.WithSidebarFooter(ui.RailLabel(htemel.Text("Ada"))), ui.SidebarCollapsible()))
	if strings.Contains(got, "in-data-ui-rail:hidden") && strings.Contains(got, `border-t border-neutral-200 px-3 py-3 dark:border-white/10 @3xl:in-data-ui-rail:hidden`) {
		t.Error("the footer must not be hidden in the rail")
	}
	if !strings.Contains(got, "@3xl:in-data-ui-rail:sr-only") {
		t.Error("RailLabel should hide its text in the rail")
	}
}
