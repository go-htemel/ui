package markdown_test

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/microcosm-cc/bluemonday"

	"github.com/go-htemel/ui"
	"github.com/go-htemel/ui/markdown"
)

var update = flag.Bool("update", false, "rewrite testdata/*.golden.html")

var _ htemel.FlowContent = markdown.Render("")

func render(t *testing.T, src string, opts ...markdown.Option) string {
	t.Helper()
	out, err := htemel.RenderString(markdown.Render(src, opts...))
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func wantAll(t *testing.T, got string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in\n%s", w, got)
		}
	}
}

func wantNone(t *testing.T, got string, bad ...string) {
	t.Helper()
	for _, b := range bad {
		if strings.Contains(strings.ToLower(got), strings.ToLower(b)) {
			t.Errorf("found %q in\n%s", b, got)
		}
	}
}

const sample = `# Journal, 31 October

Woke early; the **frost** was _back_. ~~Skipped~~ the run, read instead: https://example.com/notes.

| Day | Miles | Mood |
|:----|------:|:----:|
| Mon | 3.1   | good |
| Tue | 0     | meh  |

- [x] Write the entry
- [ ] Water the plants
  1. the fern
  2. the cactus

> Small steps, every day.

Inline ` + "`code`" + ` and a footnote[^1].

` + "```go" + `
func main() {
	fmt.Println("hi")
}
` + "```" + `

    indented code

---

![A placeholder](/img/cover.png "Cover")

[^1]: Footnotes go at the end.
`

func TestGFM(t *testing.T) {
	got := render(t, sample)
	wantAll(t, got,
		`<div class="ui-prose">`,
		`<h1 id="journal-31-october">Journal, 31 October</h1>`,
		`<strong>frost</strong>`, `<em>back</em>`, `<del>Skipped</del>`,
		`<a href="https://example.com/notes" rel="nofollow">https://example.com/notes</a>`,
		`<th align="left">Day</th>`, `<th align="right">Miles</th>`, `<th align="center">Mood</th>`, `<td align="right">3.1</td>`,
		`<ul class="contains-task-list">`,
		`<li class="task-list-item"><input checked="" disabled="" type="checkbox"> Write the entry</li>`,
		`<input disabled="" type="checkbox"> Water the plants`,
		`<ol>`, `<blockquote>`, `<code>code</code>`,
		`<sup id="fnref:1"><a href="#fn:1" class="footnote-ref" role="doc-noteref"`,
		`<div class="footnotes" role="doc-endnotes">`, `<li id="fn:1">`, `class="footnote-backref" role="doc-backlink"`,
		`<hr>`, `<img src="/img/cover.png" alt="A placeholder" title="Cover">`,
	)
}

func TestCodeBlocks(t *testing.T) {
	got := render(t, sample)
	wantAll(t, got,
		`<figure class="ui-code" aria-label="go code"><pre tabindex="0"><code class="language-go"><span class="line"><span class="tok-keyword">func</span>`,
		`<span class="tok-string">&#34;hi&#34;</span>`,
		`<figure class="ui-code" aria-label="Code"><pre tabindex="0"><code><span class="line">indented code</span>`,
	)
	// No copy button: it needs a <template> and a Datastar signal, and neither survives sanitising.
	wantNone(t, got, "<template", "data-", "<button")
	if n := strings.Count(got, `<figure class="ui-code"`); n != 2 {
		t.Errorf("%d code blocks, want 2", n)
	}
	// Unknown languages fall back to plain text.
	wantAll(t, render(t, "```nope\n<a & b>\n```"), `<code class="language-nope"><span class="line">&lt;a &amp; b&gt;</span>`)
}

