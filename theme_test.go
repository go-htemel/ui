package ui_test

import (
	"strings"
	"testing"

	"github.com/go-htemel/htemel"

	"github.com/go-htemel/ui"
)

func TestThemeClasses(t *testing.T) {
	th := ui.Theme{Primary: "indigo", Radius: "lg", Font: "font-serif"}
	tests := []struct{ in, want string }{
		{"bg-sky-600", "bg-indigo-600"},
		{"dark:hover:bg-sky-500 focus-visible:outline-sky-600", "dark:hover:bg-indigo-500 focus-visible:outline-indigo-600"},
		{"rounded-md", "rounded-lg"},
		{"*:first:rounded-l-md *:last:rounded-r-md", "*:first:rounded-l-lg *:last:rounded-r-lg"},
		{"rounded-full rounded-md-x text-md", "rounded-full rounded-md-x text-md"},
		{"  bg-red-600   ring-zinc-300 ", "bg-red-600 ring-zinc-300"},
		{"bg-skyline-600", "bg-skyline-600"},
	}
	for _, tt := range tests {
		if got := th.Classes(tt.in); got != tt.want {
			t.Errorf("Classes(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
	if got := ui.DefaultTheme.Classes("bg-sky-600  rounded-md"); got != "bg-sky-600 rounded-md" {
		t.Errorf("default theme rewrote classes: %q", got)
	}
}

func TestSetTheme(t *testing.T) {
	t.Cleanup(func() { ui.SetTheme(ui.DefaultTheme) })
	ui.SetTheme(ui.Theme{Primary: "emerald"})
	if got := ui.CurrentTheme(); got.Radius != "md" || got.Font != "font-sans" {
		t.Errorf("SetTheme did not fill defaults: %+v", got)
	}
	out, err := htemel.RenderString(ui.Button(ui.Primary, "Go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "bg-emerald-600") || strings.Contains(out, "sky") {
		t.Errorf("button does not follow the theme: %s", out)
	}
	if got := ui.BodyClasses(); !strings.HasPrefix(got, "font-sans ") {
		t.Errorf("BodyClasses() = %q", got)
	}
}
