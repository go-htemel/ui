package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
	dsse "github.com/starfederation/datastar-go/datastar"

	"github.com/go-htemel/ui"
)

// The posts list: one GET /admin/posts/list handler reads the filter, sort and page signals and
// patches three things by id, the table body, the pager and the count line. Nothing else is
// replaced, so the filter controls (which hold focus while typing) are never touched.

const (
	postsURL      = "/admin/posts/list"
	postsBodyID   = "posts-body"
	postsPagerID  = "posts-pager"
	postsCountID  = "posts-count"
	postsSelectID = "posts-selectall"
)

// clearSelection unticks every row: the selection is the object signal sel, one boolean per
// post id ("p12").
const clearSelection = "Object.keys($sel).forEach(k => $sel[k] = false)"

func postKey(id int) string { return "p" + strconv.Itoa(id) }

func postEditURL(id int) string { return "/admin/posts/" + strconv.Itoa(id) + "/edit" }

// statusBadge is the Badge of a post status: Success for published, Neutral for drafts and
// Brand for scheduled posts.
func statusBadge(status string) *html.SpanElement {
	tone := ui.Neutral
	switch status {
	case statusPublished:
		tone = ui.Success
	case statusScheduled:
		tone = ui.Brand
	}
	return ui.Badge(tone, statusLabel(status))
}

const postLinkClass = "font-medium text-neutral-900 hover:text-primary-700 focus-visible:rounded-xs focus-visible:outline-2 focus-visible:outline-offset-2 " +
	"focus-visible:outline-primary-600 dark:text-white dark:hover:text-primary-300 dark:focus-visible:outline-primary-500"

func postRow(p adminPost) *html.TrElement {
	key := postKey(p.ID)
	menu := ui.Dropdown("rowMenu"+strconv.Itoa(p.ID),
		ui.IconButton(ui.Ghost, ui.IconEllipsis(), "Actions for “"+p.Title+"”", ui.WithSize(ui.Sm)),
		ui.MenuLink("Edit", postEditURL(p.ID)),
		ui.MenuItem("Duplicate", datastar.OnClick(datastar.Post("/admin/posts/"+strconv.Itoa(p.ID)+"/duplicate"))),
		ui.MenuDivider(),
		ui.MenuItem("Delete", datastar.OnClick("$delId = "+strconv.Itoa(p.ID)+"; $delTitle = "+datastar.Quote(p.Title)+"; "+ui.Open("deleteOne"))),
	)
	meta := html.P().Class("mt-0.5 truncate text-xs text-neutral-600 dark:text-neutral-400").Text("/blog/" + p.Slug)
	if len(p.Tags) > 0 {
		meta.Text(" · " + strings.Join(p.Tags, ", "))
	}
	return html.Tr(
		ui.Td(ui.Checkbox("", datastar.Bind("sel."+key), htemel.Attr("aria-label", "Select “"+p.Title+"”"))).AddClass("w-10 pr-0"),
		ui.Td(html.Div(html.A().Href(postEditURL(p.ID)).Class(postLinkClass).Text(p.Title), meta).Class("min-w-48 max-w-md")),
		ui.Td(statusBadge(p.Status)),
		ui.Td(htemel.Text(p.Author)).AddClass("whitespace-nowrap"),
		ui.Td(adminTime(p.UpdatedAt)).AddClass("whitespace-nowrap"),
		ui.Td(html.Div(menu).Class("flex justify-end")).AddClass("w-12 pl-0"),
	)
}

// postsBody is the table body of the rows, or an EmptyState when nothing matches.
func postsBody(rows []adminPost, filtered bool) *html.TbodyElement {
	body := ui.TableBody().Id(postsBodyID)
	for _, p := range rows {
		body.Children(postRow(p))
	}
	if len(rows) == 0 {
		state := ui.EmptyState(ui.IconSearch(), "No posts match",
			"Nothing matches these filters. Clear them to see every post.",
			ui.Button(ui.Secondary, "Clear filters", ui.WithSize(ui.Sm)).
				Attrs(datastar.OnClick(clearFilters)))
		if !filtered {
			state = ui.EmptyState(ui.IconDocument(), "No posts yet", "Write your first post to get started.",
				html.Div(linkButton(true, "New post", "/admin/posts/new", ui.IconPlus())))
		}
		body.Children(ui.EmptyRow(6, state))
	}
	return body
}

