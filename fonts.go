package ui

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"strings"

	"github.com/go-htemel/htemel/html"
)

// The Geist and Geist Mono variable fonts (SIL Open Font License 1.1, copyright Vercel; the
// licence text is assets/fonts/LICENSE-Geist.txt). Each file is one variable font covering the
// weights 100-900 in a single woff2.

//go:embed assets/fonts/Geist-Variable.woff2
var geistSansWOFF2 []byte

//go:embed assets/fonts/GeistMono-Variable.woff2
var geistMonoWOFF2 []byte

// font is one embedded font file as Assets serves it.
type font struct {
	family string // CSS family name
	path   string // path under the Assets prefix
	hash   string // short content hash: the ETag and the ?v= cache buster
}

func newFont(family, path string, data []byte) font {
	sum := sha256.Sum256(data)
	return font{family: family, path: path, hash: hex.EncodeToString(sum[:])[:12]}
}

// fonts are the files FontCSS declares and FontPreload preloads, in that order.
var fonts = []font{
	newFont("Geist", "/fonts/Geist-Variable.woff2", geistSansWOFF2),
	newFont("Geist Mono", "/fonts/GeistMono-Variable.woff2", geistMonoWOFF2),
}

// url is the URL of f under prefix. FontCSS and FontPreload must produce the same string, or the
// browser fetches the file twice.
func (f font) url(prefix string) string {
	return strings.TrimRight(prefix, "/") + f.path + "?v=" + f.hash
}

// FontCSS returns the stylesheet that self-hosts Geist and Geist Mono: one @font-face rule per
// family (variable weights 100-900, font-display: swap, woff2 served by Assets under prefix),
// and an @theme block that points Tailwind's font-sans at Geist and font-mono at Geist Mono, with
// system fallbacks. prefix is where Assets is mounted ("/assets", with or without a trailing
// slash).
//
// It is opt-in. Write it into your Tailwind sources next to ThemeCSS:
//
//	css := ui.ThemeCSS() + ui.FontCSS("/assets")
//
// An application that does not include it keeps Tailwind's system font stacks, and the font
// files are never requested. Pair it with FontPreload in <head> so the text does not swap fonts
// after first paint.
//
// The weight range is the variable axis: font-medium, font-semibold and the rest select real
// weights, not synthesized ones. FontCSS leaves the body weight at the default 400: Geist reads
// well there, and the components lean on a clear step to font-medium (500) and font-semibold
// (600) for labels and headings, which a heavier body weight such as 450 would shrink.
func FontCSS(prefix string) string {
	var b strings.Builder
	for _, f := range fonts {
		b.WriteString("@font-face {\n")
		b.WriteString("  font-family: \"" + f.family + "\";\n")
		b.WriteString("  font-style: normal;\n")
		b.WriteString("  font-weight: 100 900;\n")
		b.WriteString("  font-display: swap;\n")
		b.WriteString("  src: url(\"" + f.url(prefix) + "\") format(\"woff2\");\n")
		b.WriteString("}\n\n")
	}
	b.WriteString("@theme {\n")
	b.WriteString("  --font-sans: \"Geist\", ui-sans-serif, system-ui, sans-serif;\n")
	b.WriteString("  --font-mono: \"Geist Mono\", ui-monospace, SFMono-Regular, Menlo, monospace;\n")
	b.WriteString("}\n")
	return b.String()
}

// FontPreload returns the <link rel="preload"> elements for the two font files FontCSS declares,
// with the same URLs (including the content-hash ?v=) so the browser reuses the preloaded
// response. Put them in <head>, before the stylesheet; spread them into html.Head:
//
//	for _, l := range ui.FontPreload("/assets") {
//		head.Children(l)
//	}
//
// A font preload needs crossorigin even for a same-origin file, so the element carries it.
func FontPreload(prefix string) []*html.LinkElement {
	links := make([]*html.LinkElement, 0, len(fonts))
	for _, f := range fonts {
		links = append(links, html.Link().
			Rel("preload").
			As("font").
			Type("font/woff2").
			Crossorigin().
			Href(f.url(prefix)))
	}
	return links
}
