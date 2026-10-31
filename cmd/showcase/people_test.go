package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// checkTablePatch is checkFragment for a <tbody> patch. Datastar parses patches inside a
// <template>, which accepts table parts; a plain <body> context would drop them.
func checkTablePatch(t *testing.T, label, frag string, allowed map[string]bool) {
	t.Helper()
	table := &xhtml.Node{Type: xhtml.ElementNode, Data: "table", DataAtom: atom.Table}
	nodes, err := xhtml.ParseFragment(strings.NewReader(frag), table)
	if err != nil {
		t.Fatalf("%s: parse: %v", label, err)
	}
	compareTags(t, label, frag, nodes, allowed)
}

// getSignals GETs path the way Datastar does: the signals in the datastar query parameter.
func getSignals(t *testing.T, path, signals string) (code int, body string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path+"?datastar="+url.QueryEscape(signals), nil)
	req.Header.Set("Datastar-Request", "true")
	rec := httptest.NewRecorder()
	routes().ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func TestPeopleSortAndPage(t *testing.T) {
	allowed := datastarAttributes(t)

	code, body := getSignals(t, "/demo/people", `{"sort":"name","sortDir":"desc","page":2}`)
	if code != http.StatusOK {
		t.Fatalf("status %d: %s", code, body)
	}
	var patches []string
	for _, line := range strings.Split(body, "\n") {
		if s, ok := strings.CutPrefix(line, "data: elements "); ok {
			patches = append(patches, s)
		}
	}
	if len(patches) != 2 {
		t.Fatalf("want a body and a pager patch, got %d: %s", len(patches), body)
	}
	checkTablePatch(t, "people body patch", patches[0], allowed)
	checkFragment(t, "people pager patch", patches[1], allowed)
	if !strings.Contains(patches[0], `id="people-body"`) || !strings.Contains(patches[1], `id="people-pager"`) {
		t.Errorf("patches lack their ids: %v", patches)
	}

	// Descending by name: the page-2 rows are all <= the page-1 rows, so the first row of page 1
	// is not on page 2.
	rows, page, pages := peoplePage("name", "desc", 2)
	first, _, _ := peoplePage("name", "desc", 1)
	if page != 2 || pages != 8 || len(rows) != peoplePageSize {
		t.Fatalf("page %d of %d with %d rows", page, pages, len(rows))
	}
	if rows[0].Name > first[len(first)-1].Name {
		t.Errorf("descending order broken: %q after %q", rows[0].Name, first[len(first)-1].Name)
	}
	if strings.Contains(patches[0], first[0].Name+"</td>") && first[0].Name != rows[0].Name {
		t.Errorf("page 2 shows a row of page 1")
	}
}

func TestPeoplePageIsClamped(t *testing.T) {
	_, body := getSignals(t, "/demo/people", `{"sort":"role","sortDir":"asc","page":99}`)
	if !strings.Contains(body, "datastar-patch-signals") || !strings.Contains(body, `"page":8`) {
		t.Errorf("an out-of-range page must be corrected through a signal patch: %s", body)
	}
	if code, _ := getSignals(t, "/demo/people", `not json`); code != http.StatusBadRequest {
		t.Errorf("bad signals: status %d, want 400", code)
	}
}

func TestProgressAdvances(t *testing.T) {
	events, _ := sse(t, routes(), "/demo/progress", `{"upload":90}`)
	if len(events) != 1 || !strings.Contains(events[0], `"upload":100`) {
		t.Errorf("unexpected events %q", events)
	}
}
