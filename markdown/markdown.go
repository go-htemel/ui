package markdown

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	ghtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"github.com/go-htemel/ui"
	"github.com/go-htemel/ui/code"
)

// Render converts Markdown to sanitised HTML wrapped in ui.Prose (see WithSize):
//
//	markdown.Render(post.Body, markdown.WithHeadingOffset(1))
//
// It returns the Prose <div>, so element methods still chain (AddClass("max-w-prose"), Id).
// Rendering never fails: src is UTF-8 text, and invalid bytes are replaced.
func Render(src string, opts ...Option) *html.DivElement {
	c := collect(opts)
	body := htemel.TextUnsafe(toHTML(src, c))
	switch c.size {
	case ui.Sm, ui.Xs:
		return ui.ProseSm(body)
	case ui.Lg, ui.Xl:
		return ui.ProseLg(body)
	}
	return ui.Prose(body)
}

// HTML is Render without the Prose wrapper: the sanitised HTML as a string, for storing a
// rendered copy next to the source or for feeds.
func HTML(src string, opts ...Option) string { return toHTML(src, collect(opts)) }

func toHTML(src string, c config) string {
	md := newMarkdown(c)
	var buf bytes.Buffer
	if err := md.Convert([]byte(strings.ToValidUTF8(src, "�")), &buf); err != nil {
		// Convert only fails when the writer does; a bytes.Buffer does not. Fall back to the
		// escaped source rather than dropping the content.
		buf.Reset()
		buf.WriteString("<pre>")
		buf.Write(util.EscapeHTML([]byte(src)))
		buf.WriteString("</pre>")
	}
	if c.unsafe {
		return buf.String()
	}
	p := c.policy
	if p == nil {
		p = defaultPolicy
	}
	return p.SanitizeReader(&buf).String()
}

// defaultPolicy is shared: a bluemonday policy is safe for concurrent use once built.
var defaultPolicy = Policy()

func newMarkdown(c config) goldmark.Markdown {
	parserOpts := []parser.Option{parser.WithASTTransformers(util.Prioritized(transformer{c}, 1000))}
	if !c.noHeadingIDs {
		parserOpts = append(parserOpts, parser.WithAutoHeadingID())
	}
	rendererOpts := []renderer.Option{renderer.WithNodeRenderers(
		util.Prioritized(codeRenderer{}, 100),
		util.Prioritized(taskRenderer{}, 100), // before the TaskList extension's own (500)
	)}
	if c.unsafe {
		rendererOpts = append(rendererOpts, ghtml.WithUnsafe())
	}
	return goldmark.New(
		goldmark.WithExtensions(
			// GFM (extension.GFM) with table alignment as align attributes: the sanitiser drops
			// style attributes, which is how goldmark writes alignment by default.
			extension.NewTable(extension.WithTableCellAlignMethod(extension.TableCellAlignAttribute)),
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
			extension.NewFootnote(extension.WithFootnoteIDPrefix(c.idPrefix)),
		),
		goldmark.WithParserOptions(parserOpts...),
		goldmark.WithRendererOptions(rendererOpts...),
	)
}

// transformer applies the heading offset, the id prefix and the link target, and marks task
// lists with the GitHub class names.
type transformer struct{ c config }

func (t transformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	b := reader.Source()
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Heading:
			n.Level = min(n.Level+t.c.headingOffset, 6)
			if id, ok := n.AttributeString("id"); ok && t.c.idPrefix != "" {
				if b, ok := id.([]byte); ok {
					n.SetAttributeString("id", append([]byte(t.c.idPrefix), b...))
				}
			}
		case *ast.Link, *ast.AutoLink:
			if t.c.linkTarget != "" && !inPage(n, b) {
				n.SetAttributeString("target", []byte(t.c.linkTarget))
				if t.c.linkTarget == "_blank" {
					n.SetAttributeString("rel", []byte("noopener noreferrer"))
				}
			}
		case *east.TaskCheckBox:
			if item := n.Parent(); item != nil {
				if li := item.Parent(); li != nil && li.Kind() == ast.KindListItem {
					li.SetAttributeString("class", []byte("task-list-item"))
					if list := li.Parent(); list != nil {
						list.SetAttributeString("class", []byte("contains-task-list"))
					}
				}
			}
		}
		return ast.WalkContinue, nil
	})
}

