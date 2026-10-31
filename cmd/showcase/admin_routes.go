package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-htemel/htemel"
)

// adminRoutes registers the admin demo under /admin.
func adminRoutes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			renderPageStatus(w, r, http.StatusNotFound, adminNotFound())
		})
		r.Method(http.MethodGet, "/", htemel.Handler(dashboardPage))
		r.Method(http.MethodGet, "/posts", htemel.Handler(postsPage))
		r.Get("/posts/list", postsList)
		r.Get("/posts/new", postNewPage)
		r.Get("/posts/{id}/edit", postEditPage)
		r.Post("/posts/{id}", postSave)
		r.Delete("/posts/{id}", postRemove)
		r.Post("/preview", previewHandler)
		r.Post("/posts/delete", postsDelete)
		r.Post("/posts/bulk-delete", postsBulkDelete)
		r.Post("/posts/{id}/duplicate", func(w http.ResponseWriter, r *http.Request) {
			if id, ok := postID(w, r); ok {
				postsDuplicate(w, r, id)
			}
		})
		r.Get("/settings", func(w http.ResponseWriter, r *http.Request) { renderPage(w, r, settingsPage(r, settingsGeneral)) })
		r.Get("/settings/{tab}", func(w http.ResponseWriter, r *http.Request) {
			switch tab := chi.URLParam(r, "tab"); tab {
			case settingsWriting, settingsAppearance:
				renderPage(w, r, settingsPage(r, tab))
			default:
				renderPageStatus(w, r, http.StatusNotFound, adminNotFound())
			}
		})
		r.Post("/settings/{tab}", settingsSave)
		r.Post("/reset", adminReset)
		r.Get("/notes", func(w http.ResponseWriter, r *http.Request) { renderPage(w, r, notesPage(r)) })
		r.Get("/notes/list", notesFilter)
		r.Get("/notes/{id}", notesShow)
		r.Post("/notes/{id}", notesSave)
		r.Method(http.MethodGet, "/media", htemel.Handler(func(*http.Request) htemel.Node {
			return adminStub("Media", navMedia, "Images and files for your posts.")
		}))
		r.Method(http.MethodGet, "/comments", htemel.Handler(func(*http.Request) htemel.Node {
			return adminStub("Comments", navComments, "Reader comments waiting for a decision.")
		}))
	})
}

// postID reads the {id} of the route, answering 404 when it is not a number.
func postID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return 0, false
	}
	return id, true
}
