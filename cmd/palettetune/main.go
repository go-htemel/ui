// Command palettetune derives a WCAG-tuned primary scale for a custom palette (it is not the
// default, which is stock Tailwind indigo; see the README, "Contrast"). Each step keeps its OKLCH hue and chroma (reduced only to stay inside sRGB)
// and its lightness is solved to satisfy the WCAG and spacing rules printed with the table.
//
//	go run ./cmd/palettetune           # before/after table
//	go run ./cmd/palettetune -css      # the --color-primary-* declarations for assets/ui.css
package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/go-htemel/ui/internal/wcag"
)

var steps = []string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900", "950"}

// original is the owner's palette as sRGB hex.
var original = []string{"F4FAF5", "D5F1DB", "ADE1BF", "7DCAA2", "37C392", "29A88D", "19867D", "186468", "184853", "173845", "071827"}

// toOKLCH converts sRGB to OKLCH (L in percent), the inverse of wcag.OKLCH.RGB.
func toOKLCH(c wcag.RGBA) wcag.OKLCH {
	lin := func(v uint8) float64 {
		f := float64(v) / 255
		if f <= 0.04045 {
			return f / 12.92
		}
		return math.Pow((f+0.055)/1.055, 2.4)
	}
	r, g, b := lin(c.R), lin(c.G), lin(c.B)
	l := math.Cbrt(0.4122214708*r + 0.5363325363*g + 0.0514459929*b)
	m := math.Cbrt(0.2119034982*r + 0.6806995451*g + 0.1073969566*b)
	s := math.Cbrt(0.0883024619*r + 0.2817188376*g + 0.6299787005*b)
	L := 0.2104542553*l + 0.7936177850*m - 0.0040720468*s
	a := 1.9779984951*l - 2.4285922050*m + 0.4505937099*s
	bb := 0.0259040371*l + 0.7827717662*m - 0.8086757660*s
	h := math.Atan2(bb, a) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return wcag.OKLCH{L: L * 100, C: math.Hypot(a, bb), H: h}
}

// inGamut reports whether o is inside sRGB without clamping.
func inGamut(o wcag.OKLCH) bool {
	l, c, h := o.L/100, o.C, o.H*math.Pi/180
	a, b := c*math.Cos(h), c*math.Sin(h)
	l_, m_, s_ := l+0.3963377774*a+0.2158037573*b, l-0.1055613458*a-0.0638541728*b, l-0.0894841775*a-1.2914855480*b
	lc, mc, sc := l_*l_*l_, m_*m_*m_, s_*s_*s_
	ch := []float64{
		+4.0767416621*lc - 3.3077115913*mc + 0.2309699292*sc,
		-1.2684380046*lc + 2.6097574011*mc - 0.3413193965*sc,
		-0.0041960863*lc - 0.7034186147*mc + 1.7076147010*sc,
	}
	for _, v := range ch {
		if v < -0.0005 || v > 1.0005 {
			return false
		}
	}
	return true
}

// fit reduces the chroma of o until it is inside sRGB.
func fit(o wcag.OKLCH) wcag.OKLCH {
	lo, hi := 0.0, o.C
	if inGamut(o) {
		return o
	}
	for i := 0; i < 40; i++ {
		mid := (lo + hi) / 2
		if inGamut(wcag.OKLCH{L: o.L, C: mid, H: o.H}) {
			lo = mid
		} else {
			hi = mid
		}
	}
	o.C = lo
	return o
}

var (
	white = wcag.RGBA{R: 255, G: 255, B: 255, A: 1}
	n900  = wcag.Palette["zinc-900"].RGB()
	n950  = wcag.Palette["zinc-950"].RGB()
)

func cr(a, b wcag.RGBA) float64 { return wcag.Contrast(a, b) }

// rule is a lightness predicate for one step: ok reports whether the color at L passes.
type rule func(c wcag.RGBA) bool

