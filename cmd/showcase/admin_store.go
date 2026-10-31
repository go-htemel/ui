package main

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
)

// The admin demo keeps its posts and notes in server memory, like the kanban board: every
// visitor of the showcase shares them, and POST /admin/reset restores the seed. The seed is a
// pure function of the clock (a fixed list of titles, authors and tags, dated backwards from
// "now"), so tests that inject a fixed clock see the same data every run.

// Post statuses.
const (
	statusDraft     = "draft"
	statusPublished = "published"
	statusScheduled = "scheduled"
)

var postStatuses = []string{statusPublished, statusDraft, statusScheduled}

// statusLabel is the display name of a status.
func statusLabel(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// adminAuthors are the people who write for the demo site.
var adminAuthors = []string{"Ada Lovelace", "Grace Hopper", "Margaret Hamilton", "Dennis Ritchie"}

// dateTimeLayout is the value format of <input type="datetime-local">.
const dateTimeLayout = "2006-01-02T15:04"

type adminPost struct {
	ID        int
	Title     string
	Slug      string
	Status    string
	Author    string
	Tags      []string
	UpdatedAt time.Time
	Body      string // Markdown
	// PublishAt is the datetime-local value of a scheduled post ("" otherwise).
	PublishAt string
	Comments  bool // comments are open
	// CommentCount is how many comments the post has had; the dashboard adds them up.
	CommentCount int
}

func (p adminPost) clone() adminPost {
	p.Tags = slices.Clone(p.Tags)
	return p
}

// seedTitles are the 30 posts of the seed. Slugs, tags, statuses and dates derive from the
// index, so the list has a believable spread: mostly published, some drafts, a few scheduled.
var seedTitles = []string{
	"Container queries in practice", "Writing a CMS in Go", "Notes on Datastar signals",
	"Why the viewport is the wrong question", "A tour of the new AppShell", "Server-sent events for forms",
	"Accessible tables without the pain", "Dark mode that does not flash", "Tailwind tokens, not palettes",
	"Markdown editors, honestly", "Slugs: a short history", "Keyboard shortcuts worth stealing",
	"Focus management after a swap", "Pagination that survives deletes", "Drag and drop on touch screens",
	"What a toast should announce", "Testing HTML you did not write", "Native dialogs are finally good",
	"The popover API in anger", "Contrast ratios, by hand", "Forms that validate on the server",
	"Building a tag input", "Why we self-host our fonts", "Release notes: autumn edition",
	"Designing empty states", "Small multiples for dashboards", "Reading list: October",
	"How to write a changelog", "Edge cases in bulk actions", "Hello, world (again)",
}

var seedTags = []string{"go", "css", "datastar", "tailwind", "accessibility", "writing", "design", "testing", "htmx", "performance"}

// slugify mirrors the browser's SlugField for the seed and for duplicates.
func slugify(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			dash = false
			b.WriteRune(r)
		} else {
			dash = true
		}
	}
	return b.String()
}

// seedBody is the Markdown of a seeded post: a heading, a paragraph, a list and, for every
// third post, a Go snippet, so the preview shows real structure.
func seedBody(i int, title string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", title)
	fmt.Fprintf(&b, "This post is about **%s**. It was written to show what the editor does with headings, lists, links and code.\n\n", strings.ToLower(title))
	b.WriteString("## What it covers\n\n")
	b.WriteString("- The problem, in one paragraph\n- A worked example\n- What to watch out for\n\n")
	if i%3 == 0 {
		b.WriteString("```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```\n\n")
	}
	b.WriteString("Read more in the [documentation](https://data-star.dev).\n")
	return b.String()
}

