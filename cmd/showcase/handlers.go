package main

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-htemel/htemel"
	dsse "github.com/starfederation/datastar-go/datastar"

	"github.com/go-htemel/ui"
	"github.com/go-htemel/ui/markdown"
)

func routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) { renderPage(w, indexPage()) })
	r.Get("/components/{slug}", func(w http.ResponseWriter, r *http.Request) {
		g, ok := findGroup(chi.URLParam(r, "slug"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		renderPage(w, groupPage(g))
	})
	// Bare pages of the Layouts page's phone-width iframes; registered after the group route so
	// the more specific pattern wins in chi.
	r.Get("/components/layouts/frame/{name}", layoutFrame)
	r.Get("/avatar.svg", avatarSVG)
	r.Get("/placeholder.svg", placeholderSVG)
	r.Mount("/assets", http.StripPrefix("/assets", ui.Assets()))

	r.Route("/demo", func(r chi.Router) {
		r.Post("/toast", pushToast)
		r.Post("/slow", slowSave)
		r.Post("/validate", validateUsername)
		r.Post("/kanban/move/{column}", moveCard)
		r.Post("/kanban/reset", resetBoard)
		r.Get("/people", peopleHandler)
		r.Post("/progress", progressHandler)
		r.Post("/markdown/preview", markdownPreview)
	})
	adminRoutes(r)
	return r
}

// layoutFrame serves /components/layouts/frame/{appshell|split|settings}: one layout demo at full
// size, for the iframes of the Layouts page.
func layoutFrame(w http.ResponseWriter, r *http.Request) {
	pages := map[string]func() htemel.Node{
		"appshell": phoneShellPage, "split": phoneSplitPage, "settings": phoneSettingsPage,
	}
	fn, ok := pages[chi.URLParam(r, "name")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	renderPage(w, fn())
}

func renderPage(w http.ResponseWriter, n htemel.Node) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := htemel.Render(w, n); err != nil {
		slog.Error("render", "err", err)
	}
}

// patch renders n and sends it as a Datastar patch-elements event.
func patch(sse *dsse.ServerSentEventGenerator, n htemel.Node, opts ...dsse.PatchElementOption) {
	s, err := htemel.RenderString(n)
	if err == nil {
		err = sse.PatchElements(s, opts...)
	}
	if err != nil {
		slog.Error("patch", "err", err)
	}
}

// toast appends a toast to the page's ToastRegion.
func toast(sse *dsse.ServerSentEventGenerator, t ui.Tone, title, message string) {
	patch(sse, ui.Toast(ui.NextToastID(), t, title, message, 4*time.Second),
		dsse.WithSelectorID(ui.ToastRegionID), dsse.WithModeAppend())
}

func pushToast(w http.ResponseWriter, r *http.Request) {
	tone, title, msg := ui.Success, "Saved", "Pushed by the server over SSE."
	if r.URL.Query().Get("tone") == "danger" {
		tone, title, msg = ui.Danger, "Something went wrong", "Also pushed over SSE; this one is red."
	}
	if r.URL.Query().Get("tone") == "info" {
		tone = ui.Info
	}
	if t := r.URL.Query().Get("title"); t != "" {
		title = t
	}
	if m, ok := r.URL.Query()["message"]; ok {
		msg = m[0]
	}
	toast(dsse.NewSSE(w, r), tone, title, msg)
}

func slowSave(w http.ResponseWriter, r *http.Request) {
	select {
	case <-time.After(time.Second):
	case <-r.Context().Done():
		return
	}
	toast(dsse.NewSSE(w, r), ui.Success, "Saved", "The button was disabled while the request ran.")
}

func validateUsername(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username string `json:"username"`
	}
	if err := dsse.ReadSignals(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	msg := ""
	switch name := strings.TrimSpace(in.Username); {
	case name == "":
	case len(name) < 3:
		msg = "Use at least 3 characters."
	case strings.EqualFold(name, "admin"):
		msg = "That username is taken."
	}
	if err := dsse.NewSSE(w, r).MarshalAndPatchSignals(map[string]any{"errors": map[string]any{"username": msg}}); err != nil {
		slog.Error("patch signals", "err", err)
	}
}

func moveCard(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Dragged string `json:"dragged"`
	}
	if err := dsse.ReadSignals(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c, ok := demoBoard.move(in.Dragged, chi.URLParam(r, "column"))
	sse := dsse.NewSSE(w, r)
	patch(sse, demoBoard.view())
	if ok {
		toast(sse, ui.Info, "Moved “"+c.Title+"”", "The server patched the board.")
	}
}

func resetBoard(w http.ResponseWriter, r *http.Request) {
	demoBoard.reset()
	patch(dsse.NewSSE(w, r), demoBoard.view())
}

// avatarSVG serves the placeholder picture of the Avatar example.
func avatarSVG(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` +
		`<rect width="64" height="64" fill="#0ea5e9"/><circle cx="32" cy="25" r="11" fill="#e0f2fe"/>` +
		`<path d="M12 60c2-12 10-18 20-18s18 6 20 18z" fill="#e0f2fe"/></svg>`))
}

// markdownPreview renders the text of a MarkdownEditor (the editor named by ?editor=, whose
// signal has the same name) and patches its preview.
func markdownPreview(w http.ResponseWriter, r *http.Request) {
	editor := r.URL.Query().Get("editor")
	switch editor {
	case "journal", "post":
	default:
		http.Error(w, "unknown editor", http.StatusBadRequest)
		return
	}
	var signals map[string]any
	if err := dsse.ReadSignals(r, &signals); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	src, _ := signals[editor].(string)
	// The id prefix keeps headings and footnotes from taking ids the page already uses.
	body := markdown.Render(src, markdown.WithIDPrefix(editor+"-md-"), markdown.WithSize(ui.Sm))
	patch(dsse.NewSSE(w, r), ui.MarkdownPreview(ui.MarkdownPreviewID(editor), body))
}

// placeholderSVG serves the landscape picture of the Prose examples.
func placeholderSVG(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 320">` +
		`<rect width="640" height="320" fill="#c7d2fe"/><circle cx="500" cy="80" r="36" fill="#eef2ff"/>` +
		`<path d="M0 320 180 140l120 110 90-70 250 140z" fill="#6366f1"/><path d="M0 320l260-120 380 120z" fill="#4338ca"/></svg>`))
}
