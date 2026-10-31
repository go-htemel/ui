// Package wcag measures WCAG 2.x contrast ratios for the design tokens in assets/ui.css.
//
// It converts OKLCH (the Tailwind v4 palette format) to 8-bit sRGB the way a browser paints it,
// composites translucent layers in gamma-encoded sRGB (as browsers and axe-core do), and computes
// relative luminance and contrast per WCAG 2.x. It has no dependencies.
package wcag

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// OKLCH is a color in the OKLCH space; L is a percentage (0-100), as written in CSS.
type OKLCH struct{ L, C, H float64 }

// RGBA is an 8-bit sRGB color with alpha in [0, 1].
type RGBA struct {
	R, G, B uint8
	A       float64
}

// Hex returns the color as #rrggbb (alpha is ignored).
func (c RGBA) Hex() string { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }

// Opaque returns c with alpha 1.
func (c RGBA) Opaque() RGBA { c.A = 1; return c }

// WithAlpha returns c with alpha multiplied by a.
func (c RGBA) WithAlpha(a float64) RGBA { c.A *= a; return c }

// RGB converts the OKLCH color to 8-bit sRGB. Out-of-gamut channels are clamped.
func (o OKLCH) RGB() RGBA {
	l, c, h := o.L/100, o.C, o.H*math.Pi/180
	a, b := c*math.Cos(h), c*math.Sin(h)
	// OKLab -> LMS' -> LMS -> linear sRGB (Bjorn Ottosson's reference matrices).
	l_ := l + 0.3963377774*a + 0.2158037573*b
	m_ := l - 0.1055613458*a - 0.0638541728*b
	s_ := l - 0.0894841775*a - 1.2914855480*b
	lc, mc, sc := l_*l_*l_, m_*m_*m_, s_*s_*s_
	r := +4.0767416621*lc - 3.3077115913*mc + 0.2309699292*sc
	g := -1.2684380046*lc + 2.6097574011*mc - 0.3413193965*sc
	bl := -0.0041960863*lc - 0.7034186147*mc + 1.7076147010*sc
	return RGBA{encode(r), encode(g), encode(bl), 1}
}

func encode(lin float64) uint8 {
	lin = math.Min(1, math.Max(0, lin))
	v := 12.92 * lin
	if lin > 0.0031308 {
		v = 1.055*math.Pow(lin, 1/2.4) - 0.055
	}
	return uint8(math.Round(math.Min(1, math.Max(0, v)) * 255))
}

func decode(c uint8) float64 {
	v := float64(c) / 255
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// Luminance returns the WCAG relative luminance of c (alpha ignored).
func (c RGBA) Luminance() float64 {
	return 0.2126*decode(c.R) + 0.7152*decode(c.G) + 0.0722*decode(c.B)
}

// Over composites c over the opaque backdrop bg in gamma-encoded sRGB, the way browsers blend
// translucent layers, and returns the opaque result.
func (c RGBA) Over(bg RGBA) RGBA {
	mix := func(f, b uint8) uint8 { return uint8(math.Round(c.A*float64(f) + (1-c.A)*float64(b))) }
	return RGBA{mix(c.R, bg.R), mix(c.G, bg.G), mix(c.B, bg.B), 1}
}

// Contrast returns the WCAG contrast ratio of two opaque colors, in [1, 21].
func Contrast(a, b RGBA) float64 {
	la, lb := a.Luminance(), b.Luminance()
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// ParseHex parses #rgb or #rrggbb.
func ParseHex(s string) (RGBA, error) {
	h := strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) != 6 {
		return RGBA{}, fmt.Errorf("wcag: bad hex color %q", s)
	}
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return RGBA{}, fmt.Errorf("wcag: bad hex color %q: %w", s, err)
	}
	return RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 1}, nil
}

var (
	cssDecl  = regexp.MustCompile(`--color-([a-z0-9-]+):\s*([^;]+);`)
	cssVar   = regexp.MustCompile(`^var\(--color-([a-z0-9-]+)\)$`)
	cssOKLCH = regexp.MustCompile(`^oklch\(\s*([\d.]+)%\s+([\d.]+)\s+([\d.]+)\s*\)$`)
	cssRGB   = regexp.MustCompile(`^rgb\(\s*(\d+)\s+(\d+)\s+(\d+)\s*(?:/\s*([\d.]+)\s*)?\)$`)
	darkRule = regexp.MustCompile(`(?s)\.dark\s*\{([^{}]*)\}`)
)