func seedPosts(now time.Time) []adminPost {
	base := time.Date(now.Year(), now.Month(), now.Day(), 16, 20, 0, 0, time.UTC).Add(-26 * time.Hour)
	posts := make([]adminPost, len(seedTitles))
	for i, title := range seedTitles {
		p := adminPost{
			ID:           i + 1,
			Title:        title,
			Slug:         slugify(title),
			Status:       statusPublished,
			Author:       adminAuthors[(i*3+i/4)%len(adminAuthors)],
			Tags:         []string{seedTags[i%len(seedTags)], seedTags[(i*7+3)%len(seedTags)]},
			UpdatedAt:    base.Add(-time.Duration(i) * 38 * time.Hour),
			Body:         seedBody(i, title),
			Comments:     i%5 != 4,
			CommentCount: (i * 7) % 23,
		}
		if p.Tags[0] == p.Tags[1] {
			p.Tags = p.Tags[:1]
		}
		switch {
		case i%4 == 1 || i == 6:
			p.Status = statusDraft
			p.CommentCount = 0
		case i%11 == 2:
			p.Status = statusScheduled
			p.CommentCount = 0
			p.UpdatedAt = base.Add(-time.Duration(i) * time.Hour)
			p.PublishAt = base.Add(time.Duration(48+i*9) * time.Hour).Format(dateTimeLayout)
		}
		posts[i] = p
	}
	return posts
}

// postQuery is what the posts table asks for: the signals the page sends.
type postQuery struct {
	Q       string `json:"q"`
	Status  string `json:"status"`
	Tag     string `json:"tag"`
	Sort    string `json:"sort"`
	SortDir string `json:"sortDir"`
	Page    int    `json:"page"`
}

// postsPageSize is how many posts a page of the table holds.
const postsPageSize = 10

var statusRank = map[string]int{statusPublished: 0, statusScheduled: 1, statusDraft: 2}

type adminStore struct {
	mu     sync.Mutex
	now    func() time.Time
	posts  []adminPost
	nextID int
}

func newAdminStore(now func() time.Time) *adminStore {
	s := &adminStore{now: now}
	s.reset()
	return s
}

// adminData is the posts store shared by every visitor.
var adminData = newAdminStore(time.Now)

func (s *adminStore) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.posts = seedPosts(s.now())
	s.nextID = len(s.posts) + 1
}

// run filters the posts by Q (a case-insensitive substring of the title, slug or author), Status
// and Tag (exact; "" for any), sorts them by Sort ("title", "status" or "updated") in direction
// SortDir ("desc"; anything else is ascending: A to Z, published first, oldest first) and
// returns the requested page, with the page number clamped into range, the page count (at least
// 1) and the number of posts that matched. An empty or unknown Sort is the default order, newest
// first. Equal keys fall back to newest first, then to the ID, so the order is total.
func (s *adminStore) run(q postQuery) (rows []adminPost, current, pages, total int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	needle := strings.ToLower(strings.TrimSpace(q.Q))
	var all []adminPost
	for _, p := range s.posts {
		if q.Status != "" && p.Status != q.Status {
			continue
		}
		if q.Tag != "" && !slices.Contains(p.Tags, q.Tag) {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(p.Title+" "+p.Slug+" "+p.Author), needle) {
			continue
		}
		all = append(all, p.clone())
	}
	var key func(a, b adminPost) int
	switch q.Sort {
	case "title":
		key = func(a, b adminPost) int { return cmp.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title)) }
	case "status":
		key = func(a, b adminPost) int { return cmp.Compare(statusRank[a.Status], statusRank[b.Status]) }
	case "updated":
		key = func(a, b adminPost) int { return a.UpdatedAt.Compare(b.UpdatedAt) }
	}
	desc := q.SortDir == "desc"
	slices.SortStableFunc(all, func(a, b adminPost) int {
		c := 0
		if key != nil {
			if c = key(a, b); desc {
				c = -c
			}
		}
		return cmp.Or(c, b.UpdatedAt.Compare(a.UpdatedAt), cmp.Compare(a.ID, b.ID))
	})
	total = len(all)
	pages = max((total+postsPageSize-1)/postsPageSize, 1)
	current = min(max(q.Page, 1), pages)
	start := (current - 1) * postsPageSize
	return all[start:min(start+postsPageSize, total)], current, pages, total
}