// xss is a corpus of hostile Markdown. None of it may produce script, a handler attribute, a
// dangerous URL scheme or an embedding element in the default (sanitised) mode.
var xss = []string{
	`<script>alert(1)</script>`,
	`<SCRIPT SRC=//evil.example/x.js></SCRIPT>`,
	`<img src=x onerror=alert(1)>`,
	`<img src="x" onerror="alert(1)"/>`,
	`<svg onload=alert(1)><circle r=1></svg>`,
	`<iframe src="https://evil.example"></iframe>`,
	`<iframe srcdoc="<script>alert(1)</script>"></iframe>`,
	`<object data="evil.swf"></object><embed src="evil.swf">`,
	`<a href="javascript:alert(1)">x</a>`,
	`[click](javascript:alert(1))`,
	`[click](JaVaScRiPt:alert(1))`,
	`[click](&#106;avascript:alert(1))`,
	`[click](javascript&#58;alert(1))`,
	`[click]( javascript:alert(1) )`,
	`[click](vbscript:msgbox(1))`,
	`[click](data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==)`,
	`![x](data:image/svg+xml;base64,PHN2ZyBvbmxvYWQ9YWxlcnQoMSk+)`,
	`![x](javascript:alert(1))`,
	"[ref]\n\n[ref]: javascript:alert(1)",
	`<javascript:alert(1)>`,
	`<div onclick="alert(1)">nested <b onmouseover=alert(1)>html</b></div>`,
	`<p style="background:url(javascript:alert(1))">styled</p>`,
	`<form action="/x"><input name=a><button formaction="javascript:alert(1)">go</button></form>`,
	`<math><mi xlink:href="javascript:alert(1)">x</mi></math>`,
	`<details open ontoggle=alert(1)><summary>x</summary></details>`,
	`<a href="https://ok.example" onclick="alert(1)">ok</a>`,
	"<style>body{display:none}</style>",
	"<base href=\"https://evil.example/\">",
	"<meta http-equiv=\"refresh\" content=\"0;url=https://evil.example\">",
	"<!-- <script>alert(1)</script> -->",
	"```html\n<script>alert(1)</script>\n```", // shown as code: escaped text, never a tag
	"`<script>alert(1)</script>`",
}

var (
	handlerAttr = regexp.MustCompile(`(?i)<[^>]*\son[a-z]+\s*=`)
	badScheme   = regexp.MustCompile(`(?i)(href|src|action|formaction|xlink:href)\s*=\s*["']?\s*(javascript|vbscript|data):`)
	badElement  = regexp.MustCompile(`(?i)<(script|iframe|object|embed|svg|math|style|base|meta|form|button|link)\b`)
	styleAttr   = regexp.MustCompile(`(?i)<[^>]*\sstyle\s*=`)
)

func TestXSSNeutralised(t *testing.T) {
	for _, src := range xss {
		for _, opts := range [][]markdown.Option{nil, {markdown.WithLinkTarget("_blank")}, {markdown.WithPolicy(markdown.Policy())}} {
			got := render(t, src, opts...)
			for name, re := range map[string]*regexp.Regexp{"handler attribute": handlerAttr, "dangerous scheme": badScheme, "element": badElement, "style attribute": styleAttr} {
				if m := re.FindString(got); m != "" {
					t.Errorf("%s survived (%q) for %q:\n%s", name, m, src, got)
				}
			}
		}
	}
	// Hostile code is shown, escaped.
	wantAll(t, render(t, "```html\n<script>alert(1)</script>\n```"), `&lt;`, `script`)
}

func TestUnsafeHTML(t *testing.T) {
	src := "<div class=\"callout\" data-x=\"1\">Trusted <em>HTML</em></div>\n\n<script>track()</script>"
	got := render(t, src, markdown.WithUnsafeHTML())
	wantAll(t, got, `<div class="callout" data-x="1">Trusted <em>HTML</em></div>`, `<script>track()</script>`)
	// The default drops it: goldmark omits raw HTML and the sanitiser would remove it anyway.
	wantNone(t, render(t, src), `class="callout"`, `<script`)
}

func TestCustomPolicy(t *testing.T) {
	p := bluemonday.StrictPolicy() // text only
	got := render(t, "# Hi\n\n**bold**", markdown.WithPolicy(p))
	wantAll(t, got, `<div class="ui-prose">`, "Hi", "bold")
	wantNone(t, got, "<h1", "<strong")
}

