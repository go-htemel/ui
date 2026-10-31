package ui_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/go-htemel/ui"
)

func TestFontAssets(t *testing.T) {
	h := ui.Assets()
	for _, f := range []struct{ path, file string }{
		{"/fonts/Geist-Variable.woff2", "assets/fonts/Geist-Variable.woff2"},
		{"/fonts/GeistMono-Variable.woff2", "assets/fonts/GeistMono-Variable.woff2"},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, f.path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d", f.path, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "font/woff2" {
			t.Errorf("%s: content type %q", f.path, ct)
		}
		if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") || !strings.Contains(cc, "max-age=31536000") {
			t.Errorf("%s: cache control %q", f.path, cc)
		}
		want, err := os.ReadFile(f.file)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(rec.Body.Bytes(), want) || !bytes.HasPrefix(want, []byte("wOF2")) {
			t.Errorf("%s: body is not the embedded woff2", f.path)
		}
		req := httptest.NewRequest(http.MethodGet, f.path, nil)
		req.Header.Set("If-None-Match", rec.Header().Get("ETag"))
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotModified {
			t.Errorf("%s: conditional status %d, want 304", f.path, rec.Code)
		}
	}
}

func TestFontCSS(t *testing.T) {
	css := ui.FontCSS("/assets")
	if css != ui.FontCSS("/assets/") {
		t.Error("trailing slash changes the CSS")
	}
	wantAll(t, css,
		`font-family: "Geist";`, `font-family: "Geist Mono";`,
		"font-weight: 100 900;", "font-display: swap;", `format("woff2")`,
		`url("/assets/fonts/Geist-Variable.woff2?v=`, `url("/assets/fonts/GeistMono-Variable.woff2?v=`,
		`--font-sans: "Geist", ui-sans-serif, system-ui, sans-serif;`,
		`--font-mono: "Geist Mono", ui-monospace,`)
	if !strings.Contains(ui.FontCSS("/static/ui"), `url("/static/ui/fonts/Geist-Variable`) {
		t.Error("prefix is not applied")
	}
}

// TestFontPreloadMatchesCSS checks that every preload link points at exactly the URL FontCSS
// declares, or the browser would fetch the file twice.
func TestFontPreloadMatchesCSS(t *testing.T) {
	links := ui.FontPreload("/assets")
	if len(links) != 2 {
		t.Fatalf("%d links, want 2", len(links))
	}
	css := ui.FontCSS("/assets")
	href := regexp.MustCompile(`href="([^"]+)"`)
	for _, l := range links {
		got := render(t, l)
		wantAll(t, got, `rel="preload"`, `as="font"`, `type="font/woff2"`, `crossorigin`)
		m := href.FindStringSubmatch(got)
		if m == nil || !strings.Contains(m[1], "?v=") || !strings.Contains(css, `url("`+m[1]+`")`) {
			t.Errorf("preload %s does not match the @font-face URL", got)
		}
	}
}
