package ui

import (
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
	"github.com/go-htemel/htemel/svg"
)

// FormControl is a control Field can label: Input, Textarea, Select, or any other element that is
// flow content.
type FormControl interface {
	htemel.FlowContent
	htemel.Element
}

// CompositeControl is a FormControl that wraps the element to label, such as InputGroup. Field
// renders the composite and sets the id, aria-describedby and aria-invalid on Control().
type CompositeControl interface {
	FormControl
	Control() FormControl
}

// controlBase is the look shared by Input, Textarea and Select. ui-control is not a utility: it
// marks the control for the rules in ui.css that restyle what utilities cannot reach (autofill
// colors, number spin buttons, the search clear button, the date picker icon, the color swatch).
const controlBase = "ui-control block w-full rounded-control bg-white text-base text-neutral-900 outline-1 -outline-offset-1 outline-neutral-500 " +
	"placeholder:text-neutral-500 focus:outline-2 focus:-outline-offset-2 focus:outline-primary-600 sm:text-sm/6 " +
	"disabled:cursor-not-allowed disabled:bg-neutral-50 disabled:text-neutral-500 " +
	"aria-invalid:outline-danger-500 aria-invalid:focus:outline-danger-600 " +
	"dark:bg-white/5 dark:text-white dark:outline-white/35 dark:placeholder:text-neutral-400 dark:focus:outline-primary-500 " +
	"dark:disabled:bg-white/10 dark:disabled:text-neutral-400 dark:aria-invalid:outline-danger-400"

// fileButton draws the "Choose file" button of a file input like a small Secondary button.
const fileButton = "file:mr-3 file:cursor-pointer file:rounded-control file:border-0 file:bg-white file:px-2 file:py-0.5 file:text-sm/5 file:font-semibold " +
	"file:text-neutral-900 file:ring-1 file:ring-neutral-300 file:ring-inset file:transition-colors file:duration-150 hover:file:bg-neutral-50 " +
	"dark:file:bg-white/10 dark:file:text-white dark:file:ring-white/10 dark:hover:file:bg-white/20"

// rangeInput is a slider: the native track and thumb in the primary color, no field box.
const rangeInput = "block w-full cursor-pointer accent-primary-600 focus-visible:outline-2 focus-visible:outline-offset-2 " +
	"focus-visible:outline-primary-600 disabled:cursor-not-allowed dark:accent-primary-500 dark:focus-visible:outline-primary-500"

// Input returns an <input> of type t with the given name. Bind it to a signal with
// .Attrs(datastar.Bind("email")) and put it in a Field for a label.
//
// Text-like types (text, email, password, search, number, date, time, ...) are drawn as a field.
// Native parts are restyled through ui.css: number inputs have no spin buttons (the arrow keys
// still step), the search clear button is a cross in the text color, the date and time picker
// icon follows the color scheme, and autofilled fields keep the field colors instead of the
// browser's yellow or blue. A file input's button looks like a small Secondary button, a color
// input is a swatch the height of a field, and a range input is a bare slider in the primary
// color. For checkboxes, radios and switches use Checkbox, Radio and Switch.
func Input(t html.InputType, name string) *html.InputElement {
	in := html.Input().Type(t).Name(name)
	switch t {
	case html.InputTypeColor:
		return in.Class(Classes(controlBase, "h-9 cursor-pointer p-1"))
	case html.InputTypeRange:
		return in.Class(Classes(rangeInput))
	case html.InputTypeFile:
		return in.Class(Classes(controlBase, "cursor-pointer px-1.5 py-1.5", fileButton))
	}
	return in.Class(Classes(controlBase, "px-3 py-1.5"))
}

// Textarea returns a multi-line text control with the given name and visible rows.
func Textarea(name string, rows int) *html.TextareaElement {
	return html.Textarea().Name(name).Rows(rows).Class(Classes(controlBase, "px-3 py-1.5"))
}

// Choice returns an <option> of a Select. Chain .Selected(true) to preselect it.
func Choice(value, label string) *html.OptionElement {
	return html.Option().Value(value).Text(label)
}

// FieldOption configures a Field.
type FieldOption func(*fieldConfig)

type fieldConfig struct {
	help        string
	err         string
	errorSignal string
}

// HelpText adds a hint under the control, linked with aria-describedby.
func HelpText(text string) FieldOption { return func(c *fieldConfig) { c.help = text } }

// ErrorText adds a server-rendered validation message under the control and marks the control
// aria-invalid. An empty text adds nothing, so a handler can pass its error string as is.
func ErrorText(text string) FieldOption { return func(c *fieldConfig) { c.err = text } }

// ErrorSignal shows the validation message held in a string signal (declared empty unless the
// page already has it): while the signal is non-empty, its text is shown under the control and the control is aria-invalid. A backend
// sets it by patching signals, for example {"errors": {"email": "Already taken"}} with
// ErrorSignal("errors.email").
func ErrorSignal(signal string) FieldOption { return func(c *fieldConfig) { c.errorSignal = signal } }