func (s *adminStore) get(id int) (adminPost, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.posts {
		if p.ID == id {
			return p.clone(), true
		}
	}
	return adminPost{}, false
}

// slugTaken reports whether another post than except uses slug (ignoring case).
func (s *adminStore) slugTaken(slug string, except int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.slugTakenLocked(slug, except)
}

func (s *adminStore) slugTakenLocked(slug string, except int) bool {
	return slices.ContainsFunc(s.posts, func(p adminPost) bool { return p.ID != except && strings.EqualFold(p.Slug, slug) })
}

// save stores p: a post with ID 0 is created (and gets an ID), another replaces the post with
// that ID. It stamps UpdatedAt and reports whether the ID exists.
func (s *adminStore) save(p adminPost) (adminPost, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p = p.clone()
	p.UpdatedAt = s.now().UTC()
	if p.ID == 0 {
		p.ID = s.nextID
		s.nextID++
		s.posts = append(s.posts, p)
		return p.clone(), true
	}
	i := slices.IndexFunc(s.posts, func(q adminPost) bool { return q.ID == p.ID })
	if i < 0 {
		return adminPost{}, false
	}
	p.CommentCount = s.posts[i].CommentCount
	s.posts[i] = p
	return p.clone(), true
}

// delete removes the posts with the given ids and returns how many existed.
func (s *adminStore) delete(ids ...int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.posts)
	s.posts = slices.DeleteFunc(s.posts, func(p adminPost) bool { return slices.Contains(ids, p.ID) })
	return n - len(s.posts)
}

// duplicate adds a draft copy of post id ("Title (copy)", a free slug) and returns it.
func (s *adminStore) duplicate(id int) (adminPost, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := slices.IndexFunc(s.posts, func(p adminPost) bool { return p.ID == id })
	if i < 0 {
		return adminPost{}, false
	}
	c := s.posts[i].clone()
	c.ID = s.nextID
	s.nextID++
	c.Title += " (copy)"
	c.Status, c.PublishAt, c.CommentCount = statusDraft, "", 0
	c.UpdatedAt = s.now().UTC()
	base := slugify(c.Title)
	c.Slug = base
	for n := 2; s.slugTakenLocked(c.Slug, 0); n++ {
		c.Slug = fmt.Sprintf("%s-%d", base, n)
	}
	s.posts = append(s.posts, c)
	return c.clone(), true
}

// tags returns every tag in use, sorted.
func (s *adminStore) tags() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	set := map[string]bool{}
	for _, p := range s.posts {
		for _, t := range p.Tags {
			set[t] = true
		}
	}
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	slices.Sort(out)
	return out
}

type adminCounts struct {
	Total, Published, Drafts, Scheduled              int
	PublishedThisMonth, PublishedLastMonth, Comments int
}

func sameMonth(a, b time.Time) bool { return a.Year() == b.Year() && a.Month() == b.Month() }

func (s *adminStore) counts() adminCounts {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	c := adminCounts{Total: len(s.posts)}
	for _, p := range s.posts {
		c.Comments += p.CommentCount
		switch p.Status {
		case statusPublished:
			c.Published++
			switch {
			case sameMonth(p.UpdatedAt, now):
				c.PublishedThisMonth++
			case sameMonth(p.UpdatedAt, time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)):
				c.PublishedLastMonth++
			}
		case statusDraft:
			c.Drafts++
		case statusScheduled:
			c.Scheduled++
		}
	}
	return c
}

// recent returns the n most recently updated posts.
func (s *adminStore) recent(n int) []adminPost {
	s.mu.Lock()
	defer s.mu.Unlock()
	all := make([]adminPost, len(s.posts))
	for i, p := range s.posts {
		all[i] = p.clone()
	}
	slices.SortStableFunc(all, func(a, b adminPost) int { return cmp.Or(b.UpdatedAt.Compare(a.UpdatedAt), cmp.Compare(a.ID, b.ID)) })
	return all[:min(n, len(all))]
}

