package ui

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// describedInput is an <input> that keeps its own aria-describedby (a component's hint) when
// Field adds its help and error ids.
type describedInput struct {
	*html.InputElement
	describedBy string
}

func (d *describedInput) SetAttr(name, value string) {
	if name == "aria-describedby" {
		value = strings.TrimSpace(value + " " + d.describedBy)
	}
	d.InputElement.SetAttr(name, value)
}

// TagInputOption configures TagInput.
type TagInputOption func(*tagConfig)

type tagConfig struct {
	max         int
	suggestions []string
	placeholder string
	signal      string
}

// TagMax limits the number of tags; adding one more is refused and announced.
func TagMax(n int) TagInputOption { return func(c *tagConfig) { c.max = n } }

// TagSuggestions offers values in the browser's autocomplete list (a <datalist>); picking one
// adds it as a tag at once.
func TagSuggestions(values ...string) TagInputOption {
	return func(c *tagConfig) { c.suggestions = append(c.suggestions, values...) }
}

// TagPlaceholder sets the text field's placeholder (default "Add a tag…").
func TagPlaceholder(text string) TagInputOption { return func(c *tagConfig) { c.placeholder = text } }

// TagSignal mirrors the tags into an array signal: it starts as the initial tags (unless the page
// already has it) and follows every add and remove, so an expression or a request can read it
// ($tags.length, @post sends it). The mirror is one way: setting the signal does not change the
// chips.
func TagSignal(name string) TagInputOption { return func(c *tagConfig) { c.signal = name } }

// TagInputElement is the editor TagInput returns: a <div> holding the <ui-taginput> frame. Field
// labels its text field through Control.
type TagInputElement struct {
	*html.DivElement
	control *describedInput
}

// Control returns the text field, for Field (see CompositeControl).
func (e *TagInputElement) Control() FormControl { return e.control }

const (
	tagFrame = "flex min-h-9 w-full cursor-text flex-wrap items-center gap-1.5 rounded-control bg-white px-2 py-1 outline-1 -outline-offset-1 outline-field-border " +
		"not-focus-within:hover:outline-field-border-hover focus-within:outline-2 focus-within:-outline-offset-2 focus-within:outline-primary-600 " +
		"has-aria-invalid:outline-danger-500 dark:bg-white/5 dark:focus-within:outline-primary-500 dark:has-aria-invalid:outline-danger-400"
	tagChip   = "inline-flex max-w-full items-center gap-x-1 rounded-control bg-neutral-100 py-0.5 pr-1 pl-2 text-sm/5 font-medium text-neutral-800 dark:bg-white/10 dark:text-neutral-100"
	tagRemove = "inline-flex size-4 shrink-0 cursor-pointer items-center justify-center rounded-xs text-neutral-600 transition-colors duration-150 " +
		"hover:bg-neutral-950/10 hover:text-neutral-900 focus-visible:outline-2 focus-visible:outline-primary-600 " +
		"dark:text-neutral-300 dark:hover:bg-white/15 dark:hover:text-white dark:focus-visible:outline-primary-500"
	tagField = "ui-control min-w-24 flex-1 border-0 bg-transparent px-0.5 py-0.5 text-base text-neutral-900 placeholder:text-neutral-500 focus:outline-none " +
		"sm:text-sm/6 dark:text-white dark:placeholder:text-neutral-400"
)

func tagChipNode(name, tag string) *html.LiElement {
	return html.Li(
		html.Span().Class("truncate").Attrs(htemel.Attr("data-ui-tag-text", "")).Text(tag),
		html.Button(IconX().SetClass("size-3 shrink-0")).Type(html.ButtonTypeButton).Class(Classes(tagRemove)).
			Aria("label", "Remove "+tag).Attrs(htemel.Attr("data-ui-tag-remove", "")),
		html.Input().Type(html.InputTypeHidden).Name(name).Value(tag),
	).Class(Classes(tagChip)).Attrs(htemel.Attr("data-ui-tag", tag))
}

// TagInput returns a tag editor for the form field name: the tags as chips, each with a remove
// button, and a text field to type new ones. Enter or a comma adds the typed tag (a pasted
// "a, b, c" adds three), Backspace in the empty field removes the last one, duplicates (ignoring
// case) are refused. It is a CompositeControl, so Field labels the text field:
//
//	ui.Field("tags", "Tags", ui.TagInput("tags", post.Tags, ui.TagSuggestions("go", "htmx", "css"), ui.TagMax(5)))
//
// # Form value
//
// Each tag is a hidden <input name=name value=tag> inside its chip, so a plain form post sends
// tags=go&tags=css (r.Form["tags"] in Go) with no JavaScript and no JSON to decode; text still
// in the field when the form is submitted is sent as one more tag. TagSignal also mirrors the
// tags into a Datastar array signal for requests that send signals instead of a form.
//
// The editing is done by the <ui-taginput> element of ui.js (ScriptTag). Without it the
// server-rendered tags are still submitted and shown, but cannot be changed.
//
// # Accessibility
//
// The chips are a list (role="list", "Selected tags"); each remove button is named "Remove
// <tag>". The text field is described by a visually hidden hint ("Press Enter or comma to add a
// tag…", plus the limit), alongside Field's help and error text, and adds, removals and refusals
// are announced through a polite live region. After a removal focus returns to the text field.
func TagInput(name string, tags []string, opts ...TagInputOption) *TagInputElement {
	c := tagConfig{placeholder: "Add a tag…"}
	for _, o := range opts {
		if o != nil {
			o(&c)
		}
	}
	id := domID(name)
	hint := "Press Enter or comma to add a tag; Backspace in the empty field removes the last one."
	if c.max > 0 {
		hint += " Up to " + strconv.Itoa(c.max) + " tags."
	}
	list := html.Ul().Role("list").Aria("label", "Selected tags").Class("contents")
	for _, t := range tags {
		list.Children(tagChipNode(name, t))
	}
	in := html.Input().Type(html.InputTypeText).Autocomplete("off").Placeholder(c.placeholder).
		Class(Classes(tagField)).Attrs(htemel.Attr("data-ui-tag-input", ""))
	control := &describedInput{InputElement: in, describedBy: id + "-tag-hint"}
	control.SetAttr("aria-describedby", "")
	attrs := map[string]any{"data-ui-name": name, "class": Classes(tagFrame)}
	if c.max > 0 {
		attrs["data-ui-max"] = c.max
	}
	children := []htemel.Node{
		list, in,
		html.Template(tagChipNode(name, "")).Attrs(htemel.Attr("data-ui-tag-template", "")),
		html.Span().Id(id + "-tag-hint").Class("sr-only").Text(hint),
		html.Span().Class("sr-only").Aria("live", "polite").Attrs(htemel.Attr("data-ui-tag-status", "")),
	}
	if len(c.suggestions) > 0 {
		in.List(id + "-suggestions")
		dl := html.Datalist().Id(id + "-suggestions")
		for _, s := range c.suggestions {
			dl.Children(html.Option().Value(s))
		}
		children = append(children, dl)
	}
	root := html.Div(htemel.Generic("ui-taginput", attrs, children...))
	if c.signal != "" {
		if tags == nil {
			tags = []string{}
		}
		b, _ := json.Marshal(tags)
		root.Attrs(initially(c.signal, string(b)), datastar.On("ui:tags", "$"+c.signal+" = evt.detail.tags"))
	}
	return &TagInputElement{DivElement: root, control: control}
}

var _ CompositeControl = (*TagInputElement)(nil)
