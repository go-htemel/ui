package ui

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"net/http"
	"path"
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

// themeCSSHash is a short content hash of the design tokens, used as the ETag of /ui.css.
var themeCSSHash = func() string {
	sum := sha256.Sum256([]byte(themeCSS))
	return hex.EncodeToString(sum[:])[:12]
}()

// Assets returns a handler that serves the embedded script as /ui.js and the default design
// tokens (ThemeCSS) as /ui.css. Mount it under a prefix with http.StripPrefix (or chi's Mount),
// and load the script with ScriptTag using the same prefix:
//
//	mux.Handle("/assets/", http.StripPrefix("/assets", ui.Assets()))
//
// /ui.js is cacheable for a year and immutable: ScriptTag adds a ?v=<hash> query, so a new
// version of the library gets a new URL. /ui.css is Tailwind source (an @theme block), not a
// stylesheet a browser applies on its own; it is served for tooling and the Tailwind browser
// build, so it is revalidated on every use (no-cache). /fonts/Geist-Variable.woff2 and
// /fonts/GeistMono-Variable.woff2 are the self-hosted fonts of FontCSS, served as font/woff2 with
// the same year-long immutable caching as /ui.js (FontCSS and FontPreload add the ?v=<hash>
// query); they are only requested by pages that use FontCSS. They are same-origin, so no CORS
// headers are sent. A request with a matching If-None-Match
// gets a 304.
func Assets() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /ui.js", serveAsset("ui.js", "text/javascript; charset=utf-8",
		"public, max-age=31536000, immutable", uiJSHash, uiJS))
	mux.Handle("GET /ui.css", serveAsset("ui.css", "text/css; charset=utf-8",
		"public, no-cache", themeCSSHash, []byte(themeCSS)))
	for i, data := range [][]byte{geistSansWOFF2, geistMonoWOFF2} {
		f := fonts[i]
		mux.Handle("GET "+f.path, serveAsset(path.Base(f.path), "font/woff2",
			"public, max-age=31536000, immutable", f.hash, data))
	}
	return mux
}

func serveAsset(name, contentType, cacheControl, hash string, body []byte) http.Handler {
	etag := `"` + hash + `"`
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Type", contentType)
		h.Set("Cache-Control", cacheControl)
		h.Set("ETag", etag)
		h.Set("X-Content-Type-Options", "nosniff")
		if matchETag(r.Header.Get("If-None-Match"), etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(body))
	})
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
