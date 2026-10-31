package ui

import (
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// FormControl is a control Field can label: Input, Textarea, Select, or any other element that is
// flow content.
type FormControl interface {
	htemel.FlowContent
	htemel.Element
}

const controlBase = "block w-full rounded-md bg-white text-base text-zinc-900 outline-1 -outline-offset-1 outline-zinc-300 " +
	"placeholder:text-zinc-400 focus:outline-2 focus:-outline-offset-2 focus:outline-sky-600 sm:text-sm/6 " +
	"disabled:cursor-not-allowed disabled:bg-zinc-50 disabled:text-zinc-500 " +
	"aria-invalid:outline-red-500 aria-invalid:focus:outline-red-600 " +
	"dark:bg-white/5 dark:text-white dark:outline-white/10 dark:placeholder:text-zinc-500 dark:focus:outline-sky-500 " +
	"dark:disabled:bg-white/10 dark:disabled:text-zinc-400 dark:aria-invalid:outline-red-400"

// Input returns a text-like <input> of type t with the given name. Bind it to a signal with
// .Attrs(datastar.Bind("email")) and put it in a Field for a label.
func Input(t html.InputType, name string) *html.InputElement {
	return html.Input().Type(t).Name(name).Class(Classes(controlBase, "px-3 py-1.5"))
}

// Textarea returns a multi-line text control with the given name and visible rows.
func Textarea(name string, rows int) *html.TextareaElement {
	return html.Textarea().Name(name).Rows(rows).Class(Classes(controlBase, "px-3 py-1.5"))
}

// Select returns a native <select> with the given name and options (see Choice).
func Select(name string, options ...html.SelectChild) *html.SelectElement {
	return html.Select(options...).Name(name).
		Class(Classes(controlBase, "py-1.5 pr-8 pl-3 dark:*:bg-zinc-800"))
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
// (so the label's for attribute points at it) and its aria-describedby.
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
	control.SetAttr("id", id)
	f := html.Div(
		html.Label().For(id).Class(Classes("block text-sm/6 font-medium text-zinc-900 dark:text-white")).Text(label),
		html.Div(control).Class("mt-2"),
	)
	var described []string
	if c.help != "" {
		described = append(described, id+"-help")
		f.Children(html.P().Id(id + "-help").Class(Classes("mt-2 text-sm text-zinc-500 dark:text-zinc-400")).Text(c.help))
	}
	errClass := Classes("mt-2 text-sm text-red-600 dark:text-red-400")
	if c.err != "" {
		described = append(described, id+"-error")
		control.SetAttr("aria-invalid", "true")
		f.Children(html.P().Id(id + "-error").Class(errClass).Text(c.err))
	}
	if c.errorSignal != "" {
		described = append(described, id+"-error-live")
		htemel.Apply(control, datastar.Attr("aria-invalid", "!!$"+c.errorSignal))
		f.Attrs(initially(c.errorSignal, "''"))
		f.Children(html.P().Id(id+"-error-live").Class(errClass).Style(hiddenUntilDatastar).
			Attrs(datastar.Show("!!$"+c.errorSignal), datastar.Text("$"+c.errorSignal)))
	}
	if len(described) > 0 {
		control.SetAttr("aria-describedby", strings.Join(described, " "))
	}
	return f
}

const choiceLabel = "inline-flex cursor-pointer items-center gap-x-3 text-sm/6 text-zinc-900 dark:text-white"

const choiceInput = "size-4 shrink-0 cursor-pointer accent-sky-600 focus-visible:outline-2 focus-visible:outline-offset-2 " +
	"focus-visible:outline-sky-600 disabled:cursor-not-allowed dark:accent-sky-500"

// Checkbox returns a checkbox with its label, <label><input type="checkbox">label</label>. attrs
// go on the <input>: Checkbox("Remember me", datastar.Bind("remember")).
func Checkbox(label string, attrs ...htemel.Attribute) *html.LabelElement {
	return html.Label(
		html.Input().Type(html.InputTypeCheckbox).Class(Classes(choiceInput)).Attrs(attrs...),
	).Class(Classes(choiceLabel)).Text(label)
}

// Radio returns a radio button with its label. Radios with the same name form a group; bind them
// all to one signal with datastar.Bind to read the chosen value. attrs go on the <input>.
func Radio(name, value, label string, attrs ...htemel.Attribute) *html.LabelElement {
	return html.Label(
		html.Input().Type(html.InputTypeRadio).Name(name).Value(value).Class(Classes(choiceInput)).Attrs(attrs...),
	).Class(Classes(choiceLabel)).Text(label)
}

// RadioGroup returns a <fieldset> with legend as its caption and the radios stacked below.
func RadioGroup(legend string, radios ...*html.LabelElement) *html.FieldsetElement {
	list := html.Div().Class(Classes("mt-3 flex flex-col gap-3"))
	for _, r := range radios {
		list.Children(r)
	}
	return html.Fieldset(
		html.Legend().Class(Classes("text-sm/6 font-medium text-zinc-900 dark:text-white")).Text(legend),
		list,
	)
}

// Switch returns an on/off toggle: a checkbox with role="switch" drawn as a sliding knob, and its
// label. attrs go on the <input>: Switch("Notifications", datastar.Bind("notify")).
func Switch(label string, attrs ...htemel.Attribute) *html.LabelElement {
	track := html.Span(
		html.Input().Type(html.InputTypeCheckbox).Role("switch").
			Class(Classes("absolute inset-0 size-full cursor-pointer appearance-none rounded-full focus:outline-none")).
			Attrs(attrs...),
		html.Span().Aria("hidden", "true").
			Class(Classes("pointer-events-none size-5 rounded-full bg-white shadow-xs ring-1 ring-zinc-900/5",
				"transition-transform duration-200 group-has-checked:translate-x-5")),
	).Class(Classes("group relative inline-flex h-6 w-11 shrink-0 items-center rounded-full bg-zinc-200 p-0.5",
		"transition-colors duration-200 has-checked:bg-sky-600 has-focus-visible:outline-2 has-focus-visible:outline-offset-2",
		"has-focus-visible:outline-sky-600 dark:bg-white/10 dark:has-checked:bg-sky-500"))
	return html.Label(track).Class(Classes(choiceLabel)).Text(label)
}
