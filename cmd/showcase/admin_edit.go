package main

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
	dsse "github.com/starfederation/datastar-go/datastar"

	"github.com/go-htemel/ui"
	"github.com/go-htemel/ui/markdown"
)

// The post editor: GET /admin/posts/new and /admin/posts/{id}/edit render the form, Save posts
// every signal to POST /admin/posts/{id} ("new" for a post that does not exist yet), and the
// answer is SSE: validation errors patched into the errors.* signals, or on success a toast and
// a patched "Saved" line in the page header.

const (
	savedMetaID   = "post-saved"
	statusMetaID  = "post-status-badge"
	previewBodyID = "post-preview-body"
	editorName    = "body"
)

// postForm is the signals the editor sends on Save.
type postForm struct {
	Title     string   `json:"title"`
	Slug      string   `json:"slug"`
	Body      string   `json:"body"`
	Status    string   `json:"status"`
	PublishAt string   `json:"publishAt"`
	Author    string   `json:"author"`
	Comments  bool     `json:"comments"`
	Tags      []string `json:"tags"`
}

// postErrors is the errors.* signals: a message per field, "" when the field is fine.
type postErrors struct {
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	PublishAt string `json:"publishAt"`
}

func (e postErrors) count() int {
	n := 0
	for _, s := range []string{e.Title, e.Slug, e.PublishAt} {
		if s != "" {
			n++
		}
	}
	return n
}

// validate checks f for the post with the given id (0 for a new one) against the store and
// returns its messages (all "" when it is valid) and the post to save.
func (s *adminStore) validate(id int, f postForm) (postErrors, adminPost) {
	var e postErrors
	p := adminPost{
		ID: id, Title: strings.TrimSpace(f.Title), Slug: strings.TrimSpace(f.Slug), Body: f.Body,
		Status: f.Status, Author: f.Author, Comments: f.Comments, PublishAt: strings.TrimSpace(f.PublishAt),
	}
	switch {
	case p.Title == "":
		e.Title = "Enter a title."
	case len([]rune(p.Title)) > 120:
		e.Title = "Use at most 120 characters."
	}
	switch {
	case p.Slug == "":
		e.Slug = "Enter a slug."
	case slugify(p.Slug) != p.Slug:
		e.Slug = "Use lowercase letters, numbers and single hyphens."
	default:
		s.mu.Lock()
		for _, o := range s.posts {
			if o.ID != id && strings.EqualFold(o.Slug, p.Slug) {
				e.Slug = "Another post, “" + o.Title + "”, already uses this slug."
				break
			}
		}
		s.mu.Unlock()
	}
	if !slices.Contains(postStatuses, p.Status) {
		p.Status = statusDraft
	}
	if !slices.Contains(adminAuthors, p.Author) {
		p.Author = adminAuthors[0]
	}
	if p.Status == statusScheduled {
		if _, err := time.Parse(dateTimeLayout, p.PublishAt); err != nil {
			e.PublishAt = "Choose the date and time to publish."
		}
	} else {
		p.PublishAt = ""
	}
	seen := map[string]bool{}
	for _, t := range f.Tags {
		if t = strings.TrimSpace(t); t != "" && !seen[strings.ToLower(t)] {
			seen[strings.ToLower(t)] = true
			p.Tags = append(p.Tags, t)
		}
	}
	return e, p
}

func savedMeta(p adminPost, saved bool) *html.SpanElement {
	text := "Not saved yet"
	switch {
	case saved:
		text = "Saved just now, " + p.UpdatedAt.Format("15:04")
	case p.ID != 0:
		text = "Last saved " + p.UpdatedAt.Format("Jan 2, 15:04")
	}
	return html.Span().Id(savedMetaID).Text(text)
}