// clearFilters empties the filters and reloads the first page.
var clearFilters = ui.FilterReset("q", "status", "tag") + "; $page = 1; " + clearSelection + "; " + datastar.Get(postsURL)

// selectAllHead is the header cell of the select column. Its checkbox ticks every row of the
// page shown (the ids are in the expression, so the cell is patched with the body) and follows
// the rows: it is checked while all of them are.
func selectAllHead(rows []adminPost) *html.ThElement {
	ids := make([]string, len(rows))
	for i, p := range rows {
		ids[i] = datastar.Quote(postKey(p.ID))
	}
	list := "[" + strings.Join(ids, ", ") + "]"
	box := ui.Checkbox("",
		htemel.Attr("aria-label", "Select all posts on this page"),
		datastar.Effect("const ids = "+list+"; el.checked = ids.length > 0 && ids.every(k => $sel[k])"),
		datastar.OnChange("const on = el.checked; "+list+".forEach(k => $sel[k] = on)"),
	)
	return ui.Th(box).Id(postsSelectID).AddClass("w-10 pr-0")
}

func postsPager(current, pages int) *html.NavElement {
	return ui.PaginationSignal(current, pages, "page", datastar.Get(postsURL), "Posts pages").Id(postsPagerID)
}

func postsCount(total, current, pages int) *html.PElement {
	text := "No posts"
	if total > 0 {
		from := (current-1)*postsPageSize + 1
		to := min(current*postsPageSize, total)
		text = fmt.Sprintf("Showing %d–%d of %d posts", from, to, total)
	}
	return html.P().Id(postsCountID).Tabindex(-1).Role("status").
		Class("text-sm text-neutral-600 outline-none dark:text-neutral-400").Text(text)
}

// queryFiltered reports whether q narrows the list.
func (q postQuery) filtered() bool {
	return strings.TrimSpace(q.Q) != "" || q.Status != "" || q.Tag != ""
}

// patchPosts runs q and patches the body, the select-all cell, the pager and the count. When
// the page asked for is out of range (the last post of the last page was deleted) it also
// sets the page signal to the page shown.
func patchPosts(sse *dsse.ServerSentEventGenerator, q postQuery) {
	rows, current, pages, total := adminData.run(q)
	patch(sse, postsBody(rows, q.filtered()))
	patch(sse, selectAllHead(rows))
	patch(sse, postsPager(current, pages))
	patch(sse, postsCount(total, current, pages))
	if current != q.Page {
		patchSignals(sse, map[string]any{"page": current})
	}
}

func patchSignals(sse *dsse.ServerSentEventGenerator, v any) {
	if err := sse.MarshalAndPatchSignals(v); err != nil {
		slog.Error("patch signals", "err", err)
	}
}

// postsList is GET /admin/posts/list.
func postsList(w http.ResponseWriter, r *http.Request) {
	var q postQuery
	if err := dsse.ReadSignals(r, &q); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	patchPosts(dsse.NewSSE(w, r), q)
}

// postsSignals is the signals the list page and its handlers share: the query, the selection
// and the target of the delete dialog.
type postsSignals struct {
	postQuery
	Sel      map[string]bool `json:"sel"`
	DelID    int             `json:"delId"`
	DelTitle string          `json:"delTitle"`
}

func readPostsSignals(w http.ResponseWriter, r *http.Request) (postsSignals, bool) {
	var in postsSignals
	if err := dsse.ReadSignals(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return in, false
	}
	return in, true
}

