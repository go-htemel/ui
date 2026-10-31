package code_test

import (
	"flag"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	htmlel "github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui/code"
)

var update = flag.Bool("update", false, "rewrite testdata/*.golden.html from the current output")

// The Block and Inline types are content of the right kind (compile-time checks).
var (
	_ htemel.FlowContent                   = (*code.BlockElement)(nil)
	_ htemel.PhrasingContent               = (*htmlel.CodeElement)(nil)
	_ interface{ SetAttr(string, string) } = (*code.BlockElement)(nil)
)

const goSrc = `package main

import "fmt"

// main prints a greeting.
func main() {
	fmt.Println("hello", 42)
}
`

const jsonSrc = `{"name": "ui", "tags": ["go", "htemel"], "stable": true, "n": 1.5, "x": null}`

func render(t *testing.T, n htemel.Node) string {
	t.Helper()
	s, err := htemel.RenderString(n)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestGolden(t *testing.T) {
	cases := []struct {
		name string
		node htemel.Node
	}{
		{"block-go", code.Block("go", goSrc)},
		{"block-line-numbers", code.Block("go", goSrc, code.WithLineNumbers())},
		{"block-start-highlight", code.Block("go", goSrc, code.WithLineNumbers(), code.WithStart(41),
			code.WithHighlight(46), code.WithHighlightRange(48, 49))},
		{"block-title", code.Block("go", goSrc, code.WithTitle("main.go"))},
		{"block-wrap", code.Block("go", goSrc, code.WithWrap())},
		{"block-no-copy", code.Block("go", goSrc, code.WithoutCopyButton())},
		{"block-title-no-copy", code.Block("go", goSrc, code.WithoutCopyButton(), code.WithTitle("main.go"))},
		{"block-id", code.Block("go", goSrc, code.WithID("Quick Start #1"))},
		{"block-json", code.Block("json", jsonSrc)},
		{"block-bash", code.Block("bash", "# install\ngo get github.com/go-htemel/ui@latest && echo \"done $?\"")},
		{"block-unknown-language", code.Block("no-such-lang", "a <b> & c")},
		{"block-empty-lines", code.Block("css", "a {\n\n  color: red;\n}\n", code.WithLineNumbers())},
		{"inline-go", code.Inline("go", `ui.Button(ui.Primary, "Save")`)},
	}
	seen := map[string]bool{}
	for _, tc := range cases {
		seen[tc.name] = true
		t.Run(tc.name, func(t *testing.T) {
			got := render(t, tc.node) + "\n"
			path := filepath.Join("testdata", tc.name+".golden.html")
			if *update {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%v (run go test ./code -update to create it)", err)
			}
			if got != string(want) {
				t.Errorf("output differs from %s (rerun with -update if intended)\n got: %s\nwant: %s", path, got, want)
			}
		})
	}
	if !*update {
		files, _ := filepath.Glob(filepath.Join("testdata", "*.golden.html"))
		for _, f := range files {
			if name := strings.TrimSuffix(filepath.Base(f), ".golden.html"); !seen[name] {
				t.Errorf("stale golden file %s has no case", f)
			}
		}
	}
}

var tag = regexp.MustCompile(`<[^>]*>`)

// visibleText is the text of the <code> element of a rendered block: tags stripped, entities
// decoded.
func codeText(t *testing.T, out string) string {
	t.Helper()
	i := strings.Index(out, "<code")
	j := strings.Index(out, "</code>")
	if i < 0 || j < 0 {
		t.Fatalf("no <code> in %s", out)
	}
	return html.UnescapeString(tag.ReplaceAllString(out[i:j], ""))
}

// TestEscaping: markup and ampersands in the source are escaped everywhere they are rendered
// (the highlighted lines and the copy template), in every language.
func TestEscaping(t *testing.T) {
	const src = `<script>alert("x & y")</script> &amp; </template><b>`
	for _, lang := range []string{"html", "js", "go", "bash", "json", "css", "markdown", "", "nope"} {
		out := render(t, code.Block(lang, src))
		for _, bad := range []string{"<script", "<b>", "</template><b"} {
			if strings.Contains(out, bad) {
				t.Errorf("%s: output contains raw %q:\n%s", lang, bad, out)
			}
		}
		if got := codeText(t, out); got != src {
			t.Errorf("%s: code text %q, want %q", lang, got, src)
		}
		// Exactly one template, holding the escaped source.
		if n := strings.Count(out, "<template>"); n != 1 {
			t.Errorf("%s: %d templates", lang, n)
		}
		tpl := out[strings.Index(out, "<template>")+len("<template>") : strings.Index(out, "</template>")]
		if html.UnescapeString(tpl) != src {
			t.Errorf("%s: template %q, want the source", lang, tpl)
		}
	}
	inline := render(t, code.Inline("html", `<script>&</script>`))
	if strings.Contains(inline, "<script") || !strings.Contains(inline, "&lt;") || !strings.Contains(inline, "&amp;") {
		t.Errorf("inline not escaped: %s", inline)
	}
}

// TestRoundTrip: whatever the language, the text of the rendered code is the source, and the
// source has one .line per line.
func TestRoundTrip(t *testing.T) {
	srcs := map[string]string{
		"go":       goSrc,
		"css":      "a > b { color: #fff; /* x */ }\n\n\n.c::after { content: \"\\201C\"; }",
		"html":     "<!doctype html>\n<p class=\"a\" hidden>Hi &amp; bye</p>\n<script>let a = `x\ny`;</script>",
		"js":       "const a = `multi\nline ${x}`; // c\n/* block\ncomment */\n",
		"json":     jsonSrc,
		"bash":     "echo $HOME | grep 'x' > /dev/null\ncat <<EOF\nhi\nEOF\n",
		"sql":      "SELECT a, count(*) FROM t WHERE b = 'x' GROUP BY a;",
		"yaml":     "a:\n  - b: 1\n  - c: \"d\"\n# e",
		"markdown": "# Title\n\n* item `code`\n\n```go\nx := 1\n```\n",
		"diff":     "--- a\n+++ b\n@@ -1 +1 @@\n-old\n+new\n",
		"":         "plain\n\ttext   with  spaces  ",
	}
	for lang, src := range srcs {
		out := render(t, code.Block(lang, src))
		want := strings.TrimRight(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
		if got := codeText(t, out); got != want {
			t.Errorf("%q: code text differs\n got: %q\nwant: %q", lang, got, want)
		}
		if got, want := strings.Count(out, `class="line"`)+strings.Count(out, `class="line" data-ui-hl`), strings.Count(want, "\n")+1; got != want {
			t.Errorf("%q: %d line spans, want %d", lang, got, want)
		}
	}
	// CRLF and a lone CR are line breaks.
	if got := codeText(t, render(t, code.Block("", "a\r\nb\rc\n\n"))); got != "a\nb\nc" {
		t.Errorf("line endings: %q", got)
	}
}

func TestEmptySource(t *testing.T) {
	out := render(t, code.Block("go", ""))
	if n := strings.Count(out, `class="line"`); n != 1 {
		t.Errorf("empty source has %d lines, want 1", n)
	}
}

// TestTokenKinds: the semantic classes come out for representative tokens, and only the
// documented set is ever used.
func TestTokenKinds(t *testing.T) {
	out := render(t, code.Block("go", goSrc))
	for _, want := range []string{
		`<span class="tok-keyword">package</span>`,
		`tok-string`, `tok-comment`, `tok-number`, `tok-function`, `tok-punctuation`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("go: missing %s in\n%s", want, out)
		}
	}
	if !strings.Contains(render(t, code.Block("html", `<a href="x">`)), `tok-tag`) {
		t.Error("html: no tok-tag")
	}
	if !strings.Contains(render(t, code.Block("json", jsonSrc)), `class="tok-attribute"`) {
		t.Error("json: keys are not attributes")
	}
	if !strings.Contains(render(t, code.Block("go", "var x int = nil")), `tok-type`) {
		t.Error("go: no tok-type")
	}
	allowed := map[string]bool{}
	for _, k := range strings.Fields("keyword builtin name function type string number literal attribute tag comment operator punctuation error") {
		allowed["tok-"+k] = true
	}
	for _, lang := range []string{"go", "css", "html", "js", "json", "bash", "sql", "yaml", "markdown", "diff", "python", "rust"} {
		src := goSrc + jsonSrc + "<a href=x> SELECT 1; # c\n{{ x }} $y ++ --z"
		for _, m := range regexp.MustCompile(`class="(tok-[a-z]+)"`).FindAllStringSubmatch(render(t, code.Block(lang, src)), -1) {
			if !allowed[m[1]] {
				t.Errorf("%s: unexpected class %s", lang, m[1])
			}
		}
	}
}

// TestUnknownLanguageIsPlain: an unknown or empty language renders the text unhighlighted, with
// no token spans and no error.
func TestUnknownLanguageIsPlain(t *testing.T) {
	for _, lang := range []string{"", "text", "plaintext", "definitely-not-a-language", "  "} {
		out := render(t, code.Block(lang, "func main() { return 1 }"))
		if strings.Contains(out, "tok-") {
			t.Errorf("%q: highlighted:\n%s", lang, out)
		}
		if got := codeText(t, out); got != "func main() { return 1 }" {
			t.Errorf("%q: text %q", lang, got)
		}
	}
	// Language names are case-insensitive.
	if a, b := render(t, code.Block("Go", goSrc)), render(t, code.Block("go", goSrc)); a != b || !strings.Contains(a, "tok-keyword") {
		t.Errorf("Go and go differ: %s vs %s", a, b)
	}
}

// TestDeterministic: the same source gives the same HTML and the same signal; different source
// or language gives a different one; WithID overrides.
func TestDeterministic(t *testing.T) {
	a := render(t, code.Block("go", goSrc))
	for range 3 {
		if b := render(t, code.Block("go", goSrc)); a != b {
			t.Fatal("two renders of the same block differ")
		}
	}
	sig := regexp.MustCompile(`data-signals:(_copied_[a-z0-9_]+?)__ifmissing="false"`)
	id := func(out string) string {
		m := sig.FindStringSubmatch(out)
		if m == nil {
			t.Fatalf("no copied signal in %s", out)
		}
		return m[1]
	}
	if !regexp.MustCompile(`^_copied_[0-9a-f]{8}$`).MatchString(id(a)) {
		t.Errorf("signal %q is not _copied_ + 8 hex digits", id(a))
	}
	ids := map[string]bool{
		id(a): true,
		id(render(t, code.Block("go", goSrc+"x"))):               true,
		id(render(t, code.Block("css", goSrc))):                  true,
		id(render(t, code.Block("go", goSrc, code.WithID("a")))): true,
		id(render(t, code.Block("go", goSrc, code.WithID("b")))): true,
	}
	if len(ids) != 5 {
		t.Errorf("signals collide: %v", ids)
	}
	if got := id(render(t, code.Block("go", goSrc, code.WithID("Quick Start #1")))); got != "_copied_quick_start__1" {
		t.Errorf("WithID signal %q", got)
	}
	// The handler, the icons and the live region all use that signal.
	if n := strings.Count(a, id(a)); n < 6 {
		t.Errorf("signal %s appears %d times, want it in the declaration, handler and bindings", id(a), n)
	}
	// Options that do not change the source do not change the id.
	if id(render(t, code.Block("go", goSrc, code.WithLineNumbers(), code.WithWrap()))) != id(a) {
		t.Error("layout options changed the id")
	}
}

func TestOptions(t *testing.T) {
	has := strings.Contains
	plain := render(t, code.Block("go", goSrc))
	if has(plain, "ui-code-lines") || has(plain, "ui-code-wrap") || has(plain, "data-ui-hl") || has(plain, "ui-code-title") || has(plain, "--ui-code") {
		t.Errorf("defaults carry options:\n%s", plain)
	}
	if !has(plain, `aria-label="Copy code"`) || !has(plain, `aria-live="polite"`) || !has(plain, `aria-label="Code actions"`) {
		t.Errorf("copy button or live region missing:\n%s", plain)
	}
	if !has(plain, `tabindex="0"`) {
		t.Error("the pre is not focusable")
	}

	n := render(t, code.Block("go", goSrc, code.WithLineNumbers(), code.WithStart(98)))
	if !has(n, "ui-code-lines") || !has(n, "--ui-code-start:98") || !has(n, "--ui-code-digits:3") {
		t.Errorf("line numbers: %s", n)
	}
	if o := render(t, code.Block("go", goSrc, code.WithLineNumbers())); !has(o, "--ui-code-digits:2") || has(o, "--ui-code-start") {
		t.Errorf("default start: %s", o)
	}
	if o := render(t, code.Block("go", goSrc, code.WithStart(0), code.WithStart(-3), code.WithLineNumbers())); has(o, "--ui-code-start") {
		t.Errorf("invalid start accepted: %s", o)
	}

	hl := render(t, code.Block("go", goSrc, code.WithHighlight(2, 99), code.WithHighlightRange(5, 6), code.WithStart(1)))
	if got := strings.Count(hl, "data-ui-hl"); got != 3 {
		t.Errorf("%d highlighted lines, want 3 (2, 5, 6; 99 is outside)", got)
	}
	// Highlight numbers follow WithStart.
	hl = render(t, code.Block("go", goSrc, code.WithStart(10), code.WithHighlight(10, 2)))
	if got := strings.Count(hl, "data-ui-hl"); got != 1 || !has(hl, `<span class="line" data-ui-hl>`) {
		t.Errorf("highlight with start: %d, %s", got, hl)
	}
	if !has(render(t, code.Block("go", goSrc, code.WithHighlightRange(1, 1<<62))), "data-ui-hl") {
		t.Error("huge range failed")
	}

	if o := render(t, code.Block("go", goSrc, code.WithTitle("a <b>.go"))); !has(o, `<span class="ui-code-title">a &lt;b&gt;.go</span>`) || !has(o, "ui-code-titled") || !has(o, `aria-label="a &lt;b&gt;.go"`) {
		t.Errorf("title: %s", o)
	}
	if o := render(t, code.Block("go", goSrc, code.WithWrap())); !has(o, "ui-code-wrap") {
		t.Errorf("wrap: %s", o)
	}
	nc := render(t, code.Block("go", goSrc, code.WithoutCopyButton()))
	for _, s := range []string{"<template", "<button", "toolbar", "data-signals", "aria-live"} {
		if has(nc, s) {
			t.Errorf("without copy button the output has %q:\n%s", s, nc)
		}
	}
	if o := render(t, code.Block("go", goSrc, code.WithoutCopyButton(), code.WithTitle("x"))); !has(o, "ui-code-toolbar") || has(o, "<button") {
		t.Errorf("title without button: %s", o)
	}
	// Nil options are skipped.
	render(t, code.Block("go", goSrc, nil))
}

// TestCopyButton: the click handler reads the template of its own figure and the icons fade.
func TestCopyButton(t *testing.T) {
	out := render(t, code.Block("go", goSrc))
	for _, want := range []string{
		`navigator.clipboard.writeText(el.closest(&#39;figure&#39;).querySelector(&#39;template&#39;).content.textContent)`,
		`data-on:click=`, `data-text=`, `data-class:opacity-0=`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in\n%s", want, out)
		}
	}
	// The motion rule of the library: every transition is switched off for reduced motion.
	if strings.Count(out, "transition-opacity") != strings.Count(out, "motion-reduce:transition-none") {
		t.Error("a transition lacks motion-reduce:transition-none")
	}
	if !strings.Contains(out, "duration-150") {
		t.Error("icon swap is not 150ms")
	}
}

// TestAttrsDelegate: Attrs, SetAttr and Class reach the figure.
func TestAttrsDelegate(t *testing.T) {
	b := code.Block("go", "x")
	b.SetAttr("data-x", "1")
	b.Attrs(datastar.Show("$a")).Class("mt-4")
	out := render(t, b)
	for _, want := range []string{`data-x="1"`, `data-show="$a"`, `class="ui-code mt-4"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	// A block is flow content: it fits where a Card or a paragraph does.
	render(t, htmlel.Div(code.Block("go", "x")))
}
