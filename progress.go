package ui

import (
	"strconv"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// ProgressOption configures Progress and ProgressSignal.
type ProgressOption func(*progressConfig)

type progressConfig struct {
	size  Size
	tone  Tone
	label string
}

// ProgressSize sets the bar thickness (default Md).
func ProgressSize(s Size) ProgressOption { return func(c *progressConfig) { c.size = s } }

// ProgressTone colors the bar. The default, Neutral, uses the primary color.
func ProgressTone(t Tone) ProgressOption { return func(c *progressConfig) { c.tone = t } }

// ProgressLabel names the bar for assistive technology (aria-label).
func ProgressLabel(label string) ProgressOption { return func(c *progressConfig) { c.label = label } }

var progressHeights = map[Size]string{Xs: "h-1", Sm: "h-1.5", Md: "h-2.5", Lg: "h-4", Xl: "h-6"}

var progressTones = map[Tone]string{
	Neutral: "bg-primary-600 dark:bg-primary-500",
	Info:    "bg-info-600 dark:bg-info-500",
	Success: "bg-success-700 dark:bg-success-500",
	Warning: "bg-warning-700 dark:bg-warning-500",
	Danger:  "bg-danger-600 dark:bg-danger-500",
}

func progressBar(max int, opts []ProgressOption) (*html.DivElement, *html.DivElement) {
	var c progressConfig
	for _, opt := range opts {
		if opt != nil {
			opt(&c)
		}
	}
	fill := html.Div().Class(Classes("h-full rounded-full transition-[width] duration-300", pick(progressTones, c.tone, Neutral)))
	track := html.Div(fill).
		Role("progressbar").
		Aria("valuemin", "0").
		Aria("valuemax", strconv.Itoa(max)).
		Class(Classes("w-full overflow-hidden rounded-full bg-neutral-200 dark:bg-white/10", pick(progressHeights, c.size, Md)))
	if c.label != "" {
		track.Aria("label", c.label)
	}
	return track, fill
}

// Progress returns a progress bar showing value out of max (role="progressbar" with
// aria-valuenow). value is clamped to 0..max; a max below 1 counts as 100.
//
// Options: ProgressSize, ProgressTone, ProgressLabel.
func Progress(value, max int, opts ...ProgressOption) *html.DivElement {
	if max < 1 {
		max = 100
	}
	value = min(value, max)
	if value < 0 {
		value = 0
	}
	track, fill := progressBar(max, opts)
	fill.Style("width: " + strconv.FormatFloat(float64(value)*100/float64(max), 'f', -1, 64) + "%")
	return track.Aria("valuenow", strconv.Itoa(value))
}

// ProgressSignal returns a progress bar that follows the number signal, between 0 and max (the
// signal is declared as 0 unless the page already defines it). Set it from the browser
// ($upload = 40) or have the server patch it as work advances, for example with datastar-go's
// MarshalAndPatchSignals. A max below 1 counts as 100.
func ProgressSignal(signal string, max int, opts ...ProgressOption) *html.DivElement {
	if max < 1 {
		max = 100
	}
	sig, m := "$"+signal, strconv.Itoa(max)
	track, fill := progressBar(max, opts)
	fill.Style("width: 0%").
		Attrs(datastar.Style("width", "Math.min(100, Math.max(0, "+sig+" / "+m+" * 100)) + '%'"))
	return track.Aria("valuenow", "0").
		Attrs(initially(signal, "0"), datastar.Attr("aria-valuenow", "Math.min("+m+", Math.max(0, "+sig+"))"))
}

// Skeleton returns a placeholder for text that is still loading: lines pulsing grey bars, the
// last one shorter. It is a polite status with aria-busy; the bars are hidden from assistive
// technology.
func Skeleton(lines int) *html.DivElement {
	lines = max(lines, 1)
	d := html.Div().
		Role("status").
		Aria("busy", "true").
		Aria("label", "Loading").
		Class(Classes("flex animate-pulse flex-col gap-y-3"))
	for i := range lines {
		w := "w-full"
		if i == lines-1 && lines > 1 {
			w = "w-2/3"
		}
		d.Children(html.Div().Aria("hidden", "true").Class(Classes("h-3 rounded-full bg-neutral-200 dark:bg-white/10", w)))
	}
	return d
}

var spinnerSizes = map[Size]string{Xs: "size-3", Sm: "size-4", Md: "size-5", Lg: "size-6", Xl: "size-8"}

// Spinner returns a spinning ring of the given size in the current text color, with "Loading…"
// for screen readers. It is phrasing content, so it fits inside text and buttons.
func Spinner(size Size) *html.SpanElement {
	return html.Span(
		IconSpinner().Class(Classes("shrink-0 animate-spin", pick(spinnerSizes, size, Md))),
		html.Span().Class("sr-only").Text("Loading…"),
	).Role("status").Class("inline-flex items-center")
}

// Steps returns a horizontal step indicator for a multi-step flow: current is the 1-based
// number of the active step, earlier steps show a check mark and later ones are muted. The
// active step carries aria-current="step". The <nav> landmark is named "Progress"; a page with
// two step indicators needs a distinct name for each: use StepsLabel.
func Steps(current int, labels ...string) *html.NavElement {
	return StepsLabel("Progress", current, labels...)
}

// StepsLabel is Steps with label as the name of the <nav> landmark instead of "Progress"
// ("Checkout progress", "Signup progress"), so two step indicators on a page can be told apart.
func StepsLabel(label string, current int, labels ...string) *html.NavElement {
	list := html.Ol().Class(Classes("flex items-center"))
	for i, label := range labels {
		n := i + 1
		var dot htemel.PhrasingContent
		var state string
		switch {
		case n < current:
			state = " (completed)"
			dot = html.Span(IconCheck().Class("size-4")).
				Class(Classes("flex size-7 shrink-0 items-center justify-center rounded-full bg-primary-700 text-white"))
		case n == current:
			state = " (current)"
			dot = html.Span().Text(strconv.Itoa(n)).
				Class(Classes("flex size-7 shrink-0 items-center justify-center rounded-full text-sm font-semibold text-primary-700 ring-2 ring-primary-700",
					"dark:text-primary-400 dark:ring-primary-500"))
		default:
			dot = html.Span().Text(strconv.Itoa(n)).
				Class(Classes("flex size-7 shrink-0 items-center justify-center rounded-full text-sm font-medium text-neutral-500 ring-1 ring-neutral-300",
					"dark:text-neutral-400 dark:ring-white/20"))
		}
		text := "text-neutral-500 dark:text-neutral-400"
		if n <= current {
			text = "text-neutral-900 dark:text-white"
		}
		li := html.Li(
			dot,
			html.Span(htemel.Text(label), html.Span().Class("sr-only").Text(state)).Class(Classes("text-sm font-medium whitespace-nowrap", text)),
		).Class(Classes("flex items-center gap-x-2"))
		if n == current {
			li.Aria("current", "step")
		}
		if n > 1 {
			line := "bg-neutral-200 dark:bg-white/10"
			if n <= current {
				line = "bg-primary-700 dark:bg-primary-500"
			}
			list.Children(html.Li().Aria("hidden", "true").Class(Classes("mx-3 h-px min-w-6 flex-1", line)))
		}
		list.Children(li)
	}
	return html.Nav(list).Aria("label", navLabel([]string{label}, "Progress"))
}
