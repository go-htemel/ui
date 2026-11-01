package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/go-htemel/ui"
)

// fixedNow is the clock of the store tests.
var fixedNow = time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)

func testStore() *adminStore { return newAdminStore(func() time.Time { return fixedNow }) }

// resetAdmin restores the shared stores after a test that changes them.
func resetAdmin(t *testing.T) {
	t.Helper()
	adminData.reset()
	adminNotes.reset()
	adminSettingsData.reset()
	t.Cleanup(func() {
		adminData.reset()
		adminNotes.reset()
		adminSettingsData.reset()
	})
}

func titles(rows []adminPost) []string {
	out := make([]string, len(rows))
	for i, p := range rows {
		out[i] = p.Title
	}
	return out
}

func TestSeed(t *testing.T) {
	s := testStore()
	c := s.counts()
	if c.Total != 30 || c.Published+c.Drafts+c.Scheduled != 30 {
		t.Fatalf("counts %+v, want 30 posts", c)
	}
	if c.Drafts == 0 || c.Scheduled == 0 || c.Published == 0 || c.PublishedThisMonth == 0 || c.Comments == 0 {
		t.Errorf("the seed lacks variety: %+v", c)
	}
	seen := map[string]bool{}
	for _, p := range s.recent(100) {
		if seen[p.Slug] || p.Slug != slugify(p.Title) {
			t.Errorf("post %d: slug %q duplicated or not the slug of %q", p.ID, p.Slug, p.Title)
		}
		seen[p.Slug] = true
		if (p.Status == statusScheduled) != (p.PublishAt != "") {
			t.Errorf("post %d: status %s with PublishAt %q", p.ID, p.Status, p.PublishAt)
		}
	}
	// Deterministic: the same clock gives the same posts.
	a, b := s.recent(30), testStore().recent(30)
	for i := range a {
		if a[i].ID != b[i].ID || !a[i].UpdatedAt.Equal(b[i].UpdatedAt) || a[i].Body != b[i].Body {
			t.Fatalf("seed differs between two stores at %d", i)
		}
	}
	if len(adminData.tags()) < 5 {
		t.Errorf("tags: %v", adminData.tags())
	}
}

