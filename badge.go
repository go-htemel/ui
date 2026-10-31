package ui

import "github.com/go-htemel/htemel/html"

// toneSoft is the low-contrast color set of each tone, shared by Badge and Alert.
var toneSoft = map[Tone]string{
	Neutral: "bg-neutral-100 text-neutral-700 ring-neutral-500/20 dark:bg-neutral-400/10 dark:text-neutral-300 dark:ring-neutral-400/20",
	Brand:   "bg-primary-50 text-primary-700 ring-primary-600/20 dark:bg-primary-400/10 dark:text-primary-300 dark:ring-primary-400/30",
	Info:    "bg-info-50 text-info-700 ring-info-600/20 dark:bg-info-400/10 dark:text-info-300 dark:ring-info-400/30",
	Success: "bg-success-50 text-success-700 ring-success-600/20 dark:bg-success-500/10 dark:text-success-300 dark:ring-success-500/20",
	Warning: "bg-warning-50 text-warning-800 ring-warning-600/20 dark:bg-warning-400/10 dark:text-warning-300 dark:ring-warning-400/20",
	Danger:  "bg-danger-50 text-danger-700 ring-danger-600/20 dark:bg-danger-400/10 dark:text-danger-300 dark:ring-danger-400/20",
}

// Badge returns a small pill, <span>, with label in tone t.
func Badge(t Tone, label string) *html.SpanElement {
	return html.Span().
		Class(Classes("inline-flex items-center gap-x-1 rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset",
			pick(toneSoft, t, Neutral))).
		Text(label)
}

// Kbd returns a keyboard key, <kbd>, such as Kbd("Ctrl") or Kbd("⌘K").
func Kbd(keys string) *html.KbdElement {
	return html.Kbd().
		Class(Classes("inline-flex items-center rounded-control border border-neutral-200 bg-neutral-50 px-1.5 font-mono text-xs",
			"font-medium text-neutral-600 shadow-xs dark:border-white/10 dark:bg-white/5 dark:text-neutral-300")).
		Text(keys)
}