// inPage reports whether a link points into the same document (#fragment), which never gets a
// target.
func inPage(n ast.Node, source []byte) bool {
	switch n := n.(type) {
	case *ast.Link:
		return bytes.HasPrefix(n.Destination, []byte("#"))
	case *ast.AutoLink:
		return bytes.HasPrefix(n.URL(source), []byte("#"))
	}
	return false
}

// codeRenderer renders fenced and indented code blocks with the code package (highlighted on
// the server, no copy button: the button needs a <template>, which the sanitiser removes).
type codeRenderer struct{}

func (r codeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.render)
	reg.Register(ast.KindCodeBlock, r.render)
}

func (codeRenderer) render(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	var lang string
	if f, ok := n.(*ast.FencedCodeBlock); ok {
		lang = string(f.Language(source))
	}
	var src strings.Builder
	lines := n.Lines()
	for i := range lines.Len() {
		seg := lines.At(i)
		src.Write(seg.Value(source))
	}
	out, err := htemel.RenderString(code.Block(lang, src.String(), code.WithoutCopyButton()))
	if err != nil {
		return ast.WalkStop, err
	}
	_, _ = w.WriteString(out)
	_ = w.WriteByte('\n')
	return ast.WalkSkipChildren, nil
}

// taskRenderer draws a task-list checkbox with an accessible name: an unlabelled form control
// fails WCAG 1.3.1 / 4.1.2, and "Done", checked or not, reads naturally before the item text.
type taskRenderer struct{}

func (r taskRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(east.KindTaskCheckBox, r.render)
}

func (taskRenderer) render(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	if n.(*east.TaskCheckBox).IsChecked {
		_, _ = w.WriteString(`<input type="checkbox" checked="" disabled="" aria-label="Done"> `)
	} else {
		_, _ = w.WriteString(`<input type="checkbox" disabled="" aria-label="Done"> `)
	}
	return ast.WalkContinue, nil
}

// Plain returns the text of the Markdown with the markup removed, for search indexes, meta
// descriptions and excerpts: blocks are separated by a blank line, line breaks inside a
// paragraph become spaces, code keeps its text, images keep their alt text, and raw HTML is
// dropped. The result is plain text, not HTML: escape it when you output it (htemel does).
func Plain(src string, opts ...Option) string {
	c := collect(opts)
	c.noHeadingIDs = true
	md := newMarkdown(c)
	b := []byte(strings.ToValidUTF8(src, "�"))
	doc := md.Parser().Parse(text.NewReader(b))
	var blocks []string
	var cur strings.Builder
	flush := func() {
		if s := strings.Join(strings.Fields(cur.String()), " "); s != "" {
			blocks = append(blocks, s)
		}
		cur.Reset()
	}
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n := n.(type) {
		case *ast.Text:
			if entering {
				cur.Write(n.Value(b))
				if n.SoftLineBreak() || n.HardLineBreak() {
					cur.WriteByte(' ')
				}
			}
		case *ast.String:
			if entering {
				cur.Write(n.Value)
			}
		case *ast.CodeSpan:
			// Its children are Text nodes; nothing to add.
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			if entering {
				flush()
				lines := n.Lines()
				for i := range lines.Len() {
					seg := lines.At(i)
					cur.Write(seg.Value(b))
				}
				flush()
			}
			return ast.WalkSkipChildren, nil
		case *ast.RawHTML, *ast.HTMLBlock:
			return ast.WalkSkipChildren, nil
		case *east.TaskCheckBox:
		case *east.TableCell:
			if !entering {
				cur.WriteByte(' ')
			}
		default:
			if n.Type() == ast.TypeBlock && n.Kind() != ast.KindDocument {
				flush()
			}
		}
		return ast.WalkContinue, nil
	})
	flush()
	return strings.Join(blocks, "\n\n")
}

// Excerpt returns Plain(src) as one line cut to at most n characters (runes) at a word
// boundary, with "…" appended when it was cut: a teaser for a list of posts. n below 1 returns
// "".
func Excerpt(src string, n int, opts ...Option) string {
	if n < 1 {
		return ""
	}
	s := strings.Join(strings.Fields(Plain(src, opts...)), " ")
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	cut := n
	for cut > 0 && !unicode.IsSpace(r[cut]) {
		cut--
	}
	if cut == 0 { // one long word: cut inside it
		cut = n
	}
	return strings.TrimRightFunc(string(r[:cut]), func(r rune) bool { return unicode.IsSpace(r) || unicode.IsPunct(r) }) + "…"
}