func TestPostsRun(t *testing.T) {
	s := testStore()
	all := func(q postQuery) []adminPost {
		var out []adminPost
		for p := 1; ; p++ {
			q.Page = p
			rows, cur, pages, _ := s.run(q)
			out = append(out, rows...)
			if cur >= pages {
				return out
			}
		}
	}
	tests := []struct {
		name  string
		q     postQuery
		check func(t *testing.T, rows []adminPost, cur, pages, total int)
	}{
		{"default is newest first, ten a page", postQuery{Page: 1}, func(t *testing.T, rows []adminPost, cur, pages, total int) {
			if len(rows) != 10 || cur != 1 || pages != 3 || total != 30 {
				t.Fatalf("got %d rows, page %d of %d, total %d", len(rows), cur, pages, total)
			}
			for i := 1; i < len(rows); i++ {
				if rows[i].UpdatedAt.After(rows[i-1].UpdatedAt) {
					t.Errorf("row %d is newer than row %d", i, i-1)
				}
			}
		}},
		{"title ascending", postQuery{Sort: "title", SortDir: "asc", Page: 1}, func(t *testing.T, rows []adminPost, _, _, _ int) {
			if rows[0].Title != "A tour of the new AppShell" || rows[1].Title != "Accessible tables without the pain" {
				t.Errorf("first titles: %v", titles(rows[:2]))
			}
		}},
		{"title descending", postQuery{Sort: "title", SortDir: "desc", Page: 1}, func(t *testing.T, rows []adminPost, _, _, _ int) {
			if rows[0].Title != "Writing a CMS in Go" {
				t.Errorf("first title %q", rows[0].Title)
			}
		}},
		{"updated ascending is oldest first", postQuery{Sort: "updated", SortDir: "asc", Page: 1}, func(t *testing.T, rows []adminPost, _, _, _ int) {
			if rows[0].ID != 30 {
				t.Errorf("oldest is post %d, want 30", rows[0].ID)
			}
		}},
		{"status ascending: published first", postQuery{Sort: "status", SortDir: "asc", Page: 1}, func(t *testing.T, rows []adminPost, _, _, _ int) {
			if rows[0].Status != statusPublished {
				t.Errorf("first status %s", rows[0].Status)
			}
			last := all(postQuery{Sort: "status", SortDir: "asc"})
			if last[len(last)-1].Status != statusDraft {
				t.Errorf("last status %s, want draft", last[len(last)-1].Status)
			}
		}},
		{"status filter", postQuery{Status: statusDraft, Page: 1}, func(t *testing.T, rows []adminPost, _, _, total int) {
			if total != s.counts().Drafts {
				t.Errorf("%d matches, want %d drafts", total, s.counts().Drafts)
			}
			for _, p := range rows {
				if p.Status != statusDraft {
					t.Errorf("%q is %s", p.Title, p.Status)
				}
			}
		}},
		{"tag filter", postQuery{Tag: "go", Page: 1}, func(t *testing.T, rows []adminPost, _, _, total int) {
			if total == 0 {
				t.Fatal("no post tagged go")
			}
			for _, p := range all(postQuery{Tag: "go"}) {
				found := false
				for _, tg := range p.Tags {
					found = found || tg == "go"
				}
				if !found {
					t.Errorf("%q has tags %v", p.Title, p.Tags)
				}
			}
		}},
		{"search matches title, slug and author case-insensitively", postQuery{Q: "  DATASTAR ", Page: 1}, func(t *testing.T, rows []adminPost, _, _, total int) {
			if total == 0 {
				t.Fatal("no match")
			}
			for _, p := range rows {
				if !strings.Contains(strings.ToLower(p.Title+p.Slug+p.Author), "datastar") {
					t.Errorf("%q does not match", p.Title)
				}
			}
		}},
		{"search by author", postQuery{Q: "hopper", Page: 1}, func(t *testing.T, rows []adminPost, _, _, total int) {
			if total == 0 {
				t.Fatal("no match")
			}
			for _, p := range rows {
				if p.Author != "Grace Hopper" {
					t.Errorf("%q is by %s", p.Title, p.Author)
				}
			}
		}},
		{"filters combine", postQuery{Q: "post", Status: statusPublished, Tag: "writing", Page: 1}, func(t *testing.T, rows []adminPost, _, _, _ int) {
			for _, p := range rows {
				if p.Status != statusPublished {
					t.Errorf("%q is %s", p.Title, p.Status)
				}
			}
		}},
		{"page two", postQuery{Page: 2}, func(t *testing.T, rows []adminPost, cur, _, _ int) {
			first, _, _, _ := s.run(postQuery{Page: 1})
			if cur != 2 || len(rows) != 10 || rows[0].ID == first[0].ID {
				t.Errorf("page %d, %d rows, first id %d", cur, len(rows), rows[0].ID)
			}
		}},
		{"page past the end is clamped", postQuery{Page: 99}, func(t *testing.T, rows []adminPost, cur, pages, _ int) {
			if cur != 3 || pages != 3 || len(rows) != 10 {
				t.Errorf("page %d of %d with %d rows", cur, pages, len(rows))
			}
		}},
		{"page below one is clamped", postQuery{Page: -4}, func(t *testing.T, _ []adminPost, cur, _, _ int) {
			if cur != 1 {
				t.Errorf("page %d", cur)
			}
		}},
		{"nothing matches: one empty page", postQuery{Q: "zzzz", Page: 3}, func(t *testing.T, rows []adminPost, cur, pages, total int) {
			if len(rows) != 0 || cur != 1 || pages != 1 || total != 0 {
				t.Errorf("%d rows, page %d of %d, total %d", len(rows), cur, pages, total)
			}
		}},
		{"unknown sort keeps the default order", postQuery{Sort: "nope", SortDir: "asc", Page: 1}, func(t *testing.T, rows []adminPost, _, _, _ int) {
			def, _, _, _ := s.run(postQuery{Page: 1})
			if titles(rows)[0] != titles(def)[0] {
				t.Errorf("first %q, default first %q", rows[0].Title, def[0].Title)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, cur, pages, total := s.run(tt.q)
			tt.check(t, rows, cur, pages, total)
		})
	}
}

func TestStoreMutations(t *testing.T) {
	s := testStore()
	p, ok := s.save(adminPost{Title: "New", Slug: "new", Status: statusDraft, Author: adminAuthors[0]})
	if !ok || p.ID != 31 || !p.UpdatedAt.Equal(fixedNow) {
		t.Fatalf("create: %+v %v", p, ok)
	}
	p.Title = "Renamed"
	if q, ok := s.save(p); !ok || q.ID != 31 {
		t.Fatalf("update: %+v %v", q, ok)
	}
	if g, _ := s.get(31); g.Title != "Renamed" {
		t.Errorf("stored title %q", g.Title)
	}
	if _, ok := s.save(adminPost{ID: 999}); ok {
		t.Error("saving an unknown id succeeded")
	}
	d, ok := s.duplicate(1)
	if !ok || d.Title != "Container queries in practice (copy)" || d.Status != statusDraft || d.Slug != "container-queries-in-practice-copy" {
		t.Fatalf("duplicate: %+v", d)
	}
	d2, _ := s.duplicate(1)
	if d2.Slug != "container-queries-in-practice-copy-2" {
		t.Errorf("second duplicate slug %q", d2.Slug)
	}
	if n := s.delete(1, 2, 999); n != 2 {
		t.Errorf("deleted %d, want 2", n)
	}
	if _, ok := s.get(1); ok {
		t.Error("post 1 survived")
	}
	s.reset()
	if s.counts().Total != 30 {
		t.Error("reset did not restore the seed")
	}
}

func TestValidate(t *testing.T) {
	s := testStore()
	ok := postForm{Title: "A fine title", Slug: "a-fine-title", Status: statusPublished, Author: adminAuthors[1], Tags: []string{" go ", "Go", "css", ""}}
	tests := []struct {
		name string
		id   int
		mut  func(*postForm)
		want postErrors
	}{
		{"valid", 0, func(*postForm) {}, postErrors{}},
		{"title required", 0, func(f *postForm) { f.Title = "   " }, postErrors{Title: "Enter a title."}},
		{"title too long", 0, func(f *postForm) { f.Title = strings.Repeat("x", 121) }, postErrors{Title: "Use at most 120 characters."}},
		{"slug required", 0, func(f *postForm) { f.Slug = "" }, postErrors{Slug: "Enter a slug."}},
		{"slug shape", 0, func(f *postForm) { f.Slug = "Not A Slug" }, postErrors{Slug: "Use lowercase letters, numbers and single hyphens."}},
		{"slug unique", 0, func(f *postForm) { f.Slug = "writing-a-cms-in-go" }, postErrors{Slug: "Another post, “Writing a CMS in Go”, already uses this slug."}},
		{"slug unique ignoring case is a shape error first", 0, func(f *postForm) { f.Slug = "Writing-a-cms-in-go" }, postErrors{Slug: "Use lowercase letters, numbers and single hyphens."}},
		{"a post may keep its own slug", 2, func(f *postForm) { f.Slug = "writing-a-cms-in-go" }, postErrors{}},
		{"scheduled needs a date", 0, func(f *postForm) { f.Status = statusScheduled }, postErrors{PublishAt: "Choose the date and time to publish."}},
		{"scheduled with a date", 0, func(f *postForm) { f.Status, f.PublishAt = statusScheduled, "2026-12-01T09:30" }, postErrors{}},
		{"two errors", 0, func(f *postForm) { f.Title, f.Slug = "", "" }, postErrors{Title: "Enter a title.", Slug: "Enter a slug."}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := ok
			tt.mut(&f)
			got, _ := s.validate(tt.id, f)
			if got != tt.want {
				t.Errorf("errors %+v, want %+v", got, tt.want)
			}
		})
	}
	_, p := s.validate(0, ok)
	if strings.Join(p.Tags, ",") != "go,css" {
		t.Errorf("tags %v, want trimmed and de-duplicated", p.Tags)
	}
	_, p = s.validate(0, postForm{Title: "x", Slug: "x", Status: "bogus", Author: "Nobody", PublishAt: "2026-12-01T09:30"})
	if p.Status != statusDraft || p.Author != adminAuthors[0] || p.PublishAt != "" {
		t.Errorf("normalised post %+v", p)
	}
}

