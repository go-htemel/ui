package ui

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/go-htemel/htemel/html"
)

// uiJS is the optional progressive-enhancement script: the <ui-sortable> custom element that adds
// touch dragging to Draggable and DropZone. Every component works without it.
//
//go:embed assets/ui.js
var uiJS []byte

// uiJSHash is a short content hash of uiJS, used as the ETag and as the ScriptTag cache buster.
var uiJSHash = func() string {
	sum := sha256.Sum256(uiJS)
	return hex.EncodeToString(sum[:])[:12]
}()

// Assets returns a handler that serves the embedded script as /ui.js. Mount it under a prefix
// with http.StripPrefix (or chi's Mount), and load it with ScriptTag using the same prefix:
//
//	mux.Handle("/assets/", http.StripPrefix("/assets", ui.Assets()))
//
// The response is cacheable for a year and immutable: ScriptTag adds a ?v=<hash> query, so a new
// version of the library gets a new URL. A request with a matching If-None-Match gets a 304.
func Assets() http.Handler {
	etag := `"` + uiJSHash + `"`
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ui.js", func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Type", "text/javascript; charset=utf-8")
		h.Set("Cache-Control", "public, max-age=31536000, immutable")
		h.Set("ETag", etag)
		h.Set("X-Content-Type-Options", "nosniff")
		if matchETag(r.Header.Get("If-None-Match"), etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		http.ServeContent(w, r, "ui.js", time.Time{}, bytes.NewReader(uiJS))
	})
	return mux
}

func matchETag(header, etag string) bool {
	for _, c := range strings.Split(header, ",") {
		if c = strings.TrimSpace(c); c == etag || c == "W/"+etag || c == "*" {
			return true
		}
	}
	return false
}

// ScriptTag returns the <script type="module"> element that loads the script served by Assets
// at prefix ("/assets", with or without a trailing slash). Put it in <head>; a module script is
// deferred, so it runs after the document is parsed, and it needs no inline script, so a strict
// Content-Security-Policy (script-src 'self') allows it.
func ScriptTag(prefix string) *html.ScriptElement {
	return html.Script().
		Type(html.ScriptTypeModule).
		Src(strings.TrimRight(prefix, "/") + "/ui.js?v=" + uiJSHash)
}
