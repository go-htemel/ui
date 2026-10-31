package ui

import (
	"strings"

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
// The handler then reads {"dragged": "card-42"} with datastar-go's ReadSignals.
//
// # Accessibility and touch
//
// HTML drag events come from a mouse (or pen) only: touch screens and keyboards cannot start a
// drag, and no Datastar expression can synthesize one. The fallback is "pick up, then move
// here", which runs the same action with the payload held in a signal instead of read from the
// drag data: put PickUp(signal, payload) on a button inside the draggable, and a MoveHere button
// in the DropZone. MoveHere appears only while something is picked up, so the keyboard and
// touch flow is: activate "Pick up" on a card, then activate "Move here" in the target zone.
//
//	ui.Draggable("card-42", card, ui.Button(ui.Ghost, "Pick up").Attrs(ui.PickUp("dragged", "card-42")))
//	ui.DropZone(ui.DropInto("dragged", act), ui.MoveHere("dragged", "Move here", act), column...)
//
// Touch and pen drags work too when the page loads the script from Assets and the board is
// wrapped in Sortable: Draggable and DropZone carry data-ui-drag, data-ui-payload and
// data-ui-dropzone for it, and a touch drop reaches the zone as a ui:drop event that runs the
// same action. Without the script the Pick up / Move here flow above is the touch fallback.
//
// During a mouse drag the draggable carries data-ui-dragging (and aria-grabbed="true", the
// deprecated but still read ARIA 1.0 state) and a DropZone under the pointer carries data-ui-over,
// which the built-in classes style. They are element attributes rather than Datastar signals
// because a signal needs a name unique to each zone and draggable, which these constructors
// cannot choose, and every signal would be sent to the server with each request.

// DragData is the expression that reads the dropped payload in a DropZone action.
const DragData = "evt.dataTransfer.getData('text/plain')"

// DropDetail is the expression that reads the dropped payload from the ui:drop event of a touch
// drag; DropZone substitutes it for DragData in its action.
const DropDetail = "evt.detail.payload"

// Draggable returns a draggable <div> carrying payload as its drag data.
func Draggable(payload string, children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).
		Draggable(html.DraggableTrue).
		Aria("grabbed", "false").
		Class(Classes("cursor-grab active:cursor-grabbing data-ui-dragging:opacity-50")).
		Attrs(
			htemel.Attr("data-ui-drag", ""),
			htemel.Attr("data-ui-payload", payload),
			datastar.On("dragstart",
				"evt.dataTransfer.setData('text/plain', "+datastar.Quote(payload)+"); evt.dataTransfer.effectAllowed = 'move'; "+
					"el.toggleAttribute('data-ui-dragging', true); el.setAttribute('aria-grabbed', 'true')"),
			datastar.On("dragend", "el.toggleAttribute('data-ui-dragging', false); el.setAttribute('aria-grabbed', 'false')"),
		)
}

// DropZone returns a <div> that accepts drops and runs action, a Datastar expression, on each
// drop. It cancels dragover (which is what makes an element a drop target) and the browser's
// default drop handling (which would otherwise navigate to dropped text in some browsers). While
// a drag is over the zone it has the data-ui-over attribute, styled with an outline and tint; dragenter
// and dragleave toggle it, ignoring the enter/leave pairs fired when the pointer crosses a child.
//
// The zone also listens for the ui:drop event that <ui-sortable> dispatches for touch drags (see
// Sortable): the same action runs, with the payload read from evt.detail.payload instead of the
// drag data, so DropInto works for both without change.
func DropZone(action string, children ...htemel.FlowContent) *html.DivElement {
	return html.Div(children...).
		Class(Classes("data-ui-over:outline-2 data-ui-over:outline-dashed data-ui-over:outline-primary-600 data-ui-over:bg-primary-500/10 dark:data-ui-over:outline-primary-500")).
		Attrs(
			htemel.Attr("data-ui-dropzone", ""),
			datastar.On("dragenter", "evt.preventDefault(); el.toggleAttribute('data-ui-over', true)"),
			datastar.On("dragover", "evt.preventDefault(); evt.dataTransfer.dropEffect = 'move'"),
			datastar.On("dragleave", "if (!el.contains(evt.relatedTarget)) el.toggleAttribute('data-ui-over', false)"),
			datastar.On("drop", "el.toggleAttribute('data-ui-over', false); "+action, datastar.Prevent()),
			datastar.On("ui:drop", "el.toggleAttribute('data-ui-over', false); "+strings.ReplaceAll(action, DragData, DropDetail)),
		)
}

// ViaSuffix names the companion signal of a pick-up: PickUp("dragged", ...) also sets
// $draggedVia to 'pickup', and MoveHere is shown only while it is. Mouse and touch drops (DropInto)
// set the payload signal alone, so MoveHere buttons never flash during a drop.
const ViaSuffix = "Via"

// PickUp is the click handler of a "pick up" button inside a Draggable: it stores payload in
// signal and marks the pick-up in signal+ViaSuffix ($draggedVia = 'pickup'), the touch and
// keyboard counterpart of starting a drag. Declare signal on an ancestor, as for DropInto;
// MoveHere declares the Via signal.
func PickUp(signal, payload string) htemel.Attribute {
	return datastar.OnClick("$" + signal + " = " + datastar.Quote(payload) + "; $" + signal + ViaSuffix + " = 'pickup'")
}

// MoveHere returns the button of a DropZone that drops the picked-up item there: it runs action,
// the same expression the zone's drop runs but without DropInto, since PickUp already put the
// payload in signal, and then clears the signal and its Via companion. It is shown only while
// something was picked up with PickUp ($<signal>Via === 'pickup'), not while a mouse or touch
// drop briefly holds the payload in signal. It declares the Via signal (empty unless the page
// already has it).
//
//	ui.DropZone(ui.DropInto("dragged", act), ui.MoveHere("dragged", "Move here", act))
//
// Pass the plain action (datastar.Post(...)), not the DropInto expression.
func MoveHere(signal, label, action string) *html.ButtonElement {
	via := signal + ViaSuffix
	return html.Button().
		Type(html.ButtonTypeButton).
		Class(Classes("w-full cursor-pointer rounded-control border border-dashed border-primary-600 px-3 py-2 text-sm font-medium text-primary-600",
			"transition-colors duration-150 hover:bg-primary-500/10 hover:text-primary-700 focus-visible:outline-2 focus-visible:outline-primary-600 dark:focus-visible:outline-primary-500 dark:text-primary-400 dark:hover:text-primary-400")).
		Style(hiddenUntilDatastar).
		Attrs(
			datastar.Signal(via, datastar.Quote(""), datastar.IfMissing()),
			datastar.Show("$"+via+" === 'pickup'"),
			datastar.OnClick(action+"; setTimeout(() => { $"+signal+" = ''; $"+via+" = '' })"),
		).
		Text(label)
}

// DropInto returns a drop action that stores the payload in signal, runs action and then clears
// the signal again: DropInto("dragged", datastar.Post("/move")) is
// "$dragged = <DragData>; @post('/move'); setTimeout(() => { $dragged = ” })". The request has
// already read the signal when the timeout fires. It never touches the Via companion of
// PickUp, so MoveHere stays hidden during a drop. Declare the signal on an ancestor, for example
// with datastar.Signal("dragged", datastar.Quote("")), so that it exists before the first drop.
func DropInto(signal, action string) string {
	return "$" + signal + " = " + DragData + "; " + action + "; setTimeout(() => { $" + signal + " = '' })"
}