// adminPaths are the full-page routes of the demo and the nav item each one marks current ("" for
// none).
var adminPaths = []struct{ path, nav string }{
	{"/admin", "Dashboard"},
	{"/admin/posts", "Posts"},
	{"/admin/posts/new", ""},
	{"/admin/posts/1/edit", ""},
	{"/admin/notes", "Notes"},
	{"/admin/settings", "Settings"},
	{"/admin/settings/writing", "Settings"},
	{"/admin/settings/appearance", "Settings"},
	{"/admin/media", "Media"},
	{"/admin/comments", "Comments"},
}

var inlineScript = regexp.MustCompile(`<script(?:\s[^>]*)?>`)

func TestAdminPages(t *testing.T) {
	allowed := datastarAttributes(t)
	h := routes()
	for _, p := range adminPaths {
		t.Run(p.path, func(t *testing.T) {
			code, body := get(t, h, p.path)
			if code != http.StatusOK {
				t.Fatalf("status %d", code)
			}
			checkDocument(t, p.path, body, allowed)
			boot := strings.Index(body, ui.ColorSchemeBootstrapSource())
			if boot < 0 || boot > strings.Index(body, tailwindJS) || boot > strings.Index(body, `rel="stylesheet"`) {
				t.Error("ui.ColorSchemeBootstrap is missing or comes after the Tailwind script or a stylesheet")
			}
			// Quality bar: the only inline script is the color scheme bootstrap.
			inline := 0
			for _, tag := range inlineScript.FindAllString(body, -1) {
				if !strings.Contains(tag, " src=") {
					inline++
				}
			}
			if inline != 1 {
				t.Errorf("%d inline scripts, want only ColorSchemeBootstrap", inline)
			}
			for _, want := range []string{`<main id="main"`, `href="/"`, "back to showcase", `aria-label="Color scheme"`, `id="toasts"`, datastarJS, `id="userMenu-trigger"`} {
				if !strings.Contains(body, want) {
					t.Errorf("page lacks %q", want)
				}
			}
			// Exactly one nav link (in the sidebar and again in the drawer copy) is the current page.
			// Only the nav links carry an icon, and so a title.
			var names []string
			for _, tag := range regexp.MustCompile(`<a [^>]*>`).FindAllString(body, -1) {
				if m := regexp.MustCompile(`title="([^"]*)"`).FindStringSubmatch(tag); m != nil && strings.Contains(tag, `aria-current="page"`) {
					names = append(names, m[1])
				}
			}
			switch {
			case p.nav == "" && len(names) != 0:
				t.Errorf("nav marks %v current, want none", names)
			case p.nav != "" && (len(names) != 2 || names[0] != p.nav || names[1] != p.nav):
				t.Errorf("nav marks %v current, want %s in the sidebar and the drawer", names, p.nav)
			}
		})
	}

	t.Run("editor ancestor and breadcrumbs", func(t *testing.T) {
		_, body := get(t, h, "/admin/posts/1/edit")
		for _, want := range []string{`data-ui-ancestor`, `aria-label="Breadcrumb"`, `All posts`, `Container queries in practice`, `id="post-saved"`, `id="post-status-badge"`} {
			if !strings.Contains(body, want) {
				t.Errorf("editor lacks %q", want)
			}
		}
	})
	t.Run("404", func(t *testing.T) {
		for _, p := range []string{"/admin/nope", "/admin/posts/999/edit", "/admin/settings/nope", "/admin/posts/abc/edit"} {
			code, body := get(t, h, p)
			if code != http.StatusNotFound {
				t.Errorf("%s: status %d, want 404", p, code)
			}
			if p != "/admin/posts/abc/edit" {
				checkDocument(t, p, body, allowed)
				if !strings.Contains(body, "Page not found") || !strings.Contains(body, `<main id="main"`) {
					t.Errorf("%s: not the EmptyState 404 inside the shell", p)
				}
			}
		}
	})
	t.Run("notices", func(t *testing.T) {
		_, body := get(t, h, "/admin/posts?notice=deleted")
		if !strings.Contains(body, "Post deleted") {
			t.Error("?notice=deleted shows no toast")
		}
		_, body = get(t, h, "/admin/posts?notice=%3Cimg%20src=x%3E")
		if strings.Contains(body, "<img src=x>") {
			t.Error("an unknown notice leaked into the page")
		}
	})
}

