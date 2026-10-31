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
		Class("grid gap-4 md:grid-cols-3").
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
				html.H3().Class("text-sm font-semibold text-zinc-900 dark:text-white").Text(col.title),
				ui.Badge(ui.Neutral, strconv.Itoa(len(in))),
			).Class("flex items-center justify-between px-1 pb-1"),
			htemel.Each(in, func(c kanbanCard) htemel.Node { return kanbanCardView(c) }),
		).Id("column-" + col.key).
			Class(ui.Classes("flex min-h-48 flex-col gap-2 rounded-md bg-zinc-100 p-3 dark:bg-white/5")))
	}
	return root
}

func kanbanCardView(c kanbanCard) *html.DivElement {
	return ui.Draggable(c.ID,
		ui.IconGripVertical().Class("size-4 shrink-0 text-zinc-400"),
		html.Span().Class("flex-1").Text(c.Title),
		ui.Button(ui.Ghost, "Pick up", ui.WithSize(ui.Xs)).Attrs(ui.PickUp("dragged", c.ID)),
	).Id("card-" + c.ID).
		AddClass(ui.Classes("flex items-center gap-2 rounded-md bg-white p-3 text-sm text-zinc-800 shadow-xs ring-1 ring-zinc-950/5",
			"dark:bg-zinc-800 dark:text-zinc-100 dark:ring-white/10"))
}
