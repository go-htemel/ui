package code

import (
	"crypto/sha1" //nolint:gosec // an identifier, not a security boundary
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

// copiedMs is how long the copy button shows its check mark.
const copiedMs = 1500

// BlockElement is the <figure> that Block returns. Its element methods (Attrs, SetAttr, AddClass,
// Id, ...) are the figure's, and it is flow content like any other component. Use AddClass rather
// than Class to add classes: Class replaces the ui-code classes.
//
// To cap the height and scroll vertically, add the Tailwind arbitrary property
// "[--ui-code-max-height:24rem]" with AddClass.
type BlockElement struct {
	*html.FigureElement
}

// Block returns a highlighted code block:
//
//	<figure class="ui-code ..."><div class="ui-code-toolbar">...</div><pre><code>...</code></pre><template>...</template></figure>
//
// lang is a chroma lexer name or alias; an unknown one renders plain text. src is shown exactly
// (line endings normalised to \n, trailing newlines dropped). The block scrolls horizontally
// when a line is too long; the <pre> is focusable so a keyboard can scroll it.
//
// Options: WithLineNumbers, WithStart, WithHighlight, WithHighlightRange, WithTitle, WithWrap,
// WithoutCopyButton, WithID. The copy button is on by default and needs Datastar.
func Block(lang, src string, opts ...Option) *BlockElement {
	c := collect(opts)
	src = normalize(src)
	lines := highlight(lang, src)
	langClass := token(lang)

	codeEl := html.Code()
	if langClass != "" {
		codeEl.Class("language-" + langClass)
	}
	for i, line := range lines {
		if i > 0 {
			codeEl.Children(htemel.Text("\n"))
		}
		span := html.Span(segments(line)...).Class("line")
		if c.highlighted(c.start + i) {
			span.Data("hl", "")
		}
		codeEl.Children(span)
	}

	fig := html.Figure().Class(ui.Classes(
		"ui-code",
		cond(c.lineNumbers, "ui-code-lines"),
		cond(c.wrap, "ui-code-wrap"),
		cond(c.title != "", "ui-code-titled"),
	))
	fig.Aria("label", label(c.title, lang))
	if c.lineNumbers {
		digits := max(len(strconv.Itoa(c.start+len(lines)-1)), 2)
		style := fmt.Sprintf("--ui-code-digits:%d", digits)
		if c.start != 1 {
			style += fmt.Sprintf(";--ui-code-start:%d", c.start)
		}
		fig.Style(style)
	}

	if c.title != "" || !c.noCopy {
		fig.Children(toolbar(c, lang, src))
	}
	fig.Children(html.Pre(codeEl).Tabindex(0))
	if !c.noCopy {
		// The raw source for the copy button: a template's content is inert and never rendered.
		fig.Children(html.Template(htemel.Text(src)))
		fig.Attrs(datastar.Signal(signalName(c, lang, src), "false", datastar.IfMissing()))
	}
	return &BlockElement{FigureElement: fig}
}

// Inline returns src highlighted as a <code class="ui-code-inline"> for use in running text. The
// language and the text are handled as in Block.
func Inline(lang, src string) *html.CodeElement {
	src = normalize(src)
	el := html.Code().Class("ui-code-inline")
	for i, line := range highlight(lang, src) {
		if i > 0 {
			el.Children(htemel.Text("\n"))
		}
		el.Children(segments(line)...)
	}
	return el
}

// toolbar is the title and the copy button.
func toolbar(c config, lang, src string) *html.DivElement {
	bar := html.Div().Class("ui-code-toolbar").Role("group").Aria("label", "Code actions")
	if c.title != "" {
		bar.Children(html.Span().Class("ui-code-title").Text(c.title))
	}
	if c.noCopy {
		return bar
	}
	sig := signalName(c, lang, src)
	// Two stacked icons that cross-fade in 150ms (no motion with prefers-reduced-motion).
	const iconClass = "col-start-1 row-start-1 size-4 shrink-0 transition-opacity duration-150 motion-reduce:transition-none"
	icons := html.Span(
		ui.IconCopy().Class(iconClass).Attrs(datastar.Class("opacity-0", "$"+sig)),
		ui.IconCheck().Class(ui.Classes(iconClass, "opacity-0")).Attrs(datastar.Class("opacity-0", "!$"+sig)),
	).Class("grid")
	button := ui.IconButton(ui.Ghost, icons, "Copy code", ui.WithSize(ui.Xs)).Attrs(datastar.OnClick(copyExpr(sig)))
	status := html.Span().Class("sr-only").Aria("live", "polite").Attrs(datastar.Text("$" + sig + " ? 'Copied' : ''"))
	return bar.Children(html.Div(button, status).Class("ui-code-actions"))
}

// copyExpr writes the template's text to the clipboard, then raises sig for copiedMs. The
// signal only rises when the write succeeded; el is the button.
func copyExpr(sig string) string {
	return fmt.Sprintf("navigator.clipboard.writeText(el.closest('figure').querySelector('template').content.textContent)"+
		".then(() => { $%[1]s = true; clearTimeout(el._copied); el._copied = setTimeout(() => { $%[1]s = false }, %[2]d) })", sig, copiedMs)
}

// segments renders a line's runs: plain text as text, the rest as <span class="tok-KIND">.
func segments(line []seg) []htemel.PhrasingContent {
	out := make([]htemel.PhrasingContent, 0, len(line))
	for _, s := range line {
		if s.kind == "" {
			out = append(out, htemel.Text(s.text))
		} else {
			out = append(out, html.Span(htemel.Text(s.text)).Class("tok-"+s.kind))
		}
	}
	return out
}

// normalize converts line endings to \n and drops trailing newlines.
func normalize(src string) string {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src = strings.ReplaceAll(src, "\r", "\n")
	return strings.TrimRight(src, "\n")
}

// signalName is the Datastar signal of the copy button: local to the browser (the underscore
// keeps it out of requests), named after WithID or a hash of the language and source.
func signalName(c config, lang, src string) string {
	id := c.id
	if id == "" {
		sum := sha1.Sum([]byte(strings.ToLower(strings.TrimSpace(lang)) + "\x00" + src)) //nolint:gosec // see import
		id = hex.EncodeToString(sum[:])[:8]
	}
	return "_copied_" + slug(id, '_')
}

// slug lower-cases s, keeping letters and digits and turning everything else into sep.
func slug(s string, sep byte) string {
	b := []byte(strings.ToLower(s))
	for i, ch := range b {
		if (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') {
			b[i] = sep
		}
	}
	return string(b)
}

// token reduces a language name to a class-safe token: lower-case letters and digits, with
// every other run of characters turned into a hyphen.
func token(lang string) string {
	return strings.Trim(slug(strings.TrimSpace(lang), '-'), "-")
}

func label(title, lang string) string {
	switch {
	case title != "":
		return title
	case token(lang) != "":
		return token(lang) + " code"
	}
	return "Code"
}

func cond(ok bool, class string) string {
	if ok {
		return class
	}
	return ""
}
