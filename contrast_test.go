package ui_test

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/go-htemel/ui"
	"github.com/go-htemel/ui/internal/wcag"
)

// This file measures the WCAG 2.x contrast of the shipped design tokens. The matrix below lists
// the foreground/background class pairs the components use; the colors are resolved through
// ui.ThemeCSS() (neutral-500 -> zinc-500 -> OKLCH -> sRGB; primary aliases indigo), so the matrix follows the
// defaults.
//
// Every row names the source line it was read from. TestContrastMatrixInSync checks that those
// lines still contain the classes and that the goldens still render them, so the matrix cannot
// silently drift from the components.
//
// Set CONTRAST_REPORT=path.md to also write the full table (passes and failures).

// kind classifies what a row measures, which sets its minimum ratio (WCAG 1.4.3 and 1.4.11).
type kind int

const (
	text      kind = iota // normal text: 4.5:1
	largeText             // at least 18pt, or 14pt bold (text-xl and up, semibold): 3:1
	nonText               // focus indicators, borders, icons, fills: 3:1
)

func (k kind) min() float64 {
	if k == text {
		return 4.5
	}
	return 3
}

func (k kind) String() string { return [...]string{"text", "large text", "non-text"}[k] }

// side is one color scheme of a row.
type side struct {
	fg, bg string   // color utilities; bg "" means the foreground sits directly on the backdrop
	under  string   // backdrop utility; "" means the page background of the scheme
	group  float64  // opacity of the whole element (bg and fg fade toward the backdrop), 0 = 1
	also   []string // other classes of the row that must exist in the goldens
}

func on(fg, bg string) side       { return side{fg: fg, bg: bg} }
func (s side) over(u string) side { s.under = u; return s }
func (s side) grp(a float64) side { s.group = a; return s }
func (s side) with(c ...string) side {
	s.also = append(s.also, c...)
	return s
}

// spec is one component state in both schemes.
type spec struct {
	component, state, src string
	kind                  kind
	exempt                bool // inactive controls: WCAG 1.4.3 exempts them; reported, never failing
	light, dark           side
}

const (
	pageL = "bg-neutral-50"       // BodyClasses
	pageD = "dark:bg-neutral-950" // BodyClasses
	cardL = "bg-white"
	cardD = "dark:bg-neutral-900"
)

type tone struct {
	name, line, lfg, lbg, dfg, dbg, lborder, dborder string
}

// tones is toneSoft (badge.go:7-12) and toneAccent (toast.go:41-48), in map order: accent
// borders are one step darker in light (white toast) than in dark (neutral-800 toast). Brand is
// the primary color's tints; Info aliases sky.
var tones = []tone{
	{"Neutral", "7", "text-neutral-700", "bg-neutral-100", "dark:text-neutral-300", "dark:bg-neutral-400/10", "border-neutral-500", "border-neutral-500"},
	{"Brand", "8", "text-primary-700", "bg-primary-50", "dark:text-primary-300", "dark:bg-primary-400/10", "border-primary-600", "dark:border-primary-500"},
	{"Info", "9", "text-info-700", "bg-info-50", "dark:text-info-300", "dark:bg-info-400/10", "border-info-600", "dark:border-info-500"},
	{"Success", "10", "text-success-700", "bg-success-50", "dark:text-success-300", "dark:bg-success-500/10", "border-success-600", "dark:border-success-500"},
	{"Warning", "11", "text-warning-800", "bg-warning-50", "dark:text-warning-300", "dark:bg-warning-400/10", "border-warning-600", "dark:border-warning-500"},
	{"Danger", "12", "text-danger-700", "bg-danger-50", "dark:text-danger-300", "dark:bg-danger-400/10", "border-danger-600", "dark:border-danger-500"},
}