func TestShowcaseLinksToAdmin(t *testing.T) {
	_, body := get(t, routes(), "/")
	if !strings.Contains(body, `href="/admin"`) || !strings.Contains(body, "Admin demo") {
		t.Error("the showcase sidebar has no Admin demo link")
	}
}

func TestLayoutFrames(t *testing.T) {
	allowed := datastarAttributes(t)
	h := routes()
	for _, name := range []string{"appshell", "split", "settings"} {
		code, body := get(t, h, "/components/layouts/frame/"+name)
		if code != http.StatusOK {
			t.Fatalf("%s: status %d", name, code)
		}
		checkDocument(t, name, body, allowed)
		if !strings.Contains(body, "<main") || strings.Contains(body, `id="toasts"`) {
			t.Errorf("%s: want a bare page with a <main> and no toast region", name)
		}
	}
	if code, _ := get(t, h, "/components/layouts/frame/nope"); code != http.StatusNotFound {
		t.Errorf("unknown frame: status %d, want 404", code)
	}
	_, page := get(t, h, "/components/layouts")
	if n := strings.Count(page, `<iframe`); n != 3 || !strings.Contains(page, `loading="lazy"`) || !strings.Contains(page, `title="Phone-width app shell with a menu drawer"`) {
		t.Errorf("layouts page: %d iframes, lazy and titled expected", n)
	}
}

// sseGet and sseSend call an SSE handler (signals in the datastar query parameter for a GET, in
// the JSON body otherwise) and return its patch-elements HTML and its events.
func sseSend(t *testing.T, h http.Handler, method, path string, signals any) (events, elements []string, code int) {
	t.Helper()
	raw, err := json.Marshal(signals)
	if err != nil {
		t.Fatal(err)
	}
	var req *http.Request
	if method == http.MethodGet {
		req = httptest.NewRequest(method, path+"?datastar="+url.QueryEscape(string(raw)), nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(string(raw)))
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Datastar-Request", "true")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		return nil, nil, rec.Code
	}
	for _, ev := range strings.Split(strings.TrimSpace(rec.Body.String()), "\n\n") {
		if ev == "" {
			continue
		}
		events = append(events, ev)
		var el []string
		for _, line := range strings.Split(ev, "\n") {
			if s, ok := strings.CutPrefix(line, "data: elements "); ok {
				el = append(el, s)
			}
		}
		if len(el) > 0 {
			elements = append(elements, strings.Join(el, "\n"))
		}
	}
	return events, elements, rec.Code
}