func TestHeadingOffset(t *testing.T) {
	got := render(t, "# One\n\n## Two\n\n###### Six", markdown.WithHeadingOffset(1))
	wantAll(t, got, `<h2 id="one">One</h2>`, `<h3 id="two">Two</h3>`, `<h6 id="six">Six</h6>`)
	wantNone(t, got, "<h1")
	wantAll(t, render(t, "# One", markdown.WithHeadingOffset(9)), `<h6`)
	wantAll(t, render(t, "# One", markdown.WithHeadingOffset(-1)), `<h1`)
}

func TestIDsAndTargets(t *testing.T) {
	got := render(t, "# Body preview\n\nSee [docs](https://ok.example), [below](#body-preview) and a note[^n].\n\n[^n]: x",
		markdown.WithIDPrefix("post-"), markdown.WithLinkTarget("_blank"))
	wantAll(t, got,
		`<h1 id="post-body-preview">`,
		`<a href="https://ok.example" target="_blank" rel="nofollow noopener">docs</a>`,
		`<a href="#body-preview" rel="nofollow">below</a>`, // in-page links get no target
		`id="post-fnref:1"`, `href="#post-fn:1"`, `<li id="post-fn:1">`,
	)
	wantNone(t, render(t, "# A", markdown.WithoutHeadingIDs()), "id=")
	wantNone(t, render(t, "[x](https://a.example)", markdown.WithLinkTarget("evil")), "target")
}

func TestSizes(t *testing.T) {
	wantAll(t, render(t, "x", markdown.WithSize(ui.Sm)), `<div class="ui-prose ui-prose-sm">`)
	wantAll(t, render(t, "x", markdown.WithSize(ui.Lg)), `<div class="ui-prose ui-prose-lg">`)
	wantAll(t, render(t, "x", markdown.WithSize(ui.Md)), `<div class="ui-prose">`)
}

func TestDeterministicAndConcurrent(t *testing.T) {
	want := render(t, sample)
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got, _ := htemel.RenderString(markdown.Render(sample)); got != want {
				t.Error("output differs between renders")
			}
		}()
	}
	wg.Wait()
	if a, b := markdown.HTML(sample), markdown.HTML(sample); a != b {
		t.Error("HTML is not deterministic")
	}
}

func TestInvalidUTF8(t *testing.T) {
	wantAll(t, render(t, "a\xffb"), "a\uFFFDb")
}

func TestPlain(t *testing.T) {
	got := markdown.Plain(sample)
	for _, w := range []string{
		"Journal, 31 October\n\nWoke early; the frost was back. Skipped the run",
		"Day Miles Mood", "Write the entry", "the fern", "Small steps, every day.",
		"Inline code and a footnote", "func main() {", "indented code", "A placeholder",
	} {
		if !strings.Contains(got, w) {
			t.Errorf("Plain lacks %q:\n%s", w, got)
		}
	}
	for _, bad := range []string{"**", "~~", "[x]", "|", "```", "#", "<"} {
		if strings.Contains(got, bad) {
			t.Errorf("Plain keeps markup %q:\n%s", bad, got)
		}
	}
	if got := markdown.Plain("Hi <b>there</b>\n\n<div>block</div>"); got != "Hi there" {
		t.Errorf("Plain with raw HTML = %q", got)
	}
}

func TestExcerpt(t *testing.T) {
	for _, tc := range []struct {
		src  string
		n    int
		want string
	}{
		{"# Title\n\nSome **bold** words here.", 100, "Title Some bold words here."},
		{"# Title\n\nSome **bold** words here.", 16, "Title Some bold…"},
		{"Hello, world. More text", 13, "Hello, world…"},
		{"Supercalifragilistic", 5, "Super…"},
		{"anything", 0, ""},
		{"Ünïcödé wörds ärë fïnë", 13, "Ünïcödé wörds…"},
	} {
		if got := markdown.Excerpt(tc.src, tc.n); got != tc.want {
			t.Errorf("Excerpt(%q, %d) = %q, want %q", tc.src, tc.n, got, tc.want)
		}
	}
}

func TestGoldenDocument(t *testing.T) {
	got := render(t, sample, markdown.WithIDPrefix("doc-")) + "\n"
	path := filepath.Join("testdata", "document.golden.html")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run go test -update)", err)
	}
	if got != string(want) {
		t.Errorf("output differs from %s:\n%s", path, got)
	}
}
