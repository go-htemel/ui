package ui

import "github.com/go-htemel/htemel/html"

// toneSoft is the low-contrast color set of each tone, shared by Badge and Alert.
var toneSoft = map[Tone]string{
	Neutral: "bg-zinc-100 text-zinc-700 ring-zinc-500/20 dark:bg-zinc-400/10 dark:text-zinc-300 dark:ring-zinc-400/20",
	Info:    "bg-blue-50 text-blue-700 ring-blue-600/20 dark:bg-blue-400/10 dark:text-blue-300 dark:ring-blue-400/30",
	Success: "bg-green-50 text-green-700 ring-green-600/20 dark:bg-green-500/10 dark:text-green-300 dark:ring-green-500/20",
	Warning: "bg-amber-50 text-amber-800 ring-amber-600/20 dark:bg-amber-400/10 dark:text-amber-300 dark:ring-amber-400/20",
	Danger:  "bg-red-50 text-red-700 ring-red-600/20 dark:bg-red-400/10 dark:text-red-300 dark:ring-red-400/20",
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
		Class(Classes("inline-flex items-center rounded-md border border-zinc-200 bg-zinc-50 px-1.5 font-mono text-xs",
			"font-medium text-zinc-600 shadow-xs dark:border-white/10 dark:bg-white/5 dark:text-zinc-300")).
		Text(keys)
}
