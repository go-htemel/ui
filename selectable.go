package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// selectableClass is the one selected-row treatment of the library: a primary tint with a 2px
// primary bar on the left, driven by aria-current="true" (aria-selected is not allowed on a
// button). Idle rows are transparent with a neutral hover; colors ease over 150ms.
const selectableClass = "block w-full cursor-pointer border-l-2 border-transparent px-4 py-3 text-left transition-colors duration-150 " +
	"hover:bg-neutral-50 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary-600 " +
	"aria-[current=true]:border-primary-600 aria-[current=true]:bg-primary-50 " +
	"dark:hover:bg-white/5 dark:focus-visible:outline-primary-500 dark:aria-[current=true]:border-primary-400 dark:aria-[current=true]:bg-primary-500/10"

// SelectableItem returns a full-width list row, <button type="button">, that shows the selected
// state of a list in which one row is chosen (notes, an inbox, a file list): while selectedExpr, a
// Datastar expression such as "$noteId === 3", is true the row has aria-current="true", a primary
// tint and a bar on its left edge. Text inside keeps its neutral colors (use neutral-900 / white
// titles and neutral-600 / 400 detail). Chain .Attrs(datastar.OnClick(...)) for the action, or
// use SplitListItem inside a SplitLayout.
//
//	ui.SelectableItem("$noteId === 3", html.Span().Text("Groceries")).Attrs(datastar.OnClick("$noteId = 3"))
func SelectableItem(selectedExpr string, children ...htemel.PhrasingContent) *html.ButtonElement {
	return html.Button(children...).
		Type(html.ButtonTypeButton).
		Class(Classes(selectableClass)).
		Attrs(datastar.Attr("aria-current", "("+selectedExpr+") ? 'true' : false"))
}
