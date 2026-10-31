package ui

import (
	"time"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// DateOption configures DateInput, DateTimeInput and TimeInput.
type DateOption func(*html.InputElement)

// DateMin sets the earliest value, in the input's format: "2026-01-01" (date),
// "2026-01-01T09:00" (date and time), "09:00" (time).
func DateMin(v string) DateOption { return func(in *html.InputElement) { in.Min(v) } }

// DateMax sets the latest value, in the input's format.
func DateMax(v string) DateOption { return func(in *html.InputElement) { in.Max(v) } }

// DateStep sets the step: days for DateInput ("7"), seconds for TimeInput and DateTimeInput
// ("900" for quarter hours, "1" to show seconds).
func DateStep(v string) DateOption { return func(in *html.InputElement) { in.Step(v) } }

// DateValue sets the initial value, in the input's format.
func DateValue(v string) DateOption { return func(in *html.InputElement) { in.Value(v) } }

func dateInput(t html.InputType, name string, opts []DateOption) *html.InputElement {
	in := Input(t, name)
	for _, o := range opts {
		if o != nil {
			o(in)
		}
	}
	return in
}

// DateInput returns a native date field (<input type="date">), styled like Input with the picker
// icon in the text color. Bind it with .Attrs(datastar.Bind("published")); the value is
// "YYYY-MM-DD" whatever the reader's locale shows.
func DateInput(name string, opts ...DateOption) *html.InputElement {
	return dateInput(html.InputTypeDate, name, opts)
}

// DateTimeInput returns a native local date and time field (<input type="datetime-local">),
// value "YYYY-MM-DDTHH:MM", without a time zone: convert on the server.
func DateTimeInput(name string, opts ...DateOption) *html.InputElement {
	return dateInput(html.InputTypeDatetimeLocal, name, opts)
}

// TimeInput returns a native time field (<input type="time">), value "HH:MM" (or "HH:MM:SS"
// with a step under 60).
func TimeInput(name string, opts ...DateOption) *html.InputElement {
	return dateInput(html.InputTypeTime, name, opts)
}

// slugify is the Datastar expression that turns the text in expr into a URL slug: lowercase,
// accents removed, runs of anything but letters and digits turned into one hyphen, no leading
// or trailing hyphen. It avoids regex anchors ($ would read as a signal).
func slugify(expr string) string {
	return "String(" + expr + " ?? '').toLowerCase().normalize('NFKD').replace(/[\\u0300-\\u036f]/g, '')" +
		".replace(/[^a-z0-9]+/g, '-').split('-').filter(Boolean).join('-')"
}

// SlugField returns a text Input for a URL slug that follows another signal, typically a title:
// while the reader has not edited it, its value is the slug of fromSignal ("Hello, Wörld!"
// becomes "hello-world"). Typing in it stops the following (the signal <name>Touched turns
// true); clearing it resumes. It is bound to the signal name, so a form post and a Datastar
// request both carry it. No script file is involved: it is a data-effect and an input handler.
//
//	ui.Field("title", "Title", ui.Input(html.InputTypeText, "title").Attrs(datastar.Bind("title")))
//	ui.Field("slug", "Slug", ui.SlugField("slug", "title"), ui.HelpText("Follows the title until you edit it."))
//
// The signals start empty (and Touched false) unless the page defines them: when editing a saved
// post, preset both, e.g. datastar.Signals(map[string]any{"slug": p.Slug, "slugTouched": true}),
// so the saved slug is kept.
func SlugField(name, fromSignal string) *html.InputElement {
	touched := signalID(name) + "Touched"
	return Input(html.InputTypeText, name).
		Autocomplete("off").
		Attrs(
			htemel.Attr("spellcheck", "false"),
			initially(name, "''"),
			initially(touched, "false"),
			// The input handler comes before data-bind: Datastar adds listeners in attribute order,
			// so Touched is set before the typed value reaches the signal and wakes the effect.
			datastar.OnInput("$"+touched+" = el.value !== ''"),
			datastar.Bind(name),
			datastar.Effect("if (!$"+touched+") $"+name+" = "+slugify("$"+fromSignal)),
		)
}

// FilterDebounce is how long FilterSearch waits after the last keystroke before it runs its
// action.
const FilterDebounce = 300 * time.Millisecond

// FilterBar returns a row of filter controls above a list or table: typically a FilterSearch,
// a few Selects and a reset Button. It wraps onto more lines when narrow (the search field takes
// the remaining width, at least 15rem) and is a search landmark (role="search", aria-label
// "Filters").
//
//	ui.FilterBar(
//		ui.FilterSearch("q", datastar.Get("/people")),
//		ui.Select("team", ...).Attrs(datastar.Bind("team"), datastar.OnChange(datastar.Get("/people"))),
//		ui.Button(ui.Ghost, "Reset").Attrs(datastar.OnClick(ui.FilterReset("q", "team")+"; "+datastar.Get("/people"))),
//	)
func FilterBar(controls ...htemel.FlowContent) *html.DivElement {
	return html.Div(controls...).Role("search").Aria("label", "Filters").
		Class(Classes("flex flex-wrap items-center gap-3 *:min-w-0"))
}

// FilterSearch returns the search field of a FilterBar: an <input type="search"> bound to signal
// (declared empty unless the page has it) that runs action, usually datastar.Get(url), once the
// reader has stopped typing for FilterDebounce (300ms); the request carries the signal. It is
// named "Search" (aria-label; give it a visible label with Field instead if you prefer) and
// grows to fill the row.
func FilterSearch(signal, action string) *html.InputElement {
	return Input(html.InputTypeSearch, signal).
		Placeholder("Search…").
		Aria("label", "Search").
		Autocomplete("off").
		AddClass(Classes("flex-1 basis-60")).
		Attrs(
			initially(signal, "''"),
			datastar.Bind(signal),
			datastar.OnInput(action, datastar.Debounce(FilterDebounce)),
		)
}

// FilterReset returns the expression that empties the filter signals: FilterReset("q", "team")
// is "$q = ”; $team = ”". Follow it with the action that reloads the list.
func FilterReset(signals ...string) string {
	out := ""
	for i, s := range signals {
		if i > 0 {
			out += "; "
		}
		out += "$" + s + " = ''"
	}
	return out
}
