package wcag

import (
	"math"
	"os"
	"regexp"
	"strconv"
	"testing"
)

func mustHex(t *testing.T, s string) RGBA {
	t.Helper()
	c, err := ParseHex(s)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// TestContrastPublished checks the ratio formula against numbers published with WCAG itself and
// the WebAIM contrast checker.
func TestContrastPublished(t *testing.T) {
	for _, tc := range []struct {
		fg, bg string
		want   float64
	}{
		{"#000000", "#ffffff", 21},
		{"#ffffff", "#ffffff", 1},
		{"#767676", "#ffffff", 4.54}, // the lightest gray that passes AA on white
		{"#777777", "#ffffff", 4.48}, // one step lighter fails
		{"#0000ff", "#ffffff", 8.59},
		{"#0ea5e9", "#ffffff", 2.77}, // Tailwind v3 sky-500 under white text
		{"#0284c7", "#ffffff", 4.10}, // Tailwind v3 sky-600
	} {
		got := Contrast(mustHex(t, tc.fg), mustHex(t, tc.bg))
		if math.Abs(got-tc.want) > 0.006 {
			t.Errorf("Contrast(%s, %s) = %.3f, want %.2f", tc.fg, tc.bg, got, tc.want)
		}
	}
}

// TestOKLCHMatchesTailwindHex checks the OKLCH conversion against the hex values Tailwind v4
// publishes for its palette.
func TestOKLCHMatchesTailwindHex(t *testing.T) {
	for name, want := range map[string]string{
		"sky-500": "#00a6f4", "sky-600": "#0084d1", "red-500": "#fb2c36", "zinc-500": "#71717b",
		"zinc-950": "#09090b", "white": "#ffffff", "black": "#000000", "green-600": "#00a63e",
		"amber-500": "#fe9a00", "blue-600": "#155dfc",
	} {
		c, err := NewTheme("").Color(name)
		if err != nil {
			t.Fatal(err)
		}
		if c.Hex() != want {
			t.Errorf("%s = %s, want %s", name, c.Hex(), want)
		}
	}
}

func TestThemeFollowsAliases(t *testing.T) {
	th := NewTheme("@theme {\n --color-primary-600: var(--color-sky-600);\n --color-brand: var(--color-primary-600);\n --color-custom: #123456;\n}")
	if got := th.Alias("brand"); got != "sky-600" {
		t.Errorf("Alias(brand) = %s", got)
	}
	a, _ := th.Color("brand")
	b, _ := NewTheme("").Color("sky-600")
	if a != b {
		t.Errorf("brand = %v, sky-600 = %v", a, b)
	}
	if c, err := th.Color("custom"); err != nil || c.Hex() != "#123456" {
		t.Errorf("custom = %v, %v", c, err)
	}
	if _, err := th.Color("nope-1"); err == nil {
		t.Error("unknown color resolved")
	}
}

func TestClassColor(t *testing.T) {
	th := NewTheme("")
	c, err := th.ClassColor("dark:hover:bg-white/10")
	if err != nil || c.Hex() != "#ffffff" || math.Abs(c.A-0.1) > 1e-9 {
		t.Errorf("got %v, %v", c, err)
	}
	if _, err := th.ClassColor("rounded-lg"); err == nil {
		t.Error("non-color class accepted")
	}
}

func TestOver(t *testing.T) {
	got := RGBA{255, 255, 255, 0.5}.Over(RGBA{0, 0, 0, 1})
	if got.Hex() != "#808080" {
		t.Errorf("50%% white over black = %s", got.Hex())
	}
}

// TestPaletteComplete checks that every vendored family has all eleven steps.
func TestPaletteComplete(t *testing.T) {
	for _, fam := range []string{"sky", "zinc", "green", "amber", "red", "blue", "indigo", "emerald", "stone", "rose", "slate"} {
		for _, step := range []string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900", "950"} {
			if _, ok := Palette[fam+"-"+step]; !ok {
				t.Errorf("Palette lacks %s-%s", fam, step)
			}
		}
	}
}

// TestPaletteSource, when TAILWIND_THEME_CSS points at tailwindcss@4.1's theme.css, checks that
// palette.go matches it exactly. Skipped otherwise.
func TestPaletteSource(t *testing.T) {
	path := os.Getenv("TAILWIND_THEME_CSS")
	if path == "" {
		t.Skip("TAILWIND_THEME_CSS not set")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`--color-([a-z]+-\d+): oklch\(([\d.]+)% ([\d.]+) ([\d.]+)\);`)
	n := 0
	for _, m := range re.FindAllStringSubmatch(string(b), -1) {
		want, ok := Palette[m[1]]
		if !ok {
			continue
		}
		n++
		l, _ := strconv.ParseFloat(m[2], 64)
		c, _ := strconv.ParseFloat(m[3], 64)
		h, _ := strconv.ParseFloat(m[4], 64)
		if (OKLCH{L: l, C: c, H: h}) != want {
			t.Errorf("%s = %v, theme.css has %v", m[1], want, OKLCH{L: l, C: c, H: h})
		}
	}
	if n != 121 {
		t.Errorf("compared %d colors, want 121", n)
	}
}

func TestDarkThemeAndRGB(t *testing.T) {
	css := "@theme {\n --color-edge: oklch(65% 0.015 286);\n --color-edge-hover: var(--color-neutral-500);\n --color-neutral-500: #112233;\n}\n" +
		"@layer base {\n  html.dark { color-scheme: dark; }\n  .dark {\n    --color-edge: rgb(255 255 255 / 0.35);\n  }\n}\n"
	light, dark := NewTheme(css), NewDarkTheme(css)
	if c, err := light.ClassColor("outline-edge"); err != nil || c.A != 1 || c.Hex() != (OKLCH{65, 0.015, 286}).RGB().Hex() {
		t.Errorf("light edge = %+v, %v", c, err)
	}
	c, err := dark.ClassColor("outline-edge")
	if err != nil || c.R != 255 || c.A != 0.35 {
		t.Errorf("dark edge = %+v, %v", c, err)
	}
	if c, _ := dark.ClassColor("hover:outline-edge-hover"); c.Hex() != "#112233" {
		t.Errorf("dark edge-hover = %s, want the light value (not overridden)", c.Hex())
	}
	if c, _ := dark.ClassColor("outline-edge/50"); c.A != 0.5 {
		t.Errorf("opacity modifier = %v", c.A)
	}
}