// postsDelete is POST /admin/posts/delete: it deletes the post of the dialog ($delId), then
// patches the list and toasts.
func postsDelete(w http.ResponseWriter, r *http.Request) {
	in, ok := readPostsSignals(w, r)
	if !ok {
		return
	}
	sse := dsse.NewSSE(w, r)
	if adminData.delete(in.DelID) == 0 {
		toast(sse, ui.Warning, "Already gone", "That post no longer exists.")
	} else {
		toast(sse, ui.Success, "Post deleted", "“"+in.DelTitle+"” was removed.")
	}
	patchPosts(sse, in.postQuery)
	patchSignals(sse, map[string]any{"sel": map[string]bool{postKey(in.DelID): false}})
}

// selectedIDs returns the ids ticked in sel ("p12": true), ascending.
func selectedIDs(sel map[string]bool) []int {
	var ids []int
	for k, on := range sel {
		if n, err := strconv.Atoi(strings.TrimPrefix(k, "p")); on && err == nil && strings.HasPrefix(k, "p") {
			ids = append(ids, n)
		}
	}
	slices.Sort(ids)
	return ids
}

// postsBulkDelete is POST /admin/posts/bulk-delete: it deletes every ticked post.
func postsBulkDelete(w http.ResponseWriter, r *http.Request) {
	in, ok := readPostsSignals(w, r)
	if !ok {
		return
	}
	ids := selectedIDs(in.Sel)
	n := adminData.delete(ids...)
	sse := dsse.NewSSE(w, r)
	cleared := map[string]bool{}
	for k := range in.Sel {
		cleared[k] = false
	}
	patchPosts(sse, in.postQuery)
	patchSignals(sse, map[string]any{"sel": cleared})
	toast(sse, ui.Success, fmt.Sprintf("%d %s deleted", n, plural(n, "post", "posts")), "The selected posts were removed.")
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// postsDuplicate is POST /admin/posts/{id}/duplicate.
func postsDuplicate(w http.ResponseWriter, r *http.Request, id int) {
	in, ok := readPostsSignals(w, r)
	if !ok {
		return
	}
	sse := dsse.NewSSE(w, r)
	c, found := adminData.duplicate(id)
	if !found {
		toast(sse, ui.Warning, "Not found", "That post no longer exists.")
	} else {
		toast(sse, ui.Success, "Post duplicated", "“"+c.Title+"” was added as a draft.")
	}
	patchPosts(sse, in.postQuery)
}

// postsPage is GET /admin/posts.
func postsPage(r *http.Request) htemel.Node {
	q := postQuery{Sort: "updated", SortDir: "desc", Page: 1}
	rows, current, pages, total := adminData.run(q)
	get := datastar.Get(postsURL)
	refilter := "$page = 1; " + clearSelection + "; " + get

	tagOptions := []html.SelectChild{ui.Choice("", "All tags")}
	for _, t := range adminData.tags() {
		tagOptions = append(tagOptions, ui.Choice(t, t))
	}
	filters := ui.FilterBar(
		ui.FilterSearch("q", refilter),
		ui.Select("status",
			ui.Choice("", "All statuses"), ui.Choice(statusPublished, "Published"), ui.Choice(statusDraft, "Draft"), ui.Choice(statusScheduled, "Scheduled"),
		).Attrs(datastar.Bind("status"), datastar.OnChange(refilter), htemel.Attr("aria-label", "Status")).AddClass("w-40"),
		ui.Select("tag", tagOptions...).
			Attrs(datastar.Bind("tag"), datastar.OnChange(refilter), htemel.Attr("aria-label", "Tag")).AddClass("w-40"),
		ui.Button(ui.Ghost, "Reset").Attrs(datastar.OnClick(clearFilters)),
	)

	bulk := html.Div(
		html.P().Class("text-sm font-medium text-neutral-900 dark:text-white").
			Attrs(datastar.Text("$selCount + ($selCount === 1 ? ' post selected' : ' posts selected')")).Text("0 posts selected"),
		html.Div(
			ui.Button(ui.Ghost, "Clear selection", ui.WithSize(ui.Sm)).Attrs(datastar.OnClick(clearSelection)),
			ui.Button(ui.Danger, "Delete selected", ui.WithSize(ui.Sm), ui.WithIcon(ui.IconTrash())).Attrs(ui.Opens("deleteMany")),
		).Class("flex items-center gap-x-2"),
	).Role("region").Aria("label", "Bulk actions").
		Style("display: none").
		Class("flex flex-wrap items-center justify-between gap-3 rounded-surface bg-primary-50 px-4 py-2.5 ring-1 ring-primary-600/20 dark:bg-primary-500/10 dark:ring-primary-500/30").
		Attrs(datastar.Show("$selCount > 0"))

	table := ui.Table(
		ui.TableHead(
			selectAllHead(rows),
			ui.SortableTh("Title", "title", "sort", get),
			ui.SortableTh("Status", "status", "sort", get),
			ui.Th(htemel.Text("Author")),
			ui.SortableTh("Updated", "updated", "sort", get),
			ui.Th(html.Span().Class("sr-only").Text("Actions")),
		),
		postsBody(rows, false),
		ui.TableCaption("Posts, sortable by title, status and update date"),
	)

	deleteOne := ui.Dialog("deleteOne", "Delete this post?",
		html.P().Attrs(datastar.Text("'“' + $delTitle + '” will be deleted. This cannot be undone.'")).Text("This post will be deleted."),
		ui.DialogActions(
			ui.Button(ui.Secondary, "Cancel").Attrs(datastar.OnClick(ui.Close("deleteOne"))),
			ui.Button(ui.Danger, "Delete", ui.WithLoading("deleting")).
				Attrs(datastar.OnClick(ui.Close("deleteOne")+"; "+afterRequest(datastar.Post("/admin/posts/delete"), focusCount))),
		),
	)
	deleteMany := ui.Dialog("deleteMany", "Delete the selected posts?",
		html.P().Attrs(datastar.Text("$selCount + ($selCount === 1 ? ' post' : ' posts') + ' will be deleted. This cannot be undone.'")).Text("The selected posts will be deleted."),
		ui.DialogActions(
			ui.Button(ui.Secondary, "Cancel").Attrs(datastar.OnClick(ui.Close("deleteMany"))),
			ui.Button(ui.Danger, "Delete all", ui.WithLoading("deletingMany")).
				Attrs(datastar.OnClick(ui.Close("deleteMany")+"; "+afterRequest(datastar.Post("/admin/posts/bulk-delete"), focusCount))),
		),
	)

	signals := map[string]any{
		"q": "", "status": "", "tag": "", "sort": q.Sort, "sortDir": q.SortDir, "page": 1,
		"sel": map[string]bool{}, "delId": 0, "delTitle": "",
	}
	return adminPage("Posts", navPosts, false, r.URL.Query().Get("notice"),
		adminBody(
			ui.PageHeader("Posts",
				ui.WithBreadcrumbs(ui.Breadcrumbs(ui.Crumb{Label: "Dashboard", Href: "/admin"}, ui.Crumb{Label: "Posts"})),
				ui.WithDescription("Write, schedule and publish."),
				ui.WithActions(linkButton(true, "New post", "/admin/posts/new", ui.IconPlus()))),
			html.Div(
				filters,
				bulk,
				html.Div(table, html.Div(postsPager(current, pages), postsCount(total, current, pages)).Class("mt-4 flex flex-wrap items-center justify-between gap-3")),
				deleteOne, deleteMany,
			).Class("flex flex-col gap-y-4").
				Attrs(datastar.Signals(signals), datastar.Computed("selCount", "Object.values($sel).filter(Boolean).length")),
		))
}

// focusCount moves focus to the count line after a delete: the row that held it is gone.
const focusCount = "document.getElementById('" + postsCountID + "')?.focus()"

// afterRequest runs then once the response of the action (an @get, @post...) has been applied.
// Datastar expressions are not async functions (no await), but the actions return promises.
func afterRequest(action, then string) string { return action + ".then(() => { " + then + " })" }