func signalPatches(events []string) []string {
	var out []string
	for _, ev := range events {
		if strings.Contains(ev, "event: datastar-patch-signals") {
			out = append(out, ev)
		}
	}
	return out
}

// checkPatches checks each patched fragment parses to the tags it was written with. A table
// body or a header cell only parses inside its table or row, so those get that context.
func checkPatches(t *testing.T, label string, elements []string) {
	t.Helper()
	allowed := datastarAttributes(t)
	for i, el := range elements {
		name := label + " patch " + string(rune('0'+i))
		ctx := &xhtml.Node{Type: xhtml.ElementNode, Data: "body", DataAtom: atom.Body}
		switch {
		case strings.HasPrefix(el, "<tbody"):
			ctx = &xhtml.Node{Type: xhtml.ElementNode, Data: "table", DataAtom: atom.Table}
		case strings.HasPrefix(el, "<th"):
			ctx = &xhtml.Node{Type: xhtml.ElementNode, Data: "tr", DataAtom: atom.Tr}
		}
		nodes, err := xhtml.ParseFragment(strings.NewReader(el), ctx)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		compareTags(t, name, el, nodes, allowed)
	}
}

func TestPostsListHandler(t *testing.T) {
	h := routes()
	events, elements, code := sseSend(t, h, http.MethodGet, "/admin/posts/list",
		map[string]any{"q": "", "status": "draft", "tag": "", "sort": "title", "sortDir": "asc", "page": 1})
	if code != 200 || len(elements) != 6 {
		t.Fatalf("status %d, %d element patches, want the body, the select-all cell, the pager, the count and the two nav badges", code, len(elements))
	}
	if !strings.Contains(elements[4], `id="nav-posts-badge"`) || !strings.Contains(elements[5], `id="nav-posts-badge-drawer"`) {
		t.Errorf("nav badge patches: %s %s", elements[4], elements[5])
	}
	checkPatches(t, "list", elements)
	body, head, pager, count := elements[0], elements[1], elements[2], elements[3]
	for _, want := range []string{`id="posts-body"`, `Select “`, `Actions for “`, `Draft`} {
		if !strings.Contains(body, want) {
			t.Errorf("body patch lacks %q", want)
		}
	}
	if strings.Contains(body, ">Published</span>") || strings.Contains(body, ">Scheduled</span>") {
		t.Error("the draft filter let another status through")
	}
	if !strings.HasPrefix(head, `<th `) || !strings.Contains(head, `id="posts-selectall"`) || !strings.Contains(head, "Select all posts on this page") {
		t.Errorf("select-all patch: %s", head)
	}
	if !strings.Contains(pager, `id="posts-pager"`) || !strings.Contains(pager, `aria-label="Posts pages"`) {
		t.Errorf("pager patch: %s", pager)
	}
	if !strings.Contains(count, `id="posts-count"`) || !strings.Contains(count, "Showing 1–9 of 9 posts") {
		t.Errorf("count patch: %s", count)
	}
	// Alphabetical: a row title earlier in the alphabet comes first.
	if strings.Index(body, "Accessible tables") > strings.Index(body, "Writing a CMS in Go") {
		t.Error("rows are not sorted by title")
	}
	if len(signalPatches(events)) != 0 {
		t.Error("a page in range must not patch the page signal")
	}

	// An out-of-range page is clamped and the page signal follows.
	events, _, _ = sseSend(t, h, http.MethodGet, "/admin/posts/list", map[string]any{"page": 99, "sortDir": "desc"})
	sp := signalPatches(events)
	if len(sp) != 1 || !strings.Contains(sp[0], `"page":3`) {
		t.Errorf("signal patches %q, want page 3", sp)
	}

	// Nothing matches: an EmptyState row in the body and no pager arithmetic errors.
	_, elements, _ = sseSend(t, h, http.MethodGet, "/admin/posts/list", map[string]any{"q": "zzzz", "page": 1})
	checkPatches(t, "empty", elements)
	if !strings.Contains(elements[0], "No posts match") || !strings.Contains(elements[0], `colspan="6"`) || !strings.Contains(elements[3], "No posts") {
		t.Errorf("empty patches: %s | %s", elements[0], elements[3])
	}
	// Malformed signals are a 400.
	req := httptest.NewRequest(http.MethodGet, "/admin/posts/list?datastar=%7Bnope", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("bad signals: status %d, want 400", rec.Code)
	}
}

