package ui

import (
	"strings"
	"sync"
	"sync/atomic"
)

// Theme holds the few design tokens the components read. Component classes are written for
// DefaultTheme (the sky palette and md corners); a different theme rewrites those tokens when a
// component is built:
//
//   - Primary is a Tailwind color family ("sky", "indigo", "emerald", ...). Every "-sky-" in a
//     class, such as bg-sky-600 or dark:focus-visible:outline-sky-500, becomes "-<Primary>-".
//   - Radius is a Tailwind radius suffix ("none", "sm", "md", "lg", "xl", "2xl", "full"). Every
//     rounded-md class, including the side forms (rounded-l-md, rounded-t-md, ...), uses it.
//     Pills and avatars stay rounded-full.
//   - Font is the font-family class BodyClasses puts on <body> ("font-sans", "font-serif", ...).
//
// Tailwind only generates classes it finds in your sources. With a non-default theme, generate a
// safelist for it (task tailwind-safelist SAFELIST_FLAGS="-primary indigo") and add it to your
// Tailwind sources; see the README.
type Theme struct {
	Primary string
	Radius  string
	Font    string
}

// DefaultTheme is the theme the component classes are written in.
var DefaultTheme = Theme{Primary: "sky", Radius: "md", Font: "font-sans"}

var current atomic.Pointer[Theme]

// SetTheme replaces the theme used by every component built afterwards. Call it once at
// startup, before rendering; components already built keep their classes.
func SetTheme(t Theme) {
	if t.Primary == "" {
		t.Primary = DefaultTheme.Primary
	}
	if t.Radius == "" {
		t.Radius = DefaultTheme.Radius
	}
	if t.Font == "" {
		t.Font = DefaultTheme.Font
	}
	current.Store(&t)
	classCache.Clear()
}

// CurrentTheme returns the theme set with SetTheme, or DefaultTheme.
func CurrentTheme() Theme {
	if t := current.Load(); t != nil {
		return *t
	}
	return DefaultTheme
}

// Classes joins classes written for DefaultTheme and rewrites them for the current theme. Use it
// in your own components so they follow SetTheme:
//
//	html.Div().Class(ui.Classes("rounded-md bg-sky-600", extra))
func Classes(classes ...string) string {
	return CurrentTheme().Classes(classes...)
}

// classCache memoizes Classes for the current theme; SetTheme clears it.
var classCache sync.Map

// Classes joins classes written for DefaultTheme and rewrites them for t.
func (t Theme) Classes(classes ...string) string {
	s := strings.Join(strings.Fields(strings.Join(classes, " ")), " ")
	if t == DefaultTheme {
		return s
	}
	if t == CurrentTheme() {
		if v, ok := classCache.Load(s); ok {
			return v.(string)
		}
		out := t.rewrite(s)
		classCache.Store(s, out)
		return out
	}
	return t.rewrite(s)
}

func (t Theme) rewrite(s string) string {
	fields := strings.Fields(s)
	for i, f := range fields {
		fields[i] = t.token(f)
	}
	return strings.Join(fields, " ")
}

// token rewrites one class (with its variant prefixes, such as dark:hover:) for t.
func (t Theme) token(class string) string {
	prefix, base := "", class
	if i := strings.LastIndexByte(class, ':'); i >= 0 {
		prefix, base = class[:i+1], class[i+1:]
	}
	if t.Primary != "" && t.Primary != DefaultTheme.Primary {
		base = strings.ReplaceAll(base, "-"+DefaultTheme.Primary+"-", "-"+t.Primary+"-")
	}
	if t.Radius != "" && t.Radius != DefaultTheme.Radius {
		if side, ok := radiusSide(base); ok {
			base = "rounded" + side + "-" + t.Radius
		}
	}
	return prefix + base
}

// radiusSide reports whether base is rounded-md or a side form such as rounded-tl-md, and returns
// the side part ("", "-l", "-tl", ...).
func radiusSide(base string) (string, bool) {
	rest, ok := strings.CutPrefix(base, "rounded")
	if !ok {
		return "", false
	}
	side, ok := strings.CutSuffix(rest, "-"+DefaultTheme.Radius)
	if !ok {
		return "", false
	}
	switch side {
	case "", "-t", "-r", "-b", "-l", "-s", "-e", "-tl", "-tr", "-br", "-bl", "-ss", "-se", "-es", "-ee":
		return side, true
	}
	return "", false
}

// BodyClasses returns base classes for <body>: the theme font, the zinc page colors in light and
// dark mode, and antialiasing.
func BodyClasses() string {
	t := CurrentTheme()
	return t.Font + " " + t.Classes("bg-zinc-50 text-zinc-900 antialiased has-[dialog:modal]:overflow-hidden dark:bg-zinc-950 dark:text-zinc-100")
}
