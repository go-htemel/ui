package safelist

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/go-htemel/ui"
)

const root = "../.."

// TestCommittedSafelistIsCurrent fails when a class was added to or removed from the components
// without regenerating tailwind.safelist.txt.
func TestCommittedSafelistIsCurrent(t *testing.T) {
	classes, err := Classes(os.DirFS(root), ui.DefaultTheme)
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile(filepath.Join(root, "tailwind.safelist.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(committed) != File(classes) {
		t.Error("tailwind.safelist.txt is out of date; run task tailwind-safelist")
	}
}

var (
	classAttr = regexp.MustCompile(` class="([^"]*)"`)
	classKey  = regexp.MustCompile(`&#39;([^&]*)&#39;:`) // keys of data-class object literals
)

// TestGoldenClassesAreListed checks that every class in the rendered golden output, including
// the class names inside data-class expressions, is in the safelist.
func TestGoldenClassesAreListed(t *testing.T) {
	classes, err := Classes(os.DirFS(root), ui.DefaultTheme)
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(root, "testdata", "*.golden.html"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no golden files: %v", err)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var found []string
		for _, m := range classAttr.FindAllStringSubmatch(string(b), -1) {
			found = append(found, strings.Fields(m[1])...)
		}
		for _, m := range classKey.FindAllStringSubmatch(string(b), -1) {
			found = append(found, strings.Fields(m[1])...)
		}
		for _, c := range found {
			if strings.HasPrefix(c, "fa-") {
				continue // Font Awesome classes are the caller's, passed to ui.Icon
			}
			if _, ok := slices.BinarySearch(classes, c); !ok {
				t.Errorf("%s: class %q is not in the safelist", filepath.Base(f), c)
			}
		}
	}
}

func TestThemedSafelist(t *testing.T) {
	classes, err := Classes(os.DirFS(root), ui.Theme{Primary: "indigo", Radius: "lg", Font: "font-serif"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"bg-indigo-600", "rounded-lg", "*:first:rounded-l-lg", "font-serif", "dark:bg-zinc-900"} {
		if _, ok := slices.BinarySearch(classes, want); !ok {
			t.Errorf("themed safelist lacks %q", want)
		}
	}
	for _, c := range classes {
		if strings.Contains(c, "-sky-") || strings.HasSuffix(c, "rounded-md") {
			t.Errorf("themed safelist still has %q", c)
		}
	}
}

// TestOnlyClassBearingStrings checks that strings which never reach a class-bearing call (JS
// expressions, HTTP header values) contribute nothing, and that literals, referenced constants,
// map values and datastar class names do.
func TestOnlyClassBearingStrings(t *testing.T) {
	const src = `package x

import "github.com/starfederation/datastar-go/datastar"

const base = "flex gap-2"
const unused = "unused-class"

var sizes = map[int]string{1: "text-sm", 2: "text-lg"}

func a(w http.ResponseWriter, el El) {
	w.Header().Set("Cache-Control", "immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	el.Attrs(datastar.On("toggle", "evt.newState === 'open' && close"))
	el.Class("px-2 py-1").AddClass(Classes(base, pick(sizes, 1, 2)))
	el.Attrs(datastar.Class("rotate-180", "open"))
}
`
	got, err := Classes(fstest.MapFS{"x.go": {Data: []byte(src)}}, ui.DefaultTheme)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"flex", "font-sans", "gap-2", "px-2", "py-1", "rotate-180", "text-lg", "text-sm"}
	if !slices.Equal(got, want) {
		t.Errorf("Classes = %q, want %q", got, want)
	}
}
