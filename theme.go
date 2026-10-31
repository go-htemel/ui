package ui

import (
	_ "embed"
	"strings"
)

// themeCSS holds the default design tokens; see ThemeCSS.
//
//go:embed assets/ui.css
var themeCSS string

// ThemeCSS returns the default design tokens as a Tailwind v4 stylesheet: one @theme block that
// defines the color families the components use (primary, neutral, success, warning, danger,
// info, each a 50-950 scale: primary is a tuned Emerald Sky in literal OKLCH, the others alias a
// Tailwind palette) and the radius tokens
// (--radius-control, --radius-surface). Write it into your Tailwind sources, import it after
// tailwindcss, and override the variables in your own CSS:
//
//	go run github.com/go-htemel/ui/cmd/themecss > assets/ui.theme.css
//
// Assets serves the same file as /ui.css. Because the tokens are plain CSS variables, retheming
// needs no Go change and no rebuild of the Go program; see the README.
func ThemeCSS() string { return themeCSS }

// Classes joins class lists and normalizes the whitespace: the classes in order, separated by
// single spaces, with empty arguments dropped. It does not de-duplicate or merge conflicting
// classes. The internal/safelist scanner treats its arguments as class lists.
//
//	html.Div().Class(ui.Classes("rounded-surface bg-primary-600", extra))
func Classes(classes ...string) string {
	return strings.Join(strings.Fields(strings.Join(classes, " ")), " ")
}

// bodyClasses is what BodyClasses returns.
const bodyClasses = "font-sans bg-neutral-50 text-neutral-900 antialiased has-[dialog:modal]:overflow-hidden " +
	"dark:bg-neutral-950 dark:text-neutral-100"

// BodyClasses returns base classes for <body>: font-sans (Tailwind's --font-sans), the neutral
// page colors in light and dark mode, antialiasing, and a scroll lock while a modal Dialog is
// open.
func BodyClasses() string { return bodyClasses }
