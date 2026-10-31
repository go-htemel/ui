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

// peoplePage sorts the people by key ("name", "role" or "team"; anything else keeps the
// original order) in direction dir ("desc" or ascending), and returns the requested page with
// the page number clamped into range and the page count.
func peoplePage(key, dir string, page int) (rows []person, current, pages int) {
	all := slices.Clone(demoPeople)
	field := map[string]func(person) string{
		"name": func(p person) string { return p.Name },
		"role": func(p person) string { return p.Role },
		"team": func(p person) string { return p.Team },
	}[key]
	if field != nil {
		slices.SortStableFunc(all, func(a, b person) int {
			c := cmp.Compare(strings.ToLower(field(a)), strings.ToLower(field(b)))
			if dir == "desc" {
				return -c
			}
			return c
		})
	}
	pages = (len(all) + peoplePageSize - 1) / peoplePageSize
	current = min(max(page, 1), pages)
	start := (current - 1) * peoplePageSize
	return all[start : start+peoplePageSize], current, pages
}

// peopleBody renders the rows of one page, with the id the handler patches.
func peopleBody(key, dir string, page int) *html.TbodyElement {
	rows, _, _ := peoplePage(key, dir, page)
	body := ui.TableBody().Id(peopleBodyID)
	for _, p := range rows {
		body.Children(html.Tr(
			ui.Td(htemel.Text(p.Name)),
			ui.Td(htemel.Text(p.Role)),
			ui.Td(ui.Badge(ui.Brand, p.Team)),
		))
	}
	return body
}

func peoplePager(page int) *html.NavElement {
	_, current, pages := peoplePage("", "", page)
	return ui.PaginationSignal(current, pages, "page", datastar.Get(peopleURL), "People pages").Id(peoplePagerID)
}

// peopleHandler serves GET /demo/people: the signals arrive in the datastar query parameter.
func peopleHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Sort    string `json:"sort"`
		SortDir string `json:"sortDir"`
		Page    int    `json:"page"`
	}
	if err := dsse.ReadSignals(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, page, _ := peoplePage(in.Sort, in.SortDir, in.Page)
	sse := dsse.NewSSE(w, r)
	patch(sse, peopleBody(in.Sort, in.SortDir, page))
	patch(sse, peoplePager(page))
	if page != in.Page {
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
