package main

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
	"github.com/go-htemel/ui/markdown"
)

func markdownEditorTabs() htemel.FlowContent {
	// Write / Preview tabs: choosing Preview posts the signals, and the handler patches the
	// preview with markdown.Render (see handlers.go, markdownPreview).
	return ui.Field("journal", "Journal entry",
		ui.MarkdownEditor("journal",
			ui.WithPreview(datastar.Post("/demo/markdown/preview?editor=journal")),
			ui.WithPreviewLabel("Journal preview"),
			ui.WithValue(sampleEntry),
			ui.WithRows(12),
			ui.WithMaxLength(2000),
		),
		ui.HelpText("Only you can see your journal."),
	)
}

func markdownEditorSplit() htemel.FlowContent {
	// Side by side once the editor is 36rem wide; the preview follows the text 400ms after the
	// last keystroke.
	return ui.Field("post", "Post body",
		ui.MarkdownEditor("post",
			ui.WithPreview(datastar.Post("/demo/markdown/preview?editor=post")),
			ui.WithPreviewLabel("Post preview"),
			ui.WithSplit(),
			ui.WithValue(samplePost),
			ui.WithRows(14),
			ui.WithMonospace(),
		),
	)
}

func markdownEditorCompact() htemel.FlowContent {
	// A note field: three tools, grows with its text, and a pasted or dropped image becomes
	// Markdown (a real app would upload evt.detail.file first and insert the URL it gets back).
	return ui.Field("note", "Quick note",
		ui.MarkdownEditor("note",
			ui.WithToolbar(ui.MarkdownBold, ui.MarkdownItalic, ui.MarkdownTaskList),
			ui.WithAutosize(),
			ui.WithPlaceholder("Jot something down… try “- [ ] ” and Enter"),
			ui.WithUploadAction("evt.detail.insert('![' + evt.detail.file.name + '](/uploads/' + encodeURIComponent(evt.detail.file.name) + ')')"),
			ui.WithHint("Bold, italic and task lists", ""),
		),
	)
}

func proseArticle() htemel.FlowContent {
	// Markdown from a CMS, rendered and sanitised on the server. The headings start at <h3>
	// under this page's <h2>, and the ids are prefixed so they cannot collide with the page's.
	return markdown.Render(sampleArticle,
		markdown.WithHeadingOffset(1),
		markdown.WithIDPrefix("article-"),
	).AddClass("max-w-prose")
}

func proseSizes() htemel.FlowContent {
	// ProseSm for comments and side panels, ProseLg for reading views; every margin is in em.
	para := func() *html.PElement {
		return html.P(htemel.Text("Prose sets the type of content that has no classes: "), html.Strong().Text("bold"),
			htemel.Text(", "), html.A().Href("#").Text("links"), htemel.Text(" and "), html.Code().Text("code"), htemel.Text("."))
	}
	return html.Div(
		ui.ProseSm(html.H3().Text("ProseSm"), para(), html.Ul(html.Li().Text("Compact"), html.Li().Text("For side panels"))),
		ui.ProseLg(html.H3().Text("ProseLg"), para(), html.Ul(html.Li().Text("Roomy"), html.Li().Text("For reading"))),
	).Class("grid gap-8 md:grid-cols-2")
}

func proseHandwritten() htemel.FlowContent {
	// Prose works on any HTML, not only rendered Markdown.
	return ui.Prose(
		html.P(htemel.Text("Press "), html.Kbd().Text("Ctrl"), htemel.Text(" + "), html.Kbd().Text("K"),
			htemel.Text(" to insert a link. "), html.Mark().Text("Highlighted text"), htemel.Text(" stands out.")),
		html.Dl(
			html.Dt().Text("Draft"), html.Dd().Text("Only you can see it."),
			html.Dt().Text("Published"), html.Dd().Text("Everyone with the link can read it."),
		),
		html.Figure(
			html.Img().Src("/placeholder.svg").Alt("A landscape placeholder").Width(640).Height(320),
			html.Figcaption().Text("Figures get a muted caption."),
		),
	).AddClass("max-w-prose")
}

const sampleEntry = `## Saturday, 31 October

Woke before the alarm. The **first frost** of the year; the garden looked _silvered_.

- [x] Morning run, 5 km
- [x] Call Mum
- [ ] Repot the fern

> Small steps, every day.

Wrote a tiny helper for the blog:

` + "```go" + `
func slug(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}
` + "```" + `

Read more at [the Go blog](https://go.dev/blog/).
`

const samplePost = `# Shipping a Markdown editor

We picked a **textarea** over contenteditable: it is fast, accessible, and the
browser's undo just works.

| Option           | Undo | Accessible |
|:-----------------|:----:|-----------:|
| textarea         | yes  |        yes |
| contenteditable  | some |    depends |

1. Write in the left pane
2. Watch the preview on the right
3. Publish with ` + "`ui.Button`" + `
`

const sampleArticle = `## Why we write in Markdown

Our CMS stores every page as **Markdown** and renders it on the server: the text stays
portable, diffs stay readable, and nothing but sanitised HTML reaches the browser.[^safe]

### What you can use

- Paragraphs with _emphasis_, **strong text**, ~~strikethrough~~ and ` + "`inline code`" + `
- Links such as [the CommonMark help](https://commonmark.org/help/) or bare URLs: https://go.dev
- Nested lists
  - like this one
    1. numbered inside
    2. and back out

### A checklist

- [x] Draft the post
- [x] Ask for a review
- [ ] Publish

### A table

| Element    | Light          | Dark           |
|:-----------|:---------------|:---------------|
| Body text  | neutral-700    | neutral-300    |
| Links      | primary-600    | primary-400    |
| Code chips | neutral-100 bg | white/10 bg    |

> Good typography is invisible. Bad typography is everywhere.
>
> — a designer, probably

### Some Go

` + "```go" + `
// Render turns a post into HTML that is safe to show.
func Render(p Post) htemel.Node {
	return markdown.Render(p.Body, markdown.WithHeadingOffset(1))
}
` + "```" + `

![A landscape placeholder](/placeholder.svg "Images get the surface radius")

---

That's it. Press Ctrl+K in the editor to add a link.

[^safe]: The default policy allows no scripts, styles, iframes or event handlers.
`
