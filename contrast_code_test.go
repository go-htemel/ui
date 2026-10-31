package ui_test

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The code package's syntax colors live in assets/ui.css as --color-code-* variables (light on
// :root, dark under .dark). This file adds their rows to the contrast matrix and checks that the
// classes of each row are the steps the stylesheet really assigns.

// codeToken is one --color-code-* variable: its line in the :root block of assets/ui.css and the
// Tailwind step it uses in light and in dark.
type codeToken struct {
	name        string
	line        int
	light, dark string // "family-step"
}

// codeTokens are the text colors of a code block, in the order of the stylesheet. The dark
// declaration of each is 23 lines below the light one.
var codeTokens = []codeToken{
	{"text", 349, "neutral-800", "neutral-100"},
	{"gutter", 350, "neutral-600", "neutral-400"},
	{"title", 351, "neutral-600", "neutral-400"},
	{"keyword", 352, "primary-700", "primary-300"},
	{"builtin", 353, "primary-600", "primary-400"},
	{"name", 354, "neutral-800", "neutral-200"},
	{"function", 355, "info-700", "info-300"},
	{"type", 356, "info-800", "info-200"},
	{"string", 357, "success-800", "success-300"},
	{"number", 358, "warning-800", "warning-300"},
	{"literal", 359, "warning-800", "warning-300"},
	{"attribute", 360, "info-700", "info-300"},
	{"tag", 361, "danger-700", "danger-300"},
	{"comment", 362, "neutral-600", "neutral-400"},
	{"operator", 363, "neutral-700", "neutral-300"},
	{"punctuation", 364, "neutral-600", "neutral-400"},
	{"error", 365, "danger-700", "danger-300"},
}

const (
	codeDarkOffset = 23 // .dark declarations sit this many lines below the :root ones
	codeSurfaceL   = 345
	codeHighlightL = 347
)

// codeRows measures every token on the three backgrounds a code block puts it on: the surface
// (neutral-50, black/40 over the page in dark), the same surface on a white / neutral-900 card
// (the showcase examples), and a highlighted line (a primary tint over the surface). The dark
// highlight is measured over the card without the black layer between, which only lowers the
// ratio: the text is lighter than the background.
func codeRows() []spec {
	var m []spec
	for _, t := range codeTokens {
		src := func(extra ...int) string {
			refs := []string{"assets/ui.css:" + itoa(t.line), "assets/ui.css:" + itoa(t.line+codeDarkOffset)}
			for _, l := range extra {
				refs = append(refs, "assets/ui.css:"+itoa(l), "assets/ui.css:"+itoa(l+codeDarkOffset))
			}
			return strings.Join(refs, ",")
		}
		lf, df := "text-"+t.light, "dark:text-"+t.dark
		m = append(m,
			spec{component: "Code/" + t.name, state: "on the code surface", src: src(codeSurfaceL), kind: text,
				light: on(lf, "bg-neutral-50"), dark: on(df, "dark:bg-black/40")},
			spec{component: "Code/" + t.name, state: "on the code surface over a card", src: src(codeSurfaceL), kind: text,
				light: on(lf, "bg-neutral-50").over(cardL), dark: on(df, "dark:bg-black/40").over(cardD)},
			spec{component: "Code/" + t.name, state: "on a highlighted line", src: src(codeHighlightL), kind: text,
				light: on(lf, "bg-primary-500/10").over("bg-neutral-50"), dark: on(df, "dark:bg-primary-400/15").over(cardD)},
		)
	}
	return m
}

func itoa(n int) string { return strconv.Itoa(n) }

var codeVar = regexp.MustCompile(`^\s*--color-code-([a-z-]+): var\(--color-([a-z]+-\d+)\);`)

// TestCodeTokensMatchCSS checks the table above against assets/ui.css: the first declaration of
// each variable (:root) uses the light step and the second (.dark) the dark step, at the line the
// matrix cites.
func TestCodeTokensMatchCSS(t *testing.T) {
	b, err := os.ReadFile("assets/ui.css")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string][]string{}
	lineOf := map[string]int{}
	for i, line := range strings.Split(string(b), "\n") {
		if m := codeVar.FindStringSubmatch(line); m != nil {
			if len(seen[m[1]]) == 0 {
				lineOf[m[1]] = i + 1
			}
			seen[m[1]] = append(seen[m[1]], m[2])
		}
	}
	for _, tok := range codeTokens {
		got := seen[tok.name]
		if len(got) != 2 || got[0] != tok.light || got[1] != tok.dark {
			t.Errorf("--color-code-%s: css has %v, matrix has [%s %s]", tok.name, got, tok.light, tok.dark)
		}
		if lineOf[tok.name] != tok.line {
			t.Errorf("--color-code-%s: css line %d, matrix cites %d", tok.name, lineOf[tok.name], tok.line)
		}
	}
}
