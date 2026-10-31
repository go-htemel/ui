// Package code renders syntax-highlighted code on the server, as htemel elements for the ui
// component library: no highlighter script is sent to the browser.
//
// It lives in its own package so that applications that do not import it do not link the
// highlighter (github.com/alecthomas/chroma/v2) into their binary; the root ui package does not
// depend on it.
//
//	code.Block("go", src, code.WithLineNumbers(), code.WithTitle("main.go"))
//	html.P(htemel.Text("Call "), code.Inline("go", "ui.Button(ui.Primary, \"Save\")"), htemel.Text("."))
//
// Block returns a <figure class="ui-code"> holding an optional toolbar (title and copy button)
// and a <pre><code> with one <span class="line"> per line. Inline returns a highlighted <code>.
//
// # Highlighting
//
// The language is a chroma lexer name or alias ("go", "css", "html", "js", "json", "bash",
// "sql", "yaml", "markdown", ...). An unknown or empty language, a lexer that fails, and a lexer
// whose output does not reproduce the source all fall back to plain text: rendering never fails
// and never changes the text. The tokens become <span class="tok-KIND"> where KIND is one of a
// small set: keyword, builtin, name, function, type, string, number, literal, attribute, tag,
// comment, operator, punctuation and error. Text that needs no colour (plain identifiers,
// whitespace) is emitted without a span. All text is escaped by htemel.
//
// # Styling
//
// The classes (ui-code, ui-code-lines, ui-code-wrap, line, tok-*) and the colours
// (--color-code-*, light on :root and dark under .dark) are defined in the ui design tokens
// (ui.ThemeCSS, assets/ui.css), so an application that already imports the tokens needs nothing
// more; override the variables to retheme the highlighting. The Tailwind utilities the package
// uses are in the ui safelist (cmd/safelist reads this package too).
//
// # Line numbers, highlighting and copying
//
// Line numbers are CSS counters (generated content), so selecting and copying the code never
// includes them and screen readers do not announce them. WithHighlight marks lines with a
// data-ui-hl attribute, styled with a tint and a bar. The copy button needs Datastar: it keeps the
// source in a <template> inside the figure, writes it to the clipboard on click (a secure
// context is required: https or localhost) and shows a check mark for 1.5 seconds, announced
// through a polite live region. The signal that drives it is local to the browser (its name
// starts with an underscore, so Datastar does not send it to the server) and is named after
// a hash of the language and source, so output is deterministic; see WithID for pages that show
// the same source twice.
package code
