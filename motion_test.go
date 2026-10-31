package ui_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Motion rules, checked on every element of every golden:
//
//   - Anything that transitions more than colour also carries motion-reduce:transition-none (with
//     the same variant prefix: backdrop:transition-* needs motion-reduce:backdrop:transition-none),
//     so prefers-reduced-motion turns it off.
//   - transition-colors alone is exempt. A colour or background change moves nothing on screen;
//     the reduced-motion preference (and WCAG 2.3.3) is about movement, which is what triggers
//     vestibular discomfort, and the 150ms tint is still useful feedback for hover and focus.
//   - Only compositor-friendly properties are transitioned: opacity, translate, scale, rotate,
//     transform and grid-template-rows (plus display, overlay and visibility, which flip
//     discretely so the others can run). Never width, height, inset or the catch-all
//     transition / transition-all.
//   - No duration is longer than 200ms.

// motionProps are the properties a transition-[...] list may name.
var motionProps = map[string]bool{
	"opacity": true, "translate": true, "scale": true, "rotate": true, "transform": true,
	"grid-template-rows": true, "display": true, "overlay": true, "visibility": true,
}

// namedTransitions are the transition-* utilities allowed besides transition-[...].
// transition-shadow is allowed for ring colour changes on hover (a ring is a box-shadow, which
// transition-colors does not cover); it still needs motion-reduce:transition-none.
var namedTransitions = map[string]bool{"transition-colors": true, "transition-opacity": true, "transition-transform": true, "transition-shadow": true}

func splitVariant(c string) (prefix, base string) {
	i := strings.LastIndex(c, ":")
	// Arbitrary values may contain colons only inside brackets; none of ours do before the base.
	return c[:i+1], c[i+1:]
}

func TestMotionClasses(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "*.golden.html"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no golden files: %v", err)
	}
	transitioned := 0
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range goldenClassAttr.FindAllStringSubmatch(string(b), -1) {
			classes := strings.Fields(m[1])
			has := map[string]bool{}
			for _, c := range classes {
				has[c] = true
			}
			for _, c := range classes {
				prefix, base := splitVariant(c)
				if strings.HasPrefix(base, "duration-") {
					if n, err := strconv.Atoi(strings.TrimPrefix(base, "duration-")); err != nil || n > 200 {
						t.Errorf("%s: %q is longer than 200ms (or not a number) in %q", f, c, m[1])
					}
				}
				if base != "transition" && !strings.HasPrefix(base, "transition-") {
					continue
				}
				switch base {
				case "transition-none", "transition-discrete", "transition-normal":
					continue
				}
				switch {
				case namedTransitions[base]:
				case strings.HasPrefix(base, "transition-[") && strings.HasSuffix(base, "]"):
					for _, p := range strings.Split(strings.TrimSuffix(strings.TrimPrefix(base, "transition-["), "]"), ",") {
						if !motionProps[p] {
							t.Errorf("%s: %q transitions %q, which is not compositor-only, in %q", f, c, p, m[1])
						}
					}
				default:
					t.Errorf("%s: %q transitions more than opacity/transform/grid rows, in %q", f, c, m[1])
				}
				if base == "transition-colors" {
					continue
				}
				transitioned++
				if want := "motion-reduce:" + prefix + "transition-none"; !has[want] {
					t.Errorf("%s: %q without %q in %q", f, c, want, m[1])
				}
			}
		}
	}
	if transitioned < 10 {
		t.Errorf("only %d transitioned elements in the goldens; did the motion classes go away?", transitioned)
	}
}
