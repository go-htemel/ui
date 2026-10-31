package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/go-htemel/ui"
)

// datastarAttributes reads the Datastar v1.0.4 attribute list that the htemel datastar package
// is tested against, from the htemel module this build uses.
func datastarAttributes(t *testing.T) map[string]bool {
	t.Helper()
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/go-htemel/htemel").Output()
	if err != nil {
		t.Fatalf("locating the htemel module: %v", err)
	}
	path := filepath.Join(strings.TrimSpace(string(out)), "datastar", "testdata", "v1.0.4-attributes.txt")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	names := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" && !strings.HasPrefix(line, "#") {
			names[line] = true
		}
	}
	if len(names) < 20 {
		t.Fatalf("%s lists only %d attributes", path, len(names))
	}
	return names
}

// datastarBase reduces a rendered attribute name to its plugin name: data-on:click__outside and
// data-on-interval__duration.4000ms become data-on and data-on-interval.
func datastarBase(name string) string {
	if i := strings.IndexByte(name, ':'); i >= 0 {
		name = name[:i]
	}
	if i := strings.Index(name, "__"); i >= 0 {
		name = name[:i]
	}
	return name
}

type tag struct {
	end  bool
	name string
}

// tags tokenizes doc and returns its start and end tags in order. It fails the test on a
// tokenizer error, an end tag that does not close the innermost open element, or an element
// left open; Datastar attributes outside the reference list fail it too.
func tags(t *testing.T, label, doc string, allowed map[string]bool) []tag {
	t.Helper()
	var out, stack []tag
	z := html.NewTokenizer(strings.NewReader(doc))
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if !errors.Is(z.Err(), io.EOF) {
				t.Fatalf("%s: tokenizer: %v", label, z.Err())
			}
			if len(stack) > 0 {
				t.Errorf("%s: unclosed elements %v", label, stack)
			}
			return out
		case html.StartTagToken, html.SelfClosingTagToken:
			tok := z.Token()
			for _, a := range tok.Attr {
				// data-ui-* attributes belong to assets/ui.js, and data-hl (a highlighted line) to the
				// code package, not Datastar, which ignores attributes whose plugin it does not know.
				if strings.HasPrefix(a.Key, "data-") && !strings.HasPrefix(a.Key, "data-ui-") && a.Key != "data-hl" && !allowed[datastarBase(a.Key)] {
					t.Errorf("%s: <%s %s>: %s is not a Datastar v1.0.4 attribute", label, tok.Data, a.Key, datastarBase(a.Key))
				}
			}
			out = append(out, tag{name: tok.Data})
			if tt == html.StartTagToken && !voidElements[tok.Data] {
				stack = append(stack, tag{name: tok.Data})
			}
		case html.EndTagToken:
			tok := z.Token()
			out = append(out, tag{end: true, name: tok.Data})
			if len(stack) == 0 || stack[len(stack)-1].name != tok.Data {
				t.Fatalf("%s: </%s> does not close the open element (open: %v)", label, tok.Data, stack)
			}
			stack = stack[:len(stack)-1]
		}
	}
}

var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true, "hr": true, "img": true,
	"input": true, "link": true, "meta": true, "source": true, "track": true, "wbr": true,
}

// checkDocument checks that doc tokenizes cleanly and that the HTML5 parser builds the tree the
// markup describes: re-serializing the parsed tree yields the same tags in the same order, so
// the parser neither moved, closed nor inserted any element.
func checkDocument(t *testing.T, label, doc string, allowed map[string]bool) {
	t.Helper()
	root, err := html.Parse(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("%s: parse: %v", label, err)
	}
	compareTags(t, label, doc, []*html.Node{root}, allowed)
}

// checkFragment is checkDocument for an element patch, parsed in a <body> context.
func checkFragment(t *testing.T, label, frag string, allowed map[string]bool) {
	t.Helper()
	body := &html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body}
	nodes, err := html.ParseFragment(strings.NewReader(frag), body)
	if err != nil {
		t.Fatalf("%s: parse: %v", label, err)
	}
	compareTags(t, label, frag, nodes, allowed)
}

func compareTags(t *testing.T, label, src string, parsed []*html.Node, allowed map[string]bool) {
	t.Helper()
	want := tags(t, label, src, allowed)
	var buf bytes.Buffer
	for _, n := range parsed {
		if err := html.Render(&buf, n); err != nil {
			t.Fatal(err)
		}
	}
	got := tags(t, label+" (reparsed)", buf.String(), allowed)
	if len(got) != len(want) {
		t.Fatalf("%s: the parser restructured the document: %d tags became %d", label, len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s: the parser restructured the document at tag %d: %v became %v", label, i, want[i], got[i])
		}
	}
}

func get(t *testing.T, h http.Handler, path string) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Code, rec.Body.String()
}

