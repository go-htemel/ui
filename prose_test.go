package ui_test

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

var (
	_ htemel.FlowContent = ui.Prose()
	_ htemel.FlowContent = ui.ProseSm()
	_ htemel.FlowContent = ui.ProseLg()
)

func init() {
	addGolden(
		golden{"prose", ui.Prose(
			html.H2().Text("Release notes"),
			html.P(htemel.Text("Read the "), html.A().Href("/docs").Text("docs"), htemel.Text(" and run "), html.Code().Text("task a11y"), htemel.Text(".")),
			html.Ul(html.Li().Text("One"), html.Li().Text("Two")),
			html.Blockquote(html.P().Text("Quoted.")),
		)},
		golden{"prose-sm", ui.ProseSm(html.P().Text("Small."))},
		golden{"prose-lg", ui.ProseLg(html.P().Text("Large."))},
	)
}

// The prose colors live in assets/ui.css as --ui-prose-* variables on .ui-prose (light), with
// the dark values in the @variant dark block below them. proseVars lists them in stylesheet order
// with the lines of both declarations (control-border is a token that already has a dark value).
type proseVar struct {
	name        string
	line, dline int    // lines of the light and dark declarations; dline 0: the token switches itself
	light, dark string // utility color names: "neutral-700", "white", "white/10"
}

var proseVars = []proseVar{
	{"body", 571, 590, "neutral-700", "neutral-300"},
	{"headings", 572, 591, "neutral-900", "white"},
	{"link", 573, 592, "primary-600", "primary-400"},
	{"link-hover", 574, 593, "primary-700", "primary-300"},
	{"muted", 575, 594, "neutral-600", "neutral-400"},
	{"marker", 576, 595, "neutral-500", "neutral-400"},
	{"code", 577, 596, "neutral-800", "neutral-200"},
	{"code-bg", 578, 597, "neutral-100", "white/10"},
	{"th", 579, 598, "neutral-900", "white"},
	{"th-bg", 580, 599, "neutral-50", "white/5"},
	{"rule", 581, 600, "neutral-200", "white/10"},
	{"quote-border", 582, 0, "control-border", "control-border"},
	{"check", 583, 601, "primary-600", "primary-500"},
}

func proseVarOf(name string) proseVar {
	for _, v := range proseVars {
		if v.name == name {
			return v
		}
	}
	panic("no prose variable " + name)
}

// proseRows measures the prose text colors on the two backgrounds prose sits on: a white /
// neutral-900 card (the showcase, a CMS panel) and the page background.
func proseRows() []spec {
	var m []spec
	src := func(names ...string) string {
		var refs []string
		for _, n := range names {
			v := proseVarOf(n)
			refs = append(refs, "assets/ui.css:"+itoa(v.line))
			if v.dline != 0 {
				refs = append(refs, "assets/ui.css:"+itoa(v.dline))
			}
		}
		return strings.Join(refs, ",")
	}
	fg := func(n string) (string, string) {
		v := proseVarOf(n)
		return "text-" + v.light, "dark:text-" + v.dark
	}
	bg := func(n string) (string, string) {
		v := proseVarOf(n)
		return "bg-" + v.light, "dark:bg-" + v.dark
	}
	for _, t := range []struct{ name, state string }{
		{"body", "body text"}, {"headings", "heading"}, {"link", "link"}, {"link-hover", "link, hover"},
		{"muted", "blockquote / caption / footnote text"}, {"marker", "list marker"},
	} {
		l, d := fg(t.name)
		m = append(m,
			spec{component: "Prose", state: t.state + " on a card", src: src(t.name), kind: text, light: on(l, "").over(cardL), dark: on(d, "").over(cardD)},
			spec{component: "Prose", state: t.state + " on the page", src: src(t.name), kind: text, light: on(l, ""), dark: on(d, "")},
		)
	}
	cl, cd := fg("code")
	bl, bd := bg("code-bg")
	m = append(m,
		spec{component: "Prose", state: "inline code on its chip, on a card", src: src("code", "code-bg"), kind: text, light: on(cl, bl).over(cardL), dark: on(cd, bd).over(cardD)},
		spec{component: "Prose", state: "inline code on its chip, on the page", src: src("code", "code-bg"), kind: text, light: on(cl, bl), dark: on(cd, bd)},
	)
	tl, td := fg("th")
	hl, hd := bg("th-bg")
	m = append(m, spec{component: "Prose", state: "table header on a card", src: src("th", "th-bg"), kind: text, light: on(tl, hl).over(cardL), dark: on(td, hd).over(cardD)})
	ql := "border-" + proseVarOf("quote-border").light
	m = append(m,
		spec{component: "Prose", state: "blockquote bar on a card (SC 1.4.11)", src: src("quote-border"), kind: nonText, light: on(ql, "").over(cardL), dark: on(ql, "").over(cardD)},
		spec{component: "Prose", state: "table header rule on a card (SC 1.4.11)", src: "assets/ui.css:832", kind: nonText, light: on(ql, hl).over(cardL), dark: on(ql, hd).over(cardD)},
	)
	kl, kd := bg("check")
	m = append(m,
		spec{component: "Prose", state: "task checkbox border on a card (SC 1.4.11)", src: src("quote-border"), kind: nonText, light: on(ql, "").over(cardL), dark: on(ql, "").over(cardD)},
		spec{component: "Prose", state: "checked task fill on a card", src: src("check"), kind: nonText, light: on(kl, "").over(cardL), dark: on(kd, "").over(cardD)},
		spec{component: "Prose", state: "check mark on the fill", src: src("check"), kind: nonText, light: on("bg-white", kl), dark: on("bg-white", kd)},
	)
	return m
}

var proseVarDecl = regexp.MustCompile(`^\s*--ui-prose-([a-z-]+):\s*(.+);$`)

// TestProseVarsMatchCSS checks proseVars against assets/ui.css: the first declaration of each
// variable is the light value, the second the dark one, at the lines the matrix cites.
func TestProseVarsMatchCSS(t *testing.T) {
	b, err := os.ReadFile("assets/ui.css")
	if err != nil {
		t.Fatal(err)
	}
	type decl struct {
		line  int
		value string
	}
	seen := map[string][]decl{}
	for i, line := range strings.Split(string(b), "\n") {
		if m := proseVarDecl.FindStringSubmatch(line); m != nil {
			seen[m[1]] = append(seen[m[1]], decl{i + 1, m[2]})
		}
	}
	css := func(c string) string {
		name, alpha, ok := strings.Cut(c, "/")
		if ok {
			return "color-mix(in oklab, var(--color-" + name + ") " + alpha + "%, transparent)"
		}
		return "var(--color-" + c + ")"
	}
	for _, v := range proseVars {
		got := seen[v.name]
		want := []decl{{v.line, css(v.light)}}
		if v.dline != 0 {
			want = append(want, decl{v.dline, css(v.dark)})
		} else if v.light != v.dark {
			t.Errorf("--ui-prose-%s: no dark line but light %s != dark %s", v.name, v.light, v.dark)
		}
		if len(got) != len(want) {
			t.Errorf("--ui-prose-%s: %d declarations, want %d", v.name, len(got), len(want))
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("--ui-prose-%s: css has %v, matrix has %v", v.name, got[i], want[i])
			}
		}
	}
}
