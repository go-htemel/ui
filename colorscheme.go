package ui

import (
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// ColorSchemeKey is the localStorage key that holds the chosen color scheme: "light", "dark", or
// absent (or "system") to follow the operating system. ColorSchemeBootstrap reads it and
// ColorSchemeToggle writes it.
const ColorSchemeKey = "color-scheme"

// colorSchemeScript applies the stored color scheme to <html> as the .dark class. It is inline so
// it runs while the document is still being parsed, before the first paint; see
// ColorSchemeBootstrap. It must stay free of "</script" and of anything that needs escaping in a
// <script> element.
const colorSchemeScript = `(function () {
  var root = document.documentElement;
  var query = matchMedia('(prefers-color-scheme: dark)');
  function apply() {
    var stored = null;
    try { stored = localStorage.getItem('` + ColorSchemeKey + `'); } catch (e) {}
    root.classList.toggle('dark', stored === 'dark' || (stored !== 'light' && query.matches));
  }
  apply();
  query.addEventListener('change', apply);
})();`

// ColorSchemeOption configures ColorSchemeBootstrap and ColorSchemeToggle. Each reads the options
// it documents and ignores the rest.
type ColorSchemeOption func(*colorSchemeOptions)

type colorSchemeOptions struct {
	nonce  string
	labels bool
}

// WithNonce sets the nonce attribute of the script ColorSchemeBootstrap renders, for a
// Content-Security-Policy that allows inline scripts by nonce (script-src 'nonce-...').
func WithNonce(nonce string) ColorSchemeOption {
	return func(o *colorSchemeOptions) { o.nonce = nonce }
}

// WithLabels shows the words System, Light and Dark next to the icons of ColorSchemeToggle. By
// default the labels are the buttons' accessible names and tooltips only.
func WithLabels() ColorSchemeOption {
	return func(o *colorSchemeOptions) { o.labels = true }
}

// ColorSchemeBootstrap returns the inline <script> that decides, before the page first paints,
// whether <html> carries the .dark class, which is what the library's dark: variants follow (see
// the README, "Light and dark mode"). Put it in <head> before the stylesheet or the Tailwind
// script, and render it on every page:
//
//	html.Head(ui.ColorSchemeBootstrap(), html.Link().Rel("stylesheet").Href("/static/app.css"), ...)
//
// It reads localStorage[ColorSchemeKey] ("light", "dark", or absent for the system setting),
// follows prefers-color-scheme while the setting is the system's, and tolerates storage that is
// blocked or full. It is about ten lines and needs no dependency.
//
// This is the one inline script in the library, and it has to be: an external or deferred script
// would run after the first paint, and a dark-mode visitor would see a flash of the light page.
// A Content-Security-Policy that forbids inline scripts blocks it; pass a per-response nonce with
// WithNonce and allow it with script-src 'nonce-<value>', or hash the script (its text is
// ColorSchemeBootstrapSource) and allow the 'sha256-...' source.
func ColorSchemeBootstrap(opts ...ColorSchemeOption) *html.ScriptElement {
	var o colorSchemeOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	s := html.Script().Text(colorSchemeScript)
	if o.nonce != "" {
		s.Nonce(o.nonce)
	}
	return s
}

// ColorSchemeBootstrapSource returns the text of the script ColorSchemeBootstrap renders, for a
// Content-Security-Policy hash source.
func ColorSchemeBootstrapSource() string { return colorSchemeScript }

// colorSchemeSegment is the state styling of one ColorSchemeToggle button, added to the
// Secondary variant: the selected one (aria-pressed="true") is a primary solid with white text.
const colorSchemeSegment = "aria-pressed:bg-primary-600 aria-pressed:text-white aria-pressed:hover:bg-primary-700 " +
	"dark:aria-pressed:bg-primary-600 dark:aria-pressed:hover:bg-primary-500"

// ColorSchemeToggle returns a segmented control with three buttons, System, Light and Dark, that
// sets the color scheme and remembers it in localStorage[ColorSchemeKey]. It needs
// ColorSchemeBootstrap on the page (the bootstrap applies the choice on every load) and Datastar.
//
// The current choice lives in the colorScheme signal ("system", "light" or "dark"), initialised
// from localStorage, and each button reflects it in aria-pressed, so it also drives your own
// markup: data-show="$colorScheme === 'dark'". Several toggles on one page stay in step.
//
// Options: WithLabels.
func ColorSchemeToggle(opts ...ColorSchemeOption) *html.SpanElement {
	var o colorSchemeOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	segment := func(mode, label string, icon htemel.PhrasingContent) *html.ButtonElement {
		var b *html.ButtonElement
		if o.labels {
			b = Button(Secondary, label, WithIcon(icon), WithSize(Sm), withClass(colorSchemeSegment))
		} else {
			b = IconButton(Secondary, icon, label, WithSize(Sm), withClass(colorSchemeSegment))
		}
		pressed := "false"
		if mode == "system" {
			pressed = "true" // server-rendered default until the init expression reads the storage
		}
		return b.Attrs(
			htemel.Attr("aria-pressed", pressed),
			datastar.Attr("aria-pressed", "$colorScheme === '"+mode+"' ? 'true' : 'false'"),
			datastar.OnClick(colorSchemeSet(mode)),
		)
	}
	return ButtonGroup(
		segment("system", "System", IconMonitor()),
		segment("light", "Light", IconSun()),
		segment("dark", "Dark", IconMoon()),
	).Aria("label", "Color scheme").Attrs(
		initially("colorScheme", "'system'"),
		datastar.Init("$colorScheme = ['light', 'dark'].includes(localStorage.getItem('"+ColorSchemeKey+"')) ? localStorage.getItem('"+ColorSchemeKey+"') : 'system'"),
	)
}

// colorSchemeSet is the click expression of a toggle button. The signal and the class change
// first and the storage last, so a browser that blocks localStorage still switches for the
// current page.
func colorSchemeSet(mode string) string {
	const root = "document.documentElement.classList"
	if mode == "system" {
		return "$colorScheme = 'system'; " + root + ".toggle('dark', matchMedia('(prefers-color-scheme: dark)').matches); " +
			"localStorage.removeItem('" + ColorSchemeKey + "')"
	}
	return "$colorScheme = '" + mode + "'; " + root + ".toggle('dark', " + boolLit(mode == "dark") + "); " +
		"localStorage.setItem('" + ColorSchemeKey + "', '" + mode + "')"
}

func boolLit(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