func TestPages(t *testing.T) {
	allowed := datastarAttributes(t)
	h := routes()
	paths := []string{"/"}
	for _, g := range groups {
		paths = append(paths, "/components/"+g.Slug)
	}
	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			code, body := get(t, h, p)
			if code != http.StatusOK {
				t.Fatalf("GET %s: status %d", p, code)
			}
			checkDocument(t, p, body, allowed)
			// The bootstrap must run before any stylesheet or the Tailwind script, or the first paint
			// flashes the wrong color scheme; the toggle is in the sidebar.
			boot := strings.Index(body, ui.ColorSchemeBootstrapSource())
			if boot < 0 || boot > strings.Index(body, tailwindJS) || boot > strings.Index(body, `rel="stylesheet"`) {
				t.Errorf("GET %s: ui.ColorSchemeBootstrap is missing or comes after the Tailwind script or a stylesheet", p)
			}
			if !strings.Contains(body, `aria-label="Color scheme"`) {
				t.Errorf("GET %s: the sidebar has no color scheme toggle", p)
			}
			if !strings.Contains(body, datastarJS) || !strings.Contains(body, `id="toasts"`) {
				t.Errorf("GET %s: layout is missing Datastar or the toast region", p)
			}
		})
	}
	if code, _ := get(t, h, "/components/nope"); code != http.StatusNotFound {
		t.Errorf("unknown group: status %d, want 404", code)
	}
}

func TestExamplesHaveSource(t *testing.T) {
	for _, g := range groups {
		if len(g.Examples) == 0 {
			t.Errorf("group %s has no examples", g.Slug)
		}
		for _, e := range g.Examples {
			if !strings.HasPrefix(e.Source, "func ") || !strings.Contains(e.Source, "ui.") && !strings.Contains(e.Source, "code.") && !strings.Contains(e.Source, "markdown.") {
				t.Errorf("%s / %s: unexpected source %q", g.Slug, e.Title, e.Source)
			}
		}
	}
}

// sse posts body to path and returns the HTML of the patch-elements events in the
// response, one string per event.
func sse(t *testing.T, h http.Handler, path, body string) (events, elements []string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Datastar-Request", "true")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST %s: status %d: %s", path, rec.Code, rec.Body)
	}
	for _, ev := range strings.Split(strings.TrimSpace(rec.Body.String()), "\n\n") {
		events = append(events, ev)
		var el []string
		for _, line := range strings.Split(ev, "\n") {
			if s, ok := strings.CutPrefix(line, "data: elements "); ok {
				el = append(el, s)
			}
		}
		if len(el) > 0 {
			elements = append(elements, strings.Join(el, "\n"))
		}
	}
	return events, elements
}

func TestKanbanMove(t *testing.T) {
	allowed := datastarAttributes(t)
	h := routes()
	t.Cleanup(demoBoard.reset)

	_, elements := sse(t, h, "/demo/kanban/move/done", `{"dragged":"c1"}`)
	if len(elements) != 2 {
		t.Fatalf("want a board patch and a toast, got %d patches", len(elements))
	}
	checkFragment(t, "board patch", elements[0], allowed)
	checkFragment(t, "toast patch", elements[1], allowed)
	_, done, ok := strings.Cut(elements[0], `id="column-done"`)
	if !ok || !strings.Contains(done, `id="card-c1"`) {
		t.Error("card c1 was not moved to the done column")
	}

	// An unknown card or column re-renders the board without a toast.
	if _, elements := sse(t, h, "/demo/kanban/move/nowhere", `{"dragged":"c1"}`); len(elements) != 1 {
		t.Errorf("invalid move: got %d patches, want the board only", len(elements))
	}
}

func TestToastAndValidation(t *testing.T) {
	allowed := datastarAttributes(t)
	h := routes()

	events, elements := sse(t, h, "/demo/toast?tone=danger", `{}`)
	if len(elements) != 1 || !strings.Contains(events[0], "data: selector #toasts") || !strings.Contains(events[0], "data: mode append") {
		t.Fatalf("toast: unexpected events %q", events)
	}
	checkFragment(t, "toast", elements[0], allowed)

	events, _ = sse(t, h, "/demo/validate", `{"username":"admin"}`)
	if len(events) != 1 || !strings.Contains(events[0], "event: datastar-patch-signals") || !strings.Contains(events[0], "That username is taken.") {
		t.Errorf("validate: unexpected events %q", events)
	}
}

func TestMarkdownPreview(t *testing.T) {
	allowed := datastarAttributes(t)
	h := routes()
	body := `{"journal":"# Hi\n\n<script>alert(1)</script>\n\n` + "```go\\nfunc x() {}\\n```" + `","post":"ignored"}`
	_, elements := sse(t, h, "/demo/markdown/preview?editor=journal", body)
	if len(elements) != 1 {
		t.Fatalf("want one preview patch, got %d", len(elements))
	}
	el := elements[0]
	checkFragment(t, "preview patch", el, allowed)
	for _, want := range []string{
		`<div id="journal-preview" class="min-w-0">`,
		`<div class="ui-prose ui-prose-sm">`, `<h1 id="journal-md-hi">Hi</h1>`,
		`<figure class="ui-code"`, `<span class="tok-keyword">func</span>`,
	} {
		if !strings.Contains(el, want) {
			t.Errorf("preview lacks %q:\n%s", want, el)
		}
	}
	if strings.Contains(el, "<script") || strings.Contains(el, "ignored") {
		t.Errorf("preview leaks script or another editor's text:\n%s", el)
	}

	req := httptest.NewRequest(http.MethodPost, "/demo/markdown/preview?editor=nope", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("unknown editor: status %d, want 400", rec.Code)
	}
}
