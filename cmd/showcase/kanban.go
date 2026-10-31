package main

import (
	"slices"
	"strconv"
	"sync"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

type column struct{ key, title string }

var columns = []column{{"todo", "To do"}, {"doing", "In progress"}, {"done", "Done"}}

type kanbanCard struct {
	ID, Title, Column string
}

// board is the in-memory kanban state shared by every visitor of the showcase.
type board struct {
	mu    sync.Mutex
	cards []kanbanCard
}

var demoBoard = newBoard()

func newBoard() *board {
	b := &board{}
	b.reset()
	return b
}

func (b *board) reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cards = []kanbanCard{
		{"c1", "Write the README", "todo"},
		{"c2", "Golden tests", "todo"},
		{"c3", "Theme tokens", "doing"},
		{"c4", "Kanban demo", "doing"},
		{"c5", "Module scaffold", "done"},
	}
}

// move puts card id at the end of column col. It reports false for an unknown card or column.
func (b *board) move(id, col string) (kanbanCard, bool) {
	if !slices.ContainsFunc(columns, func(c column) bool { return c.key == col }) {
		return kanbanCard{}, false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	i := slices.IndexFunc(b.cards, func(c kanbanCard) bool { return c.ID == id })
	if i < 0 {
		return kanbanCard{}, false
	}
	c := b.cards[i]
	c.Column = col
	b.cards = append(slices.Delete(b.cards, i, i+1), c)
	return c, true
}

func (b *board) snapshot() []kanbanCard {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.cards)
}

// view renders the board. Its id lets the server patch it in place after a move.
func (b *board) view() *html.DivElement {
	cards := b.snapshot()
	root := html.Div().Id("kanban").
		Class("grid grid-cols-[repeat(auto-fit,minmax(14rem,1fr))] gap-4").
		Attrs(datastar.Signal("dragged", datastar.Quote(""), datastar.IfMissing()))
	for _, col := range columns {
		var in []kanbanCard
		for _, c := range cards {
			if c.Column == col.key {
				in = append(in, c)
			}
		}
		root.Children(ui.DropZone(
			ui.DropInto("dragged", datastar.Post("/demo/kanban/move/"+col.key)),
			ui.MoveHere("dragged", "Move here", datastar.Post("/demo/kanban/move/"+col.key)),
			html.Div(
				html.H3().Class("text-sm font-semibold text-neutral-900 dark:text-white").Text(col.title),
				ui.Badge(ui.Neutral, strconv.Itoa(len(in))),
			).Class("flex items-center justify-between px-1 pb-1"),
			htemel.Each(in, func(c kanbanCard) htemel.Node { return kanbanCardView(c) }),
		).Id("column-" + col.key).
			Class(ui.Classes("flex min-h-48 flex-col gap-2 rounded-surface bg-neutral-100 p-3 dark:bg-white/5")))
	}
	return root
}

func kanbanCardView(c kanbanCard) *html.DivElement {
	// The grip doubles as the keyboard/touch "pick up" control, so the card stays a single short
	// row: an icon button, then the title, which may wrap.
	return ui.Draggable(c.ID,
		ui.IconButton(ui.Ghost, ui.IconGripVertical().Class("size-4"), "Pick up "+c.Title, ui.WithSize(ui.Xs)).
			AddClass("-ml-1 shrink-0 text-neutral-400").
			Attrs(ui.PickUp("dragged", c.ID)),
		html.Span().Class("min-w-0 flex-1 break-words").Text(c.Title),
	).Id("card-" + c.ID).
		AddClass(ui.Classes("flex items-center gap-2 rounded-surface bg-white p-2 pr-3 text-sm text-neutral-800 shadow-xs ring-1 ring-neutral-950/5",
			"dark:bg-neutral-800 dark:text-neutral-100 dark:ring-white/10"))
}
