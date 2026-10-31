package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
	dsse "github.com/starfederation/datastar-go/datastar"

	"github.com/go-htemel/ui"
	"github.com/go-htemel/ui/markdown"
)

// The notes page: a SplitLayout of journal entries. The list filters on the server
// (GET /admin/notes/list patches the <ul>), choosing a note loads it into the detail pane
// (GET /admin/notes/{id} patches #note-detail and sets the editor's signals), and the Edit
// button swaps the rendered note for a MarkdownEditor in place.

const (
	notesListID    = "notes-list"
	noteDetailID   = "note-detail"
	noteEditBtnID  = "note-edit"
	noteTitleField = "note-title"
	notesEditor    = "noteBody"
)

func noteURL(id int) string { return "/admin/notes/" + strconv.Itoa(id) }

func noteTags(tags []string) *html.SpanElement {
	g := html.Span().Class("flex flex-wrap gap-1")
	for _, t := range tags {
		g.Children(ui.Badge(ui.Neutral, t))
	}
	return g
}

func noteDate(n adminNote) *html.TimeElement {
	return html.Time().Datetime(n.Date.Format("2006-01-02")).Text(n.Date.Format("Jan 2, 2006"))
}

// notesListView is the list of notes matching q; the item for $noteId is aria-current.
func notesListView(q string) *html.UlElement {
	ul := html.Ul().Id(notesListID).Role("list").Class("divide-y divide-neutral-200 dark:divide-white/10")
	notes := adminNotes.list(q)
	for _, n := range notes {
		id := strconv.Itoa(n.ID)
		ul.Children(html.Li(html.Button(
			html.Span().Class("block truncate text-sm font-medium text-neutral-900 dark:text-white").Text(n.Title),
			html.Span().Class("mt-0.5 block text-xs text-neutral-600 dark:text-neutral-400").Children(noteDate(n)),
			noteTags(n.Tags).AddClass("mt-1.5"),
		).Type(html.ButtonTypeButton).
			Class("block w-full cursor-pointer px-4 py-3 text-left transition-colors duration-150 hover:bg-neutral-100 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-600 aria-[current=true]:bg-neutral-100 dark:hover:bg-white/5 dark:focus-visible:outline-primary-500 dark:aria-[current=true]:bg-white/10").
			Attrs(
				ui.SplitSelect(ui.SplitLayoutSignal, "$noteId = "+id+"; $noteEdit = false; "+datastar.Get(noteURL(n.ID))),
				datastar.Attr("aria-current", "$noteId === "+id+" ? 'true' : false"),
			)))
	}
	if len(notes) == 0 {
		ul.Children(html.Li(ui.EmptyState(ui.IconSearch(), "No notes match", "Try another word.", nil)))
	}
	return ul
}

// noteDetailView is the detail pane's content for n: the header with the Edit toggle, the
// rendered note, and the editor (hidden until $noteEdit).
func noteDetailView(n adminNote) *html.DivElement {
	editBtn := ui.Button(ui.Secondary, "Edit", ui.WithSize(ui.Sm), ui.WithIcon(ui.IconPencil())).Id(noteEditBtnID).
		Attrs(datastar.Show("!$noteEdit"),
			datastar.OnClick("$noteEdit = true; requestAnimationFrame(() => document.getElementById('"+noteTitleField+"')?.focus())"))
	cancel := ui.Button(ui.Ghost, "Cancel", ui.WithSize(ui.Sm)).Style("display: none").
		Attrs(datastar.Show("$noteEdit"),
			datastar.OnClick("$noteEdit = false; $noteTitle = "+datastar.Quote(n.Title)+"; $noteBody = "+datastar.Quote(n.Body)+"; requestAnimationFrame(() => document.getElementById('"+noteEditBtnID+"')?.focus())"))
	save := ui.Button(ui.Primary, "Save", ui.WithSize(ui.Sm), ui.WithLoading("noteSaving")).Style("display: none").
		Attrs(datastar.Show("$noteEdit"),
			datastar.OnClick(afterRequest(datastar.Post(noteURL(n.ID)), "document.querySelector('#"+noteDetailID+" [aria-invalid=true]')?.focus() ?? document.getElementById('"+ui.SplitLayoutSignal+"-detail')?.focus()")))

	view := html.Div(
		html.Div(noteDate(n), noteTags(n.Tags)).Class("flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-neutral-600 dark:text-neutral-400"),
		html.Div(markdown.Render(n.Body, markdown.WithSize(ui.Sm), markdown.WithHeadingOffset(1), markdown.WithIDPrefix("note-md-"))).Class("mt-4"),
	).Class("p-4 @xl:p-6").Attrs(datastar.Show("!$noteEdit"))

	edit := html.Div(ui.Stack(ui.Lg,
		ui.Field(noteTitleField, "Title", ui.Input(html.InputTypeText, "noteTitle").Attrs(datastar.Bind("noteTitle")), ui.ErrorSignal("errors.noteTitle")),
		ui.Field("note-body", "Note", ui.MarkdownEditor(notesEditor, ui.WithValue(n.Body), ui.WithRows(12),
			ui.WithPreview(datastar.Post("/admin/preview?editor="+notesEditor)), ui.WithPreviewLabel("Note preview"))),
	)).Style("display: none").Class("p-4 @xl:p-6").Attrs(datastar.Show("$noteEdit"))

	return html.Div(
		ui.SplitDetailHeader(n.Title, editBtn, cancel, save),
		view, edit,
	).Id(noteDetailID)
}

