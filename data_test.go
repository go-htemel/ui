package ui_test

import (
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

var (
	_ htemel.FlowContent = ui.Table(ui.TableHead(), ui.TableBody())
	_ htemel.FlowContent = ui.EmptyState(nil, "x", "", nil)
	_ htemel.FlowContent = ui.Pagination(1, 1, nil)
	_ htemel.FlowContent = ui.PaginationSignal(1, 1, "p", "")
	_ htemel.FlowContent = ui.Stat("a", "b")
	_ htemel.FlowContent = ui.StatGrid()
	_ htemel.FlowContent = ui.DescriptionList()
	_ htemel.FlowContent = ui.InputGroup(nil, ui.Input(html.InputTypeText, "x"), nil)

	// Only <tr> elements fit in a body and only header cells in a header; th and td are row
	// children, nothing else is.
	_ []*html.TrElement = []*html.TrElement{ui.EmptyRow(1, ui.EmptyState(nil, "x", "", nil))}
	_ html.TrChild      = ui.Th()
	_ html.TrChild      = ui.Td()
	_ html.TrChild      = ui.SortableTh("a", "a", "sort", "")
)

func people() (*html.TheadElement, *html.TbodyElement) {
	return ui.TableHead(
			ui.Th(htemel.Text("Name")),
			ui.SortableTh("Role", "role", "sort", datastar.Get("/people")),
		), ui.TableBody(
			html.Tr(ui.Td(htemel.Text("Ada")), ui.Td(htemel.Text("Engineer"))),
			html.Tr(ui.Td(htemel.Text("Grace")), ui.Td(htemel.Text("Admiral"))),
		)
}

func pageHref(p int) string { return "/people?page=" + string(rune('0'+p%10)) }

func init() {
	head, body := people()
	head2, body2 := people()
	head3, body3 := people()
	head4, _ := people()
	addGolden(
		golden{"table", ui.Table(head, body)},
		golden{"table-striped-dense", ui.Table(head2, body2, ui.Striped(), ui.Dense(), ui.TableCaption("People"))},
		golden{"table-sticky", ui.Table(head3, body3, ui.StickyHeader())},
		golden{"table-empty", ui.Table(head4, ui.TableBody(ui.EmptyRow(2,
			ui.EmptyState(ui.IconSearch(), "No people found", "Try another search.",
				ui.Button(ui.Secondary, "Clear filters")))))},
		golden{"empty-state", ui.EmptyState(nil, "Nothing here", "", nil)},
		golden{"sortable-th", ui.SortableTh("Name", "name", "sort", datastar.Get("/people"))},
		golden{"pagination-first", ui.Pagination(1, 3, pageHref)},
		golden{"pagination-middle", ui.Pagination(9, 20, pageHref)},
		golden{"pagination-last", ui.Pagination(8, 8, pageHref)},
		golden{"pagination-signal", ui.PaginationSignal(5, 12, "page", datastar.Get("/people"))},
		golden{"pagination-signal-first-nolink", ui.PaginationSignal(1, 1, "page", "")},
		golden{"stat", ui.Stat("Revenue", "$45,231", "+12.5%")},
		golden{"stat-grid", ui.StatGrid(
			ui.Stat("Users", "1,204", "+3%"), ui.Stat("Churn", "2.1%", "-0.4%"), ui.Stat("Plan", "Pro"))},
		golden{"description-list", ui.DescriptionList(
			ui.DLText("Name", "Ada Lovelace"),
			ui.DLItem{Term: "Status", Detail: ui.Badge(ui.Success, "Active")})},
		golden{"input-group", ui.InputGroup(htemel.Text("https://"), ui.Input(html.InputTypeText, "site"), htemel.Text(".com"))},
		golden{"input-group-prefix", ui.InputGroup(htemel.Text("$"), ui.Input(html.InputTypeText, "price"), nil)},
	)
}

func renderStr(t *testing.T, n htemel.Node) string {
	t.Helper()
	s, err := htemel.RenderString(n)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSortableThExpression(t *testing.T) {
	th := ui.SortableTh("Name", "name", "sort", datastar.Get("/people"))
	got := renderStr(t, th)
	for _, want := range []string{
		`scope="col"`, `aria-sort="none"`, `data-signals:sort__ifmissing=`, `data-signals:sort-dir__ifmissing=`,
		`data-attr:aria-sort=`, `@get(&#39;/people&#39;)`, `$sortDir = $sortDir === &#39;asc&#39;`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("SortableTh output lacks %s:\n%s", want, got)
		}
	}
	if ui.SortDirSignal("sort") != "sortDir" {
		t.Error("SortDirSignal")
	}
}

func TestPaginationAria(t *testing.T) {
	got := renderStr(t, ui.Pagination(2, 3, pageHref))
	if n := strings.Count(got, `aria-current="page"`); n != 1 {
		t.Errorf("want one aria-current=page, got %d", n)
	}
	if !strings.Contains(got, `aria-label="Pagination"`) || !strings.Contains(got, `rel="prev"`) || !strings.Contains(got, `rel="next"`) {
		t.Errorf("unexpected output: %s", got)
	}
	first := renderStr(t, ui.Pagination(1, 3, pageHref))
	if strings.Contains(first, `rel="prev"`) || !strings.Contains(first, `aria-disabled="true"`) {
		t.Errorf("previous should be disabled on page 1: %s", first)
	}
	// Out-of-range input is clamped rather than rendering nothing.
	if got := renderStr(t, ui.Pagination(99, 0, pageHref)); !strings.Contains(got, `aria-current="page"`) {
		t.Errorf("clamped pagination lost its current page: %s", got)
	}
}