func TestPostsActions(t *testing.T) {
	resetAdmin(t)
	h := routes()
	base := map[string]any{"q": "", "status": "", "tag": "", "sort": "updated", "sortDir": "desc", "page": 1}
	with := func(extra map[string]any) map[string]any {
		m := map[string]any{}
		for k, v := range base {
			m[k] = v
		}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}

	// Duplicate adds a draft copy and re-renders the list.
	_, elements, code := sseSend(t, h, http.MethodPost, "/admin/posts/1/duplicate", with(nil))
	if code != 200 || len(elements) != 7 { // toast, body, select-all, pager, count, two nav badges
		t.Fatalf("duplicate: status %d, %d patches", code, len(elements))
	}
	checkPatches(t, "duplicate", elements)
	if adminData.counts().Total != 31 {
		t.Error("duplicate did not add a post")
	}
	if _, _, code := sseSend(t, h, http.MethodPost, "/admin/posts/abc/duplicate", with(nil)); code != http.StatusNotFound {
		t.Errorf("duplicate of /abc: status %d, want 404", code)
	}

	// Delete the post of the dialog: a toast, the list and a signal patch that unticks it.
	events, elements, _ := sseSend(t, h, http.MethodPost, "/admin/posts/delete", with(map[string]any{"delId": 2, "delTitle": "Writing a CMS in Go", "sel": map[string]bool{}}))
	checkPatches(t, "delete", elements)
	if _, ok := adminData.get(2); ok {
		t.Error("post 2 was not deleted")
	}
	if !strings.Contains(elements[0], "Post deleted") || len(signalPatches(events)) != 1 {
		t.Errorf("delete: toast %q, signal patches %d", elements[0], len(signalPatches(events)))
	}
	// Deleting it again says so instead of failing.
	_, elements, _ = sseSend(t, h, http.MethodPost, "/admin/posts/delete", with(map[string]any{"delId": 2}))
	if !strings.Contains(elements[0], "Already gone") {
		t.Errorf("second delete: %s", elements[0])
	}

	// Bulk delete removes every ticked post and unticks them.
	events, elements, _ = sseSend(t, h, http.MethodPost, "/admin/posts/bulk-delete", with(map[string]any{"sel": map[string]bool{"p3": true, "p4": true, "p5": false, "junk": true}}))
	checkPatches(t, "bulk", elements)
	if adminData.counts().Total != 31-1-2 {
		t.Errorf("total %d, want 28", adminData.counts().Total)
	}
	if last := elements[len(elements)-1]; !strings.Contains(last, "2 posts deleted") {
		t.Errorf("bulk toast: %s", last)
	}
	if sp := signalPatches(events); len(sp) != 1 || !strings.Contains(sp[0], `"p3":false`) {
		t.Errorf("bulk signal patch %q", sp)
	}
	if ids := selectedIDs(map[string]bool{"p10": true, "p2": true, "p3": false, "x4": true, "p": true}); len(ids) != 2 || ids[0] != 2 || ids[1] != 10 {
		t.Errorf("selectedIDs: %v", ids)
	}
}