// Field lays out a label, a control and optional help and error text. It sets the control's id
// (so the label's for attribute points at it) and its aria-describedby. For a CompositeControl
// (InputGroup) those go on the control inside it.
//
//	ui.Field("email", "Email", ui.Input(html.InputTypeEmail, "email").Attrs(datastar.Bind("email")),
//		ui.HelpText("We never share it."), ui.ErrorSignal("errors.email"))
func Field(id, label string, control FormControl, opts ...FieldOption) *html.DivElement {
	var c fieldConfig
	for _, opt := range opts {
		if opt != nil {
			opt(&c)
		}
	}
	target := FormControl(control)
	if c, ok := control.(CompositeControl); ok && c.Control() != nil {
		target = c.Control()
	}
	target.SetAttr("id", id)
	f := html.Div(
		html.Label().For(id).Class(Classes("block text-sm/6 font-medium text-neutral-900 dark:text-white")).Text(label),
		html.Div(control).Class("mt-2"),
	)
	var described []string
	if c.help != "" {
		described = append(described, id+"-help")
		f.Children(html.P().Id(id + "-help").Class(Classes("mt-2 text-sm text-neutral-500 dark:text-neutral-400")).Text(c.help))
	}
	errClass := Classes("mt-2 text-sm text-danger-600 dark:text-danger-400")
	if c.err != "" {
		described = append(described, id+"-error")
		target.SetAttr("aria-invalid", "true")
		f.Children(html.P().Id(id + "-error").Class(errClass).Text(c.err))
	}
	if c.errorSignal != "" {
		described = append(described, id+"-error-live")
		htemel.Apply(target, datastar.Attr("aria-invalid", "$"+c.errorSignal+" ? 'true' : false"))
		f.Attrs(initially(c.errorSignal, "''"))
		f.Children(html.P().Id(id+"-error-live").Class(errClass).Style(hiddenUntilDatastar).
			Attrs(datastar.Show("!!$"+c.errorSignal), datastar.Text("$"+c.errorSignal)))
	}
	if len(described) > 0 {
		target.SetAttr("aria-describedby", strings.Join(described, " "))
	}
	return f
}

const choiceLabel = "inline-flex cursor-pointer items-center gap-x-3 text-sm/6 text-neutral-900 has-disabled:cursor-not-allowed " +
	"has-disabled:text-neutral-500 dark:text-white dark:has-disabled:text-neutral-400"

// choiceBox is the drawn box of Checkbox and Radio: the native input with appearance-none, so it
// keeps its semantics, focus and keyboard behaviour but none of the browser's rendering (no
// accent-color halo on hover). The mark is a sibling drawn on top (see choiceMark).
const choiceBox = "peer col-start-1 row-start-1 size-4 cursor-pointer appearance-none border border-neutral-500 bg-white " +
	"transition-colors duration-150 not-checked:enabled:hover:border-neutral-700 " +
	"checked:border-primary-600 checked:bg-primary-600 " +
	"focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600 " +
	"disabled:cursor-not-allowed disabled:border-neutral-300 disabled:bg-neutral-100 disabled:checked:border-neutral-300 disabled:checked:bg-neutral-300 " +
	"dark:border-white/35 dark:bg-white/5 dark:not-checked:enabled:hover:border-white/60 " +
	"dark:checked:border-primary-500 dark:checked:bg-primary-500 dark:focus-visible:outline-primary-500 " +
	"dark:disabled:border-white/10 dark:disabled:bg-white/10 dark:disabled:checked:border-white/10 dark:disabled:checked:bg-white/20 " +
	"forced-colors:appearance-auto"

// choiceMark is the check, dash or dot drawn over a choiceBox: hidden until the input before it
// matches (peer-checked:, peer-indeterminate:).
const choiceMark = "pointer-events-none col-start-1 row-start-1 self-center justify-self-center opacity-0 " +
	"transition-opacity duration-150 motion-reduce:transition-none forced-colors:hidden"

// choiceInput wraps a drawn checkbox or radio and its marks in a one-cell grid, so the marks sit
// exactly on the box.
func choiceInput(input *html.InputElement, marks ...htemel.PhrasingContent) *html.SpanElement {
	return html.Span(append([]htemel.PhrasingContent{input}, marks...)...).Class(Classes("grid size-4 shrink-0 grid-cols-1"))
}

// checkMark is a 14x14 check (or, with indeterminate, a dash) stroked in white.
func checkMark(d, show string) *svg.SvgElement {
	return svg.Svg(svg.Path().D(d)).ViewBox("0 0 14 14").Fill("none").Stroke("currentColor").StrokeWidth("2").
		StrokeLinecap(svg.StrokeLinecapRound).StrokeLinejoin(svg.StrokeLinejoinRound).
		Aria("hidden", "true").
		Class(Classes(choiceMark, "size-3.5 text-white peer-disabled:text-neutral-950/25 dark:peer-disabled:text-white/25", show))
}