// rules lists, per step, what its color must satisfy. Spacing is checked afterwards.
var rules = map[string]rule{
	"50":  func(c wcag.RGBA) bool { return cr(n900, c) >= 12 },
	"100": func(c wcag.RGBA) bool { return cr(n900, c) >= 12 },
	"200": func(c wcag.RGBA) bool { return cr(n900, c) >= 12 },
	"300": func(c wcag.RGBA) bool { return cr(c, n900) >= 8 },
	"400": func(c wcag.RGBA) bool { return cr(c, n900) >= 6 && cr(c, n950) >= 6 },
	"500": func(c wcag.RGBA) bool { return cr(c, white) >= 3.1 },
	"600": func(c wcag.RGBA) bool { return cr(c, white) >= 4.8 },
	"700": func(c wcag.RGBA) bool { return cr(c, white) >= 6 },
	"800": func(c wcag.RGBA) bool { return cr(c, white) >= 8 },
}

func main() {
	css := flag.Bool("css", false, "print CSS declarations only")
	flag.Parse()

	orig := make([]wcag.OKLCH, len(steps))
	for i, h := range original {
		c, _ := wcag.ParseHex(h)
		orig[i] = toOKLCH(c)
	}

	// Lightness per step (OKLCH L, percent). 600, 700 and 800 sit at the lightest value that still
	// meets their white-text rule (4.8, 6 and 8:1); the rest are spaced evenly around them:
	// at least 0.06 apart in 50-400 and 600-950, 0.08-0.09 across 400-500-600. 500 is not
	// solved for 4.5:1 with white (that would put it level with 600); it clears 3.1:1.
	target := map[string]float64{
		"50": 98.5, "100": 92.4, "200": 86.3, "300": 79.9, "400": 73.0, "500": 63.6, "600": 54.2,
		"700": 48.0, "800": 42.0, "900": 34.0, "950": orig[idx("950")].L,
	}

	out := make([]wcag.OKLCH, len(steps))
	for i, s := range steps {
		out[i] = fit(wcag.OKLCH{L: target[s], C: orig[i].C, H: orig[i].H})
		out[i].L = math.Round(out[i].L*10) / 10
		out[i].C = math.Floor(out[i].C*1000) / 1000
		out[i].H = math.Round(out[i].H*10) / 10
	}
	if *css {
		for i, s := range steps {
			fmt.Printf("  --color-primary-%s: oklch(%.1f%% %.3f %.1f);\n", s, out[i].L, out[i].C, out[i].H)
		}
		return
	}

	fmt.Println("before (owner's original) / after (tuned)")
	fmt.Printf("%-4s | %-7s %5s %5s %6s | %-7s %5s %5s %6s | %5s %5s %5s %5s\n", "step", "hex", "L", "C", "H", "hex", "L", "C", "H", "w/wht", "n900", "n950", "dL")
	for i, s := range steps {
		oc, _ := wcag.ParseHex(original[i])
		nc := out[i].RGB()
		dl := 0.0
		if i > 0 {
			dl = (out[i-1].L - out[i].L) / 100
		}
		fmt.Printf("%-4s | #%s %5.1f %5.3f %6.1f | %-7s %5.1f %5.3f %6.1f | %5.2f %5.2f %5.2f %5.3f\n",
			s, original[i], orig[i].L, orig[i].C, orig[i].H, nc.Hex(), out[i].L, out[i].C, out[i].H,
			cr(nc, white), cr(nc, n900), cr(nc, n950), dl)
		_ = oc
	}
	failed := false
	for i, s := range steps {
		if r, ok := rules[s]; ok && !r(out[i].RGB()) {
			fmt.Printf("FAIL: step %s misses its rule\n", s)
			failed = true
		}
		if i > 0 && out[i-1].L-out[i].L < 5.9 {
			fmt.Printf("FAIL: step %s is less than 0.06 L below the previous step\n", s)
			failed = true
		}
	}
	if failed {
		defer os.Exit(1)
	}
	fmt.Println("\noriginal contrast: step w/white, vs n900, vs n950, dL")
	for i, s := range steps {
		oc, _ := wcag.ParseHex(original[i])
		dl := 0.0
		if i > 0 {
			dl = (orig[i-1].L - orig[i].L) / 100
		}
		fmt.Printf("%-4s %5.2f %5.2f %5.2f %5.3f\n", s, cr(oc, white), cr(oc, n900), cr(oc, n950), dl)
	}
}

func idx(s string) int {
	for i, v := range steps {
		if v == s {
			return i
		}
	}
	return -1
}