func TestPostSave(t *testing.T) {
	resetAdmin(t)
	h := routes()
	form := func(extra map[string]any) map[string]any {
		m := map[string]any{"title": "Edited title", "slug": "edited-title", "body": "# Hi", "status": "published", "publishAt": "", "author": adminAuthors[2], "comments": false, "tags": []string{"go", "css"}}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}

	// Validation errors patch the errors signals, toast, and store nothing.
	events, elements, code := sseSend(t, h, http.MethodPost, "/admin/posts/3", form(map[string]any{"title": "", "slug": "container-queries-in-practice"}))
	if code != 200 {
		t.Fatalf("status %d", code)
	}
	sp := signalPatches(events)
	if len(sp) != 1 || !strings.Contains(sp[0], `"title":"Enter a title."`) || !strings.Contains(sp[0], "already uses this slug") {
		t.Errorf("error patch %q", sp)
	}
	if len(elements) != 1 || !strings.Contains(elements[0], "The post was not saved") || !strings.Contains(elements[0], "Fix the 2 highlighted fields") {
		t.Errorf("error toast %q", elements)
	}
	checkPatches(t, "errors", elements)
	if p, _ := adminData.get(3); p.Title == "" || p.Title == "Edited title" {
		t.Errorf("a failed save stored %q", p.Title)
	}

	// Success: errors cleared, the header's Saved line and status badge patched, a toast.
	events, elements, _ = sseSend(t, h, http.MethodPost, "/admin/posts/3", form(nil))
	checkPatches(t, "saved", elements)
	if len(elements) != 3 || !strings.Contains(elements[0], `id="post-saved"`) || !strings.Contains(elements[0], "Saved just now") ||
		!strings.Contains(elements[1], `id="post-status-badge"`) || !strings.Contains(elements[1], "Published") || !strings.Contains(elements[2], "Post saved") {
		t.Errorf("success patches: %q", elements)
	}
	if sp := signalPatches(events); len(sp) != 1 || strings.Contains(sp[0], "Enter") {
		t.Errorf("success must clear the errors: %q", sp)
	}
	p, _ := adminData.get(3)
	if p.Title != "Edited title" || p.Slug != "edited-title" || p.Author != adminAuthors[2] || p.Comments || strings.Join(p.Tags, ",") != "go,css" {
		t.Errorf("stored %+v", p)
	}

	// A new post is created, then the browser is sent to its editor.
	events, _, _ = sseSend(t, h, http.MethodPost, "/admin/posts/new", form(map[string]any{"title": "Brand new", "slug": "brand-new"}))
	joined := strings.Join(events, "\n")
	if !strings.Contains(joined, "/admin/posts/31/edit?notice=created") {
		t.Errorf("no redirect to the new post in %q", joined)
	}
	if _, ok := adminData.get(31); !ok {
		t.Error("the new post was not stored")
	}
	if _, _, code := sseSend(t, h, http.MethodPost, "/admin/posts/999", form(nil)); code != http.StatusNotFound {
		t.Errorf("unknown post: status %d, want 404", code)
	}

	// The editor's delete: a redirect to the list with a notice.
	events, _, _ = sseSend(t, h, http.MethodDelete, "/admin/posts/31", map[string]any{})
	if !strings.Contains(strings.Join(events, "\n"), "/admin/posts?notice=deleted") {
		t.Errorf("delete events %q", events)
	}
	if _, ok := adminData.get(31); ok {
		t.Error("post 31 survived DELETE")
	}
}

func TestPreviewHandler(t *testing.T) {
	h := routes()
	events, elements, _ := sseSend(t, h, http.MethodPost, "/admin/preview", map[string]any{"body": "# Title\n\n<script>alert(1)</script>\n\n**bold**", "noteBody": "other"})
	checkPatches(t, "editor preview", elements)
	if len(elements) != 1 || !strings.Contains(elements[0], `id="body-preview"`) || !strings.Contains(elements[0], `<h1 id="pv-md-title">`) ||
		!strings.Contains(elements[0], "<strong>bold</strong>") || strings.Contains(elements[0], "<script") || strings.Contains(elements[0], "other") {
		t.Errorf("editor preview: %v", elements)
	}
	_ = events

	events, elements, _ = sseSend(t, h, http.MethodPost, "/admin/preview?as=page", map[string]any{"title": "My post", "body": "Hello *there*"})
	checkPatches(t, "page preview", elements)
	if len(elements) != 1 || !strings.Contains(elements[0], `id="post-preview-body"`) || !strings.Contains(elements[0], "My post") || !strings.Contains(elements[0], "<em>there</em>") {
		t.Errorf("page preview: %v", elements)
	}
	if sp := signalPatches(events); len(sp) != 1 || !strings.Contains(sp[0], `"postPreview":true`) {
		t.Errorf("the drawer is not opened: %q", sp)
	}

	_, elements, _ = sseSend(t, h, http.MethodPost, "/admin/preview?editor=noteBody", map[string]any{"body": "x", "noteBody": "*note*"})
	if len(elements) != 1 || !strings.Contains(elements[0], `id="noteBody-preview"`) || !strings.Contains(elements[0], "<em>note</em>") {
		t.Errorf("note preview: %v", elements)
	}
	if _, _, code := sseSend(t, h, http.MethodPost, "/admin/preview?editor=nope", map[string]any{}); code != http.StatusBadRequest {
		t.Errorf("unknown editor: status %d, want 400", code)
	}
}

