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
// ui.ThemeCSS() (primary-600 -> sky-600 -> OKLCH -> sRGB), so the matrix follows the defaults.
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

// tones is toneSoft (badge.go:7-11) and toneAccent (toast.go:41-47): accent borders are one step
// darker in light (white toast) than in dark (neutral-800 toast).
var tones = []tone{
	{"Neutral", "7", "text-neutral-700", "bg-neutral-100", "dark:text-neutral-300", "dark:bg-neutral-400/10", "border-neutral-500", "border-neutral-500"},
	{"Info", "8", "text-info-700", "bg-info-50", "dark:text-info-300", "dark:bg-info-400/10", "border-info-600", "dark:border-info-500"},
	{"Success", "9", "text-success-700", "bg-success-50", "dark:text-success-300", "dark:bg-success-500/10", "border-success-600", "dark:border-success-500"},
	{"Warning", "10", "text-warning-800", "bg-warning-50", "dark:text-warning-300", "dark:bg-warning-400/10", "border-warning-600", "dark:border-warning-500"},
	{"Danger", "11", "text-danger-700", "bg-danger-50", "dark:text-danger-300", "dark:bg-danger-400/10", "border-danger-600", "dark:border-danger-500"},
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
	// Solid surfaces with white text use the 700 step (hover 800) in both schemes: light hues such
	// as sky only reach 4.5:1 against white there.
	add("Button/Primary", "default", "button.go:14", text, on("text-white", "bg-primary-700"), on("text-white", "bg-primary-700"))
	add("Button/Primary", "hover", "button.go:14", text, on("text-white", "hover:bg-primary-800"), on("text-white", "hover:bg-primary-800"))
	add("Button/Primary", "focus-visible outline", "button.go:14", nonText, on("focus-visible:outline-primary-700", ""), on("dark:focus-visible:outline-primary-500", ""))
	add("Button/Primary", "focus-visible outline on a card", "button.go:14", nonText, on("focus-visible:outline-primary-700", "").over(cardL), on("dark:focus-visible:outline-primary-500", "").over("dark:bg-neutral-800"))
	disabled("Button/Primary", "disabled", "button.go:14", on("text-white", "bg-primary-700").grp(0.5).with(dis), on("text-white", "bg-primary-700").grp(0.5).with(dis))

	add("Button/Secondary", "default", "button.go:16", text, on("text-neutral-900", "bg-white"), on("dark:text-white", "dark:bg-white/10"))
	add("Button/Secondary", "hover", "button.go:16", text, on("text-neutral-900", "hover:bg-neutral-50"), on("dark:text-white", "dark:hover:bg-white/20"))
	add("Button/Secondary", "focus-visible outline", "button.go:16", nonText, on("focus-visible:outline-neutral-500", ""), on("focus-visible:outline-neutral-500", ""))
	disabled("Button/Secondary", "disabled", "button.go:16", on("text-neutral-900", "bg-white").grp(0.5).with(dis), on("dark:text-white", "dark:bg-white/10").grp(0.5).with(dis))

	add("Button/Outline", "default", "button.go:18", text, on("text-primary-700", ""), on("dark:text-primary-400", ""))
	add("Button/Outline", "hover", "button.go:18", text, on("text-primary-700", "hover:bg-primary-50"), on("dark:text-primary-400", "dark:hover:bg-primary-500/10"))
	add("Button/Outline", "ring (boundary)", "button.go:18", nonText, on("ring-primary-700", ""), on("dark:ring-primary-500", ""))
	add("Button/Outline", "focus-visible outline", "button.go:18", nonText, on("focus-visible:outline-primary-700", ""), on("dark:focus-visible:outline-primary-500", ""))
	disabled("Button/Outline", "disabled", "button.go:18", on("text-primary-700", "").grp(0.5).with(dis), on("dark:text-primary-400", "").grp(0.5).with(dis))

	add("Button/Ghost", "default", "button.go:20", text, on("text-neutral-700", ""), on("dark:text-neutral-300", ""))
	add("Button/Ghost", "hover", "button.go:20", text, on("text-neutral-700", "hover:bg-neutral-100"), on("dark:text-neutral-300", "dark:hover:bg-white/10"))
	add("Button/Ghost", "focus-visible outline", "button.go:20", nonText, on("focus-visible:outline-neutral-500", ""), on("focus-visible:outline-neutral-500", ""))
	disabled("Button/Ghost", "disabled", "button.go:20", on("text-neutral-700", "").grp(0.5).with(dis), on("dark:text-neutral-300", "").grp(0.5).with(dis))

	add("Button/Danger", "default", "button.go:22", text, on("text-white", "bg-danger-600"), on("text-white", "bg-danger-600"))
	add("Button/Danger", "hover", "button.go:22", text, on("text-white", "hover:bg-danger-700"), on("text-white", "hover:bg-danger-700"))
	add("Button/Danger", "focus-visible outline", "button.go:22", nonText, on("focus-visible:outline-danger-600", ""), on("dark:focus-visible:outline-danger-500", ""))
	disabled("Button/Danger", "disabled", "button.go:22", on("text-white", "bg-danger-600").grp(0.5).with(dis), on("text-white", "bg-danger-600").grp(0.5).with(dis))

	add("Button/Link", "default", "button.go:24", text, on("text-primary-700", ""), on("dark:text-primary-400", ""))
	add("Button/Link", "hover", "button.go:24", text, on("hover:text-primary-800", ""), on("dark:hover:text-primary-300", ""))
	add("Button/Link", "focus-visible outline", "button.go:24", nonText, on("focus-visible:outline-primary-700", ""), on("dark:focus-visible:outline-primary-500", ""))
	disabled("Button/Link", "disabled", "button.go:24", on("text-primary-700", "").grp(0.5).with(dis), on("dark:text-primary-400", "").grp(0.5).with(dis))

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

	// Toast (toast.go:62-73). The close button is closeButton (signals.go:45), at full opacity.
	add("Toast", "title", "toast.go:62", text, on("text-neutral-900", "bg-white"), on("dark:text-white", "dark:bg-neutral-800"))
	add("Toast", "message", "toast.go:65", text, on("text-neutral-600", "bg-white"), on("dark:text-neutral-300", "dark:bg-neutral-800"))
	add("Toast", "close icon", "toast.go:69", nonText, on("text-neutral-500", "bg-white"), on("dark:text-neutral-400", "dark:bg-neutral-800"))

	// Input (form.go:25-30). Dark inputs are bg-white/5 over the page.
	add("Input", "text", "form.go:25", text, on("text-neutral-900", "bg-white"), on("dark:text-white", "dark:bg-white/5"))
	add("Input", "placeholder", "form.go:26", text, on("placeholder:text-neutral-500", "bg-white"), on("dark:placeholder:text-neutral-400", "dark:bg-white/5"))
	add("Input", "border (SC 1.4.11)", "form.go:25", nonText, on("outline-neutral-500", ""), on("dark:outline-white/35", ""))
	add("Input", "border against the field (SC 1.4.11)", "form.go:25", nonText, on("outline-neutral-500", "bg-white"), on("dark:outline-white/35", "dark:bg-white/5"))
	add("InputGroup", "affix border", "info.go:104", nonText, on("outline-neutral-500", "bg-neutral-50"), on("dark:outline-white/35", "dark:bg-white/5"))
	add("Input", "focus outline", "form.go:26", nonText, on("focus:outline-primary-700", ""), on("dark:focus:outline-primary-500", ""))
	add("Input", "invalid outline", "form.go:28", nonText, on("aria-invalid:outline-danger-500", ""), on("dark:aria-invalid:outline-danger-400", ""))
	disabled("Input", "disabled text", "form.go:27", on("disabled:text-neutral-500", "disabled:bg-neutral-50"), on("dark:disabled:text-neutral-400", "dark:disabled:bg-white/10"))

	// Field (form.go:95-103) on the page background.
	add("Field", "label", "form.go:95", text, on("text-neutral-900", ""), on("dark:text-white", ""))
	add("Field", "help text", "form.go:101", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Field", "error text", "form.go:103", text, on("text-danger-600", ""), on("dark:text-danger-400", ""))

	// Table (table.go:94-99, 55, 66): header on neutral-50 / neutral-900, body on the white card.
	add("Table", "header text", "table.go:99", text, on("text-neutral-600", "bg-neutral-50"), on("dark:text-neutral-300", "dark:bg-neutral-900"))
	add("Table", "body text", "table.go:94", text, on("text-neutral-700", "bg-white"), on("dark:text-neutral-300", "dark:bg-neutral-900"))
	add("Table", "striped body text", "table.go:94", text,
		on("text-neutral-700", "*:even:bg-neutral-50").over(cardL).with("*:even:bg-neutral-50"),
		on("dark:text-neutral-300", "dark:*:even:bg-white/5").over(cardD).with("dark:*:even:bg-white/5"))
	add("Table", "sort button focus outline", "table.go:133", nonText, on("focus-visible:outline-primary-700", "").over(cardL), on("dark:focus-visible:outline-primary-500", "").over(cardD))
	add("Table", "empty-state description", "table.go:155", text, on("text-neutral-500", "").over(cardL), on("dark:text-neutral-400", "").over(cardD))

	// Stat (info.go:50-65) on its white / neutral-900 card. The value is text-2xl semibold.
	add("Stat", "value", "info.go:50", largeText, on("text-neutral-900", "").over(cardL), on("dark:text-white", "").over(cardD))
	add("Stat", "label", "info.go:63", text, on("text-neutral-500", "").over(cardL), on("dark:text-neutral-400", "").over(cardD))
	add("Stat", "delta neutral", "info.go:53", text, on("text-neutral-500", "").over(cardL), on("dark:text-neutral-400", "").over(cardD))
	add("Stat", "delta positive", "info.go:56", text, on("text-success-700", "").over(cardL), on("dark:text-success-400", "").over(cardD))
	add("Stat", "delta negative", "info.go:58", text, on("text-danger-600", "").over(cardL), on("dark:text-danger-400", "").over(cardD))
	add("DescriptionList", "term", "info.go:95", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))

	// Progress (progress.go:32-36, 51): fill against its track; Steps (progress.go:142-159) labels.
	for _, p := range []struct{ name, line, l, d string }{
		{"Neutral", "32", "bg-primary-600", "dark:bg-primary-500"},
		{"Info", "33", "bg-info-600", "dark:bg-info-500"},
		{"Success", "34", "bg-success-700", "dark:bg-success-500"},
		{"Warning", "35", "bg-warning-700", "dark:bg-warning-500"},
		{"Danger", "36", "bg-danger-600", "dark:bg-danger-500"},
	} {
		add("Progress/"+p.name, "fill on track", "progress.go:"+p.line, nonText, on(p.l, "bg-neutral-200"), on(p.d, "dark:bg-white/10"))
	}
	add("Steps", "completed number", "progress.go:142", text, on("text-white", "bg-primary-700"), on("text-white", "bg-primary-700"))
	add("Steps", "current number", "progress.go:146", text, on("text-primary-700", ""), on("dark:text-primary-400", ""))
	add("Steps", "upcoming number", "progress.go:150", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Steps", "upcoming label", "progress.go:153", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Steps", "current label", "progress.go:155", text, on("text-neutral-900", ""), on("dark:text-white", ""))

	// Tooltip (info.go:37-39): opaque bubble.
	add("Tooltip", "text", "info.go:38", text, on("text-white", "bg-neutral-900"), on("text-white", "dark:bg-neutral-700"))

	// Nav (nav.go:22-24, 131), Breadcrumb (nav.go:179-187), Tabs, Pagination.
	add("Nav link", "inactive", "nav.go:22", text, on("text-neutral-600", "").over(cardL), on("dark:text-neutral-400", "").over(cardD))
	add("Nav link", "active", "nav.go:23", text, on("text-neutral-900", "bg-neutral-100").over(cardL), on("dark:text-white", "dark:bg-white/10").over(cardD))
	add("Nav link", "focus-visible outline", "nav.go:24", nonText, on("focus-visible:outline-primary-700", "").over(cardL), on("dark:focus-visible:outline-primary-500", "").over(cardD))
	add("Sidebar group", "heading", "nav.go:131", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Breadcrumb", "current", "nav.go:179", text, on("text-neutral-900", ""), on("dark:text-white", ""))
	add("Breadcrumb", "link", "nav.go:186", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Tabs", "active", "tabs.go:23", text, on("text-primary-700", ""), on("dark:text-primary-400", ""))
	add("Tabs", "inactive", "tabs.go:24", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))
	add("Pagination", "item", "pagination.go:12", text, on("text-neutral-700", ""), on("dark:text-neutral-300", ""))
	add("Pagination", "current page", "pagination.go:13", text, on("aria-[current=page]:text-white", "aria-[current=page]:bg-primary-700"), on("dark:aria-[current=page]:text-white", "dark:aria-[current=page]:bg-primary-700"))
	add("Pagination", "gap", "pagination.go:18", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))

	// Kbd (badge.go:25-26), Avatar initials (avatar.go:31-32), Divider label (layout.go:44).
	add("Kbd", "text", "badge.go:26", text, on("text-neutral-600", "bg-neutral-50"), on("dark:text-neutral-300", "dark:bg-white/5"))
	add("Avatar", "initials", "avatar.go:31", text, on("text-white", "bg-neutral-600"), on("text-white", "dark:bg-neutral-500"))
	add("Divider", "label", "layout.go:44", text, on("text-neutral-500", ""), on("dark:text-neutral-400", ""))

	// DropZone highlight (dnd.go:78) and MoveHere (dnd.go:106-107).
	add("DropZone", "data-over outline", "dnd.go:78", nonText, on("data-over:outline-primary-600", ""), on("data-over:outline-primary-600", ""))
	add("MoveHere", "dashed border", "dnd.go:106", nonText, on("border-primary-600", ""), on("border-primary-600", ""))
	add("MoveHere", "text", "dnd.go:106", text, on("text-primary-700", ""), on("dark:text-primary-400", ""))

	// Dialog (dialog.go:55-58) on its white / neutral-900 panel.
	add("Dialog", "close icon", "dialog.go:55", nonText, on("text-neutral-500", "bg-white"), on("dark:text-neutral-400", "dark:bg-neutral-900"))
	add("Dialog", "body text", "dialog.go:57", text, on("text-neutral-600", "bg-white"), on("dark:text-neutral-300", "dark:bg-neutral-900"))

	// Other focus rings that sit on the page background.
	add("Disclosure", "focus-visible outline", "disclosure.go:65", nonText, on("focus-visible:outline-primary-700", ""), on("dark:focus-visible:outline-primary-500", ""))
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
	th := wcag.NewTheme(ui.ThemeCSS())
	var out []result
	for _, sp := range matrix() {
		for _, sc := range []struct {
			name string
			s    side
		}{{"light", sp.light}, {"dark", sp.dark}} {
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

var srcRef = regexp.MustCompile(`^([a-z_]+\.go):(\d+)`)

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
		m := srcRef.FindStringSubmatch(sp.src)
		if m == nil {
			t.Errorf("%s/%s: bad src %q", sp.component, sp.state, sp.src)
			continue
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
			t.Errorf("%s/%s: %s is out of range", sp.component, sp.state, sp.src)
			continue
		}
		line := lines[m[1]][n-1]
		found := false
		var all []string
		for _, s := range []side{sp.light, sp.dark} {
			for _, c := range append([]string{s.fg, s.bg, s.under}, s.also...) {
				if c == "" {
					continue
				}
				all = append(all, c)
				inLine := hasClass(line, c)
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
		"primary-600": "sky-600", "neutral-500": "zinc-500", "success-600": "green-600",
		"warning-500": "amber-500", "danger-500": "red-500", "info-600": "blue-600",
	} {
		if got := th.Alias(tok); got != want {
			t.Errorf("%s resolves to %s, want %s", tok, got, want)
		}
	}
}
