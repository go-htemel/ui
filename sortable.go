package ui

import (
	"io"

	"github.com/go-htemel/htemel"
)

// SortableElement is the <ui-sortable> custom element holding typed flow children. It is flow,
// phrasing and palpable content, like any autonomous custom element.
type SortableElement struct {
	htemel.PhrasingMarker
	htemel.PalpableMarker
	children []htemel.FlowContent
}

// Render implements htemel.Node.
func (s SortableElement) Render(w io.Writer) error {
	if _, err := io.WriteString(w, `<ui-sortable style="display: contents">`); err != nil {
		return err
	}
	for _, c := range s.children {
		if err := c.Render(w); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "</ui-sortable>")
	return err
}

// Sortable wraps Draggable and DropZone components in a <ui-sortable> custom element, defined by
// the script that Assets serves, which adds touch and pen dragging: a long press (about 150 ms)
// on a Draggable lifts a copy that follows the finger, and releasing it over a DropZone runs the
// same drop action a mouse drop runs. Mouse drags stay with the browser's native drag and drop.
//
//	ui.Sortable(ui.DropZone(act, ui.Draggable("card-1", card)), ui.DropZone(act))
//
// It is progressive enhancement. Without the script the element is an unknown inline element with
// no behaviour, native drag and drop still works with a mouse, and the PickUp and MoveHere buttons
// still work for touch and keyboard users. The element generates no box of its own
// (display: contents), so it does not affect the layout of the grid or flex container it wraps.
func Sortable(children ...htemel.FlowContent) SortableElement {
	return SortableElement{children: children}
}
