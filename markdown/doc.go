// Package markdown renders Markdown to sanitised HTML for the ui component library, styled by
// ui.Prose: blog posts, CMS pages, notes and journal entries, and the server side of a
// ui.MarkdownEditor preview.
//
// It lives in its own package so that applications that do not import it do not link the
// Markdown parser (github.com/yuin/goldmark) or the sanitiser (github.com/microcosm-cc/bluemonday)
// into their binary; the root ui package depends on neither.
//
//	markdown.Render(post.Body)                           // *html.DivElement: <div class="ui-prose">...</div>
//	markdown.Render(note, markdown.WithSize(ui.Sm))      // ui.ProseSm
//	markdown.Excerpt(post.Body, 160)                     // plain-text teaser
//
// # Syntax
//
// CommonMark plus GitHub Flavored Markdown: tables (with column alignment), ~~strikethrough~~,
// task lists (- [x] done), bare URLs as links, and footnotes ([^1]). Headings get ids made from
// their text (see WithIDPrefix and WithoutHeadingIDs). Fenced and indented code blocks are
// highlighted on the server by the code package (github.com/go-htemel/ui/code), without its copy
// button: the button keeps the source in a <template>, which the sanitiser removes, and needs a
// Datastar signal.
//
// # Sanitising
//
// Output is safe to put on a page by default. Raw HTML in the Markdown is not rendered (goldmark
// omits it), dangerous link schemes are dropped, and the result then goes through a bluemonday
// policy (Policy): its UGCPolicy, which allows text, lists, tables, images and links with http,
// https, mailto and relative URLs, adds rel="nofollow" to links and allows no style attribute,
// script, iframe, object, form or event handler, extended with exactly the markup this package
// emits (code-block classes, task-list checkboxes, table alignment, footnote roles and link
// targets). WithPolicy substitutes your own policy; WithUnsafeHTML skips both steps for trusted
// authors.
//
// # Styling
//
// Render wraps its output in ui.Prose (the .ui-prose rules and --ui-prose-* variables of the ui
// design tokens, ui.ThemeCSS), and code blocks use the code package's ui-code rules from the
// same file, so an application that imports the tokens needs nothing more. The package adds no
// Tailwind utilities of its own.
package markdown