func TestNotes(t *testing.T) {
	resetAdmin(t)
	h := routes()

	_, elements, code := sseSend(t, h, http.MethodGet, "/admin/notes/list", map[string]any{"nq": "sourdough"})
	if code != 200 || len(elements) != 1 {
		t.Fatalf("filter: status %d, %d patches", code, len(elements))
	}
	checkPatches(t, "notes list", elements)
	if !strings.Contains(elements[0], `id="notes-list"`) || !strings.Contains(elements[0], "Groceries") || strings.Contains(elements[0], "Morning pages") {
		t.Errorf("filtered list: %s", elements[0])
	}
	_, elements, _ = sseSend(t, h, http.MethodGet, "/admin/notes/list", map[string]any{"nq": "zzzz"})
	if !strings.Contains(elements[0], "No notes match") {
		t.Errorf("empty list: %s", elements[0])
	}

	events, elements, _ := sseSend(t, h, http.MethodGet, "/admin/notes/2", map[string]any{})
	checkPatches(t, "note", elements)
	if len(elements) != 2 || !strings.Contains(elements[0], `id="note-detail-header"`) || !strings.Contains(elements[0], "Talk outline: layouts that adapt") ||
		!strings.Contains(elements[1], `id="note-detail"`) || !strings.Contains(elements[1], "<blockquote>") {
		t.Errorf("detail: %q", elements)
	}
	if sp := signalPatches(events); len(sp) != 1 || !strings.Contains(sp[0], `"noteId":2`) || !strings.Contains(sp[0], `"noteEdit":false`) {
		t.Errorf("note signals %q", sp)
	}
	if _, _, code := sseSend(t, h, http.MethodGet, "/admin/notes/99", map[string]any{}); code != http.StatusNotFound {
		t.Errorf("unknown note: status %d", code)
	}

	// An empty title is refused; a good one is stored and the list and detail re-rendered.
	events, elements, _ = sseSend(t, h, http.MethodPost, "/admin/notes/2", map[string]any{"noteTitle": " ", "noteBody": "x"})
	if sp := signalPatches(events); len(sp) != 1 || !strings.Contains(sp[0], "Enter a title.") || len(elements) != 1 {
		t.Errorf("invalid save: %q %q", sp, elements)
	}
	if n, _ := adminNotes.get(2); n.Body == "x" {
		t.Error("an invalid save was stored")
	}
	_, elements, _ = sseSend(t, h, http.MethodPost, "/admin/notes/2", map[string]any{"noteTitle": "Renamed", "noteBody": "New *body*", "nq": ""})
	checkPatches(t, "note saved", elements)
	if len(elements) != 4 || !strings.Contains(elements[0], "Renamed") || !strings.Contains(elements[1], "<em>body</em>") ||
		!strings.Contains(elements[2], `id="notes-list"`) || !strings.Contains(elements[3], "Note saved") {
		t.Errorf("saved patches: %q", elements)
	}
}

func TestSettingsSave(t *testing.T) {
	resetAdmin(t)
	h := routes()
	events, elements, _ := sseSend(t, h, http.MethodPost, "/admin/settings/general", map[string]any{"siteTitle": " ", "tagline": "x", "timezone": "Asia/Tokyo"})
	if sp := signalPatches(events); len(sp) != 1 || !strings.Contains(sp[0], "at least 2 characters") || len(elements) != 1 {
		t.Errorf("invalid: %q %q", sp, elements)
	}
	if adminSettingsData.get().Timezone == "Asia/Tokyo" {
		t.Error("an invalid save was stored")
	}
	_, elements, _ = sseSend(t, h, http.MethodPost, "/admin/settings/general", map[string]any{"siteTitle": "My Site", "tagline": "Hi", "timezone": "Asia/Tokyo"})
	checkPatches(t, "settings", elements)
	if v := adminSettingsData.get(); v.SiteTitle != "My Site" || v.Timezone != "Asia/Tokyo" || !strings.Contains(elements[0], "Settings saved") {
		t.Errorf("stored %+v, toast %q", v, elements)
	}
	_, _, _ = sseSend(t, h, http.MethodPost, "/admin/settings/writing", map[string]any{"defaultAuthor": adminAuthors[1], "defaultStatus": "published", "autosave": false})
	if v := adminSettingsData.get(); v.DefaultAuthor != adminAuthors[1] || v.Autosave {
		t.Errorf("writing settings %+v", v)
	}
	if _, _, code := sseSend(t, h, http.MethodPost, "/admin/settings/nope", map[string]any{}); code != http.StatusNotFound {
		t.Errorf("unknown tab: status %d", code)
	}
	// Reset restores everything.
	adminData.delete(1, 2, 3)
	events, elements, _ = sseSend(t, h, http.MethodPost, "/admin/reset", map[string]any{})
	if adminData.counts().Total != 30 || adminSettingsData.get().SiteTitle != "Acme Journal" || !strings.Contains(elements[0], "Demo data reset") || len(signalPatches(events)) != 1 {
		t.Error("reset did not restore the seed")
	}
}

// TestNoHandWrittenDataAttributes keeps the admin demo on the typed datastar helpers: no
// data-* string in its sources. (The iframe helper files carry none either.)
func TestNoHandWrittenDataAttributes(t *testing.T) {
	files, _ := filepath.Glob("admin_*.go")
	attr := regexp.MustCompile(`"data-[a-z]`)
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if attr.Match(src) {
			t.Errorf("%s writes a data-* attribute by hand; use the datastar helpers", f)
		}
	}
}
