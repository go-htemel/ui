package ui_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/go-htemel/ui"
)

var (
	goldenClassAttr = regexp.MustCompile(` class="([^"]*)"`)
	goldenClassKey  = regexp.MustCompile(`&#39;([^&]*)&#39;:`) // keys of data-class object literals
)

// goldenClasses returns every class in the golden files: class attributes and the class names
// inside data-class expressions.
func goldenClasses(t *testing.T) map[string][]string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("testdata", "*.golden.html"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no golden files: %v", err)
	}
	out := map[string][]string{}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, re := range []*regexp.Regexp{goldenClassAttr, goldenClassKey} {
			for _, m := range re.FindAllStringSubmatch(string(b), -1) {
				out[filepath.Base(f)] = append(out[filepath.Base(f)], strings.Fields(m[1])...)
			}
		}
	}
	return out
}

// TestNoRawPaletteClasses checks that components use the semantic tokens, never the Tailwind
// palette families those tokens alias by default.
func TestNoRawPaletteClasses(t *testing.T) {
	raw := regexp.MustCompile(`-(sky|zinc|green|red|amber|blue)-\d|\brounded(-[a-z]+)?-(sm|md|lg|xl)\b`)
	for file, classes := range goldenClasses(t) {
		for _, c := range classes {
			if raw.MatchString(c) {
				t.Errorf("%s: raw palette or radius class %q; use a token", file, c)
			}
		}
	}
	for _, c := range strings.Fields(ui.BodyClasses()) {
		if raw.MatchString(c) {
			t.Errorf("BodyClasses: raw class %q", c)
		}
	}
}

var (
	cssColorToken  = regexp.MustCompile(`--color-([a-z]+)-(\d+):`)
	cssRadiusToken = regexp.MustCompile(`--radius-([a-z]+):`)
	// A utility of a token color family: bg-primary-600, dark:hover:ring-danger-400/20, ...
	classColorToken  = regexp.MustCompile(`(?:^|[:-])(?:bg|text|ring|outline|border(?:-[trblxyse])?|divide|from|to|via|fill|stroke|accent|caret|decoration|shadow|placeholder|ring-offset|inset-ring|inset-shadow)-(primary|neutral|success|warning|danger|info)-(\d+)(?:/\d+)?$`)
	classRadiusToken = regexp.MustCompile(`(?:^|:)rounded(?:-[a-z]+)?-(control|surface)$`)
)

// TestTokensDefined checks that every token a component references is defined by ThemeCSS, so a
// Tailwind build with ui.theme.css generates every class.
func TestTokensDefined(t *testing.T) {
	css := ui.ThemeCSS()
	if !regexp.MustCompile(`(?m)^@theme \{$`).MatchString(css) || regexp.MustCompile(`(?m)^@theme inline`).MatchString(css) {
		t.Fatal("ThemeCSS is not a plain @theme block")
	}
	colors, radii := map[string]bool{}, map[string]bool{}
	for _, m := range cssColorToken.FindAllStringSubmatch(css, -1) {
		colors[m[1]+"-"+m[2]] = true
	}
	for _, m := range cssRadiusToken.FindAllStringSubmatch(css, -1) {
		radii[m[1]] = true
	}
	for _, fam := range []string{"primary", "neutral", "success", "warning", "danger", "info"} {
		for _, step := range []string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900", "950"} {
			if !colors[fam+"-"+step] {
				t.Errorf("ThemeCSS lacks --color-%s-%s", fam, step)
			}
		}
	}

	all := goldenClasses(t)
	all["BodyClasses"] = strings.Fields(ui.BodyClasses())
	used := 0
	for file, classes := range all {
		for _, c := range classes {
			if m := classColorToken.FindStringSubmatch(c); m != nil {
				used++
				if !colors[m[1]+"-"+m[2]] {
					t.Errorf("%s: %q uses undefined token --color-%s-%s", file, c, m[1], m[2])
				}
			}
			if m := classRadiusToken.FindStringSubmatch(c); m != nil {
				used++
				if !radii[m[1]] {
					t.Errorf("%s: %q uses undefined token --radius-%s", file, c, m[1])
				}
			}
		}
	}
	if used < 50 {
		t.Errorf("only %d token classes found in the goldens; is the pattern stale?", used)
	}
}

func TestAssetsServeThemeCSS(t *testing.T) {
	rec := httptest.NewRecorder()
	ui.Assets().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ui.css", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != ui.ThemeCSS() {
		t.Fatalf("GET /ui.css = %d, body matches ThemeCSS: %v", rec.Code, rec.Body.String() == ui.ThemeCSS())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/css; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	req := httptest.NewRequest(http.MethodGet, "/ui.css", nil)
	req.Header.Set("If-None-Match", rec.Header().Get("ETag"))
	rec = httptest.NewRecorder()
	ui.Assets().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified {
		t.Errorf("conditional GET /ui.css = %d, want 304", rec.Code)
	}
}