// ---- notes ----

type adminNote struct {
	ID    int
	Title string
	Date  time.Time
	Tags  []string
	Body  string // Markdown
}

type noteStore struct {
	mu    sync.Mutex
	now   func() time.Time
	notes []adminNote
}

func newNoteStore(now func() time.Time) *noteStore {
	s := &noteStore{now: now}
	s.reset()
	return s
}

// adminNotes is the journal shared by every visitor.
var adminNotes = newNoteStore(time.Now)

var seedNotes = []struct {
	title string
	tags  []string
	body  string
}{
	{"Morning pages", []string{"journal"}, "Woke early. Wrote three pages **by hand** before the coffee, then typed up the useful half.\n\n- The container query idea is still the right one\n- Ask about the rail state on Monday\n"},
	{"Talk outline: layouts that adapt", []string{"talks", "css"}, "## Outline\n\n1. Why the viewport is the wrong question\n2. Container queries, briefly\n3. Nesting layouts\n4. Framed previews\n\n> Most layout bugs are breakpoint bugs in disguise.\n"},
	{"Groceries", []string{"home"}, "- [x] Oat milk\n- [x] Lemons\n- [ ] Basil\n- [ ] A loaf of sourdough\n"},
	{"Book notes: Refactoring UI", []string{"reading", "design"}, "Chapter 3 argues that hierarchy comes from **weight and colour**, not size alone.\n\n| Lever | Use it for |\n| --- | --- |\n| Weight | emphasis |\n| Colour | de-emphasis |\n"},
	{"Release checklist", []string{"work", "go"}, "```sh\ngo test -race ./...\ngolangci-lint run ./...\ntask a11y\n```\n\nThen tag, then write the changelog.\n"},
	{"Idea: a command palette", []string{"ideas"}, "A palette on `Ctrl+K` that searches posts, notes and settings. Needs a combobox pattern the library does not have yet.\n"},
	{"Weekend", []string{"journal", "home"}, "Long walk, a bad film, a good soup. Fixed the shelf.\n"},
	{"Retro: the table rewrite", []string{"work", "testing"}, "What went well: patching the **tbody** only. What did not: forgetting the pager.\n"},
}

func (s *noteStore) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	day := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC)
	s.notes = nil
	for i, n := range seedNotes {
		s.notes = append(s.notes, adminNote{ID: i + 1, Title: n.title, Date: day.AddDate(0, 0, -i*2-i/3), Tags: slices.Clone(n.tags), Body: n.body})
	}
}

// list returns the notes whose title, tags or body contain q (case-insensitive), newest first.
func (s *noteStore) list(q string) []adminNote {
	s.mu.Lock()
	defer s.mu.Unlock()
	needle := strings.ToLower(strings.TrimSpace(q))
	var out []adminNote
	for _, n := range s.notes {
		if needle == "" || strings.Contains(strings.ToLower(n.Title+" "+strings.Join(n.Tags, " ")+" "+n.Body), needle) {
			out = append(out, n)
		}
	}
	slices.SortStableFunc(out, func(a, b adminNote) int { return cmp.Or(b.Date.Compare(a.Date), cmp.Compare(a.ID, b.ID)) })
	return out
}

func (s *noteStore) get(id int) (adminNote, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range s.notes {
		if n.ID == id {
			return n, true
		}
	}
	return adminNote{}, false
}

// update replaces the title and body of note id.
func (s *noteStore) update(id int, title, body string) (adminNote, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.notes {
		if s.notes[i].ID == id {
			s.notes[i].Title, s.notes[i].Body = title, body
			return s.notes[i], true
		}
	}
	return adminNote{}, false
}