// Checkbox returns a checkbox with its label, <label><input type="checkbox">label</label>. attrs
// go on the <input>: Checkbox("Remember me", datastar.Bind("remember")).
//
// The box is drawn with tokens (appearance-none on the real input): a neutral border, the primary
// fill with a white check when checked, a dash when the input is indeterminate (set
// el.indeterminate from script), a focus-visible outline, and muted disabled states. In forced
// colors mode (Windows High Contrast) the native checkbox comes back.
func Checkbox(label string, attrs ...htemel.Attribute) *html.LabelElement {
	box := choiceInput(
		html.Input().Type(html.InputTypeCheckbox).Class(Classes(choiceBox, "rounded-[0.25rem] indeterminate:border-primary-600 indeterminate:bg-primary-600",
			"dark:indeterminate:border-primary-500 dark:indeterminate:bg-primary-500")).Attrs(attrs...),
		checkMark("M3 8L6 11L11 3.5", "peer-checked:opacity-100 peer-indeterminate:opacity-0"),
		checkMark("M3 7H11", "peer-indeterminate:opacity-100"),
	)
	return html.Label(box).Class(Classes(choiceLabel)).Text(label)
}

// Radio returns a radio button with its label. Radios with the same name form a group; bind them
// all to one signal with datastar.Bind to read the chosen value. attrs go on the <input>. It is
// drawn like Checkbox, round, with a white dot when chosen.
func Radio(name, value, label string, attrs ...htemel.Attribute) *html.LabelElement {
	box := choiceInput(
		html.Input().Type(html.InputTypeRadio).Name(name).Value(value).Class(Classes(choiceBox, "rounded-full")).Attrs(attrs...),
		html.Span().Aria("hidden", "true").
			Class(Classes(choiceMark, "size-1.5 rounded-full bg-white peer-checked:opacity-100 peer-disabled:bg-neutral-950/25 dark:peer-disabled:bg-white/25")),
	)
	return html.Label(box).Class(Classes(choiceLabel)).Text(label)
}

// RadioGroup returns a <fieldset> with legend as its caption and the radios stacked below.
func RadioGroup(legend string, radios ...*html.LabelElement) *html.FieldsetElement {
	list := html.Div().Class(Classes("mt-3 flex flex-col gap-3"))
	for _, r := range radios {
		list.Children(r)
	}
	return html.Fieldset(
		html.Legend().Class(Classes("text-sm/6 font-medium text-neutral-900 dark:text-white")).Text(legend),
		list,
	)
}

// Switch returns an on/off toggle: a checkbox with role="switch" drawn as a sliding knob, and its
// label. attrs go on the <input>: Switch("Notifications", datastar.Bind("notify")).
//
// The input covers the whole track with appearance-none, so the track and knob are all that is
// drawn. Off, the track has a neutral-500 inset ring and the knob a neutral-500 ring (both 3:1
// against what is behind them); on, the track fills with the primary color. Hover darkens the
// track, keyboard focus outlines it, and a disabled switch is muted.
func Switch(label string, attrs ...htemel.Attribute) *html.LabelElement {
	track := html.Span(
		html.Input().Type(html.InputTypeCheckbox).Role("switch").
			Class(Classes("absolute inset-0 size-full cursor-pointer appearance-none rounded-full focus:outline-none disabled:cursor-not-allowed")).
			Attrs(attrs...),
		html.Span().Aria("hidden", "true").
			Class(Classes("pointer-events-none size-5 rounded-full bg-white shadow-xs ring-1 ring-neutral-500 group-has-checked:ring-primary-600",
				"transition-transform duration-200 ease-out group-has-checked:translate-x-5 motion-reduce:transition-none",
				"dark:ring-0 dark:group-has-checked:ring-0")),
	).Class(Classes("group relative inline-flex h-6 w-11 shrink-0 items-center rounded-full bg-neutral-200 p-0.5 ring-1 ring-neutral-500 ring-inset",
		"transition-colors duration-150 has-checked:bg-primary-600 has-checked:ring-primary-600",
		"not-has-checked:has-enabled:hover:bg-neutral-300 has-checked:has-enabled:hover:bg-primary-700 has-checked:has-enabled:hover:ring-primary-700",
		"has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-primary-600 has-disabled:opacity-50",
		"dark:bg-white/10 dark:ring-white/35 dark:has-checked:bg-primary-500 dark:has-checked:ring-primary-500",
		"dark:not-has-checked:has-enabled:hover:bg-white/15 dark:has-checked:has-enabled:hover:bg-primary-400 dark:has-checked:has-enabled:hover:ring-primary-400",
		"dark:has-focus-visible:outline-primary-500"))
	return html.Label(track).Class(Classes(choiceLabel)).Text(label)
}