// patchNote patches the detail pane to n and sets the editor's signals.
func patchNote(sse *dsse.ServerSentEventGenerator, n adminNote) {
	patch(sse, noteDetailView(n))
	patchSignals(sse, map[string]any{
		"noteId": n.ID, "noteEdit": false, "noteTitle": n.Title, "noteBody": n.Body,
		"errors": map[string]any{"noteTitle": ""},
	})
}

func noteIDParam(w http.ResponseWriter, r *http.Request) (adminNote, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	n, ok := adminNotes.get(id)
	if err != nil || !ok {
		http.NotFound(w, r)
		return n, false
	}
	return n, true
}

// notesShow is GET /admin/notes/{id}.
func notesShow(w http.ResponseWriter, r *http.Request) {
	if n, ok := noteIDParam(w, r); ok {
		patchNote(dsse.NewSSE(w, r), n)
	}
}

// notesFilter is GET /admin/notes/list: the list for the search signal nq.
func notesFilter(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Q string `json:"nq"`
	}
	if err := dsse.ReadSignals(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	patch(dsse.NewSSE(w, r), notesListView(in.Q))
}

// notesSave is POST /admin/notes/{id}: the title and body of the editor.
func notesSave(w http.ResponseWriter, r *http.Request) {
	n, ok := noteIDParam(w, r)
	if !ok {
		return
	}
	var in struct {
		Title string `json:"noteTitle"`
		Body  string `json:"noteBody"`
		Q     string `json:"nq"`
	}
	if err := dsse.ReadSignals(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sse := dsse.NewSSE(w, r)
	title := strings.TrimSpace(in.Title)
	if title == "" {
		patchSignals(sse, map[string]any{"errors": map[string]any{"noteTitle": "Enter a title."}})
		toast(sse, ui.Danger, "Note not saved", "Give the note a title.")
		return
	}
	n, _ = adminNotes.update(n.ID, title, in.Body)
	patchNote(sse, n)
	patch(sse, notesListView(in.Q))
	toast(sse, ui.Success, "Note saved", "“"+n.Title+"” was updated.")
}

// notesPage is GET /admin/notes.
func notesPage(r *http.Request) htemel.Node {
	all := adminNotes.list("")
	var first adminNote
	if len(all) > 0 {
		first = all[0]
	}
	listPane := html.Div(
		html.Div(
			html.H1().Class("text-base font-semibold text-neutral-900 dark:text-white").Text("Notes"),
			html.Div(ui.FilterSearch("nq", datastar.Get("/admin/notes/list"))).Class("mt-3 flex"),
		).Class("sticky top-0 z-10 border-b border-neutral-200 bg-white p-4 dark:border-white/10 dark:bg-neutral-900"),
		notesListView(""),
	)
	signals := map[string]any{
		"noteId": first.ID, "noteEdit": false, "noteTitle": first.Title, "noteBody": first.Body,
		"errors": map[string]any{"noteTitle": ""}, "nq": "",
	}
	split := ui.SplitLayout(listPane, noteDetailView(first),
		ui.WithPaneLabels("Notes list", "Note")).Attrs(datastar.Signals(signals))
	return adminPage("Notes", navNotes, false, r.URL.Query().Get("notice"), split)
}