func matrix() []spec {
	var m []spec
	add := func(component, state, src string, k kind, light, dark side) {
		m = append(m, spec{component: component, state: state, src: src, kind: k, light: light, dark: dark})
	}
	disabled := func(component, state, src string, light, dark side) {
		m = append(m, spec{component: component, state: state, src: src, kind: text, exempt: true, light: light, dark: dark})
	}
	const dis = "disabled:opacity-50" // buttonBase, button.go:10

	// Button (button.go:14-25 buttonVariants). Text is text-sm/text-xs semibold: normal text.
	// Solids with white text use the conventional steps: 600, hover 700 in light. In dark, 600 with
	// a lighter hover 500: white on indigo-500 is 4.58:1 (indigo-400 is only 3.13:1, so the
	// Tailwind UI 500 -> 400 scheme would miss 4.5:1).
	add("Button/Primary", "default", "button.go:14", text, on("text-white", "bg-primary-600"), on("text-white", "dark:bg-primary-600"))
	add("Button/Primary", "hover", "button.go:14", text, on("text-white", "hover:bg-primary-700"), on("text-white", "dark:hover:bg-primary-500"))
	add("Button/Primary", "focus-visible outline", "button.go:14", nonText, on("focus-visible:outline-primary-600", ""), on("dark:focus-visible:outline-primary-500", ""))
	add("Button/Primary", "focus-visible outline on a card", "button.go:14", nonText, on("focus-visible:outline-primary-600", "").over(cardL), on("dark:focus-visible:outline-primary-500", "").over("dark:bg-neutral-800"))
	disabled("Button/Primary", "disabled", "button.go:14", on("text-white", "bg-primary-600").grp(0.5).with(dis), on("text-white", "dark:bg-primary-600").grp(0.5).with(dis))

	add("Button/Secondary", "default", "button.go:16", text, on("text-neutral-900", "bg-white"), on("dark:text-white", "dark:bg-white/10"))
	add("Button/Secondary", "hover", "button.go:16", text, on("text-neutral-900", "hover:bg-neutral-50"), on("dark:text-white", "dark:hover:bg-white/20"))
	add("Button/Secondary", "focus-visible outline", "button.go:16", nonText, on("focus-visible:outline-neutral-500", ""), on("focus-visible:outline-neutral-500", ""))
	disabled("Button/Secondary", "disabled", "button.go:16", on("text-neutral-900", "bg-white").grp(0.5).with(dis), on("dark:text-white", "dark:bg-white/10").grp(0.5).with(dis))

	add("Button/Outline", "default", "button.go:18", text, on("text-primary-600", ""), on("dark:text-primary-400", ""))
	add("Button/Outline", "hover", "button.go:18", text, on("hover:text-primary-700", "hover:bg-primary-50"), on("dark:hover:text-primary-400", "dark:hover:bg-primary-500/10"))
	add("Button/Outline", "ring (boundary)", "button.go:18", nonText, on("ring-primary-600", ""), on("dark:ring-primary-500", ""))
	add("Button/Outline", "focus-visible outline", "button.go:18", nonText, on("focus-visible:outline-primary-600", ""), on("dark:focus-visible:outline-primary-500", ""))
	disabled("Button/Outline", "disabled", "button.go:18", on("text-primary-600", "").grp(0.5).with(dis), on("dark:text-primary-400", "").grp(0.5).with(dis))

	add("Button/Ghost", "default", "button.go:20", text, on("text-neutral-700", ""), on("dark:text-neutral-300", ""))
	add("Button/Ghost", "hover", "button.go:20", text, on("text-neutral-700", "hover:bg-neutral-100"), on("dark:text-neutral-300", "dark:hover:bg-white/10"))
	add("Button/Ghost", "focus-visible outline", "button.go:20", nonText, on("focus-visible:outline-neutral-500", ""), on("focus-visible:outline-neutral-500", ""))
	disabled("Button/Ghost", "disabled", "button.go:20", on("text-neutral-700", "").grp(0.5).with(dis), on("dark:text-neutral-300", "").grp(0.5).with(dis))

	add("Button/Danger", "default", "button.go:22", text, on("text-white", "bg-danger-600"), on("text-white", "dark:bg-danger-700"))
	add("Button/Danger", "hover", "button.go:22", text, on("text-white", "hover:bg-danger-700"), on("text-white", "dark:hover:bg-danger-600"))
	add("Button/Danger", "focus-visible outline", "button.go:22", nonText, on("focus-visible:outline-danger-600", ""), on("dark:focus-visible:outline-danger-500", ""))
	disabled("Button/Danger", "disabled", "button.go:22", on("text-white", "bg-danger-600").grp(0.5).with(dis), on("text-white", "dark:bg-danger-700").grp(0.5).with(dis))

	add("Button/Link", "default", "button.go:24", text, on("text-primary-600", ""), on("dark:text-primary-400", ""))
	add("Button/Link", "hover", "button.go:24", text, on("hover:text-primary-700", ""), on("dark:hover:text-primary-300", ""))
	add("Button/Link", "focus-visible outline", "button.go:24", nonText, on("focus-visible:outline-primary-600", ""), on("dark:focus-visible:outline-primary-500", ""))
	disabled("Button/Link", "disabled", "button.go:24", on("text-primary-600", "").grp(0.5).with(dis), on("dark:text-primary-400", "").grp(0.5).with(dis))

	// ColorSchemeToggle segments (colorscheme.go:87-88 colorSchemeSegment on top of Secondary,
	// button.go:16): the selected segment is a primary solid with white text; the others are
	// plain Secondary buttons.
	add("ColorSchemeToggle/inactive segment", "default", "button.go:16", text, on("text-neutral-900", "bg-white"), on("dark:text-white", "dark:bg-white/10"))
	add("ColorSchemeToggle/active segment", "default", "colorscheme.go:87", text, on("aria-pressed:text-white", "aria-pressed:bg-primary-600"), on("aria-pressed:text-white", "dark:aria-pressed:bg-primary-600"))
	add("ColorSchemeToggle/active segment", "hover", "colorscheme.go:87", text, on("aria-pressed:text-white", "aria-pressed:hover:bg-primary-700"), on("aria-pressed:text-white", "dark:aria-pressed:hover:bg-primary-500"))

	// Badge, Alert (title and body) and Toast per tone.
	for _, t := range tones {
		add("Badge/"+t.name, "text on tint", "badge.go:"+t.line, text, on(t.lfg, t.lbg), on(t.dfg, t.dbg))
		add("Alert/"+t.name, "title on tint", "badge.go:"+t.line, text, on(t.lfg, t.lbg), on(t.dfg, t.dbg))
		// alert.go:33 renders the body at full opacity in the tone color (it used to be opacity-90,
		// which took Success below 4.5:1).
		add("Alert/"+t.name, "body on tint", "badge.go:"+t.line, text, on(t.lfg, t.lbg), on(t.dfg, t.dbg))
		add("Toast/"+t.name, "accent border (toast.go:41)", "toast.go:"+strconv.Itoa(42+indexOf(t.name)), nonText,
			on(t.lborder, "bg-white"), on(t.dborder, "dark:bg-neutral-800"))
	}

	// Toast (toast.go:63-74). The close button is a Ghost IconButton (ghostClose, signals.go).
	add("Toast", "title", "toast.go:63", text, on("text-neutral-900", "bg-white"), on("dark:text-white", "dark:bg-neutral-800"))
	add("Toast", "message", "toast.go:66", text, on("text-neutral-600", "bg-white"), on("dark:text-neutral-300", "dark:bg-neutral-800"))
	add("Toast", "close icon", "button.go:20", nonText, on("text-neutral-700", "bg-white"), on("dark:text-neutral-300", "dark:bg-neutral-800"))
	add("Toast", "close icon, hover", "button.go:20", nonText, on("text-neutral-700", "hover:bg-neutral-100").over("bg-white"), on("dark:text-neutral-300", "dark:hover:bg-white/10").over("dark:bg-neutral-800"))

	// Input (form.go:29-34). Dark inputs are bg-white/5 over the page.
	add("Input", "text", "form.go:33", text, on("text-neutral-900", "bg-white"), on("dark:text-white", "dark:bg-white/5"))
	add("Input", "placeholder", "form.go:33", text, on("placeholder:text-neutral-500", "bg-white"), on("dark:placeholder:text-neutral-400", "dark:bg-white/5"))
	// outline-field-border is a token: --color-field-border under @theme (light) and under .dark
	// (white at 35%), so the same class is measured against the light and the dark theme. The
	// hover color is one notch darker (neutral-500) in light and brighter (white at 55%) in dark.
	add("Input", "border (SC 1.4.11)", "form.go:29", nonText, on("outline-field-border", ""), on("outline-field-border", ""))
	add("Input", "border against the field (SC 1.4.11)", "form.go:29", nonText, on("outline-field-border", "bg-white"), on("outline-field-border", "dark:bg-white/5"))
	add("Input", "border on a card (SC 1.4.11)", "form.go:29", nonText, on("outline-field-border", "").over(cardL), on("outline-field-border", "").over(cardD))
	add("Input", "border against the field on a card (SC 1.4.11)", "form.go:29", nonText, on("outline-field-border", "bg-white").over(cardL), on("outline-field-border", "dark:bg-white/5").over(cardD))
	add("Input", "border, hover", "form.go:29", nonText, on("enabled:hover:outline-field-border-hover", ""), on("enabled:hover:outline-field-border-hover", ""))
	add("Input", "border against the field, hover", "form.go:29", nonText, on("enabled:hover:outline-field-border-hover", "bg-white"), on("enabled:hover:outline-field-border-hover", "dark:bg-white/5"))
	add("InputGroup", "affix border", "info.go:105", nonText, on("outline-field-border", "bg-neutral-50"), on("outline-field-border", "dark:bg-white/5"))
	add("Input", "focus outline", "form.go:30", nonText, on("focus:outline-primary-600", ""), on("dark:focus:outline-primary-500", ""))
	add("Input", "invalid outline", "form.go:32", nonText, on("aria-invalid:outline-danger-500", ""), on("dark:aria-invalid:outline-danger-400", ""))
	disabled("Input", "disabled text", "form.go:31", on("disabled:text-neutral-500", "disabled:bg-neutral-50"), on("dark:disabled:text-neutral-400", "dark:disabled:bg-white/10"))

	// Autofill (assets/ui.css): the opaque cover and text fill that replace Chrome's autofill
	// colors. Dark uses neutral-800 because the cover must be opaque (the field is white/5).
	add("Input", "autofilled text", "assets/ui.css:176", text, on("text-neutral-900", "bg-white"), on("dark:text-white", "dark:bg-neutral-800"))

	// File input button (form.go fileButton): Secondary-like, on the field.
	add("Input/file button", "text", "form.go:39", text, on("file:text-neutral-900", "file:bg-white").over(cardL), on("dark:file:text-white", "dark:file:bg-white/10").over("dark:bg-neutral-800"))
	add("Input/file button", "hover", "form.go:38", text, on("file:text-neutral-900", "hover:file:bg-neutral-50").over(cardL), on("dark:file:text-white", "dark:hover:file:bg-white/20").over("dark:bg-neutral-800"))

	// Checkbox and Radio (form.go choiceBox, choiceMark), on the example card.
	add("Checkbox/Radio", "label", "form.go:146", text, on("text-neutral-900", "").over(cardL), on("dark:text-white", "").over(cardD))
	add("Checkbox/Radio", "unchecked border vs card (SC 1.4.11)", "form.go:152", nonText, on("border-field-border", "").over(cardL), on("border-field-border", "").over(cardD))
	add("Checkbox/Radio", "unchecked border vs box fill (SC 1.4.11)", "form.go:152,form.go:157", nonText, on("border-field-border", "bg-white").over(cardL), on("border-field-border", "dark:bg-white/5").over(cardD))
	add("Checkbox/Radio", "hover border", "form.go:153,form.go:157", nonText, on("not-checked:enabled:hover:border-field-border-hover", "bg-white").over(cardL), on("not-checked:enabled:hover:border-field-border-hover", "dark:bg-white/5").over(cardD))
	add("Checkbox/Radio", "checked fill vs card", "form.go:154", nonText, on("checked:bg-primary-600", "").over(cardL), on("dark:checked:bg-primary-500", "").over(cardD))
	add("Checkbox", "check mark on fill", "form.go:154", nonText, on("text-white", "checked:bg-primary-600"), on("text-white", "dark:checked:bg-primary-500"))
	add("Radio", "dot on fill", "form.go:154", nonText, on("bg-white", "checked:bg-primary-600"), on("bg-white", "dark:checked:bg-primary-500"))
	add("Checkbox/Radio", "focus-visible outline", "form.go:155", nonText, on("focus-visible:outline-primary-600", "").over(cardL), on("dark:focus-visible:outline-primary-500", "").over(cardD))
	disabled("Checkbox/Radio", "disabled label", "form.go:147", on("has-disabled:text-neutral-500", "").over(cardL), on("dark:has-disabled:text-neutral-400", "").over(cardD))

	// Switch (form.go Switch): off track ring and knob ring, on track, knob on the on track. The
	// track ring is field-border (3:1 against the card); the knob ring sits on the darker
	// neutral-200 track, so it takes field-border-hover (field-border would be 2.5:1 there). In
	// dark the knob is a white disc on the white/10 track.
	add("Switch", "off track ring vs card", "form.go:238", nonText, on("ring-field-border", "").over(cardL), on("ring-field-border", "").over(cardD))
	add("Switch", "knob vs off track", "form.go:235", nonText, on("ring-field-border-hover", "bg-neutral-200").over(cardL), on("bg-white", "dark:bg-white/10").over(cardD))
	add("Switch", "on track vs card", "form.go:239", nonText, on("has-checked:bg-primary-600", "").over(cardL), on("dark:has-checked:bg-primary-500", "").over(cardD))
	add("Switch", "knob vs on track", "form.go:239", nonText, on("bg-white", "has-checked:bg-primary-600"), on("bg-white", "dark:has-checked:bg-primary-500"))
	add("Switch", "knob vs on track, hover", "form.go:240", nonText, on("bg-white", "has-checked:has-enabled:hover:bg-primary-700"), on("bg-white", "dark:has-checked:has-enabled:hover:bg-primary-400"))
	add("Switch", "focus-visible outline", "form.go:241", nonText, on("has-focus-visible:outline-primary-600", "").over(cardL), on("dark:has-focus-visible:outline-primary-500", "").over(cardD))

	// Select picker (assets/ui.css, appearance: base-select): options on the picker, the hover
	// and keyboard highlight, and the check mark of the selected option.
	add("Select picker", "option", "assets/ui.css:280", text, on("text-neutral-900", "bg-white"), on("dark:text-white", "dark:bg-neutral-800"))
	add("Select picker", "option hover / focus", "assets/ui.css:296", text, on("text-white", "bg-primary-600"), on("text-white", "bg-primary-600"))
	add("Select picker", "check mark", "assets/ui.css:291", nonText, on("text-primary-600", "bg-white"), on("dark:text-primary-400", "dark:bg-neutral-800"))
	disabled("Select picker", "disabled option", "assets/ui.css:302", on("text-neutral-500", "bg-white"), on("dark:text-neutral-400", "dark:bg-neutral-800"))

	// Field (form.go:119-127) on the page background.
	add("Field", "label", "form.go:119", text, on("text-neutral-900", ""), on("dark:text-white", ""))
	add("Field", "help text", "form.go:125", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Field", "error text", "form.go:127", text, on("text-danger-600", ""), on("dark:text-danger-400", ""))

	// Table (table.go:94-99, 55, 66): header on neutral-50 / neutral-900, body on the white card.
	add("Table", "header text", "table.go:99", text, on("text-neutral-600", "bg-neutral-50"), on("dark:text-neutral-300", "dark:bg-neutral-900"))
	add("Table", "body text", "table.go:94", text, on("text-neutral-700", "bg-white"), on("dark:text-neutral-300", "dark:bg-neutral-900"))
	add("Table", "striped body text", "table.go:94", text,
		on("text-neutral-700", "*:even:bg-neutral-50").over(cardL).with("*:even:bg-neutral-50"),
		on("dark:text-neutral-300", "dark:*:even:bg-white/5").over(cardD).with("dark:*:even:bg-white/5"))
	add("Table", "sort button focus outline", "table.go:133", nonText, on("focus-visible:outline-primary-600", "").over(cardL), on("dark:focus-visible:outline-primary-500", "").over(cardD))
	add("Table", "empty-state description", "table.go:155", text, on("text-neutral-500", "").over(cardL), on("dark:text-neutral-400", "").over(cardD))

	// Stat (info.go:51-66) on its white / neutral-900 card. The value is text-2xl semibold.
	add("Stat", "value", "info.go:51", largeText, on("text-neutral-900", "").over(cardL), on("dark:text-white", "").over(cardD))
	add("Stat", "label", "info.go:64", text, on("text-neutral-500", "").over(cardL), on("dark:text-neutral-400", "").over(cardD))
	add("Stat", "delta neutral", "info.go:54", text, on("text-neutral-500", "").over(cardL), on("dark:text-neutral-400", "").over(cardD))
	add("Stat", "delta positive", "info.go:57", text, on("text-success-700", "").over(cardL), on("dark:text-success-400", "").over(cardD))
	add("Stat", "delta negative", "info.go:59", text, on("text-danger-600", "").over(cardL), on("dark:text-danger-400", "").over(cardD))
	add("DescriptionList", "term", "info.go:96", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))

	// Progress (progress.go:32-37): fill against its track; Steps (progress.go:146-163) labels.
	for _, p := range []struct{ name, line, l, d string }{
		{"Neutral", "32", "bg-primary-600", "dark:bg-primary-500"},
		{"Brand", "33", "bg-primary-600", "dark:bg-primary-500"},
		{"Info", "34", "bg-info-600", "dark:bg-info-500"},
		{"Success", "35", "bg-success-700", "dark:bg-success-500"},
		{"Warning", "36", "bg-warning-700", "dark:bg-warning-500"},
		{"Danger", "37", "bg-danger-600", "dark:bg-danger-500"},
	} {
		add("Progress/"+p.name, "fill on track", "progress.go:"+p.line, nonText, on(p.l, "bg-neutral-200"), on(p.d, "dark:bg-white/10"))
	}
	add("Steps", "completed number", "progress.go:146", text, on("text-white", "bg-primary-600"), on("text-white", "bg-primary-600"))
	add("Steps", "current number", "progress.go:150", text, on("text-primary-600", ""), on("dark:text-primary-400", ""))
	add("Steps", "upcoming number", "progress.go:157", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Steps", "upcoming label", "progress.go:157", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Steps", "current label", "progress.go:159", text, on("text-neutral-900", ""), on("dark:text-white", ""))

	// Tooltip (info.go:36-40): opaque bubble.
	add("Tooltip", "text", "info.go:37", text, on("text-white", "bg-neutral-900"), on("text-white", "dark:bg-neutral-700"))

	// Nav (nav.go:34-39, 140), Breadcrumb (nav.go:188-195), Tabs, Pagination.
	add("Nav link", "inactive", "nav.go:35", text, on("text-neutral-600", "").over(cardL), on("dark:text-neutral-400", "").over(cardD))
	add("Nav link", "active (aria-current)", "nav.go:37", text, on("aria-[current=page]:text-primary-700", "aria-[current=page]:bg-neutral-100").over(cardL), on("dark:aria-[current=page]:text-white", "dark:aria-[current=page]:bg-white/10").over(cardD))
	add("Nav link", "active on the page background", "nav.go:37", text, on("aria-[current=page]:text-primary-700", "aria-[current=page]:bg-neutral-100"), on("dark:aria-[current=page]:text-white", "dark:aria-[current=page]:bg-white/10"))
	add("Nav link", "focus-visible outline", "nav.go:36", nonText, on("focus-visible:outline-primary-600", "").over(cardL), on("dark:focus-visible:outline-primary-500", "").over(cardD))
	add("Sidebar group", "heading", "nav.go:140", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Breadcrumb", "current", "nav.go:188", text, on("text-neutral-900", ""), on("dark:text-white", ""))
	add("Breadcrumb", "link", "nav.go:195", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Tabs", "active", "tabs.go:23", text, on("text-primary-600", ""), on("dark:text-primary-400", ""))
	add("Tabs", "inactive", "tabs.go:24", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Pagination", "item", "pagination.go:12", text, on("text-neutral-700", ""), on("dark:text-neutral-300", ""))
	add("Pagination", "current page", "pagination.go:13", text, on("aria-[current=page]:text-white", "aria-[current=page]:bg-primary-600"), on("dark:aria-[current=page]:text-white", "dark:aria-[current=page]:bg-primary-600"))
	add("Pagination", "current page hover", "pagination.go:14", text, on("aria-[current=page]:text-white", "aria-[current=page]:hover:bg-primary-700"), on("dark:aria-[current=page]:text-white", "dark:aria-[current=page]:hover:bg-primary-500"))
	add("Pagination", "gap", "pagination.go:18", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))

	// Kbd (badge.go:25-26), Avatar initials (avatar.go:31-32), Divider label (layout.go:44).
	add("Kbd", "text", "badge.go:26", text, on("text-neutral-600", "bg-neutral-50"), on("dark:text-neutral-300", "dark:bg-white/5"))
	add("Avatar", "initials", "avatar.go:31", text, on("text-white", "bg-neutral-600"), on("text-white", "dark:bg-neutral-500"))
	add("Divider", "label", "layout.go:44", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))

	// DropZone highlight (dnd.go:78) and MoveHere (dnd.go:116-117).
	add("DropZone", "data-over outline", "dnd.go:78", nonText, on("data-over:outline-primary-600", ""), on("data-over:outline-primary-600", ""))
	add("MoveHere", "dashed border", "dnd.go:116", nonText, on("border-primary-600", ""), on("border-primary-600", ""))
	add("MoveHere", "text", "dnd.go:116", text, on("text-primary-600", ""), on("dark:text-primary-400", ""))
	add("MoveHere", "text on hover", "dnd.go:117", text, on("hover:text-primary-700", "hover:bg-primary-500/10"), on("dark:hover:text-primary-400", "hover:bg-primary-500/10"))

	// Dialog (dialog.go:54, dialogHeader at dialog.go:98) and Drawer on their white / neutral-900 panels.
	// The X button of Dialog and Drawer is a Ghost IconButton (ghostClose, signals.go): no ring.
	add("Dialog", "close icon", "button.go:20", nonText, on("text-neutral-700", "bg-white"), on("dark:text-neutral-300", "dark:bg-neutral-900"))
	add("Dialog", "close icon, hover", "button.go:20", nonText, on("text-neutral-700", "hover:bg-neutral-100").over(cardL), on("dark:text-neutral-300", "dark:hover:bg-white/10").over(cardD))
	add("Dialog", "body text", "dialog.go:54", text, on("text-neutral-600", "bg-white"), on("dark:text-neutral-300", "dark:bg-neutral-900"))
	add("Drawer", "body text", "drawer.go:96", text, on("text-neutral-600", "bg-white"), on("dark:text-neutral-300", "dark:bg-neutral-900"))

	// Other focus rings that sit on the page background.
	add("Disclosure", "focus-visible outline", "disclosure.go:68", nonText, on("focus-visible:outline-primary-600", ""), on("dark:focus-visible:outline-primary-500", ""))

	// Code blocks (assets/ui.css, --color-code-*): see contrast_code_test.go.
	m = append(m, codeRows()...)
	// Prose (assets/ui.css, --ui-prose-*): see prose_test.go.
	m = append(m, proseRows()...)
	return m
}

func indexOf(toneName string) int {
	for i, t := range tones {
		if t.name == toneName {
			return i
		}
	}
	return -1
}

// result is a measured row.
type result struct {
	spec         spec
	scheme       string
	s            side
	fgHex, bgHex string
	ratio, min   float64
	fix          string
}

func (r result) ok() bool { return r.ratio+1e-9 >= r.min }

func (r result) id() string { return r.spec.component + "/" + r.spec.state + "/" + r.scheme }

func (r result) String() string {
	bg := r.s.bg
	if bg == "" {
		bg = "(backdrop " + r.s.backdrop(r.scheme) + ")"
	}
	return fmt.Sprintf("%s: fg %s=%s on bg %s=%s → %.2f:1 (need %.1f:1)", r.id(), r.s.fg, r.fgHex, bg, r.bgHex, r.ratio, r.min)
}

func (s side) backdrop(scheme string) string {
	if s.under != "" {
		return s.under
	}
	if scheme == "dark" {
		return pageD
	}
	return pageL
}

// measure resolves a side to the effective foreground and background colors, as painted:
// translucent backgrounds are composited over the backdrop, then the foreground over that, and
// element opacity (group) fades both toward the backdrop.
func measure(th *wcag.Theme, scheme string, s side) (fg, bg wcag.RGBA, err error) {
	under, err := th.ClassColor(s.backdrop(scheme))
	if err != nil {
		return fg, bg, err
	}
	if under.A != 1 {
		return fg, bg, fmt.Errorf("backdrop %s is translucent", s.backdrop(scheme))
	}
	bg = under
	if s.bg != "" {
		c, err := th.ClassColor(s.bg)
		if err != nil {
			return fg, bg, err
		}
		bg = c.Over(under)
	}
	f, err := th.ClassColor(s.fg)
	if err != nil {
		return fg, bg, err
	}
	fg = f.Over(bg)
	if s.group != 0 {
		fg = fg.WithAlpha(s.group).Over(under)
		bg = bg.WithAlpha(s.group).Over(under)
	}
	return fg, bg, nil
}

var classParts = regexp.MustCompile(`^((?:[^:]+:)*)(bg|text|outline|ring|border)-([a-z]+)-(\d+)(/\d+)?$`)

// suggest finds the nearest step of the foreground family (then the background family) that
// passes, e.g. "text-neutral-600" for a failing text-neutral-500.
func suggest(th *wcag.Theme, scheme string, s side, k kind) string {
	steps := []int{50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950}
	try := func(class string, apply func(string) side) string {
		m := classParts.FindStringSubmatch(class)
		if m == nil {
			return ""
		}
		cur, _ := strconv.Atoi(m[4])
		sorted := append([]int(nil), steps...)
		sort.SliceStable(sorted, func(i, j int) bool { return abs(sorted[i]-cur) < abs(sorted[j]-cur) })
		for _, st := range sorted {
			if st == cur {
				continue
			}
			cand := m[1] + m[2] + "-" + m[3] + "-" + strconv.Itoa(st) + m[5]
			fg, bg, err := measure(th, scheme, apply(cand))
			if err == nil && wcag.Contrast(fg, bg) >= k.min() {
				return fmt.Sprintf("%s (%.1f:1)", cand, wcag.Contrast(fg, bg))
			}
		}
		return ""
	}
	if out := try(s.fg, func(c string) side { n := s; n.fg = c; return n }); out != "" {
		return out
	}
	if s.bg != "" {
		return try(s.bg, func(c string) side { n := s; n.bg = c; return n })
	}
	return ""
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func measureAll(t *testing.T) []result {
	t.Helper()
	themes := map[string]*wcag.Theme{
		"light": wcag.NewTheme(ui.ThemeCSS()),
		"dark":  wcag.NewDarkTheme(ui.ThemeCSS()), // tokens redefined under .dark (field-border)
	}
	var out []result
	for _, sp := range matrix() {
		for _, sc := range []struct {
			name string
			s    side
		}{{"light", sp.light}, {"dark", sp.dark}} {
			th := themes[sc.name]
			fg, bg, err := measure(th, sc.name, sc.s)
			if err != nil {
				t.Fatalf("%s/%s/%s: %v", sp.component, sp.state, sc.name, err)
			}
			r := result{spec: sp, scheme: sc.name, s: sc.s, fgHex: fg.Hex(), bgHex: bg.Hex(), ratio: wcag.Contrast(fg, bg), min: sp.kind.min()}
			if !r.ok() {
				r.fix = suggest(th, sc.name, sc.s, sp.kind)
			}
			out = append(out, r)
		}
	}
	return out
}

// TestContrast fails for every pair below its WCAG minimum. Disabled controls are exempt from
// WCAG 1.4.3 and only logged.
func TestContrast(t *testing.T) {
	results := measureAll(t)
	if p := os.Getenv("CONTRAST_REPORT"); p != "" {
		if err := os.WriteFile(p, []byte(contrastReport(results)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	failed := 0
	for _, r := range results {
		switch {
		case r.ok():
		case r.spec.exempt:
			t.Logf("exempt (disabled): %s", r)
		default:
			failed++
			msg := r.String()
			if r.fix != "" {
				msg += "; try " + r.fix
			}
			t.Error(msg)
		}
	}
	t.Logf("%d rows (%d pairs x 2 schemes), %d failing", len(results), len(results)/2, failed)
}

func contrastReport(results []result) string {
	var b strings.Builder
	pass, fail, exempt := 0, 0, 0
	for _, r := range results {
		switch {
		case r.ok():
			pass++
		case r.spec.exempt:
			exempt++
		default:
			fail++
		}
	}
	fmt.Fprintf(&b, "# Contrast report\n\nTailwind v4.1 palette resolved through ThemeCSS(). %d rows: %d pass, %d fail, %d disabled-exempt below minimum.\n\n", len(results), pass, fail, exempt)
	b.WriteString("| Status | Component / state | Scheme | Kind | Foreground | Background | Ratio | Need | Suggested fix | Source |\n|---|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range results {
		status := "pass"
		switch {
		case r.ok():
		case r.spec.exempt:
			status = "exempt"
		default:
			status = "**FAIL**"
		}
		bg := r.s.bg
		if bg == "" {
			bg = "(" + r.s.backdrop(r.scheme) + ")"
		}
		fmt.Fprintf(&b, "| %s | %s / %s | %s | %s | `%s` %s | `%s` %s | %.2f | %.1f | %s | %s |\n",
			status, r.spec.component, r.spec.state, r.scheme, r.spec.kind, r.s.fg, r.fgHex, bg, r.bgHex, r.ratio, r.min, r.fix, r.spec.src)
	}
	return b.String()
}

// hasClass reports whether the Go source line contains c as a whole class token.
func hasClass(line, c string) bool {
	for _, f := range strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '"' || r == '`' }) {
		if f == c {
			return true
		}
	}
	return false
}

var srcRef = regexp.MustCompile(`^([a-z_/]+\.(?:go|css)):(\d+)`)

// cssHasColor reports whether a CSS source line references the color of utility c, for rows read
// from assets/ui.css: bg-neutral-800 matches var(--color-neutral-800).
func cssHasColor(line, c string) bool {
	m := classParts.FindStringSubmatch(c)
	if m == nil {
		base := c[strings.LastIndex(c, ":")+1:]
		_, name, ok := strings.Cut(base, "-")
		return ok && strings.Contains(line, "--color-"+strings.SplitN(name, "/", 2)[0]+")")
	}
	return strings.Contains(line, "--color-"+m[3]+"-"+m[4]+")")
}

// TestContrastMatrixInSync checks that each row's source line still contains one of its classes,
// and that every class of the matrix is still rendered by the goldens (or BodyClasses), so the
// matrix cannot rot when a component changes.
func TestContrastMatrixInSync(t *testing.T) {
	rendered := map[string]bool{}
	for _, classes := range goldenClasses(t) {
		for _, c := range classes {
			rendered[c] = true
		}
	}
	for _, c := range strings.Fields(ui.BodyClasses()) {
		rendered[c] = true
	}
	lines := map[string][]string{}
	for _, sp := range matrix() {
		// src is one or more comma-separated file:line references; a class may be on any of them.
		var cited []string
		bad := false
		for _, ref := range strings.Split(sp.src, ",") {
			m := srcRef.FindStringSubmatch(ref)
			if m == nil {
				t.Errorf("%s/%s: bad src %q", sp.component, sp.state, ref)
				bad = true
				break
			}
			if lines[m[1]] == nil {
				b, err := os.ReadFile(m[1])
				if err != nil {
					t.Fatal(err)
				}
				lines[m[1]] = strings.Split(string(b), "\n")
			}
			n, _ := strconv.Atoi(m[2])
			if n < 1 || n > len(lines[m[1]]) {
				t.Errorf("%s/%s: %s is out of range", sp.component, sp.state, ref)
				bad = true
				break
			}
			cited = append(cited, m[1]+"\x00"+lines[m[1]][n-1])
		}
		if bad {
			continue
		}
		found := false
		var all []string
		for _, s := range []side{sp.light, sp.dark} {
			for _, c := range append([]string{s.fg, s.bg, s.under}, s.also...) {
				if c == "" {
					continue
				}
				all = append(all, c)
				inLine := false
				for _, cl := range cited {
					file, line, _ := strings.Cut(cl, "\x00")
					inLine = inLine || hasClass(line, c) || strings.HasSuffix(file, ".css") && cssHasColor(line, c)
				}
				// Tone-keyed maps render one tone per golden; those classes are covered by the
				// source line check alone.
				if !rendered[c] && !inLine {
					t.Errorf("%s/%s: class %q is in no golden and not in %s", sp.component, sp.state, c, sp.src)
				}
				found = found || inLine
			}
		}
		if !found {
			t.Errorf("%s/%s: %s no longer contains any of %v", sp.component, sp.state, sp.src, all)
		}
	}
}

// TestContrastTokensResolve checks that the theme aliases the families the matrix relies on.
func TestContrastTokensResolve(t *testing.T) {
	th := wcag.NewTheme(ui.ThemeCSS())
	for tok, want := range map[string]string{
		"neutral-500": "zinc-500", "field-border-hover": "zinc-500", "success-600": "green-600",
		"warning-500": "amber-500", "primary-400": "indigo-400", "primary-500": "indigo-500", "primary-600": "indigo-600", "primary-700": "indigo-700", "danger-500": "red-500", "info-600": "sky-600", "info-700": "sky-700",
	} {
		if got := th.Alias(tok); got != want {
			t.Errorf("%s resolves to %s, want %s", tok, got, want)
		}
	}
}

// TestPrimaryAliasesIndigo checks that every step of the default primary scale is a plain alias
// of the stock Tailwind indigo step (the contrast matrix is measured against it).
func TestPrimaryAliasesIndigo(t *testing.T) {
	th := wcag.NewTheme(ui.ThemeCSS())
	for _, st := range []string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900", "950"} {
		if got, want := th.Alias("primary-"+st), "indigo-"+st; got != want {
			t.Errorf("primary-%s resolves to %q, want %q", st, got, want)
		}
	}
}

// TestFieldBorderToken checks the field outline token against its stated bounds: in light it is
// the lightest shade of its hue and chroma (steps of 0.5 L) that is 3.05:1 on both white and the
// page background, so one step lighter would fall below the 3:1 of WCAG 1.4.11; the hover color
// is darker; in dark both are translucent whites over the field, the card and the page.
func TestFieldBorderToken(t *testing.T) {
	m := regexp.MustCompile(`--color-field-border:\s*oklch\(\s*([\d.]+)%\s+([\d.]+)\s+([\d.]+)\s*\)`).FindStringSubmatch(ui.ThemeCSS())
	if m == nil {
		t.Fatal("--color-field-border is not an oklch() value in ThemeCSS")
	}
	f := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
	lch := wcag.OKLCH{L: f(m[1]), C: f(m[2]), H: f(m[3])}
	th := wcag.NewTheme(ui.ThemeCSS())
	white, _ := th.Color("white")
	page, _ := th.Color("neutral-50")
	const want = 3.05
	at := func(c wcag.OKLCH) float64 { return min(wcag.Contrast(c.RGB(), white), wcag.Contrast(c.RGB(), page)) }
	if got := at(lch); got < want {
		t.Errorf("field-border is %.2f:1 on the lighter of white and neutral-50, want %.2f:1", got, want)
	}
	lighter := lch
	lighter.L += 0.5
	if got := at(lighter); got >= want {
		t.Errorf("field-border is not the lightest: L+0.5 is still %.2f:1; lighten the token", got)
	}
	hover, _ := th.ClassColor("outline-field-border-hover")
	border, _ := th.ClassColor("outline-field-border")
	if wcag.Contrast(hover, white) <= wcag.Contrast(border, white) {
		t.Error("field-border-hover is not darker than field-border in light")
	}
	dark := wcag.NewDarkTheme(ui.ThemeCSS())
	db, _ := dark.ClassColor("outline-field-border")
	dh, _ := dark.ClassColor("outline-field-border-hover")
	if db.A >= dh.A || db.A <= 0 || dh.A >= 1 {
		t.Errorf("dark alphas: border %.2f, hover %.2f; want 0 < border < hover < 1", db.A, dh.A)
	}
}