// Theme resolves design-token color names ("primary-600") to colors. It starts from Palette and
// applies the --color-* declarations of a stylesheet on top, following var() aliases, so
// primary-600 resolves to sky-600 and then to its OKLCH value.
type Theme struct {
	decl map[string]string // name -> raw CSS value
}

// NewTheme parses the --color-* declarations of css (typically ui.ThemeCSS()) as the light
// scheme: declarations inside a `.dark { ... }` rule are left out.
func NewTheme(css string) *Theme {
	t := &Theme{decl: map[string]string{}}
	t.add(darkRule.ReplaceAllString(css, ""))
	return t
}

// NewDarkTheme is NewTheme with the declarations of the `.dark { ... }` rules of css applied on
// top, i.e. the tokens as they resolve under <html class="dark">.
func NewDarkTheme(css string) *Theme {
	t := NewTheme(css)
	for _, m := range darkRule.FindAllStringSubmatch(css, -1) {
		t.add(m[1])
	}
	return t
}

func (t *Theme) add(css string) {
	for _, m := range cssDecl.FindAllStringSubmatch(css, -1) {
		t.decl[m[1]] = strings.TrimSpace(m[2])
	}
}

// Alias returns the palette name a token ultimately points at ("primary-600" -> "sky-600"), or
// the name itself when it is a palette color.
func (t *Theme) Alias(name string) string {
	for i := 0; i < 8; i++ {
		m := cssVar.FindStringSubmatch(t.decl[name])
		if m == nil {
			return name
		}
		name = m[1]
	}
	return name
}

// Color resolves a token or palette name to a color; opaque unless the token is an rgb() with
// an alpha.
func (t *Theme) Color(name string) (RGBA, error) {
	name = t.Alias(name)
	if raw, ok := t.decl[name]; ok {
		if m := cssOKLCH.FindStringSubmatch(raw); m != nil {
			l, _ := strconv.ParseFloat(m[1], 64)
			c, _ := strconv.ParseFloat(m[2], 64)
			h, _ := strconv.ParseFloat(m[3], 64)
			return OKLCH{l, c, h}.RGB(), nil
		}
		if m := cssRGB.FindStringSubmatch(raw); m != nil {
			r, _ := strconv.Atoi(m[1])
			g, _ := strconv.Atoi(m[2])
			b, _ := strconv.Atoi(m[3])
			a := 1.0
			if m[4] != "" {
				a, _ = strconv.ParseFloat(m[4], 64)
			}
			return RGBA{uint8(r), uint8(g), uint8(b), a}, nil
		}
		if strings.HasPrefix(raw, "#") {
			return ParseHex(raw)
		}
		return RGBA{}, fmt.Errorf("wcag: unsupported value %q for --color-%s", raw, name)
	}
	if o, ok := Palette[name]; ok {
		return o.RGB(), nil
	}
	return RGBA{}, fmt.Errorf("wcag: unknown color %q", name)
}

// utility matches a Tailwind color utility after its variants are stripped: text-primary-600,
// bg-white/10, outline-neutral-400, ...
var utility = regexp.MustCompile(`^(?:bg|text|outline|ring|border|accent|fill|stroke|decoration|placeholder)-([a-z]+(?:-[a-z]+)*(?:-\d+)?)(?:/(\d+))?$`)

// ClassColor resolves the color utility in class (any variant prefixes such as "dark:" or
// "hover:" are ignored) to a color with its opacity modifier as alpha. Only the last
// colon-separated segment is considered.
func (t *Theme) ClassColor(class string) (RGBA, error) {
	base := class[strings.LastIndex(class, ":")+1:]
	m := utility.FindStringSubmatch(base)
	if m == nil {
		return RGBA{}, fmt.Errorf("wcag: %q is not a color utility", class)
	}
	c, err := t.Color(m[1])
	if err != nil {
		return RGBA{}, fmt.Errorf("%s: %w", class, err)
	}
	if m[2] != "" {
		pct, _ := strconv.ParseFloat(m[2], 64)
		c.A = pct / 100
	}
	return c, nil
}
