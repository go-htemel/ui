package main

import (
	"cmp"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
	dsse "github.com/starfederation/datastar-go/datastar"

	"github.com/go-htemel/ui"
)

// The people table demo: a fixed list that the server sorts and pages. The sortable headers and
// the pager set the sort, sortDir and page signals and GET /demo/people; the handler answers by
// patching the table body and the pager.

const (
	peoplePageSize = 5
	peopleBodyID   = "people-body"
	peoplePagerID  = "people-pager"
	peopleURL      = "/demo/people"
)

type person struct{ Name, Role, Team string }

var demoPeople = func() []person {
	first := []string{"Ada", "Grace", "Linus", "Margaret", "Dennis", "Barbara", "Ken", "Frances"}
	last := []string{"Lovelace", "Hopper", "Torvalds", "Hamilton", "Ritchie"}
	roles := []string{"Engineer", "Designer", "Manager", "Analyst"}
	teams := []string{"Platform", "Growth", "Core", "Design"}
	var out []person
	for i := range 40 {
		out = append(out, person{
			Name: first[i%len(first)] + " " + last[(i/2+i)%len(last)],
			Role: roles[(i*3)%len(roles)],
			Team: teams[(i*5+i/4)%len(teams)],
		})
	}
	return out
}()

// peopleQuery is what the people table asks for: the signals the demo sends. Q and Team filter
// (the content inputs page's FilterBar); the data page leaves them empty.
type peopleQuery struct {
	Q       string `json:"q"`
	Team    string `json:"team"`
	Sort    string `json:"sort"`
	SortDir string `json:"sortDir"`
	Page    int    `json:"page"`
}

// run filters the people by Q (a case-insensitive substring of the name, role or team) and Team
// (exact), sorts them by Sort ("name", "role" or "team"; anything else keeps the original order)
// in direction SortDir ("desc" or ascending), and returns the requested page with the page
// number clamped into range and the page count (at least 1, so an empty result has one empty
// page).
func (q peopleQuery) run() (rows []person, current, pages int) {
	needle := strings.ToLower(strings.TrimSpace(q.Q))
	var all []person
	for _, p := range demoPeople {
		if q.Team != "" && p.Team != q.Team {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(p.Name+" "+p.Role+" "+p.Team), needle) {
			continue
		}
		all = append(all, p)
	}
	field := map[string]func(person) string{
		"name": func(p person) string { return p.Name },
		"role": func(p person) string { return p.Role },
		"team": func(p person) string { return p.Team },
	}[q.Sort]
	if field != nil {
		slices.SortStableFunc(all, func(a, b person) int {
			c := cmp.Compare(strings.ToLower(field(a)), strings.ToLower(field(b)))
			if q.SortDir == "desc" {
				return -c
			}
			return c
		})
	}
	pages = max((len(all)+peoplePageSize-1)/peoplePageSize, 1)
	current = min(max(q.Page, 1), pages)
	start := (current - 1) * peoplePageSize
	return all[start:min(start+peoplePageSize, len(all))], current, pages
}

// peoplePage is run for an unfiltered query.
func peoplePage(key, dir string, page int) (rows []person, current, pages int) {
	return peopleQuery{Sort: key, SortDir: dir, Page: page}.run()
}

// peopleBody renders the rows of one page, with the id the handler patches.
func peopleBody(q peopleQuery) *html.TbodyElement {
	rows, _, _ := q.run()
	body := ui.TableBody().Id(peopleBodyID)
	for _, p := range rows {
		body.Children(html.Tr(
			ui.Td(htemel.Text(p.Name)),
			ui.Td(htemel.Text(p.Role)),
			ui.Td(ui.Badge(ui.Brand, p.Team)),
		))
	}
	if len(rows) == 0 {
		body.Children(ui.EmptyRow(3, ui.EmptyState(ui.IconSearch(), "No people match", "Try another search or team.", nil)))
	}
	return body
}

func peoplePager(q peopleQuery) *html.NavElement {
	_, current, pages := q.run()
	return ui.PaginationSignal(current, pages, "page", datastar.Get(peopleURL), "People pages").Id(peoplePagerID)
}

// peopleHandler serves GET /demo/people: the signals arrive in the datastar query parameter.
func peopleHandler(w http.ResponseWriter, r *http.Request) {
	var in peopleQuery
	if err := dsse.ReadSignals(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, page, _ := in.run()
	requested := in.Page
	in.Page = page
	sse := dsse.NewSSE(w, r)
	patch(sse, peopleBody(in))
	patch(sse, peoplePager(in))
	if page != requested {
		if err := sse.MarshalAndPatchSignals(map[string]any{"page": page}); err != nil {
			slog.Error("patch signals", "err", err)
		}
	}
}

// progressHandler serves POST /demo/progress: it adds 25 to the upload signal and patches it
// back, and the ProgressSignal bar follows.
func progressHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Upload int `json:"upload"`
	}
	if err := dsse.ReadSignals(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	next := min(in.Upload+25, 100)
	if err := dsse.NewSSE(w, r).MarshalAndPatchSignals(map[string]any{"upload": next}); err != nil {
		slog.Error("patch signals", "err", err)
	}
}
