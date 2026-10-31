package ui_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

func init() {
	addGolden(
		golden{"sortable", ui.Sortable(
			ui.DropZone("@post('/m')", ui.Draggable("a", html.P().Text("A"))),
			ui.DropZone("@post('/m')"),
		)},
	)
}

func TestSortableIsFlowAndPhrasing(t *testing.T) {
	var _ htemel.PhrasingContent = ui.Sortable()
	var _ htemel.PalpableContent = ui.Sortable()
	got := render(t, ui.Sortable(html.P().Text("x")))
	wantAll(t, got, "<ui-sortable", "</ui-sortable>", "<p>x</p>")
}

func TestAssetsHandler(t *testing.T) {
	h := ui.Assets()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ui.js", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/javascript") {
		t.Errorf("content type %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age=31536000") {
		t.Errorf("cache control %q", cc)
	}
	if !strings.Contains(rec.Body.String(), "customElements.define('ui-sortable'") {
		t.Errorf("body is not ui.js")
	}
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("no ETag")
	}
	req := httptest.NewRequest(http.MethodGet, "/ui.js", nil)
	req.Header.Set("If-None-Match", etag)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified {
		t.Errorf("conditional status %d, want 304", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope.js", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("unknown path status %d", rec.Code)
	}
}

func TestScriptTag(t *testing.T) {
	a := render(t, ui.ScriptTag("/assets"))
	b := render(t, ui.ScriptTag("/assets/"))
	if a != b {
		t.Errorf("trailing slash changes the tag:\n%s\n%s", a, b)
	}
	wantAll(t, a, `type="module"`, `src="/assets/ui.js?v=`)
}

func TestEmbeddedScriptParses(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	path, err := filepath.Abs(filepath.Join("assets", "ui.js"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(node, "--check", path).CombinedOutput(); err != nil {
		t.Fatalf("node --check: %v\n%s", err, out)
	}
}