// postSave is POST /admin/posts/{id}.
func postSave(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id := 0
	if idParam != "new" {
		n, err := strconv.Atoi(idParam)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		id = n
		if _, ok := adminData.get(id); !ok {
			http.NotFound(w, r)
			return
		}
	}
	var f postForm
	if err := dsse.ReadSignals(r, &f); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sse := dsse.NewSSE(w, r)
	errs, p := adminData.validate(id, f)
	patchSignals(sse, map[string]any{"errors": errs})
	if n := errs.count(); n > 0 {
		toast(sse, ui.Danger, "The post was not saved", fmt.Sprintf("Fix the %d highlighted %s and save again.", n, plural(n, "field", "fields")))
		return
	}
	saved, _ := adminData.save(p)
	if id == 0 {
		// The URL changes with the new id, which only a navigation does.
		if err := sse.Redirect(postEditURL(saved.ID) + "?notice=created"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	patch(sse, savedMeta(saved, true))
	patch(sse, statusBadge(saved.Status).Id(statusMetaID))
	toast(sse, ui.Success, "Post saved", "“"+saved.Title+"” is "+saved.Status+".")
}

// postRemove is DELETE /admin/posts/{id}, from the editor's confirm dialog.
func postRemove(w http.ResponseWriter, r *http.Request) {
	id, ok := postID(w, r)
	if !ok {
		return
	}
	adminData.delete(id)
	if err := dsse.NewSSE(w, r).Redirect("/admin/posts?notice=deleted"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// renderPreview renders Markdown for the editor's preview, with an id prefix that keeps heading
// ids unique against the page's own.
func renderPreview(src string) *html.DivElement {
	return markdown.Render(src, markdown.WithIDPrefix("pv-md-"), markdown.WithSize(ui.Sm))
}

// previewHandler is POST /admin/preview. By default it patches the split preview of the post
// body editor (?editor=noteBody: the notes editor); with ?as=page (the Preview button) it fills the preview drawer with the post as a
// reader would see it, and opens it.
func previewHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title    string `json:"title"`
		Body     string `json:"body"`
		NoteBody string `json:"noteBody"`
	}
	if err := dsse.ReadSignals(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	editor := r.URL.Query().Get("editor")
	if editor != "" && editor != editorName && editor != notesEditor {
		http.Error(w, "unknown editor", http.StatusBadRequest)
		return
	}
	sse := dsse.NewSSE(w, r)
	if r.URL.Query().Get("as") == "page" {
		title := strings.TrimSpace(in.Title)
		if title == "" {
			title = "Untitled"
		}
		patch(sse, html.Div(
			html.H1().Class("mb-4 text-2xl font-semibold tracking-tight text-neutral-900 dark:text-white").Text(title),
			markdown.Render(in.Body, markdown.WithIDPrefix("pv-page-"), markdown.WithSize(ui.Md)),
		).Id(previewBodyID))
		patchSignals(sse, map[string]any{"postPreview": true})
		return
	}
	if editor == notesEditor {
		patch(sse, ui.MarkdownPreview(ui.MarkdownPreviewID(notesEditor), renderPreview(in.NoteBody)))
		return
	}
	patch(sse, ui.MarkdownPreview(ui.MarkdownPreviewID(editorName), renderPreview(in.Body)))
}

// editorPage renders the editor for p; p.ID 0 is a new post.
func editorPage(r *http.Request, p adminPost) htemel.Node {
	isNew := p.ID == 0
	pageTitle := "Edit post"
	crumb := p.Title
	if isNew {
		pageTitle, crumb = "New post", "New post"
	}
	postURL := "/admin/posts/new"
	if !isNew {
		postURL = "/admin/posts/" + strconv.Itoa(p.ID)
	}
	if p.Tags == nil {
		p.Tags = []string{}
	}

	header := ui.PageHeader(pageTitle,
		ui.WithBackLink("/admin/posts", "All posts"),
		ui.WithBreadcrumbs(ui.Breadcrumbs(
			ui.Crumb{Label: "Dashboard", Href: "/admin"}, ui.Crumb{Label: "Posts", Href: "/admin/posts"}, ui.Crumb{Label: crumb})),
		ui.WithMeta(statusBadge(p.Status).Id(statusMetaID), savedMeta(p, false)),
	)

	editor := ui.MarkdownEditor(editorName,
		ui.WithValue(p.Body), ui.WithRows(16), ui.WithSplit(),
		ui.WithPreview(datastar.Post("/admin/preview")), ui.WithPreviewLabel("Post preview"))
	left := ui.Stack(ui.Lg,
		ui.Field("post-title", "Title", ui.Input(html.InputTypeText, "title").Attrs(datastar.Bind("title")), ui.ErrorSignal("errors.title")),
		ui.Field("post-slug", "Slug", ui.InputGroup(htemel.Text("/blog/"), ui.SlugField("slug", "title"), nil),
			ui.HelpText("Follows the title until you edit it."), ui.ErrorSignal("errors.slug")),
		ui.Field("post-body", "Body", editor),
		ui.Field("post-tags", "Tags", ui.TagInput("tags", p.Tags, ui.TagSuggestions(adminData.tags()...), ui.TagMax(8), ui.TagSignal("tags")),
			ui.HelpText("Press Enter or comma after each tag.")),
		ui.Field("post-cover", "Cover image", ui.FileUpload("cover", ui.FileAccept("image/*"),
			ui.FileHint("Preview only: this demo does not store the file."),
			ui.FileUploadAction("$cover = evt.detail.file.name"))),
	)

	radios := []*html.LabelElement{}
	for _, s := range postStatuses {
		attrs := []htemel.Attribute{datastar.Bind("status")}
		if p.Status == s {
			attrs = append(attrs, htemel.Attr("checked", ""))
		}
		radios = append(radios, ui.Radio("status", s, statusLabel(s), attrs...))
	}
	authors := []html.SelectChild{}
	for _, a := range adminAuthors {
		authors = append(authors, ui.Choice(a, a).Selected(a == p.Author))
	}
	commentAttrs := []htemel.Attribute{datastar.Bind("comments")}
	if p.Comments {
		commentAttrs = append(commentAttrs, htemel.Attr("checked", ""))
	}
	publish := ui.Card(
		ui.CardHeader("Publish").HeadingLevel(2),
		ui.CardBody(ui.Stack(ui.Lg,
			ui.RadioGroup("Status", radios...),
			ui.Field("post-publish-at", "Publish on",
				ui.DateTimeInput("publishAt", ui.DateValue(p.PublishAt)).
					Attrs(datastar.Bind("publishAt"), datastar.Attr("disabled", "$status !== 'scheduled'")),
				ui.HelpText("Used when the status is Scheduled."), ui.ErrorSignal("errors.publishAt")),
			ui.Field("post-author", "Author", ui.Select("author", authors...).Attrs(datastar.Bind("author"))),
			ui.Switch("Allow comments", commentAttrs...),
		)),
		ui.CardFooter(
			ui.Button(ui.Secondary, "Preview", ui.WithLoading("previewing")).
				Attrs(datastar.OnClick(datastar.Post("/admin/preview?as=page"))),
			ui.Button(ui.Primary, "Save", ui.WithLoading("saving")).Type(html.ButtonTypeSubmit),
		),
	)
	side := ui.Stack(ui.Lg, publish)
	if !isNew {
		side.Children(html.Div(ui.Button(ui.Secondary, "Delete post", ui.WithIcon(ui.IconTrash())).Attrs(ui.Opens("deletePost"))).Class("flex justify-end"))
	}

	signals := map[string]any{
		"title": p.Title, "slug": p.Slug, "slugTouched": !isNew, "body": p.Body, "status": p.Status,
		"publishAt": p.PublishAt, "author": p.Author, "comments": p.Comments, "tags": p.Tags, "cover": "",
		"errors": postErrors{}, "postPreview": false, "saving": false, "previewing": false,
	}
	form := html.Form(html.Div(left, side).Class("grid items-start gap-6 @3xl:grid-cols-[minmax(0,1fr)_20rem]")).
		Id("post-form").BoolAttr("novalidate", true).
		Attrs(datastar.Signals(signals),
			datastar.OnSubmit(ui.Then(datastar.Post(postURL), "document.querySelector('#post-form [aria-invalid=true]')?.focus()"), datastar.Prevent()))

	previewDrawer := ui.DrawerSized("postPreview", ui.DrawerRight, ui.Lg, "Preview",
		html.Div(htemel.Text("Nothing to preview yet.")).Id(previewBodyID))
	overlays := []htemel.FlowContent{previewDrawer}
	if !isNew {
		overlays = append(overlays, ui.Dialog("deletePost", "Delete this post?",
			html.P().Text("“"+p.Title+"” will be deleted. This cannot be undone."),
			ui.DialogActions(
				ui.Button(ui.Secondary, "Cancel").Attrs(datastar.OnClick(ui.Close("deletePost"))),
				ui.Button(ui.Danger, "Delete", ui.WithLoading("deleting")).Attrs(datastar.OnClick(datastar.Delete(postURL))),
			)))
	}

	content := ui.StackedLayout(header, htemel.Group(append([]htemel.Node{form}, nodes(overlays)...)...))
	return adminPage(pageTitle, navPosts, true, r.URL.Query().Get("notice"), content)
}

func nodes(in []htemel.FlowContent) []htemel.Node {
	out := make([]htemel.Node, len(in))
	for i, n := range in {
		out[i] = n
	}
	return out
}

func postEditPage(w http.ResponseWriter, r *http.Request) {
	id, ok := postID(w, r)
	if !ok {
		return
	}
	p, found := adminData.get(id)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		renderPage(w, adminNotFound())
		return
	}
	renderPage(w, editorPage(r, p))
}

func postNewPage(w http.ResponseWriter, r *http.Request) {
	renderPage(w, editorPage(r, adminPost{Status: statusDraft, Author: adminAuthors[0], Comments: true, Tags: []string{}}))
}
