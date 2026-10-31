package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// Drag and drop with the HTML drag-and-drop API and Datastar, no custom JavaScript: a Draggable
// puts its payload (usually a record id) in the drag data when the drag starts, and a DropZone
// runs an action expression when something is dropped on it. DragData reads the payload inside
// that expression; DropInto stores it in a signal so a backend action can read it:
//
//	ui.Draggable("card-42", card)
//	ui.DropZone(ui.DropInto("dragged", datastar.Post("/board/done")), column...)
//
// The handler then reads {"dragged": "card-42"} with datastar-go's ReadSignals. Touch screens do
// not fire HTML drag events; offer a button or menu as well where touch matters.

// DragData is the expression that reads the dropped payload in a DropZone action.
const DragData = "evt.dataTransfer.getData('text/plain')"

// Draggable returns a draggable <div> carrying payload as its drag data.
func Draggable(payload string, children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).
		Draggable(html.DraggableTrue).
		Class(Classes("cursor-grab active:cursor-grabbing")).
		Attrs(datastar.On("dragstart",
			"evt.dataTransfer.setData('text/plain', "+datastar.Quote(payload)+"); evt.dataTransfer.effectAllowed = 'move'"))
}

// DropZone returns a <div> that accepts drops and runs action, a Datastar expression, on each
// drop. It cancels dragover (which is what makes an element a drop target) and the browser's
// default drop handling (which would otherwise navigate to dropped text in some browsers).
func DropZone(action string, children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).
		Attrs(
			datastar.On("dragover", "evt.preventDefault()"),
			datastar.On("drop", action, datastar.Prevent()),
		)
}

// DropInto returns a drop action that stores the payload in signal and then runs action:
// DropInto("dragged", datastar.Post("/move")) is "$dragged = <DragData>; @post('/move')".
// Declare the signal on an ancestor, for example with datastar.Signal("dragged", datastar.Quote("")), so
// that it exists before the first drop.
func DropInto(signal, action string) string {
	return "$" + signal + " = " + DragData + "; " + action
}
